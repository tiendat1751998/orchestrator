# Micro-Tasks 1.23 → 1.34: Remaining Contracts

## Thứ tự: 1.23 → 1.24 → 1.25 → 1.26 → 1.27 → 1.28 → 1.29 → 1.30 → 1.31 → 1.32 → 1.33 → 1.34

> Mỗi file = 1 package = 1 interface chính + supporting types.

---

# Micro-Task 1.23: contracts/event/event.go

```go
package event

import (
	"context"
	"time"
)

// Event represents something that happened in the system.
type Event struct {
	ID        string    `json:"id"`        // Unique event ID
	Type      string    `json:"type"`      // Dot notation: "task.started", "agent.error"
	Source    string    `json:"source"`    // Who emitted: "kernel", "agent:backend"
	Payload   any       `json:"payload"`   // Event-specific data
	Timestamp time.Time `json:"timestamp"` // When it happened
}

// Bus provides publish/subscribe functionality.
type Bus interface {
	// Publish emits an event to all matching subscribers.
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for events matching the given pattern.
	// Pattern supports wildcards: "task.*" matches "task.started", "task.completed".
	// Returns an unsubscribe function. MUST be called to avoid memory leaks.
	Subscribe(pattern string, handler func(Event)) (unsubscribe func(), err error)
}

// Common event types
const (
	EventTaskStarted    = "task.started"
	EventTaskCompleted  = "task.completed"
	EventTaskFailed     = "task.failed"
	EventMissionStarted = "mission.started"
	EventMissionDone    = "mission.completed"
	EventAgentError     = "agent.error"
	EventKernelStarted  = "kernel.started"
	EventKernelStopped  = "kernel.stopped"
)
```

## ⚠️ Pitfalls
1. **Subscribe trả về `unsubscribe func()`**: Consumer PHẢI gọi unsubscribe → tránh memory leak.
2. **Payload là `any`**: Mỗi event type có payload khác nhau. Consumer dùng type assertion.
3. **Wildcard matching**: `"task.*"` phải match `"task.started"` VÀ `"task.completed"`. Implementation ở Phase 2.

---

# Micro-Task 1.24: contracts/plugin/plugin.go

```go
package plugin

import "context"

// Type identifies the category of a plugin.
type Type string

const (
	TypeAgent    Type = "agent"
	TypeProvider Type = "provider"
	TypeTool     Type = "tool"
	TypeSearch   Type = "search"
	TypeMemory   Type = "memory"
	TypeWorkflow Type = "workflow"
	TypeContext  Type = "context"
)

// Plugin is the lifecycle interface for all pluggable components.
//
// Every agent, provider, tool, etc. that registers with the kernel
// must implement this interface for lifecycle management.
//
// Lifecycle order:
//   1. Init()  — Load config, validate, allocate resources (NO connections yet)
//   2. Start() — Open connections, start goroutines, become operational
//   3. Health() — Periodic health checks (called by supervisor)
//   4. Stop()  — Close connections, stop goroutines, release resources
//
// Stop order is REVERSE of Start order (LIFO):
//   Start: EventBus → Registry → Provider → Agent
//   Stop:  Agent → Provider → Registry → EventBus
type Plugin interface {
	Name() string
	Type() Type
	Version() string
	Init(ctx context.Context, config map[string]any) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) error
}
```

## ⚠️ Pitfalls
1. **Init vs Start TÁCH RIÊNG**: Init = load config (no I/O). Start = connect/run (I/O). Tách ra để init tất cả trước, rồi start theo dependency order.
2. **Stop REVERSE order**: Agent stop trước Provider. Nếu ngược lại → Agent gọi Provider đã chết → panic.
3. **Health returns error**: nil = healthy. non-nil = unhealthy. Circuit breaker dùng để decide.

---

# Micro-Task 1.25: contracts/memory/memory.go

```go
package memory

import (
	"context"
	"time"
)

// Store provides persistent key-value storage with search.
type Store interface {
	Save(ctx context.Context, key string, value any, opts ...SaveOption) error
	Load(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	Search(ctx context.Context, query string, limit int) ([]Entry, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

// Entry represents a stored item returned by Search.
type Entry struct {
	Key       string    `json:"key"`
	Value     any       `json:"value"`
	Score     float64   `json:"score,omitempty"` // Relevance score (0-1)
	CreatedAt time.Time `json:"created_at"`
}

// SaveOption configures Save behavior (functional options pattern).
type SaveOption func(*saveOptions)

type saveOptions struct {
	TTL  time.Duration
	Tags []string
}

// WithTTL sets a time-to-live for the entry.
func WithTTL(d time.Duration) SaveOption {
	return func(o *saveOptions) { o.TTL = d }
}

// WithTags adds searchable tags to the entry.
func WithTags(tags ...string) SaveOption {
	return func(o *saveOptions) { o.Tags = tags }
}

// ApplySaveOptions processes options into saveOptions struct.
func ApplySaveOptions(opts ...SaveOption) saveOptions {
	var o saveOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
```

