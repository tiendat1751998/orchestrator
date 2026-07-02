# ChatGPT Transfer Prompt — Orchestrator Project

> **How to use**: Copy everything below the `---` line and paste it into ChatGPT.
> Then ask: "Create micro-tasks for Phase 3" (or Phase 4, 5, 6).
> Each phase should be a separate conversation to avoid context limits.

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
