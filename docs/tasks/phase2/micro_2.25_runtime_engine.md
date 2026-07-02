# Micro-Task 2.25: Create kernel/runtime/runtime.go

## Info
- **File**: `kernel/runtime/runtime.go`
- **Package**: `runtime`
- **Depends on**: 2.22-2.24
- **Time**: 20 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Runtime engine. Ties Executor + Pool + Dispatcher together.
Provides Start/Stop lifecycle for the kernel.

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
	// The callback runs in a dedicated goroutine — it can block without
	// affecting task execution.
	OnResult func(TaskResult)
}

// New creates a new Runtime.
//
// Parameters:
//   - reg: plugin registry (for finding agents)
//   - bus: event bus (for emitting events)
//   - logger: structured logger
//   - cfg: runtime configuration
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

	r.logger.Info("runtime started",
		"max_workers", r.pool.MaxWorkers(),
	)

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

	r.logger.Info("runtime stopping...")

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
		r.logger.Info("all workers completed")
	case <-ctx.Done():
		r.logger.Warn("shutdown timeout reached, some tasks may be interrupted")
	}

	// Step 3: Stop result processor
	if r.resultCancel != nil {
		r.resultCancel()
	}
	r.resultWg.Wait()

	// Step 4: Log statistics
	stats := r.pool.Stats()
	r.logger.Info("runtime stopped",
		"total_submitted", stats.TotalSubmitted,
		"total_completed", stats.TotalCompleted,
	)

	return nil
}

// Stats returns current runtime statistics.
func (r *Runtime) Stats() PoolStats {
	return r.pool.Stats()
}
```

## Pitfalls

### Pitfall 1: Drain results before exit
```go
// After ctx.Done(), drain remaining results
for {
    select {
    case result := <-r.dispatcher.Results():
        onResult(result)
    default:
        return // Channel empty, exit
    }
}
```
Without draining → results from final tasks are lost → caller never knows they completed.

### Pitfall 2: Stop is idempotent
```go
if !r.running {
    r.mu.Unlock()
    return nil // Already stopped
}
```
Multiple calls to Stop() should not panic or error.

### Pitfall 3: Shutdown timeout for pool.Wait()
```go
select {
case <-done:       // Workers finished
case <-ctx.Done(): // Timeout reached
}
```
Without timeout → a stuck agent blocks shutdown forever.

### Pitfall 4: Dispatch checks running state
```go
if !r.running {
    return fmt.Errorf("runtime: not running")
}
```
Dispatching to a stopped runtime → error. Clear signal to caller.

### Pitfall 5: Result processor goroutine lifecycle
```go
r.resultWg.Add(1)
go func() {
    defer r.resultWg.Done()
    r.processResults(ctx, onResult)
}()
```
`resultWg.Wait()` in Stop() ensures the result processor is fully stopped before returning.

## Checklist
- [ ] File `kernel/runtime/runtime.go` exists
- [ ] Runtime struct: executor, pool, dispatcher, result processor
- [ ] Config struct: MaxWorkers, DefaultTimeout, ResultBufferSize, OnResult
- [ ] `New()` constructor — wires executor + pool + dispatcher
- [ ] `Start(onResult)` — launches result processor goroutine
- [ ] `Dispatch(ctx, task)` — sends task to dispatcher
- [ ] `Stop(ctx)` — 4-step shutdown: stop dispatcher, wait pool, stop processor, log stats
- [ ] `Stats()` — returns PoolStats
- [ ] Result drain on shutdown
- [ ] Idempotent Stop
- [ ] Running state check in Dispatch
- [ ] Shutdown timeout via context
- [ ] `go build ./kernel/runtime/...` no errors
