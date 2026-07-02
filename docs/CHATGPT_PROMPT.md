# GEMINI 2.5 Flash / Pro Transfer Prompt — Orchestrator Project

> **How to use**: Copy everything below the `---` line and paste into Gemini (AI Studio or App).
> Then ask ONE of:
> - "Enhance Phase 1 contracts to production quality" (creates enhancement micro-tasks)
> - "Enhance Phase 2 kernel to production quality" (creates enhancement micro-tasks)
> - "Create micro-tasks for Phase 3" (new phase)
> Each phase should be a separate conversation.

---

## ROLE

You are a senior Go architect creating ultra-detailed micro-task files for an AI orchestrator project. Each micro-task is a standalone markdown file that another AI (or developer) can read and implement **exactly** — copy-paste the code, run the verify command, done.

## PROJECT CONTEXT

I'm building an **orchestrator** in Go 1.26 that coordinates AI agents (like Antigravity, Gemini, etc.) to execute complex multi-step tasks autonomously.

**GitHub**: `github.com/tiendat1751998/orchestrator`

### Architecture (6 layers)

```
orchestrator/
├── cmd/                    # CLI entry point
├── kernel/                 # Core engine (Phase 2 ✅ DONE)
│   ├── config/             # YAML config loading
│   ├── logger/             # slog-based structured logging
│   ├── eventbus/           # Async pub/sub with wildcards
│   ├── registry/           # Plugin management + lifecycle
│   ├── runtime/            # Task execution (executor + pool + dispatcher)
│   ├── scheduler/          # Priority queue + dependency tracking
│   └── lifecycle/          # OS signal handling
├── contracts/              # Interfaces (Phase 1 ✅ DONE)
│   ├── agent/              # Agent interface
│   ├── provider/           # AI provider interface
│   ├── tool/               # Tool interface
│   ├── workflow/           # Workflow interface
│   ├── plugin/             # Plugin interface
│   ├── context/            # Execution context
│   ├── search/             # Search interface
│   ├── memory/             # Memory interface
│   └── event/              # Event types
├── sdk/                    # Developer helpers (Phase 3)
├── plugins/                # Implementations (Phase 4)
│   ├── agents/
│   ├── providers/
│   ├── tools/
│   ├── search/
│   ├── memory/
│   └── workflow/
├── modules/                # High-level features (Phase 5)
├── api/                    # HTTP/gRPC API (Phase 6)
└── web/                    # Dashboard (Phase 6)
```

### Import Rules (CRITICAL — violating = build failure)
```
contracts/ ← kernel/ ← sdk/ ← plugins/ ← modules/ ← api/
↑ NEVER imports from anything to the right
```

- `contracts/` imports NOTHING from the project (only stdlib)
- `kernel/` imports from `contracts/` only
- `sdk/` imports from `contracts/` and `kernel/`
- `plugins/` imports from `contracts/`, `kernel/`, and `sdk/`
- `modules/` imports from everything except `api/`
- `api/` imports from everything

### Key Contracts (already implemented)

```go
// contracts/agent/agent.go
type Agent interface {
    plugin.Plugin
    Capabilities() []Capability
    CanHandle(task *Task) bool
    Execute(ctx context.Context, task *Task) (*Result, error)
    Manifest() Manifest
}

// contracts/provider/provider.go
type Provider interface {
    plugin.Plugin
    Complete(ctx context.Context, req Request) (*Response, error)
    Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
    Models() []string
}

// contracts/tool/tool.go
type Tool interface {
    plugin.Plugin
    Schema() Schema
    Execute(ctx context.Context, input map[string]any) (*Result, error)
}

// contracts/plugin/plugin.go
type Plugin interface {
    Name() string
    Type() string
    Version() string
    Init(ctx context.Context, config map[string]any) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health(ctx context.Context) error
}

// contracts/event/event.go
type Event struct {
    ID        string
    Type      string
    Source    string
    Payload   any
    Timestamp time.Time
}

type Bus interface {
    Publish(ctx context.Context, evt Event) error
    Subscribe(pattern string, handler func(Event)) (func(), error)
}
```

### Key Kernel Components (already implemented)

