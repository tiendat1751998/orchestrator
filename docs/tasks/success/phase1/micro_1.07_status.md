# Micro-Task 1.07: Create contracts/status.go

## Info
- **File**: `contracts/status.go`
- **Package**: `contracts`
- **Depends on**: 1.05
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Defines execution states (e.g. `pending`, `running`, `success`) as string constants to allow clean serializations, logs, and state transition checks in the orchestrator pipeline.

## EXACT code to create

```go
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
```

## ⚠️ Pitfalls

### Pitfall 1: Using `iota` integers for Status enums
```go
const (
    StatusPending Status = "pending" // Serializes as readable strings.
)
```
Using strings makes database records, JSON APIs, and debug log files much easier to read and troubleshoot.

### Pitfall 2: Forgetting to update Terminal checks on new Status additions
If you introduce a new terminal state (e.g. `skipped`) but forget to add it to the `IsTerminal` switch, the scheduler will block indefinitely waiting for a transition that will never happen.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/status.go` exists
- [ ] Package: `contracts`
- [ ] Status constants are declared as string types
- [ ] `IsTerminal()` evaluates terminal flags correctly
- [ ] `IsValidStatus` checks values against the list of valid entries
- [ ] `go build ./contracts/...` passes
