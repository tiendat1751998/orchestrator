# Dead Letter Queue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a thread-safe circular Dead Letter Queue (DLQ) ring buffer that captures failed event invocations, records panic stack diagnostics along with the original event context, and updates the core Bus implementation to leverage this safety mechanism.

**Architecture:** DeadLetterQueue is implemented as a thread-safe circular buffer (ring buffer) of fixed capacity. Bus struct holds a pointer to DeadLetterQueue, passes it to safeHandler, and provides a DLQ() getter. safeHandler recovers from handler panics and adds entries to DLQ.

**Tech Stack:** Go 1.26, contracts/event.

## Global Constraints

- docs/adp.md (Architectural Decision Principles)
- .agents/rules/ai_rules.md (Go code conventions and complexity budgets)
- .agents/rules/ponytail.md (Simplest minimal solution)
- .agents/rules/superpowers.md (Process discipline)
- Verify changes by running `go test -v ./kernel/eventbus/...`

---

### Task 1: Create `kernel/eventbus/dlq.go`

**Files:**
- Create: `kernel/eventbus/dlq.go`
- Test: `kernel/eventbus/dlq_test.go`

**Interfaces:**
- Consumes: `github.com/tiendat1751998/orchestrator/contracts/event.Event`
- Produces: `DeadLetterEntry` structure, `DeadLetterQueue` structure, `NewDeadLetterQueue(maxSize int) *DeadLetterQueue`, `Add(evt event.Event, errStr string)`, `Entries() []DeadLetterEntry`, `Len() int`, `Clear()`

- [ ] **Step 1: Write the failing test**
  Create `kernel/eventbus/dlq_test.go` containing:
  ```go
  package eventbus_test

  import (
  	"testing"
  	"time"

  	"github.com/tiendat1751998/orchestrator/contracts/event"
  	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
  )

  func TestDLQ_BasicAddAndRetrieve(t *testing.T) {
  	q := eventbus.NewDeadLetterQueue(3)
  	if q.Len() != 0 {
  		t.Errorf("expected length 0, got %d", q.Len())
  	}

  	evt := event.Event{ID: "e1", Type: "test.event"}
  	q.Add(evt, "panic error")

  	if q.Len() != 1 {
  		t.Errorf("expected length 1, got %d", q.Len())
  	}

  	entries := q.Entries()
  	if len(entries) != 1 {
  		t.Fatalf("expected 1 entry, got %d", len(entries))
  	}
  	if entries[0].Event.ID != "e1" {
  		t.Errorf("expected event ID e1, got %s", entries[0].Event.ID)
  	}
  	if entries[0].Error != "panic error" {
  		t.Errorf("expected error 'panic error', got %s", entries[0].Error)
  	}
  }
  ```

- [ ] **Step 2: Run test to verify it fails**
  Run: `go test -v -run TestDLQ_BasicAddAndRetrieve ./kernel/eventbus/...`
  Expected: FAIL (compilation error: NewDeadLetterQueue not defined)

- [ ] **Step 3: Write minimal implementation**
  Create `kernel/eventbus/dlq.go` containing:
  ```go
  package eventbus

  import (
  	"sync"
  	"time"

  	"github.com/tiendat1751998/orchestrator/contracts/event"
  )

  type DeadLetterEntry struct {
  	Event     event.Event `json:"event"`
  	Error     string      `json:"error"`
  	Timestamp time.Time   `json:"timestamp"`
  }

  type DeadLetterQueue struct {
  	mu       sync.RWMutex
  	entries  []DeadLetterEntry
  	maxSize  int
  	writeIdx int
  	count    int
  }

  func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
  	if maxSize <= 0 {
  		maxSize = 100
  	}
  	return &DeadLetterQueue{
  		entries: make([]DeadLetterEntry, maxSize),
  		maxSize: maxSize,
  	}
  }

  func (q *DeadLetterQueue) Add(evt event.Event, errStr string) {
  	q.mu.Lock()
  	defer q.mu.Unlock()

  	// ponytail: simple thread-safe queue addition
  	q.entries[q.writeIdx] = DeadLetterEntry{
  		Event:     evt,
  		Error:     errStr,
  		Timestamp: time.Now(),
  	}

  	q.writeIdx = (q.writeIdx + 1) % q.maxSize
  	if q.count < q.maxSize {
  		q.count++
  	}
  }

  func (q *DeadLetterQueue) Entries() []DeadLetterEntry {
  	q.mu.RLock()
  	defer q.mu.RUnlock()

  	result := make([]DeadLetterEntry, q.count)
  	if q.count == 0 {
  		return result
  	}

  	startIdx := 0
  	if q.count == q.maxSize {
  		startIdx = q.writeIdx
  	}

  	for i := 0; i < q.count; i++ {
  		readIdx := (startIdx + i) % q.maxSize
  		result[i] = q.entries[readIdx]
  	}

  	return result
  }

  func (q *DeadLetterQueue) Len() int {
  	q.mu.RLock()
  	defer q.mu.RUnlock()
  	return q.count
  }

  func (q *DeadLetterQueue) Clear() {
  	q.mu.Lock()
  	defer q.mu.Unlock()

  	q.entries = make([]DeadLetterEntry, q.maxSize)
  	q.writeIdx = 0
  	q.count = 0
  }
  ```

- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v -run TestDLQ_BasicAddAndRetrieve ./kernel/eventbus/...`
  Expected: PASS

- [ ] **Step 5: Commit**
  Run: `git add kernel/eventbus/dlq.go kernel/eventbus/dlq_test.go`
  Run: `git commit -m "feat: add dead letter queue implementation and basic test"`

---

### Task 2: Write unit tests verifying DLQ circular buffer and chronological order in `kernel/eventbus/dlq_test.go`

**Files:**
- Modify: `kernel/eventbus/dlq_test.go`

**Interfaces:**
- Consumes: `DeadLetterQueue` APIs from Task 1

- [ ] **Step 1: Write the failing test**
  Add `TestDLQ_CircularBufferAndSorting` to `kernel/eventbus/dlq_test.go`:
  ```go
  func TestDLQ_CircularBufferAndSorting(t *testing.T) {
  	q := eventbus.NewDeadLetterQueue(3)

  	q.Add(event.Event{ID: "e1"}, "err1")
  	q.Add(event.Event{ID: "e2"}, "err2")
  	q.Add(event.Event{ID: "e3"}, "err3")

  	// At this point q is full: entries are e1, e2, e3
  	entries := q.Entries()
  	if len(entries) != 3 {
  		t.Fatalf("expected 3 entries, got %d", len(entries))
  	}
  	if entries[0].Event.ID != "e1" || entries[1].Event.ID != "e2" || entries[2].Event.ID != "e3" {
  		t.Errorf("wrong initial order: %v, %v, %v", entries[0].Event.ID, entries[1].Event.ID, entries[2].Event.ID)
  	}

  	// Add e4: should overwrite e1
  	q.Add(event.Event{ID: "e4"}, "err4")

  	if q.Len() != 3 {
  		t.Errorf("expected length 3 after overflow, got %d", q.Len())
  	}

  	entries = q.Entries()
  	if len(entries) != 3 {
  		t.Fatalf("expected 3 entries after overflow, got %d", len(entries))
  	}

  	// Oldest first: e2, e3, e4
  	if entries[0].Event.ID != "e2" || entries[1].Event.ID != "e3" || entries[2].Event.ID != "e4" {
  		t.Errorf("wrong chronological order: got %v, %v, %v; want e2, e3, e4", entries[0].Event.ID, entries[1].Event.ID, entries[2].Event.ID)
  	}

  	q.Clear()
  	if q.Len() != 0 {
  		t.Errorf("expected 0 after Clear, got %d", q.Len())
  	}
  }
  ```

- [ ] **Step 2: Run test to verify it passes**
  Run: `go test -v -run TestDLQ_CircularBufferAndSorting ./kernel/eventbus/...`
  Expected: PASS

- [ ] **Step 3: Commit**
  Run: `git add kernel/eventbus/dlq_test.go`
  Run: `git commit -m "test: add circular buffer and chronological order verification tests for DLQ"`

---

### Task 3: Update `kernel/eventbus/subscriber.go`

**Files:**
- Modify: `kernel/eventbus/subscriber.go`

**Interfaces:**
- Consumes: `DeadLetterQueue` pointer parameter in `safeHandler`
- Produces: Updated signature `safeHandler(handler func(event.Event), dlq *DeadLetterQueue, logger *slog.Logger) func(event.Event)`

- [ ] **Step 1: Modify code**
  Update `safeHandler` signature and recovery block in `kernel/eventbus/subscriber.go`.
  Wait, modifying `safeHandler` signature will break the compilation of `kernel/eventbus/bus.go` since it calls it. This is expected.
  Let's check the diff:
  ```go
  func safeHandler(handler func(event.Event), dlq *DeadLetterQueue, logger *slog.Logger) func(event.Event) {
  	return func(evt event.Event) {
  		defer func() {
  			if r := recover(); r != nil {
  				errStr := fmt.Sprintf("%v", r)
  				if logger != nil {
  					logger.Error("event handler panicked",
  						"event_type", evt.Type,
  						"event_source", evt.Source,
  						"panic", errStr,
  					)
  				}
  				if dlq != nil {
  					dlq.Add(evt, errStr)
  				}
  			}
  		}()
  		handler(evt)
  	}
  }
  ```

- [ ] **Step 2: Run verification**
  Run: `go build ./kernel/eventbus/...`
  Expected: FAIL (bus.go compilation error due to mismatching safeHandler signature)

---

### Task 4: Update `kernel/eventbus/bus.go`

**Files:**
- Modify: `kernel/eventbus/bus.go`

**Interfaces:**
- Consumes: Updated `safeHandler` from Task 3
- Produces: `dlq` field in `Bus`, getter method `DLQ() *DeadLetterQueue`, initializes with default size 100 in `New`

- [ ] **Step 1: Modify code**
  Update `Bus` struct, `New` constructor, `Subscribe` method, and add `DLQ() *DeadLetterQueue` method:
  ```go
  type Bus struct {
  	subscribers *subscriberMap
  	logger      *slog.Logger
  	dlq         *DeadLetterQueue
  	wg          sync.WaitGroup
  	closed      bool
  	mu          sync.RWMutex
  }

  func New(logger *slog.Logger) *Bus {
  	return &Bus{
  		subscribers: newSubscriberMap(),
  		logger:      logger,
  		dlq:         NewDeadLetterQueue(100),
  	}
  }

  // inside (b *Bus) Subscribe:
  safe := safeHandler(handler, b.dlq, b.logger)

  // new method:
  func (b *Bus) DLQ() *DeadLetterQueue {
  	return b.dlq
  }
  ```

- [ ] **Step 2: Run verification**
  Run: `go test -v ./kernel/eventbus/...`
  Expected: PASS

- [ ] **Step 3: Commit**
  Run: `git add kernel/eventbus/subscriber.go kernel/eventbus/bus.go`
  Run: `git commit -m "feat: integrate dead letter queue into eventbus and safeHandler"`

---

### Task 5: Add integration unit tests verifying DLQ panic capture in `kernel/eventbus/bus_test.go`

**Files:**
- Modify: `kernel/eventbus/bus_test.go`

**Interfaces:**
- Consumes: EventBus publish/subscribe with panic handlers, DLQ getter

- [ ] **Step 1: Write the test**
  Add `TestBus_HandlerPanic_EnqueuesDLQ` to `kernel/eventbus/bus_test.go`:
  ```go
  func TestBus_HandlerPanic_EnqueuesDLQ(t *testing.T) {
  	bus := eventbus.New(nil)
  	defer bus.Close()

  	panicDone := make(chan struct{}, 1)

  	unsub, err := bus.Subscribe("task.panic", func(evt event.Event) {
  		defer func() { panicDone <- struct{}{} }()
  		panic("panic error message")
  	})
  	if err != nil {
  		t.Fatalf("Subscribe: %v", err)
  	}
  	defer unsub()

  	evt := testEvent("task.panic")
  	bus.Publish(context.Background(), evt)

  	select {
  	case <-panicDone:
  		// Wait a small duration for safeHandler recovery logic to finish adding to DLQ
  		time.Sleep(100 * time.Millisecond)
  	case <-time.After(2 * time.Second):
  		t.Fatal("timeout waiting for panic handler to execute")
  	}

  	dlq := bus.DLQ()
  	if dlq == nil {
  		t.Fatal("expected non-nil DLQ")
  	}

  	if dlq.Len() != 1 {
  		t.Fatalf("expected 1 error in DLQ, got %d", dlq.Len())
  	}

  	entries := dlq.Entries()
  	if len(entries) != 1 {
  		t.Fatalf("expected 1 entry in DLQ, got %d", len(entries))
  	}

  	entry := entries[0]
  	if entry.Event.Type != "task.panic" {
  		t.Errorf("expected event type task.panic, got %s", entry.Event.Type)
  	}
  	if entry.Error != "panic error message" {
  		t.Errorf("expected panic error message, got %s", entry.Error)
  	}
  }
  ```

- [ ] **Step 2: Run verification**
  Run: `go test -v ./kernel/eventbus/...`
  Expected: PASS

- [ ] **Step 3: Commit**
  Run: `git add kernel/eventbus/bus_test.go`
  Run: `git commit -m "test: add integration test for handler panic enqueuing in DLQ"`
