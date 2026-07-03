# Micro-Task 2.25: Create kernel/runtime/runtime.go

## Info
- **File**: `kernel/runtime/runtime.go`
- **Package**: `runtime`
- **Depends on**: 2.22-2.24
- **Time**: 20 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Implements the high-level execution orchestrator (`Runtime`, `Config`, and constructors) that manages the lifecycle of the executor, worker pool, and task dispatcher, routing completed events to system consumer callbacks.

## EXACT code to create

```go
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// Runtime is the task execution engine.
//
// Lifecycle:
//   New() → Start() → Dispatch tasks → Stop()
//
// Components:
//   Executor   — runs a single task (timeout, panic recovery)
//   Pool       — limits concurrent workers
//   Dispatcher — routes tasks, collects results
type Runtime struct {
	executor   *Executor
	pool       *Pool
	dispatcher *Dispatcher
	logger     *slog.Logger

	// resultProcessor handles incoming task results.
	// It reads from Dispatcher.Results() and processes them.
	resultCancel context.CancelFunc
	resultWg     sync.WaitGroup

	mu      sync.Mutex
	running bool
}

// Config configures the Runtime.
type Config struct {
	// MaxWorkers limits concurrent task execution.
	// Default: 5
	MaxWorkers int

	// DefaultTimeout for tasks without explicit timeout.
	// Default: 120s
	DefaultTimeout time.Duration

	// ResultBufferSize for the result channel.
	// Default: 100
	ResultBufferSize int

	// OnResult is called when a task completes.
	// If nil, results are logged but not processed.
	OnResult func(TaskResult)
}

// New creates a new Runtime.
func New(reg *registry.Registry, bus event.Bus, logger *slog.Logger, cfg Config) *Runtime {
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 5
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 120 * time.Second
	}

	executor := NewExecutor(reg, bus, logger, ExecutorConfig{
		DefaultTimeout: cfg.DefaultTimeout,
	})

	pool := NewPool(cfg.MaxWorkers, logger)

	dispatcher := NewDispatcher(executor, pool, logger, DispatcherConfig{
		ResultBufferSize: cfg.ResultBufferSize,
	})

	return &Runtime{
		executor:   executor,
		pool:       pool,
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// Start begins processing task results.
//
// Must be called before Dispatch().
// Call Stop() to shut down.
//
// Parameters:
//   - onResult: callback for completed tasks (nil = log only)
func (r *Runtime) Start(onResult func(TaskResult)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return fmt.Errorf("runtime: already running")
	}

	// Start result processor goroutine
	ctx, cancel := context.WithCancel(context.Background())
	r.resultCancel = cancel
	r.running = true

	r.resultWg.Add(1)
	go func() {
		defer r.resultWg.Done()
		r.processResults(ctx, onResult)
	}()

	if r.logger != nil {
		r.logger.Info("runtime started",
			"max_workers", r.pool.MaxWorkers(),
		)
	}

	return nil
}

// processResults reads from the results channel and invokes the callback.
func (r *Runtime) processResults(ctx context.Context, onResult func(TaskResult)) {
	for {
		select {
		case <-ctx.Done():
			// Drain remaining results before exiting
			for {
				select {
				case result := <-r.dispatcher.Results():
					if onResult != nil {
						onResult(result)
					} else {
						r.logResult(result)
					}
				default:
					return
				}
			}
		case result := <-r.dispatcher.Results():
			if onResult != nil {
				onResult(result)
			} else {
				r.logResult(result)
			}
		}
	}
}

// logResult logs a task result (used when no OnResult callback is set).
func (r *Runtime) logResult(tr TaskResult) {
	if r.logger == nil {
		return
	}
	if tr.Error != nil {
		r.logger.Error("task result: failed",
			"task_id", string(tr.TaskID),
			"error", tr.Error,
		)
	} else {
		r.logger.Info("task result: completed",
			"task_id", string(tr.TaskID),
			"status", tr.Result.Status,
		)
	}
}

// Dispatch sends a task for execution.
//
// The task will be:
//   1. Queued in the worker pool (blocks if pool is full)
//   2. Executed by the Executor (find agent, timeout, panic recovery)
//   3. Result sent to the OnResult callback
//
// Returns error only if the task cannot be queued (runtime stopped or ctx cancelled).
func (r *Runtime) Dispatch(ctx context.Context, task *agent.Task) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return fmt.Errorf("runtime: not running")
	}
	r.mu.Unlock()

	return r.dispatcher.Dispatch(ctx, task)
}

// Stop gracefully shuts down the runtime.
//
// Shutdown sequence:
//   1. Stop dispatcher (reject new tasks)
//   2. Wait for all in-flight tasks to complete
//   3. Stop result processor
//   4. Log final statistics
//
// Parameters:
//   - ctx: shutdown deadline (e.g., 30-second timeout)
func (r *Runtime) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil // Already stopped, no-op
	}
	r.running = false
	r.mu.Unlock()

	if r.logger != nil {
		r.logger.Info("runtime stopping...")
	}

	// Step 1: Stop accepting new tasks
	r.dispatcher.Stop()

	// Step 2: Wait for in-flight tasks (with timeout)
	done := make(chan struct{})
	go func() {
		r.pool.Wait()
		close(done)
	}()

	select {
	case <-done:
		if r.logger != nil {
			r.logger.Info("all workers completed")
		}
	case <-ctx.Done():
		if r.logger != nil {
			r.logger.Warn("shutdown timeout reached, some tasks may be interrupted")
		}
	}

	// Step 3: Stop result processor
	if r.resultCancel != nil {
		r.resultCancel()
	}
	r.resultWg.Wait()

	// Step 4: Log statistics
	stats := r.pool.Stats()
	if r.logger != nil {
		r.logger.Info("runtime stopped",
			"total_submitted", stats.TotalSubmitted,
			"total_completed", stats.TotalCompleted,
		)
	}

	return nil
}

// Stats returns current runtime statistics.
func (r *Runtime) Stats() PoolStats {
	return r.pool.Stats()
}
```

