# Micro-Task 2.23: Create kernel/runtime/pool.go

## Info
- **File**: `kernel/runtime/pool.go`
- **Package**: `runtime`
- **Depends on**: none (standalone)
- **Time**: 20 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Implements the worker execution pool (`Pool` and constructors) that enforces concurrency limits using buffered channels as counting semaphores, preventing resource exhaustion during parallel tasks execution.

## EXACT code to create

```go
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

// Pool manages a bounded set of worker goroutines.
//
// Architecture:
//   Pool uses a semaphore (buffered channel) to limit concurrency.
//   Each Submit() acquires a semaphore slot before launching a goroutine.
//   When the goroutine completes, it releases the slot.
//
// Why channel-based semaphore instead of sync.Semaphore?
//   → Channel supports context cancellation via select{}
//   → sync.Semaphore doesn't natively support context
//
// Example:
//
//	pool := NewPool(5, logger)  // max 5 concurrent workers
//	pool.Submit(ctx, func(ctx context.Context) {
//	    // do work
//	})
//	pool.Wait()  // wait for all workers to finish
type Pool struct {
	// sem is a buffered channel acting as a counting semaphore.
	// Buffer size = max concurrency.
	// Each goroutine writes to sem before work and reads after work.
	sem chan struct{}

	// wg tracks in-flight goroutines for Wait().
	wg sync.WaitGroup

	// activeCount tracks how many workers are currently running.
	// Used for monitoring/metrics.
	activeCount atomic.Int32

	// totalSubmitted tracks total tasks submitted (including queued).
	totalSubmitted atomic.Int64

	// totalCompleted tracks total tasks completed.
	totalCompleted atomic.Int64

	logger     *slog.Logger
	maxWorkers int
}

// NewPool creates a new worker pool.
//
// Parameters:
//   - maxWorkers: maximum concurrent goroutines (must be >= 1)
//   - logger: for diagnostics (can be nil)
//
// If maxWorkers < 1, defaults to 1 (sequential execution).
func NewPool(maxWorkers int, logger *slog.Logger) *Pool {
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	return &Pool{
		sem:        make(chan struct{}, maxWorkers),
		logger:     logger,
		maxWorkers: maxWorkers,
	}
}

// Submit schedules a function to run in the pool.
//
// Behavior:
//   1. If a worker slot is available → launches immediately in a goroutine
//   2. If all slots are busy → blocks until a slot opens or ctx is cancelled
//   3. If ctx is cancelled while waiting → returns error without executing
//
// The function receives a context that is cancelled when the pool is shut down.
//
// IMPORTANT: Submit does NOT wait for the function to complete.
// Call Wait() to block until all submitted functions finish.
//
// Thread-safety: safe for concurrent use.
func (p *Pool) Submit(ctx context.Context, fn func(ctx context.Context)) error {
	p.totalSubmitted.Add(1)

	// Acquire semaphore slot (blocks if pool is full)
	select {
	case p.sem <- struct{}{}:
		// Slot acquired — proceed
	case <-ctx.Done():
		// Context cancelled while waiting for a slot
		return fmt.Errorf("pool: submit cancelled: %w", ctx.Err())
	}

	// Launch worker goroutine
	p.wg.Add(1)
	p.activeCount.Add(1)

	go func() {
		defer func() {
			<-p.sem // Release semaphore slot
			p.activeCount.Add(-1)
			p.totalCompleted.Add(1)
			p.wg.Done()
		}()

		fn(ctx)
	}()

	return nil
}

// Wait blocks until all submitted functions complete.
//
// Call this during shutdown to ensure all work is done.
// After Wait returns, no more goroutines are running.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// ActiveWorkers returns the number of currently running workers.
func (p *Pool) ActiveWorkers() int {
	return int(p.activeCount.Load())
}

// MaxWorkers returns the maximum concurrency limit.
func (p *Pool) MaxWorkers() int {
	return p.maxWorkers
}

// Stats returns pool statistics.
type PoolStats struct {
	MaxWorkers     int   `json:"max_workers"`
	ActiveWorkers  int   `json:"active_workers"`
	TotalSubmitted int64 `json:"total_submitted"`
	TotalCompleted int64 `json:"total_completed"`
	QueuedTasks    int64 `json:"queued_tasks"` // Submitted - Completed - Active
}

// Stats returns current pool statistics.
func (p *Pool) Stats() PoolStats {
	active := int(p.activeCount.Load())
	submitted := p.totalSubmitted.Load()
	completed := p.totalCompleted.Load()

	return PoolStats{
		MaxWorkers:     p.maxWorkers,
		ActiveWorkers:  active,
		TotalSubmitted: submitted,
		TotalCompleted: completed,
		QueuedTasks:    submitted - completed - int64(active),
	}
}
```

## Rules
1. **Context-Aware Semaphore blocks**: Implement semaphores using buffered channels inside `select` blocks to allow context cancellation requests (`<-ctx.Done()`) to break deadlocks when all workers are busy.
2. **Deterministic WaitGroup Adds**: Ensure `wg.Add(1)` is executed synchronously on the scheduler thread before launching goroutines, preventing `Wait()` race bugs.
3. **Guard Concurrency limits**: Enforce boundaries checking when initializing max worker sizes. If `maxWorkers` is set to less than 1, override the value to 1.
4. **Panic-Safe semaphore releases**: Place channel reads (`<-p.sem`) and WG done signals (`wg.Done()`) inside deferred cleanup callbacks to avoid leaking worker slots if client functions panic.

## ⚠️ Pitfalls

### Pitfall 1: Deadlocking unbuffered channels when maxWorkers is zero
Initializing channels with size 0 (`make(chan struct{}, 0)`) results in unbuffered channels, causing all worker submissions to block indefinitely since there are no active reader threads pulling from the queue. Enforce `maxWorkers >= 1` limits.

### Pitfall 2: Launching workers before updating WaitGroup counts
If `wg.Add(1)` is placed inside the spawned goroutine rather than synchronously on the scheduler thread, calling `Wait()` immediately after submissions can evaluate to true and return prematurely. Always call `wg.Add(1)` before calling `go`.

## Verify
```bash
go build ./kernel/runtime/...
```

## Checklist
- [ ] File `kernel/runtime/pool.go` exists
- [ ] Package: `runtime`
- [ ] Concurrency channel sem initialized with buffer size matching max worker count
- [ ] Max workers parameter is guarded to always be >= 1
- [ ] `Submit` handles queue blocks and checks for context cancellations
- [ ] Spawner threads call `wg.Add(1)` before calling `go`
- [ ] Deferred closures release semaphore slots and decrement counters
- [ ] Counters use thread-safe `atomic` operations
- [ ] `Wait` blocks execution until all workers finish
- [ ] `go build ./kernel/runtime/...` passes
