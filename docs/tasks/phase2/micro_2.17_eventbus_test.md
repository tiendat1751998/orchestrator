# Micro-Task 2.17: Create kernel/eventbus/bus_test.go

## Info
- **File**: `kernel/eventbus/bus_test.go`
- **Package**: `eventbus_test`
- **Depends on**: 2.12-2.16
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/eventbus/...`

## Purpose
Comprehensive tests for EventBus: publish/subscribe, wildcard matching,
concurrency safety, panic recovery, unsubscribe, and close behavior.

## EXACT code to create

```go
package eventbus_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
)

// =============================================================================
// Helper: create a test event
// =============================================================================

func testEvent(eventType string) event.Event {
	return event.Event{
		ID:        "test-evt-001",
		Type:      eventType,
		Source:    "test",
		Payload:   "test-payload",
		Timestamp: time.Now(),
	}
}

// =============================================================================
// Basic Publish/Subscribe
// =============================================================================

func TestBus_PublishSubscribe_ExactMatch(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	// Channel to receive events (acts as synchronization point)
	received := make(chan event.Event, 1)

	unsub, err := bus.Subscribe("task.started", func(evt event.Event) {
		received <- evt
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer unsub()

	// Publish
	err = bus.Publish(context.Background(), testEvent("task.started"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Wait for handler (with timeout to avoid hanging test)
	select {
	case evt := <-received:
		if evt.Type != "task.started" {
			t.Errorf("event type: got %q, want %q", evt.Type, "task.started")
		}
		if evt.Source != "test" {
			t.Errorf("event source: got %q, want %q", evt.Source, "test")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: handler was not called within 2 seconds")
	}
}

func TestBus_PublishSubscribe_NoMatch(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	called := make(chan struct{}, 1)

	unsub, err := bus.Subscribe("task.started", func(evt event.Event) {
		called <- struct{}{}
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer unsub()

	// Publish a DIFFERENT event type
	bus.Publish(context.Background(), testEvent("task.completed"))

	// Handler should NOT be called
	select {
	case <-called:
		t.Error("handler should NOT be called for non-matching event")
	case <-time.After(200 * time.Millisecond):
		// Expected: no call within timeout
	}
}

// =============================================================================
// Wildcard Matching
// =============================================================================

func TestBus_WildcardSubscription(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var count atomic.Int32

	unsub, err := bus.Subscribe("task.*", func(evt event.Event) {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer unsub()

	// Publish multiple task events
	bus.Publish(context.Background(), testEvent("task.started"))
	bus.Publish(context.Background(), testEvent("task.completed"))
	bus.Publish(context.Background(), testEvent("task.failed"))

	// Publish non-matching event
	bus.Publish(context.Background(), testEvent("mission.started"))

	// Wait for handlers to complete
	time.Sleep(500 * time.Millisecond)

	got := count.Load()
	if got != 3 {
		t.Errorf("wildcard handler called %d times, want 3", got)
	}
}

func TestBus_GlobalWildcard(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var count atomic.Int32

	unsub, err := bus.Subscribe("*", func(evt event.Event) {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer unsub()

	bus.Publish(context.Background(), testEvent("task.started"))
	bus.Publish(context.Background(), testEvent("mission.completed"))
	bus.Publish(context.Background(), testEvent("kernel.stopped"))

	time.Sleep(500 * time.Millisecond)

	got := count.Load()
	if got != 3 {
		t.Errorf("global wildcard handler called %d times, want 3", got)
	}
}

// =============================================================================
// Unsubscribe
// =============================================================================

func TestBus_Unsubscribe_StopsDelivery(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var count atomic.Int32

	unsub, err := bus.Subscribe("task.started", func(evt event.Event) {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Publish before unsubscribe — should be received
	bus.Publish(context.Background(), testEvent("task.started"))
	time.Sleep(200 * time.Millisecond)

	if count.Load() != 1 {
		t.Fatalf("expected 1 call before unsubscribe, got %d", count.Load())
	}

	// Unsubscribe
	unsub()

	// Publish after unsubscribe — should NOT be received
	bus.Publish(context.Background(), testEvent("task.started"))
	time.Sleep(200 * time.Millisecond)

	if count.Load() != 1 {
		t.Errorf("expected still 1 call after unsubscribe, got %d", count.Load())
	}
}

func TestBus_Unsubscribe_Idempotent(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	unsub, err := bus.Subscribe("task.started", func(evt event.Event) {})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Call unsubscribe multiple times — should NOT panic
	unsub()
	unsub()
	unsub()
}

func TestBus_SubscriberCount(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	if bus.SubscriberCount() != 0 {
		t.Errorf("initial count: got %d, want 0", bus.SubscriberCount())
	}

	unsub1, _ := bus.Subscribe("task.started", func(evt event.Event) {})
	unsub2, _ := bus.Subscribe("task.completed", func(evt event.Event) {})

	if bus.SubscriberCount() != 2 {
		t.Errorf("after 2 subscribes: got %d, want 2", bus.SubscriberCount())
	}

	unsub1()

	if bus.SubscriberCount() != 1 {
		t.Errorf("after 1 unsubscribe: got %d, want 1", bus.SubscriberCount())
	}

	unsub2()

	if bus.SubscriberCount() != 0 {
		t.Errorf("after all unsubscribed: got %d, want 0", bus.SubscriberCount())
	}
}

// =============================================================================
// Concurrent Safety (CRITICAL — must pass go test -race)
// =============================================================================

func TestBus_ConcurrentPublishSubscribe(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var wg sync.WaitGroup
	var totalReceived atomic.Int64

	// 10 subscribers, each listening to different patterns
	for i := 0; i < 10; i++ {
		pattern := fmt.Sprintf("topic%d.*", i)
		unsub, err := bus.Subscribe(pattern, func(evt event.Event) {
			totalReceived.Add(1)
		})
		if err != nil {
			t.Fatalf("Subscribe %d: %v", i, err)
		}
		defer unsub()
	}

	// 10 publishers, each publishing 100 events concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(topicID int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				evt := testEvent(fmt.Sprintf("topic%d.event", topicID))
				bus.Publish(context.Background(), evt)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(1 * time.Second) // Wait for all handlers to complete

	// Each publisher publishes 100 events, each to its own topic
	// Each topic has 1 matching subscriber → 10 * 100 = 1000 deliveries
	got := totalReceived.Load()
	if got != 1000 {
		t.Errorf("concurrent: got %d deliveries, want 1000", got)
	}
}

func TestBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var wg sync.WaitGroup

	// 100 goroutines each subscribe then unsubscribe
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			unsub, err := bus.Subscribe(fmt.Sprintf("topic%d.*", id), func(evt event.Event) {})
			if err != nil {
				return
			}
			time.Sleep(time.Duration(id%10) * time.Millisecond)
			unsub()
		}(i)
	}

	wg.Wait()
	// If no panic and no race condition → PASS
}

// =============================================================================
// Panic Recovery
// =============================================================================

func TestBus_HandlerPanic_DoesNotCrash(t *testing.T) {
	bus := eventbus.New(nil) // nil logger — panic is silently recovered
	defer bus.Close()

	panicDone := make(chan struct{}, 1)
	normalDone := make(chan struct{}, 1)

	// Subscriber 1: PANICS
	unsub1, _ := bus.Subscribe("task.started", func(evt event.Event) {
		defer func() { panicDone <- struct{}{} }()
		panic("intentional panic in test")
	})
	defer unsub1()

	// Subscriber 2: normal handler
	unsub2, _ := bus.Subscribe("task.started", func(evt event.Event) {
		normalDone <- struct{}{}
	})
	defer unsub2()

	// Publish — panicking handler should NOT prevent normal handler from running
	bus.Publish(context.Background(), testEvent("task.started"))

	// Both handlers should complete (one via panic recovery)
	select {
	case <-normalDone:
		// OK: normal handler was called
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: normal handler was not called (panicking handler may have crashed the bus)")
	}
}

// =============================================================================
// Close Behavior
// =============================================================================

func TestBus_Close_RejectsNewPublishes(t *testing.T) {
	bus := eventbus.New(nil)

	bus.Close()

	err := bus.Publish(context.Background(), testEvent("task.started"))
	if err == nil {
		t.Error("expected error when publishing to closed bus")
	}
}

func TestBus_Close_WaitsForInFlightHandlers(t *testing.T) {
	bus := eventbus.New(nil)

	handlerStarted := make(chan struct{})
	handlerDone := make(chan struct{})

	unsub, _ := bus.Subscribe("task.started", func(evt event.Event) {
		close(handlerStarted)
		time.Sleep(500 * time.Millisecond) // Simulate slow handler
		close(handlerDone)
	})
	defer unsub()

	// Publish an event
	bus.Publish(context.Background(), testEvent("task.started"))

	// Wait for handler to start
	<-handlerStarted

	// Close should block until handler completes
	closeDone := make(chan struct{})
	go func() {
		bus.Close()
		close(closeDone)
	}()

	// Close should NOT return before handler finishes
	select {
	case <-closeDone:
		// Verify handler actually completed
		select {
		case <-handlerDone:
			// OK: Close waited for handler
		default:
			t.Error("Close returned before handler completed")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: Close() hung")
	}
}

// =============================================================================
// Context Cancellation
// =============================================================================

func TestBus_Publish_RespectsContext(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := bus.Publish(ctx, testEvent("task.started"))
	if err == nil {
		t.Error("expected error when publishing with cancelled context")
	}
}

// =============================================================================
// Input Validation
// =============================================================================

func TestBus_Subscribe_EmptyPattern(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	_, err := bus.Subscribe("", func(evt event.Event) {})
	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestBus_Subscribe_NilHandler(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	_, err := bus.Subscribe("task.started", nil)
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestBus_Subscribe_InvalidPattern(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	_, err := bus.Subscribe(".started", func(evt event.Event) {})
	if err == nil {
		t.Error("expected error for pattern starting with dot")
	}
}

// =============================================================================
// Pattern Matching (internal matchPattern via integration)
// =============================================================================

func TestBus_PatternMatching_SegmentCount(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	received := make(chan struct{}, 1)

	// "task.*" should NOT match "task" (different segment count)
	unsub, _ := bus.Subscribe("task.*", func(evt event.Event) {
		received <- struct{}{}
	})
	defer unsub()

	bus.Publish(context.Background(), testEvent("task")) // 1 segment, pattern has 2

	select {
	case <-received:
		t.Error("task.* should NOT match 'task' (different segment count)")
	case <-time.After(200 * time.Millisecond):
		// Expected: no match
	}
}

// =============================================================================
// Multiple Subscribers
// =============================================================================

func TestBus_MultipleSubscribers_AllReceive(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var count atomic.Int32

	// 5 subscribers all listening to the same event
	for i := 0; i < 5; i++ {
		unsub, _ := bus.Subscribe("task.started", func(evt event.Event) {
			count.Add(1)
		})
		defer unsub()
	}

	bus.Publish(context.Background(), testEvent("task.started"))
	time.Sleep(500 * time.Millisecond)

	if count.Load() != 5 {
		t.Errorf("expected all 5 subscribers to receive, got %d", count.Load())
	}
}
```

## Pitfalls

### Pitfall 1: Time-based synchronization in tests
```go
// BAD — flaky:
time.Sleep(10 * time.Millisecond)

// GOOD — deterministic:
select {
case evt := <-received:
    // check evt
case <-time.After(2 * time.Second):
    t.Fatal("timeout")
}
```
Channel + select + timeout = deterministic. Pure sleep = flaky on slow CI.

### Pitfall 2: Atomic counters for concurrent tests
```go
var count atomic.Int32  // Thread-safe counter
count.Add(1)            // Safe from multiple goroutines
count.Load()            // Safe read
```
Regular `int` + `count++` = data race = `go test -race` fails.

### Pitfall 3: Buffered channels for test synchronization
```go
received := make(chan event.Event, 1)  // buffer=1 prevents goroutine leak
```
If handler sends to unbuffered channel and nobody reads → goroutine hangs forever.

### Pitfall 4: defer unsub() prevents leak
Every `Subscribe()` in test MUST have `defer unsub()`.
Otherwise: subscription stays alive → handler fires on future events → test contamination.

## Verify
```bash
go test -v -race -count=1 ./kernel/eventbus/...
# Expected: ALL PASS, ≥ 16 test functions
# -race: detects data races in concurrent tests
# -count=1: disables test caching
```

## Checklist
- [ ] File `kernel/eventbus/bus_test.go` exists
- [ ] Package: `eventbus_test`
- [ ] ≥ 16 test functions
- [ ] Tests: exact match, no match, wildcard, global wildcard
- [ ] Tests: unsubscribe stops delivery, idempotent unsubscribe
- [ ] Tests: subscriber count tracking
- [ ] Tests: concurrent publish/subscribe (10 publishers × 100 events)
- [ ] Tests: concurrent subscribe/unsubscribe (100 goroutines)
- [ ] Tests: handler panic does NOT crash bus
- [ ] Tests: Close rejects new publishes
- [ ] Tests: Close waits for in-flight handlers
- [ ] Tests: context cancellation
- [ ] Tests: input validation (empty pattern, nil handler, invalid pattern)
- [ ] Tests: segment count mismatch
- [ ] Tests: multiple subscribers all receive
- [ ] Uses `atomic.Int32` for concurrent counters
- [ ] Uses buffered channels for synchronization
- [ ] Uses `time.After()` timeouts (not bare `time.Sleep`)
- [ ] `go test -v -race ./kernel/eventbus/...` ALL PASS