## Rules
1. **Draining Queue Results on Shutdown**: When context is canceled in `processResults`, implement a non-blocking drain loop using a `select` block with a `default` path. This prevents finished task results from being discarded when shutting down the processor.
2. **Idempotence of Stop Calls**: Make calls to `Stop` idempotent. Calling shutdown multiple times should execute as a no-op instead of returning errors.
3. **Graceful Timeout Coordination**: Coordinate the worker pool teardown (`pool.Wait()`) using context deadline select statements. This prevents broken agent loops from blocking system shutdowns indefinitely.

## ⚠️ Pitfalls

### Pitfall 1: Discarding queued results during processor shutdowns
If you exit the result processor loop immediately upon receiving context cancellation signals, completed tasks that are already in the dispatcher queue will be lost. Ensure you drain the channel using a non-blocking loop before returning.

### Pitfall 2: Re-registering active wait groups during processor shutdowns
If you forget to call `resultWg.Wait()` inside the `Stop()` sequence, the processor goroutine might access deactivated logger fields or run after shutdown returns. Ensure the worker WaitGroup completes before Stop finishes.

## Verify
```bash
go build ./kernel/runtime/...
```

## Checklist
- [ ] File `kernel/runtime/runtime.go` exists
- [ ] Package: `runtime`
- [ ] `Runtime` encapsulates executor, pool, and dispatcher components
- [ ] `Config` declares workers limits, default timeouts, and consumer callbacks
- [ ] `Start` spawns the background result processor loop
- [ ] `processResults` drains channels non-blockingly during shutdown
- [ ] `Dispatch` rejects tasks with error returns when running is false
- [ ] `Stop` implements a 4-step graceful shutdown workflow
- [ ] Shutdown limits wait times using context deadlines
- [ ] `go build ./kernel/runtime/...` passes
