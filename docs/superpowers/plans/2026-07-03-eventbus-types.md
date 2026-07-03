# EventBus Types Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `kernel/eventbus/types.go` to support internal event subscription management, complete with thread-safe concurrency controls, atomic subscription activation state, and correct array memory cleanup.

**Architecture:** Implement `subscription` and `subscriberMap` structs inside the internal `eventbus` package. Use `atomic.Bool` to track subscription active states, `sync.RWMutex` for concurrent access protection, and slice index swapping with tail pointer clearing to prevent memory leaks during subscriber removal.

**Tech Stack:** Go 1.26 stdlib (sync, sync/atomic).

## Global Constraints

- Use named field initialization for all Go structs (AEOS AI Coding Rules §4).
- Do not spawn unbounded goroutines (AEOS AI Coding Rules §4).
- Keep file size <= 300 LOC (AEOS AI Coding Rules §2).
- Keep function length <= 80 LOC (AEOS AI Coding Rules §2).
- Zero circular imports; layer constraint: contracts/ -> kernel/ -> sdk/ (AEOS AI Coding Rules §3).

---

### Task 1: Create Types and Implementation in kernel/eventbus/types.go

**Files:**
- Create: `kernel/eventbus/types.go`
- Test: `kernel/eventbus/types_test.go`

**Interfaces:**
- Consumes: `contracts/event` package
- Produces: `subscription`, `subscriberMap` internal structures and methods, `matchPattern` placeholder.

- [ ] **Step 1: Write the code for kernel/eventbus/types.go**

Create `kernel/eventbus/types.go` containing:
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

// matchPattern is a temporary placeholder to ensure compilation.
func matchPattern(pattern, eventType string) bool {
	return false
}
```

- [ ] **Step 2: Run verification command to build the package**

Run: `go build ./kernel/eventbus/...`
Expected: Compilation passes successfully with no warnings.

- [ ] **Step 3: Write tests in kernel/eventbus/types_test.go**

Create `kernel/eventbus/types_test.go` to verify the subscriberMap and subscription functionality, ensuring full coverage of concurrency safety, atomic operations, and memory cleanup (i.e. correct slice manipulation):
```go
package eventbus

import (
	"sync"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

func TestSubscriptionLifecycle(t *testing.T) {
	handler := func(e event.Event) {}
	sub := newSubscription(42, "test.pattern", handler)

	if sub.id != 42 {
		t.Errorf("expected subscription ID 42, got %d", sub.id)
	}
	if sub.pattern != "test.pattern" {
		t.Errorf("expected pattern 'test.pattern', got %q", sub.pattern)
	}
	if !sub.isActive() {
		t.Error("expected new subscription to be active")
	}

	sub.deactivate()
	if sub.isActive() {
		t.Error("expected subscription to be inactive after deactivation")
	}
}

func TestSubscriberMapAddAndRemove(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	sub1 := sm.add("evt.1", handler)
	sub2 := sm.add("evt.2", handler)

	if sub1.id == 0 || sub2.id == 0 {
		t.Error("expected non-zero IDs assigned")
	}
	if sub1.id == sub2.id {
		t.Errorf("expected unique IDs, got identical ID: %d", sub1.id)
	}

	if sm.count() != 2 {
		t.Errorf("expected count to be 2, got %d", sm.count())
	}

	sm.remove(sub1.id)
	if sm.count() != 1 {
		t.Errorf("expected count to be 1 after remove, got %d", sm.count())
	}

	if sub1.isActive() {
		t.Error("expected removed subscription to be deactivated")
	}

	// Verify the remaining subscriber is sub2
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if len(sm.subs) != 1 || sm.subs[0].id != sub2.id {
		t.Errorf("expected only sub2 to remain in slice, got %v", sm.subs)
	}
}

func TestSubscriberMapMatchingPlaceholder(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	sm.add("evt.*", handler)

	// Since matchPattern placeholder always returns false, matching must return empty
	matched := sm.matching("evt.test")
	if len(matched) != 0 {
		t.Errorf("expected matching to return empty list with placeholder matchPattern, got %d items", len(matched))
	}
}

func TestSubscriberMapConcurrency(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	var wg sync.WaitGroup
	workers := 10
	addsPerWorker := 50

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var subs []*subscription
			for i := 0; i < addsPerWorker; i++ {
				sub := sm.add("test", handler)
				subs = append(subs, sub)
			}
			// Concurrently remove half of them
			for i := 0; i < addsPerWorker; i += 2 {
				sm.remove(subs[i].id)
			}
		}()
	}

	wg.Wait()

	expectedCount := workers * (addsPerWorker / 2)
	if sm.count() != expectedCount {
		t.Errorf("expected count %d, got %d", expectedCount, sm.count())
	}
}
```

- [ ] **Step 4: Run unit tests to verify correctness**

Run: `go test -v ./kernel/eventbus/...`
Expected: All tests pass.

- [ ] **Step 5: Run quality gate validations**

Run:
```bash
go fmt ./...
go vet ./...
golangci-lint run ./kernel/eventbus/...
go test -race ./kernel/eventbus/...
```
Expected: Zero warnings, zero errors.
