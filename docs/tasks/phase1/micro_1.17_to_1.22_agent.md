# Micro-Tasks 1.17 → 1.22: Agent Contracts

## Thứ tự thực hiện
1.17 → 1.18 → 1.19 → 1.20 → 1.21 → 1.22

---

# Micro-Task 1.17: contracts/agent/capability.go

## Nội dung CHÍNH XÁC

```go
// Package agent defines the contract for AI agents.
// An agent is a specialized AI persona that can execute tasks.
package agent

// Capability represents a specific skill that an agent possesses.
// The orchestrator uses capabilities to match tasks to agents.
type Capability string

const (
	CapCodeGeneration Capability = "code_generation"
	CapCodeReview     Capability = "code_review"
	CapArchitecture   Capability = "architecture"
	CapTesting        Capability = "testing"
	CapDocumentation  Capability = "documentation"
	CapDeployment     Capability = "deployment"
	CapDebugging      Capability = "debugging"
	CapRefactoring    Capability = "refactoring"
	CapDataAnalysis   Capability = "data_analysis"
	CapResearch       Capability = "research"
)

// String returns the string representation.
func (c Capability) String() string { return string(c) }

// IsValid checks if the capability is one of the defined constants.
func (c Capability) IsValid() bool {
	switch c {
	case CapCodeGeneration, CapCodeReview, CapArchitecture, CapTesting,
		CapDocumentation, CapDeployment, CapDebugging, CapRefactoring,
		CapDataAnalysis, CapResearch:
		return true
	default:
		return false
	}
}

// HasCapability checks if a capability exists in a slice.
func HasCapability(caps []Capability, target Capability) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}
```

## ⚠️ Pitfall: Dùng `string` constants, KHÔNG `iota`. Capabilities serialize vào YAML/JSON.

---

# Micro-Task 1.18: contracts/agent/task.go

## Nội dung CHÍNH XÁC

```go
package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Task represents a unit of work assigned to an agent.
type Task struct {
	// ID uniquely identifies this task.
	ID contracts.TaskID `json:"id"`

	// Name is a short descriptive name for this task.
	// Example: "implement_user_handler", "write_unit_tests"
	Name string `json:"name"`

	// Description explains what needs to be done in detail.
	// This is included in the prompt sent to the AI.
	Description string `json:"description"`

	// Type categorizes the task for agent matching.
	// Example: "code_generation", "code_review", "testing"
	Type string `json:"type"`

	// Input contains task-specific input data.
	// Each task type has different inputs:
	//   code_generation: {"language": "go", "framework": "gin"}
	//   code_review: {"diff": "...", "focus": "security"}
	//
	// WHY map[string]any instead of typed struct?
	// → Each task type has different inputs. A generic map is flexible.
	// → Type-safe input structs can be added per-task-type in the future.
	Input map[string]any `json:"input,omitempty"`

	// Context provides additional information for the agent.
	// Examples: relevant files, previous task outputs, project docs.
	Context []ContextItem `json:"context,omitempty"`

	// Dependencies lists task IDs that must complete before this task.
	// The scheduler uses this to determine execution order.
	//
	// WHY []string instead of []*Task?
	// → Pointer references create circular dependency risk → memory leak.
	// → String IDs are serializable and DAG-friendly.
	Dependencies []contracts.TaskID `json:"dependencies,omitempty"`

	// Priority determines execution order (higher = run first).
	// Range: 1 (lowest) to 10 (highest). Default: 5.
	Priority int `json:"priority"`

	// Timeout is the maximum time allowed for this task.
	// If the agent doesn't finish within this time, the task is cancelled.
	Timeout time.Duration `json:"timeout"`

	// Metadata contains arbitrary key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ContextItem provides additional context to an agent.
type ContextItem struct {
	// Type identifies the kind of context.
	// Values: "file", "snippet", "url", "memory", "task_output"
	Type string `json:"type"`

	// Content is the actual context data.
	Content string `json:"content"`

	// Source identifies where this context came from.
	// Example: "/src/main.go", "task-123-output", "https://docs.example.com"
	Source string `json:"source,omitempty"`
}

// NewTask creates a new Task with a generated ID and default values.
func NewTask(name, description, taskType string) *Task {
	return &Task{
		ID:          contracts.NewTaskID(),
		Name:        name,
		Description: description,
		Type:        taskType,
		Priority:    5,               // Default: medium priority
		Timeout:     5 * time.Minute, // Default: 5 minutes
	}
}

// AddDependency adds a task dependency.
func (t *Task) AddDependency(taskID contracts.TaskID) *Task {
	t.Dependencies = append(t.Dependencies, taskID)
	return t
}

// AddContext adds a context item.
func (t *Task) AddContext(ctxType, content, source string) *Task {
	t.Context = append(t.Context, ContextItem{
		Type:    ctxType,
		Content: content,
		Source:  source,
	})
	return t
}

// HasDependencies returns true if this task depends on other tasks.
func (t *Task) HasDependencies() bool {
	return len(t.Dependencies) > 0
}
```

