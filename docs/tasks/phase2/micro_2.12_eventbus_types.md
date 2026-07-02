# Micro-Task 2.12: Tạo kernel/eventbus/types.go

## Thông tin
- **File tạo**: `kernel/eventbus/types.go`
- **Package**: `eventbus`
- **Dependencies trước**: Phase 1 (contracts/event)
- **Thời gian**: 10 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Mục đích
Định nghĩa internal types cho EventBus implementation.
Tách khỏi contracts/event vì đây là implementation details.

## Nội dung CHÍNH XÁC cần tạo

```go
// Package eventbus implements the event.Bus contract.
// It provides an in-memory, thread-safe, asynchronous publish/subscribe system.
//
// Key properties:
//   - Thread-safe: all methods can be called from multiple goroutines
//   - Asynchronous: Publish() returns immediately, handlers run in goroutines
//   - Wildcard: Subscribe("task.*") matches "task.started", "task.completed"
//   - Panic-safe: handler panics are recovered, logged, and don't crash the system
//   - Ordered: events are delivered in publish order WITHIN a single subscriber
package eventbus

import (
	"sync"
	"sync/atomic"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// subscription represents a single event subscriber.
type subscription struct {
	// id uniquely identifies this subscription (for unsubscribe).
	id uint64

	// pattern is the event type pattern to match.
	// Examples: "task.started" (exact), "task.*" (wildcard), "*" (all)
	pattern string

	// handler is the function to call when a matching event is published.
	// handler is called in a separate goroutine for each event.
	handler func(event.Event)

	// active indicates whether this subscription is still active.
	// Set to false by unsubscribe. Checked before invoking handler.
	//
	// WHY atomic.Bool instead of regular bool?
	// → Unsubscribe() may be called from goroutine A while handler is
	//   being invoked in goroutine B. Without atomic, data race.
	active atomic.Bool
}

// newSubscription creates a new active subscription.
func newSubscription(id uint64, pattern string, handler func(event.Event)) *subscription {
	s := &subscription{
		id:      id,
		pattern: pattern,
		handler: handler,
	}
	s.active.Store(true)
	return s
}

// isActive returns whether this subscription is still active.
func (s *subscription) isActive() bool {
	return s.active.Load()
}

// deactivate marks this subscription as inactive (unsubscribed).
// Safe to call multiple times (idempotent).
func (s *subscription) deactivate() {
	s.active.Store(false)
}

// subscriberMap provides thread-safe storage for subscriptions.
//
// WHY sync.RWMutex instead of sync.Mutex?
// → Publish needs to READ the subscriber list (RLock — multiple readers).
// → Subscribe/Unsubscribe need to WRITE the subscriber list (Lock — exclusive).
// → RWMutex allows concurrent publishes without blocking each other.
type subscriberMap struct {
	mu   sync.RWMutex
	subs []*subscription
	// nextID is atomic to avoid locking for ID generation.
	nextID atomic.Uint64
}

// newSubscriberMap creates a new subscriber map.
func newSubscriberMap() *subscriberMap {
	return &subscriberMap{
		subs: make([]*subscription, 0),
	}
}

// add registers a new subscription and returns it.
func (sm *subscriberMap) add(pattern string, handler func(event.Event)) *subscription {
	id := sm.nextID.Add(1)
	sub := newSubscription(id, pattern, handler)

	sm.mu.Lock()
	sm.subs = append(sm.subs, sub)
	sm.mu.Unlock()

	return sub
}

// remove deactivates a subscription and removes it from the list.
// Safe to call multiple times.
func (sm *subscriberMap) remove(id uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, sub := range sm.subs {
		if sub.id == id {
			sub.deactivate()
			// Remove from slice by swapping with last element
			sm.subs[i] = sm.subs[len(sm.subs)-1]
			sm.subs[len(sm.subs)-1] = nil // Allow GC
			sm.subs = sm.subs[:len(sm.subs)-1]
			return
		}
	}
}

// matching returns all active subscriptions that match the given event type.
// Returns a COPY of matching subscriptions (safe to iterate without lock).
func (sm *subscriberMap) matching(eventType string) []*subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var matched []*subscription
	for _, sub := range sm.subs {
		if sub.isActive() && matchPattern(sub.pattern, eventType) {
			matched = append(matched, sub)
		}
	}
	return matched
}

// count returns the number of active subscriptions.
func (sm *subscriberMap) count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	count := 0
	for _, sub := range sm.subs {
		if sub.isActive() {
			count++
		}
	}
	return count
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: atomic.Bool cho subscription.active
Unsubscribe goroutine A sets `active = false`.
Handler goroutine B checks `active` before invoking.
Without atomic → data race → `go test -race` fails.

### Pitfall 2: Slice remove pattern
```go
// Swap with last element + shrink slice
sm.subs[i] = sm.subs[len(sm.subs)-1]
sm.subs[len(sm.subs)-1] = nil  // IMPORTANT: allow GC to collect the subscription
sm.subs = sm.subs[:len(sm.subs)-1]
```
Setting last element to `nil` prevents memory leak (slice still holds reference).

### Pitfall 3: matching() returns COPY
```go
// Returns new slice → caller can iterate without holding the lock
var matched []*subscription
```
Nếu trả về slice reference từ subscriberMap → concurrent modification risk.

## Checklist
- [ ] File `kernel/eventbus/types.go` tồn tại
- [ ] `subscription` struct: id, pattern, handler, active (atomic.Bool)
- [ ] `newSubscription()` constructor sets active=true
- [ ] `isActive()` dùng atomic.Load
- [ ] `deactivate()` dùng atomic.Store, idempotent
- [ ] `subscriberMap` struct: RWMutex + subs slice + atomic nextID
- [ ] `add()` — Lock for write, generates unique ID
- [ ] `remove()` — Lock for write, swap+shrink pattern, nil for GC
- [ ] `matching()` — RLock for read, returns COPY
- [ ] `count()` — RLock for read
- [ ] `go build ./kernel/eventbus/...` không lỗi
