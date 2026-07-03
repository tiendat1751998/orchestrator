# Micro-Task 2.17: Create kernel/eventbus/bus_test.go

## Info
- **File**: `kernel/eventbus/bus_test.go`
- **Package**: `eventbus_test`
- **Depends on**: 2.12-2.16
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/eventbus/...`

## Purpose
Implements integration and concurrent safety unit tests for the EventBus. It verifies topic matching, multi-goroutine subscription allocations, panic isolation boundaries, and channel shutdown routines.

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

	received := make(chan event.Event, 1)

	unsub, err := bus.Subscribe("task.started", func(evt event.Event) {
		received <- evt
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer unsub()

	err = bus.Publish(context.Background(), testEvent("task.started"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

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

	bus.Publish(context.Background(), testEvent("task.completed"))

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

	bus.Publish(context.Background(), testEvent("task.started"))
	bus.Publish(context.Background(), testEvent("task.completed"))
	bus.Publish(context.Background(), testEvent("task.failed"))
	bus.Publish(context.Background(), testEvent("mission.started"))

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

	bus.Publish(context.Background(), testEvent("task.started"))
	time.Sleep(200 * time.Millisecond)

	if count.Load() != 1 {
		t.Fatalf("expected 1 call before unsubscribe, got %d", count.Load())
	}

	unsub()

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
// Concurrent Safety
// =============================================================================

func TestBus_ConcurrentPublishSubscribe(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var wg sync.WaitGroup
	var totalReceived atomic.Int64

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

	got := totalReceived.Load()
	if got != 1000 {
		t.Errorf("concurrent: got %d deliveries, want 1000", got)
	}
}

func TestBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var wg sync.WaitGroup

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
}

// =============================================================================
// Panic Recovery
// =============================================================================

func TestBus_HandlerPanic_DoesNotCrash(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	panicDone := make(chan struct{}, 1)
	normalDone := make(chan struct{}, 1)

	unsub1, _ := bus.Subscribe("task.started", func(evt event.Event) {
		defer func() { panicDone <- struct{}{} }()
		panic("intentional panic in test")
	})
	defer unsub1()

	unsub2, _ := bus.Subscribe("task.started", func(evt event.Event) {
		normalDone <- struct{}{}
	})
	defer unsub2()

	bus.Publish(context.Background(), testEvent("task.started"))

	select {
	case <-normalDone:
		// OK
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
		time.Sleep(500 * time.Millisecond)
		close(handlerDone)
	})
	defer unsub()

	bus.Publish(context.Background(), testEvent("task.started"))

	<-handlerStarted

	closeDone := make(chan struct{})
	go func() {
		bus.Close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
		select {
		case <-handlerDone:
			// OK
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
	cancel()

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

	unsub, _ := bus.Subscribe("task.*", func(evt event.Event) {
		received <- struct{}{}
	})
	defer unsub()

	bus.Publish(context.Background(), testEvent("task"))

	select {
	case <-received:
		t.Error("task.* should NOT match 'task' (different segment count)")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}
}

// =============================================================================
// Multiple Subscribers
// =============================================================================

func TestBus_MultipleSubscribers_AllReceive(t *testing.T) {
	bus := eventbus.New(nil)
	defer bus.Close()

	var count atomic.Int32

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

## Rules
1. **Race Detector Compatibility**: All counters matching state changes within concurrent test loops must use standard atomic types (`atomic.Int32` / `atomic.Int64`) to pass `go test -race` checks.
2. **Subscription Lifecycle Cleaning**: Subscribe calls inside tests must use deferred unsubscribe invocations (`defer unsub()`) to prevent handlers from leaking across multiple test cases.
3. **Synchronization Channels**: Coordinate test expectations using channels and select statements with timeout limits (`time.After`) rather than using arbitrary sleep durations.

## ⚠️ Pitfalls

### Pitfall 1: Leaving dangling event subscriptions active between test runs
Failing to clean up subscriptions using `unsub()` leaves handlers bound to the event bus memory, triggering test pollution side effects when subsequent tests publish events. Always call `defer unsub()`.

### Pitfall 2: Using unbuffered channels to sync worker threads
If a handler writes to an unbuffered synchronization channel but the main test thread fails to read it (for example, if another assertion fails early), the worker thread hangs permanently. Always size synchronization channels to at least `buffer=1`.

## Verify
```bash
go test -v -race -count=1 ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/bus_test.go` exists
- [ ] Package: `eventbus_test` (external testing package)
- [ ] At least 16 test functions are implemented
- [ ] Tests verify exact, wildcard, and global wildcard mappings
- [ ] Concurrent publications tests (10 writers x 100 events) run cleanly under `-race` checks
- [ ] Concurrent subscription allocation tests execute safely
- [ ] Panic recovery checks verify that bad handlers do not crash the bus
- [ ] Shutdown testing verifies that `Close` waits for in-flight tasks
- [ ] Channels are buffered to prevent thread leak hangs
- [ ] `go test -v -race ./kernel/eventbus/...` passes