## ⚠️ Pitfalls
1. **Dependencies dùng `[]contracts.TaskID`** (named type), KHÔNG `[]string`. Type safety.
2. **Timeout per task**: Mỗi task có timeout riêng. KHÔNG dùng global timeout cho tất cả.
3. **Input dùng `map[string]any`**: Flexible cho nhiều loại task. Parse vào typed struct ở agent implementation.
4. **Builder pattern**: `NewTask()` + `AddDependency()` + `AddContext()` cho clean API.

---

# Micro-Task 1.19: contracts/agent/result.go

## Nội dung CHÍNH XÁC

```go
package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// Result represents the output of a task executed by an agent.
type Result struct {
	// TaskID links this result to the task that was executed.
	TaskID contracts.TaskID `json:"task_id"`

	// AgentName identifies which agent executed the task.
	AgentName string `json:"agent_name"`

	// Status indicates the outcome.
	Status contracts.Status `json:"status"`

	// Output is the primary text output from the agent.
	Output string `json:"output"`

	// Artifacts are files, diffs, or other outputs produced.
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Error is a human-readable error description.
	// Populated when Status is Failed or Timeout.
	//
	// WHY string instead of error interface?
	// → error interface cannot be JSON serialized.
	// → String is portable across JSON, gRPC, databases.
	Error string `json:"error,omitempty"`

	// Duration is how long the task took to execute.
	Duration time.Duration `json:"duration"`

	// Usage tracks token consumption from the AI provider.
	// Nil if the task didn't use a provider (e.g., pure tool execution).
	Usage *provider.Usage `json:"usage,omitempty"`

	// Metadata for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Artifact represents a file or data output from an agent.
type Artifact struct {
	// Name is a descriptive name (e.g., "main.go", "api_diff", "test_report")
	Name string `json:"name"`

	// Type categorizes the artifact.
	// Values: "file", "diff", "log", "report", "image"
	Type string `json:"type"`

	// Path is the filesystem path (for "file" type artifacts).
	Path string `json:"path,omitempty"`

	// Content is inline content (for small artifacts like diffs or reports).
	// Use either Path or Content, not both.
	Content string `json:"content,omitempty"`
}

// IsSuccess returns true if the task completed successfully.
func (r *Result) IsSuccess() bool {
	return r.Status.IsSuccess()
}

// IsFailed returns true if the task failed.
func (r *Result) IsFailed() bool {
	return r.Status.IsFailed()
}

// SuccessResult creates a successful result.
func SuccessResult(taskID contracts.TaskID, agentName, output string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusSuccess,
		Output:    output,
	}
}

// FailedResult creates a failed result.
func FailedResult(taskID contracts.TaskID, agentName, errMsg string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusFailed,
		Error:     errMsg,
	}
}
```

## ⚠️ Pitfalls
1. **Error là `string`**: `error` interface KHÔNG serialize được.
2. **Usage là `*provider.Usage` (pointer)**: Nil khi task không dùng provider.
3. **Artifact: Path vs Content**: Chỉ dùng 1. Path cho files lớn, Content cho diffs/reports nhỏ.
4. **Import cycle risk**: `agent` import `provider` cho Usage type. OK vì contracts packages không import lẫn nhau theo vòng.

---

# Micro-Task 1.20: contracts/agent/manifest.go

## Nội dung CHÍNH XÁC