## ⚠️ Pitfall: Functional options pattern (`SaveOption`) cho phép thêm options mà không break interface.

---

# Micro-Task 1.26: contracts/search/search.go

```go
package search

import "context"

// Engine provides search and indexing capabilities.
type Engine interface {
	Index(ctx context.Context, items []Indexable) error
	Search(ctx context.Context, query string, opts ...SearchOption) ([]Result, error)
}

// Indexable is anything that can be indexed for search.
type Indexable interface {
	SearchID() string
	SearchContent() string
	SearchMetadata() map[string]string
}

// Result represents a search match.
type Result struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Score    float64           `json:"score"` // 0-1, higher = more relevant
	Metadata map[string]string `json:"metadata"`
}

// SearchOption configures search behavior.
type SearchOption func(*searchOptions)

type searchOptions struct {
	MaxResults int
	MinScore   float64
	Filters    map[string]string
}

func WithMaxResults(n int) SearchOption {
	return func(o *searchOptions) { o.MaxResults = n }
}

func WithMinScore(score float64) SearchOption {
	return func(o *searchOptions) { o.MinScore = score }
}
```

---

# Micro-Task 1.27: contracts/workflow/workflow.go

```go
package workflow

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Workflow defines a sequence of steps to execute.
type Workflow interface {
	Name() string
	Steps() []Step
	Execute(ctx context.Context, input map[string]any) (*Result, error)
}

// Step defines a single step in a workflow.
type Step struct {
	Name       string   `yaml:"name" json:"name"`
	Agent      string   `yaml:"agent" json:"agent"`           // Agent name to use
	Task       string   `yaml:"task" json:"task"`             // Task type
	DependsOn  []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Condition  string   `yaml:"condition,omitempty" json:"condition,omitempty"` // "previous.status == 'success'"
	OnFailure  string   `yaml:"on_failure,omitempty" json:"on_failure,omitempty"` // "retry", "skip", "abort"
	MaxRetries int      `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
}

// Result represents the outcome of a workflow execution.
type Result struct {
	Status   contracts.Status         `json:"status"`
	Steps    map[string]*StepResult   `json:"steps"`
	Duration time.Duration            `json:"duration"`
}

// StepResult is the outcome of a single workflow step.
type StepResult struct {
	Status   contracts.Status `json:"status"`
	Output   any              `json:"output"`
	Error    string           `json:"error,omitempty"`
	Duration time.Duration    `json:"duration"`
}
```

---

# Micro-Task 1.28: contracts/context/context.go

```go
package agentcontext

import "context"

// Builder constructs the context window for an AI agent.
// It decides what information to include in the prompt.
type Builder interface {
	// Build creates context from various sources.
	Build(ctx context.Context, opts ...BuildOption) ([]Item, error)
}

// Item is a piece of context to include in the prompt.
type Item struct {
	Type     string `json:"type"`     // "file", "snippet", "search_result", "memory"
	Content  string `json:"content"`
	Source   string `json:"source"`
	Priority int    `json:"priority"` // Higher = more important = include first
	Tokens   int    `json:"tokens"`   // Estimated token count
}

// BuildOption configures context building.
type BuildOption func(*buildOptions)

type buildOptions struct {
	MaxTokens int
	Sources   []string
}

func WithMaxTokens(n int) BuildOption {
	return func(o *buildOptions) { o.MaxTokens = n }
}
```

## ⚠️ Pitfall: Package name `agentcontext` (NOT `context`) — tránh conflict với Go standard `context` package.

---

# Micro-Task 1.29: contracts/planner/planner.go

```go
package planner

import (
	"context"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// Planner decomposes missions into executable task graphs.
type Planner interface {
	// Plan decomposes a mission into a list of tasks with dependencies.
	Plan(ctx context.Context, mission *Mission) ([]*agent.Task, error)

	// Replan creates a new plan when a task fails.
	Replan(ctx context.Context, mission *Mission, failedTask *agent.Task, err error) ([]*agent.Task, error)
}

// Mission is a high-level goal from the user.
type Mission struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Constraints []string          `json:"constraints,omitempty"` // "use Go", "no external deps"
	Metadata    map[string]string `json:"metadata,omitempty"`
}
```

---

# Micro-Task 1.30: contracts/orchestrator/orchestrator.go

```go
package orchestrator

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/planner"
)

