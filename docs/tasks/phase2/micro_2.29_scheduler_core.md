# Micro-Task 2.29: Create kernel/scheduler/scheduler.go

## Info
- **File**: `kernel/scheduler/scheduler.go`
- **Package**: `scheduler`
- **Depends on**: 2.27 (queue.go), 2.28 (deps.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/scheduler/...`

## Purpose
Core scheduler loop. Dequeues ready tasks and dispatches them.
Bridges the priority queue, dependency tracker, and runtime dispatcher.

## EXACT code to create

```go
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// DispatchFunc is a callback the scheduler uses to send ready tasks for execution.
//
// Why a function instead of importing runtime.Dispatcher?
//   → Avoids circular dependency: scheduler → runtime → scheduler
//   → Allows testing scheduler with a mock dispatch function
//   → The kernel wires the real dispatcher function at bootstrap time
type DispatchFunc func(ctx context.Context, task *agent.Task) error

// Scheduler manages task scheduling with priority and dependencies.
//
// Architecture:
//   Submit(task) → add to priority queue → scheduler loop dequeues →
//   check dependencies → if ready → dispatch(task) → if not ready → re-enqueue
//
// The scheduler runs a loop in a separate goroutine (started by Run()).
// Stop the loop by cancelling the context.
type Scheduler struct {
	queue    *PriorityQueue
	deps     *DependencyTracker
	dispatch DispatchFunc
	logger   *slog.Logger

	// pending tracks tasks that are waiting for dependencies.
	// Key: task ID. Value: task + priority.
	// When a dependency completes, we check if any pending tasks become ready.
	mu      sync.Mutex
	pending map[contracts.TaskID]*pendingTask

	// readySignal is used to wake the scheduler loop when new ready tasks appear.
	readySignal chan struct{}
}

// pendingTask tracks a task waiting for dependencies.
type pendingTask struct {
	task     *agent.Task
	priority int
}

// New creates a new Scheduler.
func New(dispatch DispatchFunc, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		queue:       NewPriorityQueue(),
		deps:        NewDependencyTracker(),
		dispatch:    dispatch,
		logger:      logger,
		pending:     make(map[contracts.TaskID]*pendingTask),
		readySignal: make(chan struct{}, 1),
	}
}

// Submit adds a task to the scheduler.
//
// If the task has no dependencies → enqueued immediately.
// If the task has dependencies → held in pending until all deps complete.
//
// Parameters:
//   - task: the task to schedule
//   - priority: scheduling priority (higher = dequeued first)
//
// Returns error if:
//   - Task has circular dependencies
//   - Task ID is already submitted
func (s *Scheduler) Submit(task *agent.Task, priority int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate
	if _, exists := s.pending[task.ID]; exists {
		return fmt.Errorf("scheduler: task %q already submitted", task.ID)
	}

	// Register dependencies
	if len(task.Dependencies) > 0 {
		if err := s.deps.AddDependencies(task.ID, task.Dependencies); err != nil {
			return fmt.Errorf("scheduler: %w", err)
		}
	}

	// Check if ready immediately
	if s.deps.IsReady(task.ID) {
		s.queue.Enqueue(task, priority)
		s.logger.Debug("task enqueued (no dependencies)",
			"task_id", string(task.ID),
			"task_name", task.Name,
			"priority", priority,
		)
	} else {
		// Hold in pending
		s.pending[task.ID] = &pendingTask{task: task, priority: priority}
		pending := s.deps.PendingDependencies(task.ID)
		s.logger.Debug("task pending (waiting for dependencies)",
			"task_id", string(task.ID),
			"task_name", task.Name,
			"pending_deps", fmt.Sprintf("%v", pending),
		)
	}

	return nil
}

// NotifyCompleted is called when a task completes execution.
//
// This method:
//   1. Marks the task as completed in the dependency tracker
//   2. Checks if any pending tasks are now ready
//   3. Enqueues newly-ready tasks
//
// Called by the runtime result processor.
func (s *Scheduler) NotifyCompleted(taskID contracts.TaskID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deps.MarkCompleted(taskID)

	// Remove from pending if it was there
	delete(s.pending, taskID)

	// Check if any pending tasks are now ready
	nowReady := make([]*pendingTask, 0)
	for id, pt := range s.pending {
		if s.deps.IsReady(id) {
			nowReady = append(nowReady, pt)
			delete(s.pending, id)
		}
	}

	// Enqueue newly-ready tasks
	for _, pt := range nowReady {
		s.queue.Enqueue(pt.task, pt.priority)
		s.logger.Info("task became ready (dependencies completed)",
			"task_id", string(pt.task.ID),
			"task_name", pt.task.Name,
		)
	}

	// Signal the scheduler loop
	if len(nowReady) > 0 {
		select {
		case s.readySignal <- struct{}{}:
		default:
		}
	}
}

