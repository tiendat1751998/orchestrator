# Micro-Task 2.15: Tạo kernel/eventbus/bus.go

## Thông tin
- **File tạo**: `kernel/eventbus/bus.go`
- **Package**: `eventbus`
- **Dependencies trước**: 2.12 (types.go), 2.13 (matcher.go), 2.14 (subscriber.go)
- **Thời gian**: 25 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Mục đích
Core EventBus implementation. Thread-safe, async publish/subscribe.
Implements `event.Bus` contract từ Phase 1.

## Nội dung CHÍNH XÁC cần tạo

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
	//
	// WHY sync.WaitGroup?
	// → When shutting down, we need to wait for all running handlers to finish.
	// → Without WaitGroup, Close() returns immediately → handlers may access
	//   freed resources → crash.
	wg sync.WaitGroup

	// closed prevents new publishes after Close().
	// Using atomic bool for lock-free checking in hot path.
	closed bool
	mu     sync.RWMutex // Protects 'closed' field
}

// Compile-time check: Bus implements event.Bus
var _ event.Bus = (*Bus)(nil)

// New creates a new EventBus.
//
// Parameters:
//   - logger: for logging handler panics and errors. Can be nil (errors silently ignored).
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
//
// Usage:
//
//	unsub, err := bus.Subscribe("task.*", func(evt event.Event) {
//	    log.Info("task event", "type", evt.Type)
//	})
//	if err != nil { ... }
//	defer unsub() // Clean up when done
//
// Thread-safety: safe for concurrent use.
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
// Useful for debugging and metrics.
func (b *Bus) SubscriberCount() int {
	return b.subscribers.count()
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: wg.Add(1) TRƯỚC go func
```go
// ❌ SAI:
go func() {
    b.wg.Add(1)    // May execute AFTER Close() calls wg.Wait()
    defer b.wg.Done()
    ...
}()

// ✅ ĐÚNG:
b.wg.Add(1)        // Add BEFORE launching goroutine
go func() {
    defer b.wg.Done()
    ...
}()
```
Nếu `wg.Add(1)` bên trong goroutine → `Close()` gọi `wg.Wait()` trước khi `Add(1)` → Wait returns sớm → handler truy cập freed resources.

### Pitfall 2: Double-check active status
```go
// After matching(), subscriber may have been unsubscribed
// Check again inside the goroutine
if !s.isActive() {
    return
}
```
Race window: matching() returns sub → unsub() called → goroutine starts → handler fires on unsubscribed subscription.

### Pitfall 3: Compile-time interface check
```go
var _ event.Bus = (*Bus)(nil)
```
Nếu Bus thiếu method → compiler error TẠI đây (not at runtime).

### Pitfall 4: Context cancellation check
```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```
Non-blocking check. Nếu ctx đã cancelled → không publish. Nếu chưa → continue.

### Pitfall 5: Closed bus error
Publish() sau Close() → error. KHÔNG silent drop.
Caller needs to know the event was not delivered.

## Checklist
- [ ] File `kernel/eventbus/bus.go` tồn tại
- [ ] Bus struct implements event.Bus (compile-time check)
- [ ] `New(logger)` constructor
- [ ] `Publish()` — async dispatch, context check, closed check
- [ ] `Subscribe()` — pattern validation, nil handler check, panic recovery wrap
- [ ] `Close()` — sets closed=true, waits for wg
- [ ] `SubscriberCount()` for debugging
- [ ] wg.Add(1) BEFORE go func() (not inside)
- [ ] Double-check active status in goroutine
- [ ] Thread-safe: RWMutex for closed flag
- [ ] Godoc comments
- [ ] `go build ./kernel/eventbus/...` không lỗi