```go
package agent

// Manifest describes an agent's configuration, loaded from YAML.
//
// Example agent.yaml:
//
//	name: backend
//	version: "0.1.0"
//	role: "Backend Developer"
//	description: "Generates Go backend code"
//	capabilities:
//	  - code_generation
//	  - testing
//	provider: antigravity
//	model: gemini-2.5-pro
//	tools:
//	  - read_file
//	  - write_file
//	  - git_commit
//	prompt_file: prompts/system.md
//	temperature: 0.3
//	max_tokens: 8192
type Manifest struct {
	Name         string       `yaml:"name" json:"name"`
	Version      string       `yaml:"version" json:"version"`
	Role         string       `yaml:"role" json:"role"`
	Description  string       `yaml:"description" json:"description"`
	Capabilities []Capability `yaml:"capabilities" json:"capabilities"`
	Provider     string       `yaml:"provider" json:"provider"`
	Model        string       `yaml:"model,omitempty" json:"model,omitempty"`
	Tools        []string     `yaml:"tools,omitempty" json:"tools,omitempty"`
	SystemPrompt string       `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	PromptFile   string       `yaml:"prompt_file,omitempty" json:"prompt_file,omitempty"`
	Temperature  float64      `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxTokens    int          `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
}
```

## ⚠️ Pitfalls
1. **SystemPrompt vs PromptFile**: Hỗ trợ cả 2. Short prompts inline, long prompts in file.
2. **Provider là string name**: Registry resolve name → instance at runtime.
3. **Tools là `[]string`**: Tool names, NOT tool instances. Registry resolve at runtime.

---

# Micro-Task 1.21: contracts/agent/agent.go

## Nội dung CHÍNH XÁC

```go
package agent

import "context"

// Agent is the core interface that all AI agents must implement.
//
// An agent is a specialized AI persona with specific capabilities.
// Examples: Backend Developer, Code Reviewer, DevOps Engineer.
//
// Each agent:
//   - Has a role and capabilities (defined in Manifest)
//   - Uses a provider to communicate with an AI model
//   - Can use tools to interact with the outside world
//   - Receives Tasks and returns Results
//
// Lifecycle:
//   Agent lifecycle (Init, Start, Stop) is managed by the Plugin interface.
//   This interface only defines runtime behavior.
type Agent interface {
	// Name returns the unique identifier (e.g., "backend", "reviewer").
	Name() string

	// Role returns the human-readable role (e.g., "Backend Developer").
	Role() string

	// Capabilities returns the list of capabilities this agent has.
	Capabilities() []Capability

	// Execute performs a task and returns the result.
	//
	// Implementation steps (typical):
	//   1. Build prompt from system prompt + task description + context
	//   2. Call provider.Send() with the prompt
	//   3. If response has tool calls → execute tools → send results back to AI
	//   4. Repeat step 2-3 until AI returns final response (no more tool calls)
	//   5. Build Result from final response
	//
	// Error handling:
	//   - System errors (network, panic) → return (nil, error)
	//   - Business errors (AI can't do task) → return (Result{Status: Failed}, nil)
	//   - NEVER return (non-nil Result, non-nil error) simultaneously
	Execute(ctx context.Context, task *Task) (*Result, error)

	// CanHandle checks if this agent can handle the given task.
	//
	// The orchestrator calls this to find the right agent for a task.
	// Default implementation: check if task.Type matches any Capability.
	//
	// Override for custom matching logic:
	//   - Agent "backend" handles tasks with Type="code_generation" AND Input["language"]="go"
	//   - Agent "reviewer" handles any task with Type="code_review"
	CanHandle(task *Task) bool
}
```

## ⚠️ Pitfalls
1. **Execute return convention**: `(nil, error)` = system error. `(result, nil)` = always, even on failure. NEVER `(result, error)` together.
2. **CanHandle**: Orchestrator gọi mỗi agent.CanHandle() để tìm match. PHẢI nhanh (no I/O).
3. **No Init/Start/Stop**: Lifecycle = Plugin interface.

---

# Micro-Task 1.22: contracts/agent/agent_test.go

## Tests cần viết
1. **TestCapability_IsValid**: Valid/invalid capabilities
2. **TestHasCapability**: Tìm capability trong slice
3. **TestNewTask**: Default values (priority=5, timeout=5m)
4. **TestTask_AddDependency**: Chain builder
5. **TestTask_AddContext**: Context items
6. **TestResult_IsSuccess**: Status checks
7. **TestSuccessResult**: Constructor helper
8. **TestFailedResult**: Constructor helper
9. **TestResult_JSONRoundTrip**: Serialize/deserialize
10. **TestManifest_Fields**: Verify struct fields exist and tags correct

## Lệnh verify
```bash
go test -v ./contracts/agent/...
```

## Checklist cho tất cả 1.17-1.22
- [ ] 6 files tồn tại trong `contracts/agent/`
- [ ] 10 capabilities defined
- [ ] Task với dependencies (named type `[]contracts.TaskID`)
- [ ] Result với Artifacts slice
- [ ] Agent interface với 4 methods
- [ ] Manifest struct với 12 fields
- [ ] ≥ 10 test functions
- [ ] `go build ./contracts/...` không lỗi
- [ ] `go test ./contracts/agent/...` ALL PASS
