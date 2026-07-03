# Micro-Task 2.38: Create kernel/eventbus/dlq.go (Event Bus Dead Letter Queue)

## Info
- **File created**: `kernel/eventbus/dlq.go`
- **Files updated**:
  - `kernel/eventbus/subscriber.go` (Update `safeHandler` to publish to DLQ on panic)
  - `kernel/eventbus/bus.go` (Add DLQ struct field, constructor allocation, and getter methods)
- **Package**: `eventbus`
- **Depends on**: 2.12 (types.go), 2.14 (subscriber.go), 2.15 (bus.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Implements a thread-safe circular Dead Letter Queue (DLQ) ring buffer that captures failed event invocations, records panic stack diagnostics along with the original event context, and updates the core `Bus` implementation to leverage this safety mechanism.

## EXACT code to create

### Part 1: Create `kernel/eventbus/dlq.go`

```go
package eventbus

import (
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// DeadLetterEntry represents a failed event execution log.
type DeadLetterEntry struct {
	Event     event.Event `json:"event"`
	Error     string      `json:"error"`
	Timestamp time.Time   `json:"timestamp"`
}

// DeadLetterQueue is a thread-safe circular buffer (ring buffer) for failed events.
// It stores up to maxSize failures, discarding oldest entries when full.
type DeadLetterQueue struct {
	mu       sync.RWMutex
	entries  []DeadLetterEntry
	maxSize  int
	writeIdx int
	count    int
}

// NewDeadLetterQueue creates a new DLQ with the specified capacity.
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &DeadLetterQueue{
		entries: make([]DeadLetterEntry, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new failure entry into the circular buffer.
func (q *DeadLetterQueue) Add(evt event.Event, errStr string) {
	q.mu.Lock()
	defer q.mu.Unlock()

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

// Entries returns a copy of all active entries in chronological order (oldest first).
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

// Len returns the current number of failures in the queue.
func (q *DeadLetterQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count
}

// Clear resets the queue.
func (q *DeadLetterQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.entries = make([]DeadLetterEntry, q.maxSize)
	q.writeIdx = 0
	q.count = 0
}
```

---

### Part 2: Update `kernel/eventbus/subscriber.go`

In `kernel/eventbus/subscriber.go`, update the `safeHandler` function signature and implementation:

```diff
-func safeHandler(handler func(event.Event), logger *slog.Logger) func(event.Event) {
+func safeHandler(handler func(event.Event), dlq *DeadLetterQueue, logger *slog.Logger) func(event.Event) {
 	return func(evt event.Event) {
 		defer func() {
 			if r := recover(); r != nil {
+				errStr := fmt.Sprintf("%v", r)
 				if logger != nil {
 					logger.Error("event handler panicked",
 						"event_type", evt.Type,
 						"event_source", evt.Source,
-						"panic", fmt.Sprintf("%v", r),
+						"panic", errStr,
 					)
 				}
+				if dlq != nil {
+					dlq.Add(evt, errStr)
+				}
 			}
 		}()
 		handler(evt)
 	}
 }
```

---

### Part 3: Update `kernel/eventbus/bus.go`

In `kernel/eventbus/bus.go`, update the `Bus` struct, constructor, `Subscribe` method, and add a `DLQ` getter:

```diff
 type Bus struct {
 	subscribers *subscriberMap
 	logger      *slog.Logger
+	dlq         *DeadLetterQueue
 	wg          sync.WaitGroup
 	closed      bool
 	mu          sync.RWMutex
 }

 func New(logger *slog.Logger) *Bus {
 	return &Bus{
 		subscribers: newSubscriberMap(),
 		logger:      logger,
+		dlq:         NewDeadLetterQueue(100),
 	}
 }

 func (b *Bus) Subscribe(pattern string, handler func(event.Event)) (func(), error) {
 	if err := validatePattern(pattern); err != nil {
 		return nil, err
 	}

 	b.mu.RLock()
 	if b.closed {
 		b.mu.RUnlock()
 		return nil, fmt.Errorf("eventbus: bus is closed")
 	}
 	b.mu.RUnlock()

-	safe := safeHandler(handler, b.logger)
+	safe := safeHandler(handler, b.dlq, b.logger)
 	sub := b.subscribers.add(pattern, safe)

 	return makeUnsubscribe(b.subscribers, sub), nil
 }

+func (b *Bus) DLQ() *DeadLetterQueue {
+	return b.dlq
+}
```

## Rules
1. **Thread-Safe Ring Buffer Limits**: Dead Letter Queues must utilize circular buffers (ring buffers) with fixed maximum capacities (`maxSize`) rather than unbounded slices. This avoids memory exhaustion under massive panic incidents.
2. **Correct Chronological Exports**: The `Entries()` method must extract items in strict chronological sequence (oldest first). Ensure starting indices are computed as `writeIdx` when the ring is full.
3. **Panic Logging Integration**: The `safeHandler` function must publish the recovered panic string alongside original event payloads directly to the active DLQ.

## ⚠️ Pitfalls

### Pitfall 1: Accumulating logs in unbounded slices
```go
```
Always allocate fixed-size structures and wrap indexes (`writeIdx = (writeIdx + 1) % maxSize`).

### Pitfall 2: Copying circular buffer items using linear index offsets when full
If a ring buffer is full and has wrapped around, using `0` as the starting index inside the copy loop output returns mismatched event histories. Calculate the start index from the current write index location.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [] File `kernel/eventbus/dlq.go` exists
- [] `DeadLetterQueue` implements circular ring buffer logic
- [] `safeHandler` accepts `DeadLetterQueue` parameters and publishes recovery logs on panics
- [] `Bus` struct integrates `dlq` fields
- [ ] `New` initializes DLQ instances with safe defaults
- [] `DLQ()` accessor returns reference pointers
- [] `Entries` recovers and sorts events chronological (oldest first)
- [] All locks release cleanly via defer calls
- [] `go build ./kernel/eventbus/...` passes