// Orchestrator coordinates the entire mission execution flow.
type Orchestrator interface {
	// ExecuteMission runs a mission from start to finish.
	ExecuteMission(ctx context.Context, mission *planner.Mission) (*MissionResult, error)

	// Status returns the current status of a mission.
	Status(missionID string) (*MissionStatus, error)

	// Cancel cancels a running mission.
	Cancel(missionID string) error
}

// MissionResult is the final output of a completed mission.
type MissionResult struct {
	MissionID  string                   `json:"mission_id"`
	Status     contracts.Status         `json:"status"`
	Tasks      map[string]*agent.Result `json:"tasks"`
	Summary    string                   `json:"summary"`
	Artifacts  []agent.Artifact         `json:"artifacts"`
	Duration   time.Duration            `json:"duration"`
}

// MissionStatus is the real-time status of a running mission.
type MissionStatus struct {
	MissionID    string           `json:"mission_id"`
	Status       contracts.Status `json:"status"`
	CurrentTask  string           `json:"current_task,omitempty"`
	TotalTasks   int              `json:"total_tasks"`
	DoneTasks    int              `json:"done_tasks"`
	FailedTasks  int              `json:"failed_tasks"`
	Elapsed      time.Duration    `json:"elapsed"`
}
```

---

# Micro-Task 1.31: contracts/resilience/resilience.go

```go
package resilience

import "context"

// CircuitBreaker prevents cascading failures.
type CircuitBreaker interface {
	Execute(fn func() error) error
	State() string // "closed", "open", "half-open"
	Reset()
}

// RetryPolicy defines retry behavior.
type RetryPolicy interface {
	Execute(ctx context.Context, fn func() error) error
}

// Fallback provides alternative execution paths.
type Fallback interface {
	Execute(primary func() error, fallback func() error) error
}
```

---

# Micro-Task 1.32: contracts/security/security.go

```go
package security

import "context"

// PermissionManager controls what agents can do.
type PermissionManager interface {
	CanUseTool(agentName, toolName string) bool
	CanAccessPath(agentName, path string) bool
	CanRunCommand(agentName, command string) bool
}

// AuditLogger records all agent actions for security review.
type AuditLogger interface {
	Log(ctx context.Context, entry AuditEntry) error
	Query(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// AuditEntry is a single audit log record.
type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	Agent     string `json:"agent"`
	Action    string `json:"action"` // "tool_call", "file_read", "file_write", "command"
	Target    string `json:"target"` // tool name, file path, command
	Allowed   bool   `json:"allowed"`
	Details   string `json:"details,omitempty"`
}

// AuditFilter for querying audit logs.
type AuditFilter struct {
	Agent  string `json:"agent,omitempty"`
	Action string `json:"action,omitempty"`
	Since  string `json:"since,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}
```

---

# Micro-Task 1.33: contracts/gateway/gateway.go

```go
package gateway

import "context"

// Gateway is the unified entry point for external requests.
type Gateway interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Address() string // Returns the listen address (e.g., ":8080")
}
```

---

# Micro-Task 1.34: contracts/feedback/feedback.go

```go
package feedback

import "context"

// Evaluator assesses the quality of agent outputs.
type Evaluator interface {
	Evaluate(ctx context.Context, input EvalInput) (*EvalResult, error)
}

// EvalInput is what to evaluate.
type EvalInput struct {
	TaskType   string `json:"task_type"`
	AgentName  string `json:"agent_name"`
	Output     string `json:"output"`
	Expected   string `json:"expected,omitempty"` // If available
}

// EvalResult is the evaluation outcome.
type EvalResult struct {
	Score    float64 `json:"score"` // 0-1
	Feedback string  `json:"feedback"`
	Pass     bool    `json:"pass"`
}

// Scorer tracks agent performance over time.
type Scorer interface {
	RecordResult(agentName, taskType string, success bool, duration float64)
	GetScore(agentName, taskType string) float64
	GetRanking(taskType string) []AgentScore
}

// AgentScore represents an agent's performance score.
type AgentScore struct {
	AgentName  string  `json:"agent_name"`
	Score      float64 `json:"score"`
	TaskCount  int     `json:"task_count"`
	SuccessRate float64 `json:"success_rate"`
}
```

---

## Checklist tổng cho 1.23 → 1.34
- [ ] 12 files tồn tại (1 file per package)
- [ ] Mỗi file có package declaration đúng
- [ ] Mỗi file có ít nhất 1 interface
- [ ] Tất cả sử dụng `context.Context` cho I/O methods
- [ ] Functional options pattern cho extensible methods
- [ ] KHÔNG có import cycles
- [ ] `go build ./contracts/...` không lỗi
