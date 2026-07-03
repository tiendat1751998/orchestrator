# Micro-Task 2.29: Create kernel/scheduler/scheduler.go

## Info
- **File**: `kernel/scheduler/scheduler.go`
- **Package**: `scheduler`
- **Depends on**: 2.27 (queue.go), 2.28 (deps.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/scheduler/...`

## Purpose
Implements the core scheduler loop (`Scheduler`, `DispatchFunc`, `pendingTask` and constructors) that orchestrates task submissions, resolves task dependency requirements, and dispatches ready tasks.

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
		if s.logger != nil {
			s.logger.Debug("task enqueued (no dependencies)",
				"task_id", string(task.ID),
				"task_name", task.Name,
				"priority", priority,
			)
		}
	} else {
		// Hold in pending
		s.pending[task.ID] = &pendingTask{task: task, priority: priority}
		pending := s.deps.PendingDependencies(task.ID)
		if s.logger != nil {
			s.logger.Debug("task pending (waiting for dependencies)",
				"task_id", string(task.ID),
				"task_name", task.Name,
				"pending_deps", fmt.Sprintf("%v", pending),
			)
		}
	}

	return nil
}

// NotifyCompleted is called when a task completes execution.
//
// This method:
//   1. Marks the task as completed in the dependency tracker
//   2. Checks if any pending tasks are now ready
//   3. Enqueues newly-ready tasks
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
		}
	}

	// Enqueue newly-ready tasks
	for _, pt := range nowReady {
		delete(s.pending, pt.task.ID)
		s.queue.Enqueue(pt.task, pt.priority)
		if s.logger != nil {
			s.logger.Info("task became ready (dependencies completed)",
				"task_id", string(pt.task.ID),
				"task_name", pt.task.Name,
			)
		}
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
// This method BLOCKS until ctx is cancelled. Run it in a goroutine.
func (s *Scheduler) Run(ctx context.Context) {
	if s.logger != nil {
		s.logger.Info("scheduler started")
		defer s.logger.Info("scheduler stopped")
	}

	for {
		// Wait for a ready task (blocks until available or ctx cancelled)
		task, ok := s.queue.DequeueWait(ctx)
		if !ok {
			return // Context cancelled
		}

		if s.logger != nil {
			s.logger.Debug("dispatching task",
				"task_id", string(task.ID),
				"task_name", task.Name,
			)
		}

		if err := s.dispatch(ctx, task); err != nil {
			if s.logger != nil {
				s.logger.Error("task dispatch failed",
					"task_id", string(task.ID),
					"error", err,
				)
			}
			// Re-enqueue with lower priority (retry)
			s.queue.Enqueue(task, 0)
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
	s.queue = NewPriorityQueue()
}
```

## Rules
1. **Loose Coupling (DispatchFunc)**: Use function signatures (`DispatchFunc`) rather than directly importing runtime dispatcher structs. This avoids circular packages importing loops and makes testing simple.
2. **Batch Task Readiness Scans**: During completion callbacks (`NotifyCompleted`), scan all pending tasks in the tracker rather than looking up single child nodes. Completing a task can unlock multiple nodes in the dependency graph.
3. **Queue Re-enqueueing**: If the dispatch callback fails, re-enqueue the task with low priority to prevent silent drops.

## ⚠️ Pitfalls

### Pitfall 1: Creating circular package imports during execution loops
If `scheduler` directly imports `runtime` to access dispatchers, and `runtime` imports `scheduler` to register statuses, the Go compiler rejects the packages. Solve this by defining `DispatchFunc` and passing closures at bootstrap.

### Pitfall 2: Re-enqueuing failed dispatches without logging warnings
Silently discarding tasks when dispatching fails makes troubleshooting difficult. Always log errors and re-enqueue tasks with low priority.

## Verify
```bash
go build ./kernel/scheduler/...
```

## Checklist
- [ ] File `kernel/scheduler/scheduler.go` exists
- [ ] Package: `scheduler`
- [ ] `DispatchFunc` signature declared to avoid circular imports
- [ ] `Scheduler` holds priority queue, tracker, pending maps, and signaling channels
- [ ] `Submit` catches duplicate submissions and enqueues immediately if ready
- [ ] `NotifyCompleted` triggers dependency updates and enqueues unblocked tasks
- [ ] `Run` runs in blocks, calling `DequeueWait`
- [ ] `Reset` clears queue and pending maps between runs
- [ ] `go build ./kernel/scheduler/...` passes