- **Config**: YAML loading with env var resolution (`${ENV_VAR:default}`)
- **Logger**: `log/slog` wrapper with redaction, pretty formatter, field helpers
- **EventBus**: Async pub/sub, wildcard patterns (`task.*`), panic recovery
- **Registry**: Thread-safe plugin storage, agent finder, lifecycle management (Init→Start→Stop LIFO)
- **Runtime**: Executor (panic recovery, timeout), Pool (semaphore), Dispatcher (result channel)
- **Scheduler**: Priority queue (`container/heap`), dependency tracker (DFS cycle detection)
- **Kernel**: State machine (Created→Initializing→Running→ShuttingDown→Stopped), Start/Stop lifecycle

## REMAINING PHASES

### Phase 3 — SDK (Developer Helpers)
Purpose: Make it easy for developers to create agents, providers, and tools.

Components to create:
1. **sdk/baseagent/**: Base agent struct with common logic (embed in custom agents)
2. **sdk/baseprovider/**: Base provider struct
3. **sdk/basetool/**: Base tool struct
4. **sdk/helpers/**: ID generation, retry logic, rate limiter
5. **sdk/testing/**: Mock agent, mock provider, mock eventbus for unit tests
6. **sdk/middleware/**: Logging middleware, metrics middleware, retry middleware
7. Tests for each component

### Phase 4 — Plugins (Implementations)
Purpose: Real agent/provider/tool implementations.

Components to create:
1. **plugins/providers/antigravity/**: Provider that calls Antigravity CLI (`antigravity` binary)
2. **plugins/providers/gemini/**: Provider that calls Gemini API (REST)
3. **plugins/providers/router/**: Routes requests to the best provider
4. **plugins/agents/coder/**: Agent that generates code
5. **plugins/agents/reviewer/**: Agent that reviews code
6. **plugins/agents/planner/**: Agent that breaks tasks into subtasks
7. **plugins/tools/filesystem/**: Read/write/list files
8. **plugins/tools/shell/**: Execute shell commands
9. **plugins/tools/git/**: Git operations
10. **plugins/search/semantic/**: Semantic code search
11. **plugins/memory/local/**: File-based conversation memory
12. **plugins/workflow/sequential/**: Run tasks in sequence
13. **plugins/workflow/parallel/**: Run tasks in parallel
14. Tests for each component

### Phase 5 — Modules (High-Level Features)
Purpose: Business logic that combines kernel + plugins.

Components to create:
1. **modules/mission/**: Mission planner — breaks user request into task graph
2. **modules/feedback/**: Quality scoring and agent performance tracking
3. **modules/session/**: Conversation session management
4. **modules/orchestration/**: Main orchestration loop (mission → plan → execute → verify → report)
5. Tests for each component

### Phase 6 — API & CLI
Purpose: User-facing interfaces.

Components to create:
1. **cmd/orchestrator/**: Main CLI entry point
2. **cmd/orchestrator/commands/**: CLI commands (run, status, config, version)
3. **api/http/**: REST API server
4. **api/http/handlers/**: API route handlers
5. **api/http/middleware/**: Auth, CORS, logging
6. **web/**: Dashboard (optional, can be Phase 7)
7. Tests for each component

## MICRO-TASK FILE FORMAT

Every micro-task file MUST follow this EXACT format:

```markdown
# Micro-Task X.XX: [Create/Update] [file path]

## Info
- **File**: `[relative path from project root]`
- **Package**: `[Go package name]`
- **Depends on**: [list of prior micro-task numbers]
- **Time**: [estimated minutes]
- **Verify**: `[go build/test command]`

## Purpose
[2-3 sentences explaining WHY this file exists and WHAT it does]

## EXACT code to create

\```go
[COMPLETE, COMPILABLE Go code]
[EVERY import]
[EVERY function]
[EVERY type]
[Godoc comments on exported types/functions]
[Internal comments explaining non-obvious decisions]
\```

## Pitfalls

### Pitfall 1: [Name]
\```go
// WRONG:
[bad code example]

// CORRECT:
[good code example]
\```
[Explanation of WHY the wrong version is wrong]

### Pitfall 2: [Name]
[repeat pattern]

## Verify
\```bash
[exact command to verify]
# Expected: [what success looks like]
\```

## Checklist
- [ ] File exists at correct path
- [ ] Package name is correct
- [ ] All exported types have Godoc
- [ ] [specific items for this file]
- [ ] Build command passes
```

## RULES (CRITICAL)

1. **One file per micro-task**. Never put 2 files in 1 task.
2. **EXACT code**. The code must compile as-is. No pseudocode, no `...`, no TODOs.
3. **ALL imports listed**. Every import path must be complete.
4. **Pitfalls section mandatory**. At least 2 pitfalls per file explaining common mistakes.
5. **Thread-safety**. Any struct accessed from multiple goroutines MUST use sync.Mutex or sync.RWMutex.
6. **Error messages include context**. `fmt.Errorf("component: action: %w", err)` — always wrap with component name.
7. **No global variables**. All state via constructor injection (DI).
8. **Test files use `_test` package**. `package foo_test`, not `package foo`.
9. **Verify command at the end**. Must be a single `go build` or `go test` command.
10. **English only**. All code, comments, and documentation in English.
11. **Go 1.26**. Use modern Go features (generics OK, slog OK, atomic types OK).
12. **Constructor pattern**: `New(deps) *Type` — every struct has a constructor.
13. **snake_case for JSON/YAML tags and log keys**. NOT camelCase.
14. **Godoc on all exported identifiers**. Functions, types, constants, variables.
15. **Tests use table-driven pattern** where applicable.

## STYLE EXAMPLES

### Good error message:
```go
return fmt.Errorf("registry: plugin %q already registered", name)
```

### Good constructor:
```go
func New(logger *slog.Logger, cfg Config) *Service {
    if cfg.Timeout <= 0 {
        cfg.Timeout = 30 * time.Second
    }
    return &Service{logger: logger, cfg: cfg}
}
```

### Good thread-safety:
```go
type Registry struct {
    mu    sync.RWMutex
    items map[string]*Item
}

func (r *Registry) Get(name string) (*Item, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ...
}

func (r *Registry) Add(item *Item) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ...
}
```

### Good test:
```go
func TestService_Create_DuplicateName(t *testing.T) {
    svc := New(nil, Config{})
    svc.Add("foo")
    
    err := svc.Add("foo")
    if err == nil {
        t.Error("expected error for duplicate name")
    }
}
```

## HOW TO START

When I ask you to "create micro-tasks for Phase X", you should:

1. First, output the **index.md** file listing ALL micro-tasks for that phase with a dependency graph (mermaid diagram).
2. Then output each micro-task file **one at a time** in the exact format above.
3. Number them sequentially: Phase 3 = `3.01`, `3.02`, ...; Phase 4 = `4.01`, etc.
4. Group by component (e.g., all baseagent files together, then baseprovider, etc.)
5. End with a `X.XX_verify.md` that has the full verification checklist.

If the conversation gets long, I'll say "continue" and you pick up where you left off.

## CONTEXT FROM COMPLETED PHASES

### Phase 1 contracts (36 micro-tasks) — contracts/ directory
All interface files are implemented and tested. Key types:
- `contracts.TaskID`, `contracts.MissionID` (string types)
- `contracts.StatusPending`, `StatusRunning`, `StatusCompleted`, `StatusFailed` (status constants)
- `agent.Capability` type with constants like `CapabilityCodeGeneration`, `CapabilityCodeReview`, `CapabilityDeployment`
- `agent.Task` struct with `ID`, `Name`, `Type`, `Input`, `Dependencies []TaskID`, `Timeout time.Duration`
- `agent.Result` struct with `TaskID`, `Status`, `Output any`, `Error string`
- `provider.Request` with `Model`, `Messages []Message`, `Tools []ToolDef`, `Temperature`, `MaxTokens`
- `provider.Response` with `Content`, `ToolCalls []ToolCall`, `Usage TokenUsage`
- `event.Event` struct and `event.Bus` interface
- Standard event type constants: `EventTaskStarted`, `EventTaskCompleted`, `EventTaskFailed`, `EventKernelStarted`, `EventKernelStopped`

### Phase 2 kernel (35 micro-tasks) — kernel/ directory
All kernel components implemented and tested. Key APIs:
- `config.Load(path)` → `*Config`
- `logger.New(Options{Level, Format, Output})` → `*Logger`
- `eventbus.New(slogger)` → `*Bus` (implements `event.Bus`)
- `registry.New(slogger)` → `*Registry`
- `registry.Register(plugin)`, `GetAgent(name)`, `FindAgentForTask(task)`
- `runtime.New(reg, bus, logger, Config)` → `*Runtime`
- `scheduler.New(dispatchFn, logger)` → `*Scheduler`
- `kernel.New(cfg)` → `*Kernel`
- `kernel.RegisterPlugin(p)`, `kernel.Start(ctx)`, `kernel.Stop(ctx)`

---

Now create micro-tasks for **Phase [NUMBER]**.

---
---

# APPENDIX: PRODUCTION HARDENING (Phase 1 & 2 Enhancement)

> Use this section when asking: "Enhance Phase 1/2 to production quality"

## WHAT'S MISSING from current Phase 1 & 2

The current micro-tasks cover basic functionality but are NOT production-ready. Here's exactly what's missing and needs to be added as new micro-tasks:

---

### GAP 1: Structured Error Types (current: raw fmt.Errorf strings)

**Problem**: Errors are plain strings. Callers can't programmatically distinguish between "not found", "timeout", "validation", "permission denied".

**What to add** (contracts/errors/ package):
```go
// contracts/errors/errors.go

// Error categories — callers use errors.Is() to check
type NotFoundError struct { Resource, Name string }
type ValidationError struct { Field, Message string }
type TimeoutError struct { Duration time.Duration }
type ConflictError struct { Resource, Name string }
type PermissionError struct { Action, Resource string }
type RetryableError struct { Err error; RetryAfter time.Duration }

// IsRetryable(err) bool — checks if error should be retried
// IsNotFound(err) bool — checks if resource was not found
// Wrap(err, "context") — adds context while preserving error type
```

**Why production needs this**:
- Retry logic: only retry `RetryableError`, not `ValidationError`
- HTTP API: `NotFoundError` → 404, `ValidationError` → 400, `TimeoutError` → 504
- Logging: different severity for different error types
- Circuit breaker: only trip on `RetryableError`

---

### GAP 2: Context Propagation (current: ctx passed but not enriched)

**Problem**: Context is passed through but doesn't carry request-scoped data (trace ID, task ID, deadline info).

**What to add** (contracts/context/ package enhancement):
```go
// contracts/context/context.go

type Key string
const (
    KeyTraceID   Key = "trace_id"
    KeyTaskID    Key = "task_id"
    KeyMissionID Key = "mission_id"
    KeyAgentName Key = "agent_name"
)

func WithTraceID(ctx context.Context, traceID string) context.Context
func TraceIDFrom(ctx context.Context) string
func WithTaskID(ctx context.Context, taskID string) context.Context
func TaskIDFrom(ctx context.Context) string

// Every log line automatically includes trace_id + task_id
// Every error wraps with trace_id for correlation
```

**Why production needs this**:
- Distributed tracing across agent calls
- Log correlation: find ALL logs for a single task
- Deadline propagation: parent timeout flows to child operations

---

### GAP 3: Retry & Circuit Breaker (current: no retry logic)

**Problem**: When a provider call fails, the system just returns the error. No retry, no backoff, no circuit breaker.

**What to add** (kernel/resilience/ package):
```go
// kernel/resilience/retry.go
type RetryConfig struct {
    MaxAttempts int           // Default: 3
    InitialWait time.Duration // Default: 1s
    MaxWait     time.Duration // Default: 30s
    Multiplier  float64       // Default: 2.0 (exponential backoff)
    Jitter      bool          // Default: true (randomize wait ±25%)
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error)

// kernel/resilience/circuitbreaker.go
type CircuitBreaker struct { ... }
// States: Closed (normal) → Open (failing) → HalfOpen (testing)
// Trips after N consecutive failures
// Resets after timeout
```

**Why production needs this**:
- API rate limits → retry with backoff
- Transient network errors → retry
- Provider down → circuit breaker prevents hammering
- Jitter prevents thundering herd

---

### GAP 4: Metrics & Observability Hooks (current: only logging)

**Problem**: No metrics collection. Can't answer: "How many tasks/sec?", "What's P99 latency?", "Which agent is slowest?"

**What to add** (kernel/metrics/ package):
```go
// kernel/metrics/metrics.go
type Metrics interface {
    Counter(name string, tags map[string]string) CounterMetric
    Histogram(name string, tags map[string]string) HistogramMetric
    Gauge(name string, tags map[string]string) GaugeMetric
}

type CounterMetric interface { Inc(); Add(float64) }
type HistogramMetric interface { Observe(float64) }
type GaugeMetric interface { Set(float64) }

// In-memory implementation for Phase 2
// Prometheus/OpenTelemetry adapter for Phase 5

// Standard metrics to track:
// - orchestrator_tasks_total (counter, by status)
// - orchestrator_task_duration_seconds (histogram, by agent)
// - orchestrator_active_workers (gauge)
// - orchestrator_queue_depth (gauge)
// - orchestrator_provider_requests_total (counter, by provider+model)
// - orchestrator_provider_tokens_total (counter, by provider)
// - orchestrator_provider_latency_seconds (histogram, by provider)
```

---

### GAP 5: Graceful Degradation (current: fail-fast on any error)

**Problem**: If ONE provider fails, the entire system stops. No fallback, no degradation.

**What to add**:
```go
// kernel/runtime/executor.go — enhance ExecuteTask

// Current: one agent fails → task fails → done
// Production: try fallback agents, emit degradation events

func (e *Executor) ExecuteTask(ctx, task) (*Result, error) {
    agents := e.registry.FindAllAgentsForTask(task)
    
    for i, agent := range agents {
        result, err := e.tryAgent(ctx, agent, task)
        if err == nil {
            return result, nil
        }
        
        if !errors.IsRetryable(err) {
            return nil, err // Non-retryable → fail immediately
        }
        
        // Log fallback attempt
        e.logger.Warn("agent failed, trying fallback",
            "failed_agent", agent.Name(),
            "attempt", i+1,
            "total_agents", len(agents),
        )
    }
    return nil, fmt.Errorf("all %d agents failed for task %q", len(agents), task.Name)
}
```

---

### GAP 6: Input Validation on All Structs (current: minimal)

**Problem**: Structs accept any input. Invalid data propagates deep into the system before failing with confusing errors.

**What to add**: `Validate() error` method on EVERY struct that holds user input.

```go
// contracts/agent/task.go
func (t *Task) Validate() error {
    if t.ID == "" { return errors.NewValidation("task", "id", "required") }
    if t.Name == "" { return errors.NewValidation("task", "name", "required") }
    if t.Type == "" { return errors.NewValidation("task", "type", "required") }
    if t.Timeout < 0 { return errors.NewValidation("task", "timeout", "must be >= 0") }
    // Check dependencies: no duplicates, no self-reference
    seen := make(map[TaskID]bool)
    for _, dep := range t.Dependencies {
        if dep == t.ID { return errors.NewValidation("task", "dependencies", "self-dependency") }
        if seen[dep] { return errors.NewValidation("task", "dependencies", "duplicate: "+string(dep)) }
        seen[dep] = true
    }
    return nil
}

// contracts/provider/request.go
func (r *Request) Validate() error {
    if r.Model == "" { return errors.NewValidation("request", "model", "required") }
    if len(r.Messages) == 0 { return errors.NewValidation("request", "messages", "at least one required") }
    if r.Temperature < 0 || r.Temperature > 2 { return errors.NewValidation("request", "temperature", "must be 0-2") }
    if r.MaxTokens < 0 { return errors.NewValidation("request", "max_tokens", "must be >= 0") }
    return nil
}
```

---

### GAP 7: Event Bus — Dead Letter Queue (current: events silently dropped if handler panics)

**Problem**: If an event handler panics, the event is lost. No record of failure. No retry.

**What to add**:
```go
// kernel/eventbus/dlq.go (Dead Letter Queue)
type DeadLetterEntry struct {
    Event     event.Event
    Error     string
    Handler   string // handler function name
    Timestamp time.Time
    Attempts  int
}

type DeadLetterQueue struct {
    mu      sync.Mutex
    entries []DeadLetterEntry
    maxSize int // Default: 1000 (ring buffer)
}

func (dlq *DeadLetterQueue) Add(entry DeadLetterEntry)
func (dlq *DeadLetterQueue) Entries() []DeadLetterEntry
func (dlq *DeadLetterQueue) Len() int
func (dlq *DeadLetterQueue) Clear()
```

---

### GAP 8: Config Hot-Reload (current: load once at startup)

**Problem**: Changing log level or max workers requires restarting the kernel.

**What to add**:
```go
// kernel/config/watcher.go
type Watcher struct { ... }

func NewWatcher(path string, onChange func(*Config)) *Watcher
func (w *Watcher) Start(ctx context.Context) error // Watches file for changes
func (w *Watcher) Stop()

// Hot-reloadable fields (safe to change at runtime):
// - log_level
// - max_concurrent_tasks
// - provider timeout
//
// NON-reloadable fields (require restart):
// - provider type
// - provider binary path
// - listen address
```

---

### GAP 9: Health Check Depth (current: shallow Health() method)

**Problem**: `Health()` returns nil or error. No structured health info.

**What to add**:
```go
// contracts/plugin/health.go
type HealthStatus string
const (
    HealthOK       HealthStatus = "ok"
    HealthDegraded HealthStatus = "degraded"
    HealthDown     HealthStatus = "down"
)

type HealthReport struct {
    Status    HealthStatus          `json:"status"`
    Message   string                `json:"message,omitempty"`
    Details   map[string]any        `json:"details,omitempty"`
    Children  map[string]HealthReport `json:"children,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
    Duration  time.Duration         `json:"duration"`
}

// Kernel aggregates health from all plugins into a tree:
// {
//   "status": "degraded",
//   "children": {
//     "provider:antigravity": {"status": "ok", "duration": "5ms"},
//     "provider:gemini": {"status": "down", "message": "connection refused"},
//     "agent:coder": {"status": "ok"}
//   }
// }
```

---

### GAP 10: Resource Cleanup Guarantees (current: defer but no verification)

**Problem**: If Stop() panics or hangs, resources leak. No verification that cleanup actually happened.

**What to add**:
```go
// kernel/runtime/runtime.go — enhance Stop()
func (r *Runtime) Stop(ctx context.Context) error {
    // ... existing code ...
    
    // After stopping, verify no goroutine leaks
    // Check pool stats
    stats := r.pool.Stats()
    if stats.ActiveWorkers > 0 {
        r.logger.Error("goroutine leak detected",
            "active_workers", stats.ActiveWorkers,
            "submitted", stats.TotalSubmitted,
            "completed", stats.TotalCompleted,
        )
    }
    
    // Verify result channel is drained
    remaining := 0
    for range r.dispatcher.Results() {
        remaining++
    }
    if remaining > 0 {
        r.logger.Warn("undrained results", "count", remaining)
    }
}
```

---

## HOW TO ASK GEMINI FOR ENHANCEMENTS

Use this exact prompt pattern:

```
Based on the gaps listed in "APPENDIX: PRODUCTION HARDENING",
create enhancement micro-tasks for Phase [1 or 2].

For Phase 1:
- Number them 1.37, 1.38, ... (continuing from existing 1.36)
- Focus on: structured errors, input validation, health reports, context propagation

For Phase 2:
- Number them 2.36, 2.37, ... (continuing from existing 2.35)
- Focus on: retry/circuit breaker, metrics hooks, dead letter queue, config hot-reload, graceful degradation, resource cleanup

Each enhancement micro-task must:
1. Show the EXACT file to modify or create
2. Show the COMPLETE updated code (not just the diff)
3. Explain what was added and WHY
4. Include tests for the new functionality
5. Have a verify command
```

---

## PRODUCTION QUALITY CHECKLIST

When reviewing any micro-task output, verify it covers:

- [ ] **Error types**: Uses structured errors, not string formatting
- [ ] **Input validation**: Every public function validates its inputs
- [ ] **Context propagation**: trace_id and task_id flow through all calls
- [ ] **Timeout handling**: Every blocking operation has a timeout
- [ ] **Retry logic**: Transient failures are retried with backoff
- [ ] **Circuit breaker**: Repeated failures trip a breaker
- [ ] **Metrics**: Key operations emit counters/histograms
- [ ] **Health checks**: Every component reports structured health
- [ ] **Graceful degradation**: Fallback agents, not immediate failure
- [ ] **Resource cleanup**: Verified in Stop(), goroutine leak detection
- [ ] **Dead letter queue**: Failed events are captured, not dropped
- [ ] **Hot-reload**: Runtime config changes without restart
- [ ] **Thread-safety**: Race detector passes (`go test -race`)
- [ ] **Idempotent operations**: Start/Stop/Subscribe safe to call multiple times
- [ ] **Panic recovery**: Never crash the process, always recover + log
- [ ] **Backpressure**: Bounded channels/queues, no unbounded growth
- [ ] **Deterministic ordering**: Registration order preserved (not map iteration)
- [ ] **Error wrapping**: `fmt.Errorf("component: action: %w", err)` — preserves chain
