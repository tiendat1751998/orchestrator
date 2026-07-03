# Micro-Task 2.22: Create kernel/runtime/executor.go

## Info
- **File**: `kernel/runtime/executor.go`
- **Package**: `runtime`
- **Depends on**: Phase 1 contracts, 2.18 registry
- **Time**: 25 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Implements the task execution engine (`Executor`, `ExecutorConfig`, and constructors) that isolates task runs, handles agent matching lookups, applies timeout contexts, recovers panics, and publishes telemetry events.

## EXACT code to create

```go
// Package runtime provides the task execution engine.
//
// Architecture:
//   Executor   — runs a single task on an agent (timeout, panic recovery, events)
//   Pool       — limits concurrent goroutines (semaphore pattern)
//   Dispatcher — routes tasks from scheduler to workers
//   Runtime    — ties everything together (Start/Stop lifecycle)
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// Executor runs a single task on a suitable agent.
//
// Execution steps:
//   1. Find agent from registry (via FindAgentForTask)
//   2. Create child context with timeout
//   3. Emit "task.started" event
//   4. Call agent.Execute(ctx, task) with panic recovery
//   5. Emit "task.completed" or "task.failed" event
//   6. Return result or error
//
// Thread-safety: Executor is stateless and safe for concurrent use.
// Multiple goroutines can call ExecuteTask simultaneously.
type Executor struct {
	registry *registry.Registry
	eventBus event.Bus
	logger   *slog.Logger

	// defaultTimeout is used when task.Timeout is not set.
	defaultTimeout time.Duration
}

// ExecutorConfig configures the Executor.
type ExecutorConfig struct {
	// DefaultTimeout for tasks without an explicit timeout.
	// Default: 120s (matching the provider timeout).
	DefaultTimeout time.Duration
}

// NewExecutor creates a new Executor.
func NewExecutor(reg *registry.Registry, bus event.Bus, logger *slog.Logger, cfg ExecutorConfig) *Executor {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 120 * time.Second
	}

	return &Executor{
		registry:       reg,
		eventBus:       bus,
		logger:         logger,
		defaultTimeout: cfg.DefaultTimeout,
	}
}

// ExecuteTask runs a single task end-to-end.
//
// This is the core execution function. It handles:
//   - Agent selection (from registry)
//   - Timeout enforcement (child context)
//   - Panic recovery (prevents agent crash from killing the system)
//   - Event emission (task.started, task.completed, task.failed)
//   - Duration measurement
//
// The caller (Dispatcher) runs this in a goroutine from the worker pool.
//
// Parameters:
//   - ctx: parent context (may have deadline from scheduler)
//   - task: the task to execute
//
// Returns:
//   - *agent.Result: on success
//   - error: on failure (timeout, agent error, panic, no agent found)
func (e *Executor) ExecuteTask(ctx context.Context, task *agent.Task) (result *agent.Result, err error) {
	startTime := time.Now()
	taskID := string(task.ID)

	// Step 1: Find agent
	selectedAgent, err := e.registry.FindAgentForTask(task)
	if err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}
	agentName := selectedAgent.Name()

	// Step 2: Create child context with timeout
	timeout := e.defaultTimeout
	if task.Timeout > 0 {
		timeout = task.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Step 3: Emit task.started
	if e.eventBus != nil {
		eventbus.PublishTaskStarted(e.eventBus, taskID, agentName)
	}

	if e.logger != nil {
		e.logger.Info("task execution started",
			"task_id", taskID,
			"task_name", task.Name,
			"agent", agentName,
			"timeout", timeout,
		)
	}

	// Step 4: Execute with panic recovery
	//
	// WHY defer + recover instead of running in a sub-goroutine?
	// → Simpler control flow (no channel needed for result/error)
	// → recover() only works in the SAME goroutine as the panic
	// → The agent.Execute call runs synchronously within this goroutine
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("executor: agent %q panicked: %v\n%s", agentName, r, stack)
			result = nil

			if e.logger != nil {
				e.logger.Error("agent panicked during execution",
					"task_id", taskID,
					"agent", agentName,
					"panic", fmt.Sprintf("%v", r),
					"stack", stack,
				)
			}
		}

		// Step 5: Emit completion event
		duration := time.Since(startTime)

		if err != nil {
			if e.eventBus != nil {
				eventbus.PublishTaskFailed(e.eventBus, taskID, err)
			}
			if e.logger != nil {
				e.logger.Error("task execution failed",
					"task_id", taskID,
					"agent", agentName,
					"duration", duration,
					"error", err,
				)
			}
		} else {
			if e.eventBus != nil {
				eventbus.PublishTaskCompleted(e.eventBus, taskID, result)
			}
			if e.logger != nil {
				e.logger.Info("task execution completed",
					"task_id", taskID,
					"agent", agentName,
					"duration", duration,
					"status", result.Status,
				)
			}
		}
	}()

	// Execute the task on the agent
	result, err = selectedAgent.Execute(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("executor: agent %q failed: %w", agentName, err)
	}

	return result, nil
}
```

## Rules
1. **Named Return values for Panic Recoveries**: The `ExecuteTask` method signature must declare named return values `(result *agent.Result, err error)`. This is required to allow the deferred recovery function block to alter the return values on panic.
2. **Deterministic Event Emissions**: Place the completion/failure event emission logic inside the deferred cleanup closure block to ensure events are published even if the agent panics.
3. **Trace diagnostics**: Include `debug.Stack()` captures inside panic recoveries to print detailed diagnostic stack logs.
4. **Context Leak Safeguards**: Timeout contexts (`context.WithTimeout`) must call `cancel()` via a deferred statement to release memory.

## ⚠️ Pitfalls

### Pitfall 1: Omiting named returns inside functions using panic recoveries
Without named returns, the deferred `recover()` block can catch panics but has no way to assign the error back to the caller. This results in the function returning `nil, nil` to the caller, masking the crash.

### Pitfall 2: Forgetting to defer cancel functions inside contexts
Leaving `cancel()` uncalled creates a resource leak. Go's runtime context remains allocated in memory until the timeout expires. Always call `defer cancel()`.

## Verify
```bash
go build ./kernel/runtime/...
```

## Checklist
- [ ] File `kernel/runtime/executor.go` exists
- [ ] Package: `runtime`
- [ ] `ExecutorConfig` defines `DefaultTimeout` settings
- [ ] `NewExecutor` configures fallbacks when zero parameters are passed
- [ ] `ExecuteTask` returns named outputs `(result *agent.Result, err error)`
- [ ] Agent matched using registry lookups
- [ ] Execution context maps to child timeouts with deferred `cancel()`
- [ ] Task start, completed, and failure events are published
- [ ] Deferred closures catch panics and record stack traces using `debug.Stack()`
- [ ] `go build ./kernel/runtime/...` passes
