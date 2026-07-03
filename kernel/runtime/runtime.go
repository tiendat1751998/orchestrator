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
//
//	New() → Start() → Dispatch tasks → Stop()
//
// Components:
//
//	Executor   — runs a single task (timeout, panic recovery)
//	Pool       — limits concurrent workers
//	Dispatcher — routes tasks, collects results
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
//  1. Queued in the worker pool (blocks if pool is full)
//  2. Executed by the Executor (find agent, timeout, panic recovery)
//  3. Result sent to the OnResult callback
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

// Stop gracefully shuts down the runtime and verifies resource release.
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

	// Step 4: Verification of resources (Leak Detection)
	stats := r.pool.Stats()
	if stats.ActiveWorkers > 0 {
		if r.logger != nil {
			r.logger.Error("resource leak: worker goroutines are still active after shutdown",
				"active_workers", stats.ActiveWorkers,
				"submitted", stats.TotalSubmitted,
				"completed", stats.TotalCompleted,
			)
		}
	}

	// Verify the result channel is drained
	undrainedCount := 0
DrainLoop:
	for {
		select {
		case <-r.dispatcher.Results():
			undrainedCount++
		default:
			break DrainLoop
		}
	}
	if undrainedCount > 0 {
		if r.logger != nil {
			r.logger.Warn("resource warning: undrained task results remaining in channel",
				"count", undrainedCount,
			)
		}
	}

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
