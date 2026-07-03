# Micro-Task 2.24: Create kernel/runtime/dispatcher.go

## Info
- **File**: `kernel/runtime/dispatcher.go`
- **Package**: `runtime`
- **Depends on**: 2.22 (executor.go), 2.23 (pool.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Implements the task dispatcher (`Dispatcher`, `TaskResult`, `DispatcherConfig` and constructors) that coordinates task submission from scheduler channels into worker execution pools, collecting results through structured channels.

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

	if d.logger != nil {
		d.logger.Debug("dispatching task",
			"task_id", string(task.ID),
			"task_name", task.Name,
		)
	}

	return d.pool.Submit(ctx, func(ctx context.Context) {
		result, err := d.executor.ExecuteTask(ctx, task)

		// Send result back (blocking with context abort fallback to prevent hangs or drops)
		select {
		case d.results <- TaskResult{
			TaskID: task.ID,
			Result: result,
			Error:  err,
		}:
		case <-ctx.Done():
			if d.logger != nil {
				d.logger.Warn("context cancelled, result send aborted",
					"task_id", string(task.ID),
					"error", err,
				)
			}
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

	if d.logger != nil {
		d.logger.Info("dispatcher stopped")
	}
}
```

## Rules
1. **Context-Aware Result Sends**: Write result transmissions (`d.results <-`) inside `select` blocks containing context cancellation checks (`<-ctx.Done()`). This avoids deadlocks during system shutdowns while preventing silent event drops.
2. **Dispatch Rejection Guards**: Block incoming tasks with error returns if the dispatcher is stopped.
3. **Queue Size Configuration**: Set dispatcher result channel buffers to at least the configured worker pool capacity to prevent backpressure.

## ⚠️ Pitfalls

### Pitfall 1: Dropping completed task results silently on full channels
Using a `default` case to drop results when the results channel buffer is full causes the scheduler to wait forever for tasks that have already completed. Always block on the write channel, but include a `<-ctx.Done()` branch to allow aborting during shutdowns.

### Pitfall 2: Permitting dispatches after dispatcher deactivations
Failing to check `d.stopped` under read locks during `Dispatch()` calls allows tasks to be submitted after deactivation. This leaves tasks hanging in the queue, never executing.

## Verify
```bash
go build ./kernel/runtime/...
```

## Checklist
- [ ] File `kernel/runtime/dispatcher.go` exists
- [ ] Package: `runtime`
- [ ] `TaskResult` holds TaskID, Result, and Error properties
- [ ] `NewDispatcher` initializes result channels with configurable buffers
- [ ] `Dispatch` rejects tasks with error returns after stopping
- [ ] Worker routines dispatch results using context-aware channels
- [ ] `Results` returns a read-only channel
- [ ] `Stop` changes stopped state under sync mutex locks
- [ ] `go build ./kernel/runtime/...` passes
