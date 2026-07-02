# Micro-Task 2.24: Create kernel/runtime/dispatcher.go

## Info
- **File**: `kernel/runtime/dispatcher.go`
- **Package**: `runtime`
- **Depends on**: 2.22 (executor.go), 2.23 (pool.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Routes tasks from the scheduler to the worker pool. Bridges scheduler and executor.

## EXACT code to create

```go
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// TaskResult pairs a task ID with its execution result.
// Used to communicate results back from worker goroutines.
type TaskResult struct {
	TaskID contracts.TaskID
	Result *agent.Result
	Error  error
}

// Dispatcher routes tasks from the scheduler to the worker pool.
//
// Architecture:
//   Scheduler → Dispatcher.Dispatch(task) → Pool.Submit → Executor.ExecuteTask
//
// Results are collected via a channel. The scheduler or orchestrator reads
// results to update task status and handle dependencies.
//
// Thread-safety: safe for concurrent use.
type Dispatcher struct {
	executor *Executor
	pool     *Pool
	logger   *slog.Logger

	// results receives completed task results.
	// Buffer size should be >= max concurrent tasks to avoid backpressure.
	results chan TaskResult

	// mu protects the stopped flag.
	mu      sync.RWMutex
	stopped bool
}

// DispatcherConfig configures the Dispatcher.
type DispatcherConfig struct {
	// ResultBufferSize is the size of the results channel buffer.
	// Default: 100
	ResultBufferSize int
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(executor *Executor, pool *Pool, logger *slog.Logger, cfg DispatcherConfig) *Dispatcher {
	if cfg.ResultBufferSize <= 0 {
		cfg.ResultBufferSize = 100
	}

	return &Dispatcher{
		executor: executor,
		pool:     pool,
		logger:   logger,
		results:  make(chan TaskResult, cfg.ResultBufferSize),
	}
}

// Dispatch sends a task to the worker pool for execution.
//
// This method is NON-BLOCKING:
//   - If a worker slot is available → task starts immediately
//   - If all slots are busy → blocks until a slot opens or ctx is cancelled
//
// Results are sent to the Results() channel.
//
// Parameters:
//   - ctx: parent context (cancelled on shutdown)
//   - task: the task to execute
//
// Returns error only if the task cannot be submitted (pool full + ctx cancelled).
// Execution errors are reported via the Results() channel.
func (d *Dispatcher) Dispatch(ctx context.Context, task *agent.Task) error {
	d.mu.RLock()
	if d.stopped {
		d.mu.RUnlock()
		return fmt.Errorf("dispatcher: stopped, cannot dispatch task %q", task.Name)
	}
	d.mu.RUnlock()

	d.logger.Debug("dispatching task",
		"task_id", string(task.ID),
		"task_name", task.Name,
	)

	return d.pool.Submit(ctx, func(ctx context.Context) {
		result, err := d.executor.ExecuteTask(ctx, task)

		// Send result back (non-blocking if buffer is large enough)
		select {
		case d.results <- TaskResult{
			TaskID: task.ID,
			Result: result,
			Error:  err,
		}:
		default:
			// Buffer full — log warning (results are being consumed too slowly)
			d.logger.Warn("result buffer full, dropping result",
				"task_id", string(task.ID),
				"error", err,
			)
		}
	})
}

// Results returns the channel that receives task execution results.
//
// The caller should read from this channel continuously to avoid
// backpressure on the worker pool.
//
// Channel is closed when Stop() is called and all workers finish.
func (d *Dispatcher) Results() <-chan TaskResult {
	return d.results
}

// Stop prevents new tasks from being dispatched.
// Does NOT wait for in-flight tasks — call pool.Wait() separately.
func (d *Dispatcher) Stop() {
	d.mu.Lock()
	d.stopped = true
	d.mu.Unlock()

	d.logger.Info("dispatcher stopped")
}
```

## Pitfalls

### Pitfall 1: Non-blocking result send
```go
select {
case d.results <- result:  // Send if buffer has space
default:                    // Drop if buffer full
    d.logger.Warn("result buffer full")
}
```
If the results channel blocks → worker goroutine hangs → pool slot not released → deadlock.

### Pitfall 2: Stopped check before dispatch
```go
d.mu.RLock()
if d.stopped {
    d.mu.RUnlock()
    return fmt.Errorf("dispatcher: stopped")
}
d.mu.RUnlock()
```
After Stop(), new dispatches should be rejected. Otherwise → tasks queued after shutdown → never complete.

### Pitfall 3: Result buffer size
Default 100. If you have maxWorkers=5 and consumer reads results quickly → 100 is plenty.
If consumer is slow → results accumulate → buffer fills → results dropped.
Monitor `dropped result` log warnings in production.

### Pitfall 4: TaskResult contains Error
```go
type TaskResult struct {
    TaskID contracts.TaskID
    Result *agent.Result  // nil on error
    Error  error           // nil on success
}
```
Caller MUST check Error first. If Error != nil, Result is nil.

## Checklist
- [ ] File `kernel/runtime/dispatcher.go` exists
- [ ] TaskResult struct with TaskID, Result, Error
- [ ] Dispatcher struct with executor, pool, results channel
- [ ] `NewDispatcher()` with default buffer size
- [ ] `Dispatch(ctx, task)` — submits to pool, sends results via channel
- [ ] `Results()` — returns receive-only channel
- [ ] `Stop()` — prevents new dispatches
- [ ] Non-blocking result send (select + default)
- [ ] Stopped check before dispatch
- [ ] `go build ./kernel/runtime/...` no errors
