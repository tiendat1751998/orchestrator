# Micro-Task 5.18: Create kernel/resilience/timeout.go

## Info
- **File**: `kernel/resilience/timeout.go`
- **Package**: `resilience`
- **Depends on**: 5.17
- **Time**: 15 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements the cascading timeout context calculator (`CascadingTimeoutContext`) to adjust sub-task timeout budgets based on the remaining mission duration.

## EXACT code to create

```go
package resilience

import (
	"context"
	"time"
)

// CascadingTimeoutContext derives a sub-context with a timeout bounded by the parent context's deadline.
//
// Parameters:
//   - parent: parent context tracking mission deadline.
//   - desiredTaskTimeout: default timeout limit allocated to task.
//
// Returns derived context and cancel callback.
func CascadingTimeoutContext(parent context.Context, desiredTaskTimeout time.Duration) (context.Context, context.CancelFunc) {
	deadline, hasDeadline := parent.Deadline()
	if !hasDeadline {
		// Parent has no deadline: apply default task timeout
		return context.WithTimeout(parent, desiredTaskTimeout)
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		// Parent has already timed out
		ctx, cancel := context.WithCancel(parent)
		cancel() // Cancel immediately
		return ctx, cancel
	}

	// Task timeout is capped by remaining parent duration
	actualTimeout := desiredTaskTimeout
	if remaining < desiredTaskTimeout {
		actualTimeout = remaining
	}

	return context.WithTimeout(parent, actualTimeout)
}
```

## Pitfalls

### Pitfall 1: Allocating budgets exceeding parent limits
```go
// WRONG:
ctx, cancel := context.WithTimeout(parent, 5*time.Minute) // If parent has 2 minutes remaining, task gets cut off early by parent anyway but task context thinks it has 5 minutes!
```
Failing to cap child timeouts to parent remaining times leaves workers running tasks after the parent context has canceled. Ensure child timeouts are capped.

### Pitfall 2: Leaking context resources on successful execution
Failing to call the returned `CancelFunc` leaves context timer resources allocated, causing memory leaks. Always call `defer cancel()`.

## Verify
```bash
go build ./kernel/resilience/...
# Expected: clean compilation without errors
```

## Checklist
- [x] File exists at `kernel/resilience/timeout.go`
- [x] Package name is `resilience`
- [x] All exported types have Godoc
- [x] Cascading timeouts are derived correctly
- [x] Cancel callbacks are returned along with contexts
- [x] Build command passes
