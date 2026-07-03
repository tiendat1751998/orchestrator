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
