# RFC-0001: Kernel Architecture — Three Runtimes

- **Status**: PROPOSED → **REVISED**
- **Priority**: P0 — Foundation
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Revised**: 2026-07-03 (Review fixes: A4, A5, C2, C3)
- **Depends on**: RFC-0000 (State Machine)

## Summary

The kernel is organized into **three specialized runtimes** and **two shared services**, each with a clear responsibility boundary. This replaces the current flat kernel structure with a layered architecture inspired by OS kernel design.

## Motivation

The current kernel is a flat collection of packages (config, logger, eventbus, registry, runtime, scheduler, lifecycle). This creates unclear boundaries:
- Where does "brain logic" live vs "execution logic"?
- Who owns process lifecycle — provider or runtime?
- Where does plugin management end and execution begin?

## Design

### Architecture Overview

```
                              Kernel (FSM: Created→Booting→Running→Stopped)
                                │
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
    3 Runtimes              2 Shared Services       Infrastructure
         │                      │                      │
   ┌─────┼─────┐          ┌─────┼─────┐          ┌────┼────┐
   ▼     ▼     ▼          ▼           ▼          ▼    ▼    ▼
Execution Brain Plugin  Knowledge  History    Config Logger EventBus
Runtime  Runtime Runtime  Engine   Timeline
```

### Three Runtimes

```
Execution Runtime             Brain Runtime              Plugin Runtime
(kernel/execution/)           (kernel/brain/)            (kernel/plugin/)
─────────────────────         ──────────────────         ─────────────────
Executor                      Decision Engine            Registry[T]
Worker Pool                   Planning Engine            Lifecycle Manager
Task Dispatcher               Policy Engine              Capability Index
Result Collector              Context Engine             Health Monitor
Process Manager               (Cognitive — Phase 5+)     Discovery
Resource Manager
```

### Two Shared Services (Fix for Issue A4)

> [!IMPORTANT]
> Knowledge Engine and History Timeline are **shared kernel services**, NOT owned by any single runtime. All three runtimes access them via `contracts/` interfaces.

```
Shared Services
─────────────────────────────
Knowledge Engine (kernel/knowledge/)
  → Brain reads templates, patterns, facts for planning/decisions
  → Execution writes outcomes for learning
  → Plugin queries capabilities

History Timeline (kernel/history/)
  → FSM auto-appends transitions (via OnTransition callback)
  → Brain queries past decisions for learning
  → Execution queries task history for retry logic
  → All: audit, replay, debugging
```

### 1. Execution Runtime (`kernel/execution/`)

**Responsibility**: Execute tasks. Manage processes, workers, resources. The "muscle".

```
kernel/execution/
├── runtime.go          # Execution Runtime lifecycle (FSM: Created→Running→Stopped)
├── executor.go         # Execute a single task on a single agent
├── pool.go             # Worker pool (bounded goroutines via semaphore)
├── dispatcher.go       # Dispatch tasks from queue to workers
├── collector.go        # Collect results from workers
├── process/
│   ├── manager.go      # Spawn/monitor/kill OS processes
│   ├── process.go      # Process abstraction (FSM: Started→Running→Exited)
│   └── stdio.go        # Stdin/stdout/stderr management
└── resource/
    ├── manager.go       # Track resource usage (sessions, CPU, RAM, quotas)
    ├── quota.go         # API quota tracking per provider
    └── monitor.go       # System resource monitoring
```

#### Task Queue Ownership (Fix for Issue C2)

> [!IMPORTANT]
> **Brain produces, Execution consumes.**
> - Brain Runtime determines task order (DAG topological sort + priority)
> - Brain pushes ordered tasks into a **shared task queue** (channel or priority queue)
> - Execution Runtime pops from queue → dispatches to workers
> - Queue is owned by Kernel, injected into both runtimes

```go
// kernel/kernel.go
type Kernel struct {
    // ...
    taskQueue chan *brain.TaskSpec  // Brain produces, Execution consumes
}
```

### 2. Brain Runtime (`kernel/brain/`)

**Responsibility**: Think. Make decisions. Plan. The "brain".

