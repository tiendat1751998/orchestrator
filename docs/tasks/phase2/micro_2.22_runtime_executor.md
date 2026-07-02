# Micro-Task 2.22: Create kernel/runtime/executor.go

## Info
- **File**: `kernel/runtime/executor.go`
- **Package**: `runtime`
- **Depends on**: Phase 1 contracts, 2.18 registry
- **Time**: 25 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Execute a single task: find agent → set timeout → call agent.Execute → recover panics → emit events.

## EXACT code to create

```go
// Package runtime provides the task execution engine.
//
// Architecture:
//   Executor  — runs a single task on an agent (timeout, panic recovery, events)
//   Pool      — limits concurrent goroutines (semaphore pattern)
//   Dispatcher — routes tasks from scheduler to workers
//   Runtime   — ties everything together (Start/Stop lifecycle)
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

	e.logger.Info("task execution started",
		"task_id", taskID,
		"task_name", task.Name,
		"agent", agentName,
		"timeout", timeout,
	)

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

			e.logger.Error("agent panicked during execution",
				"task_id", taskID,
				"agent", agentName,
				"panic", fmt.Sprintf("%v", r),
				"stack", stack,
			)
		}

		// Step 5: Emit completion event
		duration := time.Since(startTime)

		if err != nil {
			if e.eventBus != nil {
				eventbus.PublishTaskFailed(e.eventBus, taskID, err)
			}
			e.logger.Error("task execution failed",
				"task_id", taskID,
				"agent", agentName,
				"duration", duration,
				"error", err,
			)
		} else {
			if e.eventBus != nil {
				eventbus.PublishTaskCompleted(e.eventBus, taskID, result)
			}
			e.logger.Info("task execution completed",
				"task_id", taskID,
				"agent", agentName,
				"duration", duration,
				"status", result.Status,
			)
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

## Pitfalls

### Pitfall 1: Named return values for panic recovery
```go
func (e *Executor) ExecuteTask(ctx context.Context, task *agent.Task) (result *agent.Result, err error) {
```
Named returns (`result`, `err`) allow the deferred recover() function to SET the return values.
Without named returns → recover() cannot communicate error back to caller.

### Pitfall 2: debug.Stack() for panic diagnostics
```go
stack := string(debug.Stack())
```
Stack trace is CRITICAL for debugging agent panics. Without it, you only know "something panicked" but not WHERE.

### Pitfall 3: Child context with timeout
```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```
`defer cancel()` is MANDATORY. Without it → context leak → resource leak.
Go vet catches this: "the cancel function is not used on all paths".

### Pitfall 4: Task timeout vs default timeout
```go
timeout := e.defaultTimeout  // 120s
if task.Timeout > 0 {
    timeout = task.Timeout    // Task-specific override
}
```
Task.Timeout = 0 means "use default". NOT "no timeout" (that would be dangerous).

### Pitfall 5: Event emission in defer (ALWAYS fires)
```go
defer func() {
    // This ALWAYS runs, even on panic
    if err != nil {
        PublishTaskFailed(...)
    } else {
        PublishTaskCompleted(...)
    }
}()
```
If we put events in the normal flow (not defer), a panic would skip the event → listeners never know the task failed.

## Checklist
- [ ] File `kernel/runtime/executor.go` exists
- [ ] Package: `package runtime`
- [ ] ExecutorConfig struct with DefaultTimeout
- [ ] `NewExecutor()` constructor with defaults
- [ ] `ExecuteTask()` with named returns (for panic recovery)
- [ ] Step 1: Find agent from registry
- [ ] Step 2: Child context with timeout
- [ ] Step 3: Emit task.started event
- [ ] Step 4: Panic recovery with debug.Stack()
- [ ] Step 5: Emit task.completed or task.failed (in defer)
- [ ] Duration measurement
- [ ] Structured logging at each step
- [ ] `defer cancel()` for timeout context
- [ ] `go build ./kernel/runtime/...` no errors
