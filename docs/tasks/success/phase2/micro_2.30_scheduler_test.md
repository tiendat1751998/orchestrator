# Micro-Task 2.30: Create kernel/scheduler/scheduler_test.go

## Info
- **File**: `kernel/scheduler/scheduler_test.go`
- **Package**: `scheduler_test`
- **Depends on**: 2.27-2.29
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/scheduler/...`

## Purpose
Implements integration unit tests for the scheduler package. It verifies priority ordering, FIFO tie-breaking, wait-signaling under context cancellations, dependency resolution trees, self-dependency blocks, circular dependency detection, and diamond-shaped dependency configurations.

## EXACT code to create

```go
package scheduler_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/scheduler"
)

// =============================================================================
// Priority Queue Tests
// =============================================================================

func TestPriorityQueue_HigherPriorityFirst(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	q.Enqueue(&agent.Task{ID: "low", Name: "low-priority"}, 1)
	q.Enqueue(&agent.Task{ID: "high", Name: "high-priority"}, 10)
	q.Enqueue(&agent.Task{ID: "mid", Name: "mid-priority"}, 5)

	task1 := q.Dequeue()
	if task1.ID != "high" {
		t.Errorf("first dequeue: got %q, want %q", task1.ID, "high")
	}

	task2 := q.Dequeue()
	if task2.ID != "mid" {
		t.Errorf("second dequeue: got %q, want %q", task2.ID, "mid")
	}

	task3 := q.Dequeue()
	if task3.ID != "low" {
		t.Errorf("third dequeue: got %q, want %q", task3.ID, "low")
	}
}

func TestPriorityQueue_FIFOForEqualPriority(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	q.Enqueue(&agent.Task{ID: "first"}, 5)
	time.Sleep(1 * time.Millisecond)
	q.Enqueue(&agent.Task{ID: "second"}, 5)
	time.Sleep(1 * time.Millisecond)
	q.Enqueue(&agent.Task{ID: "third"}, 5)

	t1 := q.Dequeue()
	t2 := q.Dequeue()
	t3 := q.Dequeue()

	if t1.ID != "first" || t2.ID != "second" || t3.ID != "third" {
		t.Errorf("FIFO order: got %s, %s, %s", t1.ID, t2.ID, t3.ID)
	}
}

func TestPriorityQueue_DequeueEmpty(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	task := q.Dequeue()
	if task != nil {
		t.Errorf("Dequeue empty: got %v, want nil", task)
	}
}

func TestPriorityQueue_DequeueWait_ContextCancel(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	task, ok := q.DequeueWait(ctx)
	if ok {
		t.Error("DequeueWait should return false on context cancellation")
	}
	if task != nil {
		t.Error("task should be nil on cancellation")
	}
}

