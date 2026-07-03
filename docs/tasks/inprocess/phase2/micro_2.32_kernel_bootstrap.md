# Micro-Task 2.32: Create kernel/kernel.go

## Info
- **File**: `kernel/kernel.go`
- **Package**: `kernel`
- **Depends on**: 2.01-2.31 (all kernel components)
- **Time**: 25 min
- **Verify**: `go build ./kernel/...`

## Purpose
Implements the main orchestrator kernel (`Kernel` and constructors) that serves as the bootstrap coordinator. It instantiates the subsystems (logger, eventbus, registry, runtime, scheduler) in their strict dependency order and manages their execution lifecycles.

## EXACT code to create

```go
package kernel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
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
//   Config    → loads settings
//   Logger    → structured logging
//   EventBus  → publish/subscribe events
//   Registry  → plugin management
//   Runtime   → task execution (executor + pool + dispatcher)
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
		MaxWorkers:       cfg.Orchestrator.MaxConcurrentTasks,
		DefaultTimeout:   cfg.Orchestrator.ShutdownTimeout,
		ResultBufferSize: 100,
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

## Rules
1. **Import Aliases**: Aliases `klogger` and `kruntime` are required to avoid namespace clashes with standard Go library packages (`log` / `runtime`).
2. **Order of Dependency Instantiation**: Subsystems must be initialized in strict sequence: Logger first, then EventBus, Registry, Runtime, and finally Scheduler. Breaking this sequence will lead to nil pointer dereference panics.
3. **Best-Effort Shutdown Teardown**: Subsystem stops must continue even if one of them fails. Log any errors and continue clearing components.

## ⚠️ Pitfalls

### Pitfall 1: Incorrect dependency instantiation sequence
Attempting to instantiate components out of order (such as creating the Registry before the Logger is initialized) triggers panic crashes during bootstrap. Follow the documented initialization sequence.

### Pitfall 2: Forgetting import libraries for the agent package
The `kernel.go` file uses `agent.Task` pointers inside the scheduler dispatch function closure. Ensure `"github.com/tiendat1751998/orchestrator/contracts/agent"` is explicitly included in the imports block.

## Verify
```bash
go build ./kernel/...
```

## Checklist
- [ ] File `kernel/kernel.go` exists
- [ ] Package: `kernel`
- [ ] Kernel aggregates configuration settings, log systems, states, event busses, registries, runtimes, and schedulers
- [ ] Subsystems are wired sequentially in `New`
- [ ] `RegisterPlugin` accepts plugins only when the kernel is in the `StateCreated` state
- [ ] `Start` validates settings, launches loops, and sets states
- [ ] `Stop` executes LIFO shutdown sequences
- [ ] Import aliases are configured for conflicting package paths
- [ ] `go build ./kernel/...` passes
