# Micro-Task 2.32: Create kernel/kernel.go

## Info
- **File**: `kernel/kernel.go`
- **Package**: `kernel`
- **Depends on**: 2.01-2.31 (all kernel components)
- **Time**: 25 min
- **Verify**: `go build ./kernel/...`

## Purpose
Main kernel struct. Wires all components together.
Provides Start/Stop lifecycle that orchestrates all subsystems.

## EXACT code to create

```go
package kernel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
	klogger "github.com/tiendat1751998/orchestrator/kernel/logger"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
	kruntime "github.com/tiendat1751998/orchestrator/kernel/runtime"
	"github.com/tiendat1751998/orchestrator/kernel/scheduler"
)

// Kernel is the orchestrator core.
//
// It wires together all subsystems:
//   Config   → loads settings
//   Logger   → structured logging
//   EventBus → publish/subscribe events
//   Registry → plugin management
//   Runtime  → task execution (executor + pool + dispatcher)
//   Scheduler → task scheduling (priority queue + dependencies)
//
// Lifecycle:
//   New(cfg)                → create kernel
//   RegisterPlugin(plugin)  → add plugins
//   Start(ctx)              → init plugins, start subsystems
//   [operate]               → submit tasks, handle events
//   Stop(ctx)               → graceful shutdown
//
// State machine: Created → Initializing → Running → ShuttingDown → Stopped
type Kernel struct {
	cfg       *config.Config
	log       *klogger.Logger
	state     *StateMachine
	eventBus  *eventbus.Bus
	registry  *registry.Registry
	runtime   *kruntime.Runtime
	scheduler *scheduler.Scheduler
}

// New creates a new Kernel from the given config.
//
// This does NOT start the kernel. Call Start() after registering plugins.
//
// Initialization order:
//   1. Logger (needed by all other components for logging)
//   2. EventBus (needed by registry and runtime for event emission)
//   3. Registry (needed by runtime for agent lookup)
//   4. Runtime (needs registry and eventbus)
//   5. Scheduler (needs runtime dispatcher function)
//
// Why this order? Each component depends on the previous one.
// Changing the order → nil pointer panic.
func New(cfg *config.Config) (*Kernel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kernel: config is nil")
	}

	// 1. Create logger
	log := klogger.New(klogger.Options{
		Level:  cfg.Orchestrator.LogLevel,
		Format: cfg.Orchestrator.LogFormat,
	})

	// 2. Create event bus
	bus := eventbus.New(log.Slog())

	// 3. Create registry
	reg := registry.New(log.Slog())

	// 4. Create runtime
	rt := kruntime.New(reg, bus, log.Slog(), kruntime.Config{
		MaxWorkers:     cfg.Orchestrator.MaxConcurrentTasks,
		DefaultTimeout: cfg.Orchestrator.ShutdownTimeout,
	})

	// 5. Create scheduler
	// The scheduler's dispatch function calls runtime.Dispatch.
	// This bridges scheduler → runtime without circular imports.
	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error {
			return rt.Dispatch(ctx, task)
		},
		log.Slog(),
	)

	return &Kernel{
		cfg:       cfg,
		log:       log,
		state:     NewStateMachine(),
		eventBus:  bus,
		registry:  reg,
		runtime:   rt,
		scheduler: sched,
	}, nil
}

// RegisterPlugin adds a plugin to the kernel.
//
// Must be called BEFORE Start().
// Calling after Start() returns an error.
//
// Plugins include: agents, providers, tools, and other extensions.
func (k *Kernel) RegisterPlugin(p plugin.Plugin) error {
	if !k.state.Is(StateCreated) {
		return fmt.Errorf("kernel: cannot register plugin in state %v (must be in Created state)", k.state.Current())
	}

	return k.registry.Register(p)
}

// Start initializes and starts all kernel subsystems.
//
// Startup sequence:
//   1. Validate config
//   2. Transition to Initializing state
//   3. Init all plugins (via registry)
//   4. Start all plugins (via registry)
//   5. Start runtime (result processor)
//   6. Start scheduler loop
//   7. Transition to Running state
//   8. Emit "kernel.started" event
//
// If any step fails, the kernel transitions to Stopped and returns the error.
// Already-started components are cleaned up.
func (k *Kernel) Start(ctx context.Context) error {
	// Validate config
	if err := config.Validate(k.cfg); err != nil {
		return fmt.Errorf("kernel: %w", err)
	}

	// Transition: Created → Initializing
	if err := k.state.Transition(StateInitializing); err != nil {
		return err
	}

	k.log.Info("kernel starting",
		"name", k.cfg.Orchestrator.Name,
		"max_workers", k.cfg.Orchestrator.MaxConcurrentTasks,
	)

	// Init plugins
	if err := k.registry.InitAll(ctx, nil); err != nil {
		k.state.Transition(StateStopped)
		return fmt.Errorf("kernel: init plugins: %w", err)
	}

	// Start plugins
	if err := k.registry.StartAll(ctx); err != nil {
		k.state.Transition(StateStopped)
		return fmt.Errorf("kernel: start plugins: %w", err)
	}

	// Start runtime
	k.runtime.Start(func(result kruntime.TaskResult) {
		// Notify scheduler that task completed (for dependency resolution)
		k.scheduler.NotifyCompleted(result.TaskID)
	})

	// Start scheduler loop in background goroutine
	go k.scheduler.Run(ctx)

	// Transition: Initializing → Running
	if err := k.state.Transition(StateRunning); err != nil {
		k.state.Transition(StateStopped)
		return err
	}

	// Emit startup event
	eventbus.PublishKernelStarted(k.eventBus)

	k.log.Info("kernel started successfully",
		"name", k.cfg.Orchestrator.Name,
		"plugins", k.registry.Count(),
		"agents", len(k.registry.ListAgents()),
		"providers", len(k.registry.ListProviders()),
	)

	return nil
}

// Stop gracefully shuts down the kernel.
//
// Shutdown sequence (REVERSE of startup):
//   1. Transition to ShuttingDown state
//   2. Emit "kernel.stopping" event
//   3. Stop runtime (wait for in-flight tasks)
//   4. Stop all plugins (reverse order)
//   5. Close event bus
//   6. Transition to Stopped state
//   7. Log final message
//
// Idempotent: calling Stop on an already-stopped kernel is a no-op.
func (k *Kernel) Stop(ctx context.Context) error {
	if k.state.IsStopped() {
		return nil // Already stopped
	}

	// Transition: Running → ShuttingDown
	if err := k.state.Transition(StateShuttingDown); err != nil {
		// If not in Running state, force stop
		k.log.Warn("kernel stop from unexpected state",
			"current_state", k.state.Current().String(),
		)
	}

	k.log.Info("kernel stopping...")

	// Stop runtime (wait for tasks to complete)
	if err := k.runtime.Stop(ctx); err != nil {
		k.log.Error("runtime stop error", "error", err)
	}

	// Stop plugins (reverse registration order)
	if err := k.registry.StopAll(ctx); err != nil {
		k.log.Error("plugin stop error", "error", err)
	}

	// Close event bus (wait for in-flight events)
	k.eventBus.Close()

	// Transition: ShuttingDown → Stopped
	k.state.Transition(StateStopped)

	k.log.Info("kernel stopped")

	return nil
}

// =============================================================================
// Accessor Methods
// =============================================================================

// EventBus returns the kernel's event bus.
// Used by plugins to subscribe to events.
func (k *Kernel) EventBus() event.Bus {
	return k.eventBus
}

// Registry returns the kernel's plugin registry.
func (k *Kernel) Registry() *registry.Registry {
	return k.registry
}

// Logger returns the kernel's logger.
func (k *Kernel) Logger() *klogger.Logger {
	return k.log
}

// State returns the kernel's current state.
func (k *Kernel) State() State {
	return k.state.Current()
}

// Config returns the kernel's config (read-only).
func (k *Kernel) Config() *config.Config {
	return k.cfg
}
```