func TestPriorityQueue_DequeueWait_GetsSignaled(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	var received *agent.Task
	done := make(chan struct{})

	go func() {
		task, ok := q.DequeueWait(context.Background())
		if ok {
			received = task
		}
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	q.Enqueue(&agent.Task{ID: "signaled", Name: "test"}, 1)

	select {
	case <-done:
		if received == nil || received.ID != "signaled" {
			t.Errorf("received: got %v, want task 'signaled'", received)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: DequeueWait was not signaled")
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	if q.Len() != 0 {
		t.Errorf("empty Len: got %d", q.Len())
	}

	q.Enqueue(&agent.Task{ID: "a"}, 1)
	q.Enqueue(&agent.Task{ID: "b"}, 2)

	if q.Len() != 2 {
		t.Errorf("Len after 2 enqueues: got %d", q.Len())
	}

	q.Dequeue()

	if q.Len() != 1 {
		t.Errorf("Len after 1 dequeue: got %d", q.Len())
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	q.Enqueue(&agent.Task{ID: "keep"}, 5)
	q.Enqueue(&agent.Task{ID: "remove-me"}, 10)

	removed := q.Remove("remove-me")
	if !removed {
		t.Error("Remove should return true for existing task")
	}

	if q.Len() != 1 {
		t.Errorf("Len after remove: got %d, want 1", q.Len())
	}

	task := q.Dequeue()
	if task.ID != "keep" {
		t.Errorf("remaining task: got %q, want %q", task.ID, "keep")
	}
}

func TestPriorityQueue_Peek(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	if q.Peek() != nil {
		t.Error("Peek empty should return nil")
	}

	q.Enqueue(&agent.Task{ID: "first"}, 10)
	q.Enqueue(&agent.Task{ID: "second"}, 5)

	peeked := q.Peek()
	if peeked.ID != "first" {
		t.Errorf("Peek: got %q, want %q (highest priority)", peeked.ID, "first")
	}

	if q.Len() != 2 {
		t.Error("Peek should not remove the task")
	}
}

func TestPriorityQueue_ConcurrentAccess(t *testing.T) {
	q := scheduler.NewPriorityQueue()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			q.Enqueue(&agent.Task{ID: contracts.TaskID(fmt.Sprintf("t-%d", id))}, id)
		}(i)
	}

	wg.Wait()

	if q.Len() != 50 {
		t.Errorf("Len after concurrent enqueue: got %d, want 50", q.Len())
	}
}

// =============================================================================
// Dependency Tracker Tests
// =============================================================================

func TestDependencyTracker_NoDependencies(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	if !dt.IsReady("task-1") {
		t.Error("task with no dependencies should be ready")
	}
}

func TestDependencyTracker_WaitForDependency(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	dt.AddDependency("task-B", "task-A")

	if dt.IsReady("task-B") {
		t.Error("task-B should NOT be ready (task-A not completed)")
	}

	dt.MarkCompleted("task-A")

	if !dt.IsReady("task-B") {
		t.Error("task-B should be ready (task-A completed)")
	}
}

func TestDependencyTracker_MultipleDependencies(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	dt.AddDependency("task-C", "task-A")
	dt.AddDependency("task-C", "task-B")

	dt.MarkCompleted("task-A")
	if dt.IsReady("task-C") {
		t.Error("task-C should NOT be ready (task-B not completed)")
	}

	dt.MarkCompleted("task-B")
	if !dt.IsReady("task-C") {
		t.Error("task-C should be ready (both A and B completed)")
	}
}

func TestDependencyTracker_SelfDependency(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	err := dt.AddDependency("task-A", "task-A")
	if err == nil {
		t.Error("expected error for self-dependency")
	}
}

func TestDependencyTracker_CircularDependency(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	dt.AddDependency("task-A", "task-B")
	err := dt.AddDependency("task-B", "task-A")
	if err == nil {
		t.Error("expected error for circular dependency A→B, B→A")
	}
}

func TestDependencyTracker_CircularDependency_ThreeNodes(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	dt.AddDependency("A", "B")
	dt.AddDependency("B", "C")
	err := dt.AddDependency("C", "A")
	if err == nil {
		t.Error("expected error for circular dependency A→B→C→A")
	}
}

func TestDependencyTracker_PendingDependencies(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	dt.AddDependency("task-C", "task-A")
	dt.AddDependency("task-C", "task-B")

	pending := dt.PendingDependencies("task-C")
	if len(pending) != 2 {
		t.Errorf("PendingDependencies: got %d, want 2", len(pending))
	}

	dt.MarkCompleted("task-A")
	pending = dt.PendingDependencies("task-C")
	if len(pending) != 1 {
		t.Errorf("PendingDependencies after A completed: got %d, want 1", len(pending))
	}
}

func TestDependencyTracker_DiamondDependency(t *testing.T) {
	dt := scheduler.NewDependencyTracker()

	// Diamond: D depends on B and C. B and C both depend on A.
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	dt.AddDependency("B", "A")
	dt.AddDependency("C", "A")
	dt.AddDependency("D", "B")
	dt.AddDependency("D", "C")

	if dt.IsReady("D") {
		t.Error("D should not be ready")
	}

	dt.MarkCompleted("A")
	if !dt.IsReady("B") || !dt.IsReady("C") {
		t.Error("B and C should be ready after A")
	}

	if dt.IsReady("D") {
		t.Error("D should not be ready (B and C not completed)")
	}

	dt.MarkCompleted("B")
	dt.MarkCompleted("C")

	if !dt.IsReady("D") {
		t.Error("D should be ready after B and C")
	}
}

// =============================================================================
// Scheduler Integration Tests
// =============================================================================

func TestScheduler_SubmitAndRun(t *testing.T) {
	var dispatched atomic.Int32

	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error {
			dispatched.Add(1)
			return nil
		},
		nil,
	)

	for i := 0; i < 3; i++ {
		task := &agent.Task{
			ID:   contracts.TaskID(fmt.Sprintf("t-%d", i)),
			Name: fmt.Sprintf("task-%d", i),
			Type: "test",
		}
		sched.Submit(task, i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go sched.Run(ctx)
	time.Sleep(500 * time.Millisecond)
	cancel()

	if dispatched.Load() != 3 {
		t.Errorf("dispatched: got %d, want 3", dispatched.Load())
	}
}

func TestScheduler_DependencyWait(t *testing.T) {
	dispatched := make([]contracts.TaskID, 0)
	mu := sync.Mutex{}

	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error {
			mu.Lock()
			dispatched = append(dispatched, task.ID)
			mu.Unlock()
			return nil
		},
		nil,
	)

	taskA := &agent.Task{ID: "A", Name: "task-A", Type: "test"}
	taskB := &agent.Task{
		ID:           "B",
		Name:         "task-B",
		Type:         "test",
		Dependencies: []contracts.TaskID{"A"},
	}

	sched.Submit(taskB, 10)
	sched.Submit(taskA, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go sched.Run(ctx)
	time.Sleep(300 * time.Millisecond)

	sched.NotifyCompleted("A")
	time.Sleep(500 * time.Millisecond)
	cancel()

	mu.Lock()
	defer mu.Unlock()

	if len(dispatched) < 2 {
		t.Fatalf("expected 2 dispatches, got %d", len(dispatched))
	}
	if dispatched[0] != "A" {
		t.Errorf("first dispatch: got %q, want %q", dispatched[0], "A")
	}
	if dispatched[1] != "B" {
		t.Errorf("second dispatch: got %q, want %q", dispatched[1], "B")
	}
}

func TestScheduler_DuplicateSubmit(t *testing.T) {
	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error { return nil },
		nil,
	)

	task := &agent.Task{ID: "dup", Name: "duplicate", Type: "test"}
	sched.Submit(task, 1)

	err := sched.Submit(task, 1)
	if err == nil {
		t.Error("expected error for duplicate submit")
	}
}

func TestScheduler_CircularDependencyRejected(t *testing.T) {
	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error { return nil },
		nil,
	)

	taskA := &agent.Task{
		ID:           "A",
		Name:         "A",
		Type:         "test",
		Dependencies: []contracts.TaskID{"B"},
	}
	taskB := &agent.Task{
		ID:           "B",
		Name:         "B",
		Type:         "test",
		Dependencies: []contracts.TaskID{"A"},
	}

	sched.Submit(taskA, 1)
	err := sched.Submit(taskB, 1)
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestScheduler_QueueAndPendingLen(t *testing.T) {
	sched := scheduler.New(
		func(ctx context.Context, task *agent.Task) error { return nil },
		nil,
	)

	sched.Submit(&agent.Task{ID: "ready", Name: "ready", Type: "test"}, 1)

	sched.Submit(&agent.Task{
		ID:           "pending",
		Name:         "pending",
		Type:         "test",
		Dependencies: []contracts.TaskID{"ready"},
	}, 1)

	if sched.QueueLen() != 1 {
		t.Errorf("QueueLen: got %d, want 1", sched.QueueLen())
	}
	if sched.PendingLen() != 1 {
		t.Errorf("PendingLen: got %d, want 1", sched.PendingLen())
	}
}
```

## Rules
1. **Sync Wait Operations**: Coordinate DequeueWait test routines using buffered channels to synchronize test completion, avoiding fragile timeout sleeps.
2. **Atomic Verification Counters**: Counters updated concurrently within parallel test loops must utilize `atomic.Int32`/`atomic.Int64` types to pass `go test -race` checks.
3. **Graph Integrity Checks**: Check for self-dependencies, circular chains, and diamond structures to verify the scheduler doesn't deadlock.

## ⚠️ Pitfalls

### Pitfall 1: Relying on sleep intervals for worker thread synchronization
Using hardcoded sleeps (e.g. `time.Sleep(500 * time.Millisecond)`) inside tests makes them fragile in slow environments (like CI pipelines). Always use select blocks combined with channel notifications.

### Pitfall 2: Sharing un-synchronized variables across threads
Updating test counters using normal `int` increments inside goroutines triggers race conditions. Use atomic integers for concurrent counting.

## Verify
```bash
go test -v -race ./kernel/scheduler/...
```

## Checklist
- [ ] File `kernel/scheduler/scheduler_test.go` exists
- [ ] Package: `scheduler_test` (external testing package)
- [ ] At least 20 test functions are defined
- [ ] Queue tests verify priorities, FIFO tie-breakers, and wait signals
- [ ] Dependency tests verify self-reference blocks, circular graphs, and diamond flows
- [ ] Integration tests verify submissions, completions, and duplicate guards
- [ ] All concurrent assertions use atomic variables
- [ ] `go test -v -race ./kernel/scheduler/...` passes