```
kernel/brain/
├── runtime.go          # Brain Runtime lifecycle
├── decision/
│   ├── engine.go       # Rule-based decision engine
│   ├── rules.go        # Built-in decision rules
│   └── strategy.go     # Executor selection strategies
├── planning/
│   ├── planner.go      # Template-first planner
│   ├── templates.go    # Plan template library
│   ├── dag.go          # DAG construction and validation
│   └── replanner.go    # Replan on failure (deterministic rules)
├── policy/
│   ├── engine.go       # Policy evaluation engine
│   ├── security.go     # Security policies
│   └── resource.go     # Resource usage policies
├── context/
│   ├── engine.go       # Context Engine (assemble context for AI calls)
│   ├── builder.go      # Build context from multiple sources
│   ├── ranker.go       # Rank relevance of context items
│   ├── compressor.go   # Compress to fit token window
│   ├── window.go       # Token window management
│   └── cache.go        # Cache assembled contexts
└── cognitive/          # Phase 5+ (interfaces defined now, implemented later)
    ├── reflection.go
    ├── learning.go
    └── optimization.go
```

**Key principle**: Brain NEVER calls AI for decisions. Brain uses deterministic Go logic. AI is only called (via Execution Runtime) when content generation is needed.

### 3. Plugin Runtime (`kernel/plugin/`)

**Responsibility**: Manage plugin lifecycle. Registration, discovery, health monitoring.

```
kernel/plugin/
├── runtime.go          # Plugin Runtime lifecycle
├── registry.go         # Generic plugin registry (Go generics)
├── lifecycle.go        # Init → Start → Stop lifecycle management
├── capability.go       # Capability indexing and matching
├── health.go           # Plugin health monitoring
└── discovery.go        # Plugin auto-discovery
```

**Registry design** (generic, not 6 separate registries):

```go
// One generic Registry with type-safe operations.
type Registry[T plugin.Plugin] struct {
    mu    sync.RWMutex
    items map[string]T
    caps  map[string][]string  // capability → [plugin names]
}

func (r *Registry[T]) Register(name string, p T) error
func (r *Registry[T]) Get(name string) (T, bool)
func (r *Registry[T]) List() []T
func (r *Registry[T]) FindByCapability(cap string) []T
func (r *Registry[T]) Unregister(name string) error
```

### Provider Interface (Fix for Issue A5)

> [!IMPORTANT]
> Provider is split into **base + specialized** interfaces. API providers handle HTTP calls directly. CLI providers define commands — Execution Runtime spawns the process.

```go
// contracts/provider/provider.go — Base interface (all providers)
type Provider interface {
    // Info returns provider metadata.
    Info() ProviderInfo
    // Capabilities returns what this provider can do.
    Capabilities() []Capability
    // ModelConfig returns model-specific configuration.
    ModelConfig() ModelConfig
}

// contracts/provider/api.go — API providers (Gemini REST, OpenAI, etc.)
// These providers handle HTTP/gRPC calls directly.
type APIProvider interface {
    Provider
    // Complete sends a request and returns a complete response.
    Complete(ctx context.Context, req Request) (*Response, error)
    // Stream sends a request and returns a streaming response.
    Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
}

// contracts/provider/cli.go — CLI providers (Antigravity CLI, Claude CLI, etc.)
// These providers define commands — Execution Runtime spawns processes.
type CLIProvider interface {
    Provider
    // BuildCommand creates the OS command to execute.
    // The Execution Runtime's Process Manager will spawn and manage it.
    BuildCommand(req Request) (*CLICommand, error)
    // ParseOutput interprets stdout/stderr from the completed process.
    ParseOutput(result *ProcessResult) (*Response, error)
}

// CLICommand describes a command to be spawned by Execution Runtime.
type CLICommand struct {
    Program string
    Args    []string
    Env     map[string]string
    Dir     string
    Stdin   string  // Optional input to send via stdin
}

// ProcessResult is what the Process Manager returns after a command completes.
type ProcessResult struct {
    ExitCode int
    Stdout   []byte
    Stderr   []byte
    Duration time.Duration
}
```

**Why this matters**: 
- CLI provider doesn't need to handle process lifecycle → simpler, testable
- API provider doesn't need fake Translate/Parse methods → clean interface
- Execution Runtime's Process Manager is the ONLY place that spawns OS processes
- Testing: mock Process Manager, no real processes needed

### Kernel Bootstrap (Fix for Issue C3)

