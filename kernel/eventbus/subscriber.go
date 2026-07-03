package eventbus

import (
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// safeHandler wraps a handler function with panic recovery.
//
// WHY?
// → If a handler panics, it would crash the goroutine running it.
// → Without recovery, the panic propagates up and crashes the process.
// → With recovery, the panic is logged and the bus continues operating.
// → This is critical: ONE bad handler should NOT bring down the entire system.
func safeHandler(handler func(event.Event), dlq *DeadLetterQueue, logger *slog.Logger) func(event.Event) {
	return func(evt event.Event) {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("%v", r)
				if logger != nil {
					logger.Error("event handler panicked",
						"event_type", evt.Type,
						"event_source", evt.Source,
						"panic", errStr,
					)
				}
				if dlq != nil {
					dlq.Add(evt, errStr)
				}
			}
		}()
		handler(evt)
	}
}

// makeUnsubscribe creates an idempotent unsubscribe function.
//
// The returned function:
//   - First call: deactivates the subscription and removes from map
//   - Subsequent calls: no-op (safe to call multiple times)
//
// Pattern: defer unsub() — cleanup guaranteed even on panic.
func makeUnsubscribe(sm *subscriberMap, sub *subscription) func() {
	return func() {
		sub.deactivate()
		sm.remove(sub.id)
	}
}

// validatePattern checks if a subscription pattern is valid.
//
// Valid patterns:
//   - "*"             (global wildcard)
//   - "task.started"  (exact type)
//   - "task.*"        (wildcard segment)
//   - "*.started"     (prefix wildcard)
//
// Invalid patterns:
//   - ""              (empty)
//   - "."             (no segments)
//   - ".started"      (starts with dot)
//   - "task."         (ends with dot)
func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("eventbus: empty pattern")
	}
	if pattern == "." {
		return fmt.Errorf("eventbus: invalid pattern %q (no segments)", pattern)
	}
	if len(pattern) > 1 && (pattern[0] == '.' || pattern[len(pattern)-1] == '.') {
		return fmt.Errorf("eventbus: invalid pattern %q (starts or ends with dot)", pattern)
	}
	return nil
}
