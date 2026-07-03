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
