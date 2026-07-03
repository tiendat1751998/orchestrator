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
	dlq         *DeadLetterQueue

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
		dlq:         NewDeadLetterQueue(100),
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
	safe := safeHandler(handler, b.dlq, b.logger)

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

func (b *Bus) DLQ() *DeadLetterQueue {
	return b.dlq
}