```go
// kernel/kernel.go
type Kernel struct {
    fsm        fsm.Machine       // Kernel FSM: Created→Booting→Running→ShuttingDown→Stopped
    config     *Config
    logger     *slog.Logger
    eventBus   event.Bus
    timeline   history.Timeline

    execution  *execution.Runtime
    brain      *brain.Runtime
    plugin     *plugin.Runtime
    knowledge  *knowledge.Engine   // Shared service
    
    taskQueue  chan *brain.TaskSpec // Brain → Execution pipeline
}

func (k *Kernel) Start(ctx context.Context) error {
    // 1. Fire FSM: Created → Booting
    if err := k.fsm.Fire(ctx, "boot", nil); err != nil {
        return fmt.Errorf("kernel boot transition failed: %w", err)
    }

    // 2. Start shared services first
    if err := k.knowledge.Start(ctx); err != nil {
        k.fsm.Fire(ctx, "fail", err)
        return fmt.Errorf("knowledge engine start failed: %w", err)
    }

    // 3. Start Plugin Runtime (load + init plugins)
    if err := k.plugin.Start(ctx); err != nil {
        k.fsm.Fire(ctx, "fail", err)
        return fmt.Errorf("plugin runtime start failed: %w", err)
    }

    // 4. Start Brain Runtime (load rules, templates, knowledge)
    if err := k.brain.Start(ctx); err != nil {
        k.fsm.Fire(ctx, "fail", err)
        return fmt.Errorf("brain runtime start failed: %w", err)
    }

    // 5. Start Execution Runtime (spawn worker pool)
    if err := k.execution.Start(ctx); err != nil {
        k.fsm.Fire(ctx, "fail", err)
        return fmt.Errorf("execution runtime start failed: %w", err)
    }

    // 6. Fire FSM: Booting → Running
    if err := k.fsm.Fire(ctx, "ready", nil); err != nil {
        return fmt.Errorf("kernel ready transition failed: %w", err)
    }
    
    k.logger.Info("kernel started successfully")
    return nil
}

func (k *Kernel) Shutdown(ctx context.Context) error {
    k.fsm.Fire(ctx, "shutdown", nil)
    
    // Reverse order: execution → brain → plugin → shared services
    k.execution.Stop(ctx)
    k.brain.Stop(ctx)
    k.plugin.Stop(ctx)
    k.knowledge.Stop(ctx)
    
    k.fsm.Fire(ctx, "stopped", nil)
    return nil
}
```

### Import Rules (Updated)

```
contracts/        ← imports NOTHING (only stdlib)
    ↓
kernel/fsm/       ← imports contracts/fsm/
kernel/knowledge/ ← imports contracts/knowledge/, contracts/history/
kernel/brain/     ← imports contracts/brain/, contracts/knowledge/ (read-only)
kernel/execution/ ← imports contracts/agent/, contracts/provider/, contracts/knowledge/ (write outcomes)
kernel/plugin/    ← imports contracts/plugin/, contracts/tool/
    ↓
sdk/              ← imports contracts/
    ↓
plugins/          ← imports contracts/ + sdk/
    ↓
modules/          ← imports contracts/ + kernel/
    ↓
cmd/              ← imports everything (wires the system together)
```

## Impact

### New Packages
- `kernel/execution/` — Replaces current `kernel/runtime/`
- `kernel/execution/process/` — Process management (extracted from providers)
- `kernel/execution/resource/` — Resource tracking
- `kernel/brain/` — Replaces standalone `kernel/planner/`, `kernel/orchestrator/`
- `kernel/brain/decision/` — Decision engine
- `kernel/brain/planning/` — Planner (moved from `kernel/planner/`)
- `kernel/brain/policy/` — Policy engine (absorbs `kernel/security/`)
- `kernel/brain/context/` — Context engine
- `kernel/brain/cognitive/` — Cognitive layer (Phase 5+)
- `kernel/plugin/` — Replaces current `kernel/registry/`
- `kernel/knowledge/` — Shared service
- `kernel/history/` — Shared service

### Removed/Merged Packages
- `kernel/runtime/` → merged into `kernel/execution/`
- `kernel/registry/` → merged into `kernel/plugin/`
- `kernel/planner/` → merged into `kernel/brain/planning/`
- `kernel/orchestrator/` → orchestration distributed across brain + execution
- `kernel/security/` → merged into `kernel/brain/policy/`
- `kernel/feedback/` → merged into `kernel/brain/cognitive/`

## Open Questions

1. ~~**Orchestrator as concept**~~ **RESOLVED**: No standalone orchestrator. Brain decides WHAT, Execution does HOW. Kernel coordinates them via shared task queue.

2. ~~**Context Engine placement**~~ **RESOLVED**: Under Brain Runtime. Execution Runtime accesses it via `contracts/brain/` ContextEngine interface when needed.

3. ~~**Task queue ownership**~~ **RESOLVED**: Brain produces task order, Execution consumes. Queue owned by Kernel, injected into both runtimes.

4. ~~**Provider interface**~~ **RESOLVED**: Split into `APIProvider` (handles HTTP calls) and `CLIProvider` (defines commands, Runtime spawns).
