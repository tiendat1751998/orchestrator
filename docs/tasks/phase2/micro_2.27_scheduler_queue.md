# Micro-Task 2.27: Create kernel/scheduler/queue.go

## Info
- **File**: `kernel/scheduler/queue.go`
- **Package**: `scheduler`
- **Depends on**: Phase 1 contracts
- **Time**: 20 min
- **Verify**: `go build ./kernel/scheduler/...`

## Purpose
Thread-safe priority queue backed by `container/heap`.
Higher priority tasks are dequeued first. Equal priority → FIFO order.

## EXACT code to create

```go
// Package scheduler provides task scheduling with priority queuing and
// dependency resolution.
package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// queueItem wraps a task with scheduling metadata.
type queueItem struct {
	task     *agent.Task
	priority int       // Higher = dequeued first
	enqueued time.Time // For FIFO ordering when priorities are equal
	index    int       // Managed by container/heap (DO NOT set manually)
}

// priorityQueue implements container/heap.Interface.
//
// This is the INTERNAL, UNSYNCHRONIZED queue.
// All external access goes through PriorityQueue (which adds mutex).
//
// Heap invariant: parent priority >= child priority (max-heap).
type priorityQueue []*queueItem

// Len returns the number of items in the queue.
func (pq priorityQueue) Len() int { return len(pq) }

// Less reports whether element i should be dequeued before element j.
//
// Ordering:
//   1. Higher priority first (priority 10 before priority 5)
//   2. If equal priority: earlier enqueue time first (FIFO)
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].priority != pq[j].priority {
		return pq[i].priority > pq[j].priority // Higher priority first
	}
	return pq[i].enqueued.Before(pq[j].enqueued) // FIFO for equal priority
}

// Swap swaps two elements and updates their indices.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an element to the heap.
// Called by heap.Push — DO NOT call directly.
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*queueItem)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the highest-priority element.
// Called by heap.Pop — DO NOT call directly.
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // Allow GC
	item.index = -1 // Mark as removed
	*pq = old[:n-1]
	return item
}

// =============================================================================
// Thread-safe wrapper
// =============================================================================

// PriorityQueue is a thread-safe priority queue for tasks.
//
// Thread-safety: all methods use sync.Mutex.
// container/heap is NOT thread-safe by itself.
//
// Signal mechanism: uses sync.Cond to notify waiting consumers
// when new items are enqueued (avoids busy-waiting / polling).
type PriorityQueue struct {
	mu   sync.Mutex
	cond *sync.Cond
	pq   priorityQueue
}

// NewPriorityQueue creates an empty priority queue.
func NewPriorityQueue() *PriorityQueue {
	q := &PriorityQueue{
		pq: make(priorityQueue, 0),
	}
	q.cond = sync.NewCond(&q.mu)
	heap.Init(&q.pq)
	return q
}

// Enqueue adds a task with the given priority.
//
// Thread-safe. Signals waiting consumers via Cond.
//
// Parameters:
//   - task: the task to enqueue
//   - priority: higher = dequeued first (range: 0-100 recommended)
func (q *PriorityQueue) Enqueue(task *agent.Task, priority int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := &queueItem{
		task:     task,
		priority: priority,
		enqueued: time.Now(),
	}
	heap.Push(&q.pq, item)

	// Signal ONE waiting consumer (Dequeue or DequeueWait)
	q.cond.Signal()
}

// Dequeue removes and returns the highest-priority task.
//
// Returns nil if the queue is empty (non-blocking).
func (q *PriorityQueue) Dequeue() *agent.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.pq.Len() == 0 {
		return nil
	}

	item := heap.Pop(&q.pq).(*queueItem)
	return item.task
}

// DequeueWait removes and returns the highest-priority task.
// If the queue is empty, blocks until a task is available or ctx is cancelled.
//
// This avoids busy-waiting (polling with time.Sleep).
// Uses sync.Cond.Wait() which releases the mutex while waiting.
//
// Returns (nil, false) if ctx is cancelled while waiting.
func (q *PriorityQueue) DequeueWait(ctx context.Context) (*agent.Task, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Start a goroutine to broadcast on context cancellation
	// This wakes up the Cond.Wait() below so we can check ctx.Done()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			q.cond.Broadcast() // Wake all waiters
		case <-done:
			// Normal exit
		}
	}()
	defer close(done)

	for q.pq.Len() == 0 {
		// Check context before waiting
		select {
		case <-ctx.Done():
			return nil, false
		default:
		}

		q.cond.Wait() // Releases mutex, waits for signal, re-acquires mutex

		// Check context after waking
		select {
		case <-ctx.Done():
			return nil, false
		default:
		}
	}

	item := heap.Pop(&q.pq).(*queueItem)
	return item.task, true
}

// Len returns the number of queued tasks.
func (q *PriorityQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len()
}

// Peek returns the highest-priority task without removing it.
// Returns nil if empty.
func (q *PriorityQueue) Peek() *agent.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.pq.Len() == 0 {
		return nil
	}
	return q.pq[0].task
}

// Remove removes a specific task by ID.
// Returns true if found and removed, false if not found.
//
// Used when cancelling a queued (not yet running) task.
func (q *PriorityQueue) Remove(taskID contracts.TaskID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.pq {
		if item.task.ID == taskID {
			heap.Remove(&q.pq, i)
			return true
		}
	}
	return false
}
```

## Pitfalls

### Pitfall 1: container/heap is NOT thread-safe
```go
// WRONG: calling heap.Push without mutex
heap.Push(&q.pq, item) // Data race!

// CORRECT: wrap with mutex
q.mu.Lock()
heap.Push(&q.pq, item)
q.mu.Unlock()
```

### Pitfall 2: sync.Cond for signaling (not busy-wait)
```go
// WRONG: polling (wastes CPU)
for q.pq.Len() == 0 {
    time.Sleep(100 * time.Millisecond)
}

// CORRECT: condition variable (zero CPU when idle)
for q.pq.Len() == 0 {
    q.cond.Wait() // Releases mutex, sleeps until Signal()
}
```

### Pitfall 3: Context cancellation with sync.Cond
`sync.Cond.Wait()` does NOT support context. Workaround:
- Launch a goroutine that watches ctx.Done()
- On cancel, call cond.Broadcast() to wake all waiters
- After waking, check ctx.Done() to determine if it's a real signal or cancel

### Pitfall 4: Nil GC in Pop
```go
old[n-1] = nil  // Allow GC to collect the removed item
```
Without this → slice still holds a reference → item is never garbage collected.

### Pitfall 5: FIFO for equal priorities
```go
if pq[i].priority != pq[j].priority {
    return pq[i].priority > pq[j].priority
}
return pq[i].enqueued.Before(pq[j].enqueued) // FIFO tiebreaker
```
Without FIFO tiebreaker → equal-priority tasks are dequeued in random order.

## Checklist
- [ ] File `kernel/scheduler/queue.go` exists
- [ ] `priorityQueue` implements `heap.Interface` (Len, Less, Swap, Push, Pop)
- [ ] Higher priority dequeued first
- [ ] FIFO ordering for equal priorities
- [ ] `PriorityQueue` thread-safe wrapper with mutex
- [ ] `Enqueue(task, priority)` with cond.Signal()
- [ ] `Dequeue()` non-blocking (returns nil if empty)
- [ ] `DequeueWait(ctx)` blocking with context cancellation
- [ ] `Len()`, `Peek()`, `Remove(taskID)`
- [ ] Nil GC in Pop
- [ ] sync.Cond for signaling (not busy-wait)
- [ ] `go build ./kernel/scheduler/...` no errors
