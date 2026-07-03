# Micro-Task 2.15: Create kernel/eventbus/bus.go

## Info
- **File**: `kernel/eventbus/bus.go`
- **Package**: `eventbus`
- **Depends on**: 2.12 (types.go), 2.13 (matcher.go), 2.14 (subscriber.go)
- **Time**: 25 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Implements the core EventBus system (`Bus` and constructors) that conforms to the `event.Bus` contract, supporting asynchronous event publication, handler panic recovery, and safe shut down procedures.

## EXACT code to create

```go
package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// Bus is the in-memory event bus implementation.
//
// It implements the event.Bus contract from contracts/event.
//
// Architecture:
//   - Publish() fans out events to all matching subscribers asynchronously
//   - Each subscriber handler runs in its own goroutine
//   - Handlers are wrapped with panic recovery (see safeHandler)
//   - Subscriber management uses RWMutex for concurrent publish/subscribe
//
// Thread-safety:
//   - Subscribe/Unsubscribe: acquire write lock (exclusive)
//   - Publish: acquires read lock (shared with other publishers)
//   - This means multiple Publish() calls can run concurrently
//
// Lifecycle:
//   - New() creates the bus
//   - Subscribe/Publish operate normally
//   - Close() stops accepting new events and waits for in-flight handlers
type Bus struct {
	subscribers *subscriberMap
	logger      *slog.Logger

	// wg tracks in-flight handler goroutines.
	// Close() waits for all handlers to complete.
	wg sync.WaitGroup

	// closed prevents new publishes after Close().
	closed bool
	mu     sync.RWMutex // Protects 'closed' field
}

// Compile-time check: Bus implements event.Bus
var _ event.Bus = (*Bus)(nil)

// New creates a new EventBus.
//
// Parameters:
//   - logger: for logging handler panics and errors. Can be nil.
func New(logger *slog.Logger) *Bus {
	return &Bus{
		subscribers: newSubscriberMap(),
		logger:      logger,
	}
}

// Publish emits an event to all matching subscribers.
//
// This method is ASYNCHRONOUS: it returns immediately after dispatching
// events to handler goroutines. It does NOT wait for handlers to complete.
//
// If the bus is closed, returns an error.
//
// Thread-safety: safe for concurrent use. Multiple goroutines can Publish simultaneously.
func (b *Bus) Publish(ctx context.Context, evt event.Event) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if bus is closed
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("eventbus: bus is closed")
	}
	b.mu.RUnlock()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find all matching subscribers
	matches := b.subscribers.matching(evt.Type)

	// Dispatch to each subscriber in a separate goroutine
	for _, sub := range matches {
		if !sub.isActive() {
			continue // Skip unsubscribed handlers
		}

		b.wg.Add(1)
		go func(s *subscription) {
			defer b.wg.Done()

			// Double-check active status (may have been unsubscribed after matching)
			if !s.isActive() {
				return
			}

			// Handler is already wrapped with panic recovery (safeHandler)
			s.handler(evt)
		}(sub)
	}

	return nil
}

// Subscribe registers a handler for events matching the given pattern.
//
// Pattern matching rules (see matcher.go):
//   - "*"           → matches all events
//   - "task.*"      → matches "task.started", "task.completed", etc.
//   - "task.started" → matches only "task.started"
//
// Returns an unsubscribe function. MUST be called when done to prevent memory leaks.
func (b *Bus) Subscribe(pattern string, handler func(event.Event)) (func(), error) {
	// Validate pattern
	if err := validatePattern(pattern); err != nil {
		return nil, err
	}

	// Validate handler
	if handler == nil {
		return nil, fmt.Errorf("eventbus: handler cannot be nil")
	}

	// Wrap handler with panic recovery
	safe := safeHandler(handler, b.logger)

	// Register subscription
	sub := b.subscribers.add(pattern, safe)

	// Create unsubscribe function
	unsub := makeUnsubscribe(b.subscribers, sub)

	if b.logger != nil {
		b.logger.Debug("subscription added",
			"pattern", pattern,
			"subscriber_id", sub.id,
			"total_subscribers", b.subscribers.count(),
		)
	}

	return unsub, nil
}

// Close stops the bus from accepting new events and waits for
// all in-flight handlers to complete.
//
// After Close():
//   - Publish() returns an error
//   - Subscribe() still works (but handlers won't fire for new events)
//
// This should be called during graceful shutdown.
func (b *Bus) Close() {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()

	// Wait for all in-flight handlers to complete
	b.wg.Wait()

	if b.logger != nil {
		b.logger.Info("eventbus closed")
	}
}

// SubscriberCount returns the number of active subscribers.
// Used for debugging and metrics.
func (b *Bus) SubscriberCount() int {
	return b.subscribers.count()
}
```

## Rules
1. **WaitGroup Ordering**: Always execute `wg.Add(1)` *before* starting the goroutine (`go func()`). Adding counts inside the spawned thread creates race conditions where the supervisor can call `Wait()` and return before the goroutine starts.
2. **Double-Check Active States**: The worker thread must check `isActive()` *inside* the goroutine again. This protects against executing handlers for subscriptions that were unsubscribed during the scheduling gap.
3. **Graceful Shutdown**: The `Close` method must mark the bus as closed to block new events, and call `wg.Wait()` to block execution until all in-flight subscriber handlers have run to completion.
4. **Context Safety**: Since event publication helpers frequently pass a `nil` context value (for fire-and-forget logging), the `Publish` implementation must explicitly fall back to `context.Background()` if `ctx == nil` is passed, avoiding nil pointer dereference panics.

## ⚠️ Pitfalls

### Pitfall 1: Calling wg.Add(1) inside spawned threads
Always call `wg.Add` synchronously before launching threads to guarantee tracking.

### Pitfall 2: Silent event drops after bus closures
When publishers send events to a closed bus, returning a `nil` error masks the failure. Always return a clear error (e.g. `eventbus: bus is closed`) so the calling application knows the delivery failed.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/bus.go` exists
- [ ] Package: `eventbus`
- [ ] `Bus` struct implements the `event.Bus` interface
- [ ] `Publish` checks if closed, and checks for context cancellation
- [ ] Goroutines call `wg.Add(1)` on the dispatcher thread
- [ ] Spawners check `isActive()` inside worker threads before invoking handlers
- [ ] `Subscribe` validates pattern layouts and rejects `nil` handlers
- [ ] `Close` stops publications and waits for `wg.Wait()`
- [ ] Context parameter handles `nil` checks by assigning `context.Background()`
- [ ] `go build ./kernel/eventbus/...` passes