## Pitfalls

### Pitfall 1: Import alias to avoid collision
```go
import (
    klogger "github.com/.../kernel/logger"   // alias: klogger
    kruntime "github.com/.../kernel/runtime"  // alias: kruntime
)
```
`logger` and `runtime` collide with Go standard library names. Aliases avoid confusion.

### Pitfall 2: Scheduler dispatch function closure
```go
sched := scheduler.New(
    func(ctx context.Context, task *agent.Task) error {
        return rt.Dispatch(ctx, task)
    },
    log.Slog(),
)
```
This closure bridges scheduler → runtime. Scheduler package does NOT import runtime package.

### Pitfall 3: Runtime result → scheduler notification
```go
k.runtime.Start(func(result kruntime.TaskResult) {
    k.scheduler.NotifyCompleted(result.TaskID)
})
```
When a task completes, the scheduler is notified so dependent tasks can be unblocked.
This is the critical link between execution and scheduling.

### Pitfall 4: Stop is best-effort
```go
if err := k.runtime.Stop(ctx); err != nil {
    k.log.Error("runtime stop error", "error", err)
    // Continue stopping other components
}
```
If runtime stop fails → still stop plugins → still close event bus.
Don't abort shutdown because one component had trouble.

### Pitfall 5: Missing import for agent package
The kernel.go file uses `agent.Task` inside the scheduler dispatch function.
Make sure to add:
```go
import "github.com/tiendat1751998/orchestrator/contracts/agent"
```

## Checklist
- [ ] File `kernel/kernel.go` exists
- [ ] Package: `package kernel`
- [ ] Kernel struct with: cfg, log, state, eventBus, registry, runtime, scheduler
- [ ] `New(cfg)` — wires all components in correct order
- [ ] `RegisterPlugin(p)` — only in Created state
- [ ] `Start(ctx)` — validates config, inits/starts plugins, starts runtime/scheduler
- [ ] `Stop(ctx)` — reverse order shutdown, idempotent
- [ ] Accessor methods: EventBus(), Registry(), Logger(), State(), Config()
- [ ] Import aliases for kernel/logger and kernel/runtime
- [ ] Scheduler dispatch closure (no circular import)
- [ ] Runtime result → scheduler NotifyCompleted
- [ ] State machine transitions at each lifecycle step
- [ ] Best-effort shutdown (continues on errors)
- [ ] `go build ./kernel/...` no errors
