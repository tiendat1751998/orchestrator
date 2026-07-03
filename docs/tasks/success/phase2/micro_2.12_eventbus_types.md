# Micro-Task 2.12: Create kernel/eventbus/types.go

## Info
- **File**: `kernel/eventbus/types.go`
- **Package**: `eventbus`
- **Depends on**: Phase 1 completed (contracts/event package)
- **Time**: 10 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Defines internal structures (`subscription`, `subscriberMap`) that support the EventBus implementation details. These types are kept within the implementation package rather than the public contracts package.

## EXACT code to create

```go
package eventbus

import (
	"sync"
	"sync/atomic"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// subscription represents a single event subscriber.
type subscription struct {
	id      uint64
	pattern string
	handler func(event.Event)
	active  atomic.Bool
}

func newSubscription(id uint64, pattern string, handler func(event.Event)) *subscription {
	s := &subscription{
		id:      id,
		pattern: pattern,
		handler: handler,
	}
	s.active.Store(true)
	return s
}

func (s *subscription) isActive() bool {
	return s.active.Load()
}

func (s *subscription) deactivate() {
	s.active.Store(false)
}

type subscriberMap struct {
	mu     sync.RWMutex
	subs   []*subscription
	nextID atomic.Uint64
}

func newSubscriberMap() *subscriberMap {
	return &subscriberMap{
		subs: make([]*subscription, 0),
	}
}

func (sm *subscriberMap) add(pattern string, handler func(event.Event)) *subscription {
	id := sm.nextID.Add(1)
	sub := newSubscription(id, pattern, handler)

	sm.mu.Lock()
	sm.subs = append(sm.subs, sub)
	sm.mu.Unlock()

	return sub
}

func (sm *subscriberMap) remove(id uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, sub := range sm.subs {
		if sub.id == id {
			sub.deactivate()
			// Swap with last element and slice
			sm.subs[i] = sm.subs[len(sm.subs)-1]
			sm.subs[len(sm.subs)-1] = nil
			sm.subs = sm.subs[:len(sm.subs)-1]
			return
		}
	}
}

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

## Rules
1. **Concurrency Control (RWMutex)**: Use `sync.RWMutex` to allow concurrent `matching` check lookups (which only read data) without blocking publish actions.
2. **Atomic States**: The `active` boolean inside the `subscription` struct must use `atomic.Bool` to avoid data races when unsubscribe actions are called concurrently with event deliveries.
3. **Array Memory Cleanup**: Swapping the deleted index with the tail element is standard in Go. Make sure to set the trailing index to `nil` before slicing to allow the garbage collector to release memory.

## Pitfalls

### Pitfall 1: Data race on subscriber active checks
If a subscriber calls `Unsubscribe()` from goroutine A while goroutine B is pushing events, checking `active` via normal boolean assignments triggers a race condition. Always use `atomic.Bool` for safe state checks.

### Pitfall 2: Memory leaks from hanging slice references
```go
// WRONG:
sm.subs[i] = sm.subs[len(sm.subs)-1]
sm.subs = sm.subs[:len(sm.subs)-1] // Erased from slice, but the last index memory still holds a pointer to the sub, blocking GC!

// CORRECT:
sm.subs[i] = sm.subs[len(sm.subs)-1]
sm.subs[len(sm.subs)-1] = nil // Set index pointer to nil to permit garbage collection
sm.subs = sm.subs[:len(sm.subs)-1]
```
Always set discarded tail elements to `nil` when cutting slices.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/types.go` exists
- [ ] Package: `eventbus`
- [ ] `subscription` struct uses `atomic.Bool` for the active field
- [ ] `subscriberMap` struct uses `sync.RWMutex` to synchronize slice operations
- [ ] `add()` assigns unique IDs atomically using `atomic.Uint64` counters
- [ ] `remove()` correctly swaps and clears tail pointers to prevent memory leaks
- [ ] `matching()` reads lists safely under `RLock` and returns copies
- [ ] `go build ./kernel/eventbus/...` passes