// Run starts the scheduler loop.
//
// The loop:
//   1. Dequeue highest-priority ready task
//   2. Dispatch it for execution
//   3. Repeat until ctx is cancelled
//
// This method BLOCKS until ctx is cancelled. Run it in a goroutine:
//
//	go scheduler.Run(ctx)
func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("scheduler started")
	defer s.logger.Info("scheduler stopped")

	for {
		// Wait for a ready task (blocks until available or ctx cancelled)
		task, ok := s.queue.DequeueWait(ctx)
		if !ok {
			return // Context cancelled
		}

		// Dispatch the task
		s.logger.Debug("dispatching task",
			"task_id", string(task.ID),
			"task_name", task.Name,
		)

		if err := s.dispatch(ctx, task); err != nil {
			s.logger.Error("task dispatch failed",
				"task_id", string(task.ID),
				"error", err,
			)
			// Re-enqueue with lower priority (retry)
			// Note: In production, add retry count and max retries
			s.queue.Enqueue(task, 0) // Priority 0 = low priority retry
		}
	}
}

// QueueLen returns the number of tasks in the ready queue.
func (s *Scheduler) QueueLen() int {
	return s.queue.Len()
}

// PendingLen returns the number of tasks waiting for dependencies.
func (s *Scheduler) PendingLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending)
}

// Reset clears all state (queue, pending, dependencies).
// Used between missions.
func (s *Scheduler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deps.Reset()
	s.pending = make(map[contracts.TaskID]*pendingTask)
	// Note: PriorityQueue doesn't have a Reset — create a new one
	s.queue = NewPriorityQueue()
}
```

## Pitfalls

### Pitfall 1: DispatchFunc avoids circular imports
```
scheduler imports runtime → runtime imports scheduler → CIRCULAR
```
Solution: scheduler defines `DispatchFunc` type. Kernel injects the real function at bootstrap.

### Pitfall 2: NotifyCompleted must check ALL pending tasks
```go
for id, pt := range s.pending {
    if s.deps.IsReady(id) {
        nowReady = append(nowReady, pt)
    }
}
```
When task A completes, it might unblock tasks B, C, and D. Must check ALL pending tasks.

### Pitfall 3: Don't delete from map while iterating
```go
// First: collect ready tasks
for id, pt := range s.pending { ... }
// Then: delete them
for _, pt := range nowReady { delete(s.pending, pt.task.ID) }
```
Actually in Go, deleting from map during `range` IS safe. But collecting first is clearer.

### Pitfall 4: Failed dispatch → re-enqueue
```go
if err := s.dispatch(ctx, task); err != nil {
    s.queue.Enqueue(task, 0) // Retry with low priority
}
```
Don't drop the task silently. Re-enqueue it so it can be retried.

### Pitfall 5: DequeueWait replaces busy-loop
```go
// OLD (bad): scheduler loop polls with sleep
for {
    task := s.queue.Dequeue()
    if task == nil {
        time.Sleep(100 * time.Millisecond) // Wastes CPU
        continue
    }
}

// NEW (good): blocks until task available
task, ok := s.queue.DequeueWait(ctx)
if !ok { return } // ctx cancelled
```

## Checklist
- [ ] File `kernel/scheduler/scheduler.go` exists
- [ ] DispatchFunc type (avoids circular imports)
- [ ] Scheduler struct: queue, deps, dispatch, pending map
- [ ] `New(dispatch, logger)` constructor
- [ ] `Submit(task, priority)` — duplicate check, dep registration, ready check
- [ ] `NotifyCompleted(taskID)` — marks done, checks pending, enqueues ready
- [ ] `Run(ctx)` — blocking loop with DequeueWait
- [ ] `QueueLen()`, `PendingLen()`, `Reset()`
- [ ] Failed dispatch → re-enqueue
- [ ] DispatchFunc pattern avoids circular imports
- [ ] `go build ./kernel/scheduler/...` no errors
