# Micro-Task 2.14: Create kernel/eventbus/subscriber.go

## Info
- **File**: `kernel/eventbus/subscriber.go`
- **Package**: `eventbus`
- **Depends on**: 2.12 (types.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Implements subscriber utilities (`safeHandler`, `makeUnsubscribe`, `validatePattern`) that handle event patterns validation, panic recovery wrapper setups, and cleanup functions.

## EXACT code to create

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
```

## Rules
1. **Panic Recovery Safeguards**: All subscriber execution callbacks must be wrapped inside deferred recovery blocks (`safeHandler`). Recovered panics must be logged to standard diagnostics without terminating the application process.
2. **Idempotence**: Returned unsubscribe closures must be idempotent, allowing clients to invoke them multiple times safely.
3. **Regex Pattern Validations**: Enforce format checks on target paths (e.g. rejecting leading/trailing dots or empty paths).

## ⚠️ Pitfalls

### Pitfall 1: Crashing runtime processes from unhandled subscriber panics
If a custom event subscriber panics (e.g. nil pointer dereferences or map access conflicts), it will crash the system if the event dispatcher routine does not run recovery checks. Enforce `safeHandler` wrapping globally.

### Pitfall 2: Re-locking maps inside recursive unsubscribe calls
Calling unsubscribe must not dead-lock mapping components. Since `deactivate` uses atomic booleans and `sm.remove` synchronizes inside its lock bounds, unsubscribe functions compile and run safely.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/subscriber.go` exists
- [ ] Package: `eventbus`
- [ ] `safeHandler` wraps handlers in defer/recover statements
- [ ] Recovered panics write warning details to log outputs
- [ ] `makeUnsubscribe` returns a callable closure function
- [ ] Unsubscribe closures call `deactivate` and `remove`
- [ ] `validatePattern` catches empty string inputs, leading dots, and trailing dots
- [ ] `go build ./kernel/eventbus/...` passes
