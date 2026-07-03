package contracts

// Status represents the current state of a task, mission, or agent.
type Status string

const (
	// StatusPending indicates the item is queued but not yet started.
	StatusPending Status = "pending"

	// StatusRunning indicates the item is currently being executed.
	StatusRunning Status = "running"

	// StatusSuccess indicates the item completed successfully.
	StatusSuccess Status = "success"

	// StatusFailed indicates the item completed with an error.
	StatusFailed Status = "failed"

	// StatusCancelled indicates the item was cancelled before completion
	// (e.g., user pressed Ctrl+C, or parent context was cancelled).
	StatusCancelled Status = "cancelled"

	// StatusRetrying indicates the item failed and is being retried.
	StatusRetrying Status = "retrying"

	// StatusSkipped indicates the item was skipped
	// (e.g., a dependency failed and the item cannot run).
	StatusSkipped Status = "skipped"

	// StatusTimeout indicates the item exceeded its time limit.
	StatusTimeout Status = "timeout"
)

// IsTerminal returns true if the status represents a final state.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled, StatusSkipped, StatusTimeout:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the status indicates successful completion.
func (s Status) IsSuccess() bool {
	return s == StatusSuccess
}

// IsFailed returns true if the status indicates a failure.
func (s Status) IsFailed() bool {
	return s == StatusFailed || s == StatusTimeout
}

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// ValidStatuses returns all valid status values.
func ValidStatuses() []Status {
	return []Status{
		StatusPending,
		StatusRunning,
		StatusSuccess,
		StatusFailed,
		StatusCancelled,
		StatusRetrying,
		StatusSkipped,
		StatusTimeout,
	}
}

// IsValidStatus checks if a string is a valid Status value.
func IsValidStatus(s string) bool {
	for _, valid := range ValidStatuses() {
		if string(valid) == s {
			return true
		}
	}
	return false
}
