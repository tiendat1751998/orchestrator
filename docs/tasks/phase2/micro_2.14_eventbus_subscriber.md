# Micro-Task 2.14: Tạo kernel/eventbus/subscriber.go

## Thông tin
- **File tạo**: `kernel/eventbus/subscriber.go`
- **Package**: `eventbus`
- **Dependencies trước**: 2.12 (types.go)
- **Thời gian**: 15 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Mục đích
Subscriber management: tạo, quản lý, và cleanup subscriptions.

## Nội dung CHÍNH XÁC cần tạo

```go
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
func safeHandler(handler func(event.Event), logger *slog.Logger) func(event.Event) {
	return func(evt event.Event) {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					logger.Error("event handler panicked",
						"event_type", evt.Type,
						"event_source", evt.Source,
						"panic", fmt.Sprintf("%v", r),
					)
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
//
// WHY sync.Once instead of just checking a bool?
// → Multiple goroutines might call unsub() concurrently.
// → sync.Once guarantees exactly-once execution, thread-safe.
func makeUnsubscribe(sm *subscriberMap, sub *subscription) func() {
	// sync.Once is NOT used here because subscription.deactivate() is already
	// idempotent (atomic.Store). The remove() operation is also safe to call
	// multiple times (no-op if subscription is already removed).
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
//   - "task.*.deep.*" (OK actually — multi-segment wildcard)
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
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Panic recovery trong handler
```go
defer func() {
    if r := recover(); r != nil {
        logger.Error("handler panicked", ...)
    }
}()
```
PHẢI recover. Một handler lỗi KHÔNG được crash toàn bộ event bus.

### Pitfall 2: Unsubscribe idempotent
User gọi `unsub()` 2 lần → không panic, không error.
`atomic.Store(false)` idempotent. `remove()` idempotent (no-op if not found).

## Checklist
- [ ] File `kernel/eventbus/subscriber.go` tồn tại
- [ ] `safeHandler()` — wraps handler with panic recovery
- [ ] `makeUnsubscribe()` — returns idempotent unsubscribe function
- [ ] `validatePattern()` — validates event pattern
- [ ] Panic recovery logs error (does NOT re-panic)
- [ ] `go build ./kernel/eventbus/...` không lỗi
