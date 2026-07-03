# Micro-Task 1.37: Update contracts/errors.go (Structured Error Types)

## Info
- **File**: `contracts/errors.go` (Update existing file to add structured errors)
- **Package**: `contracts`
- **Depends on**: 1.05
- **Time**: 20 min
- **Verify**: `go build ./contracts/...`

## Purpose
Enhances the system's error handling by adding structured error types (`ValidationError`, `NotFoundError`, `TimeoutError`, `ConflictError`, `PermissionError`, and `RetryableError`) alongside sentinel errors. This allows calling components (such as HTTP APIs, schedulers, and resilience policies) to inspect error details programmatically instead of relying on string comparisons.

## EXACT code to create

Append these declarations to the end of [contracts/errors.go](file:///d:/project/orchestrator/contracts/errors.go) (retaining all the sentinel error definitions from Micro-Task 1.05):

```go
import (
	"errors"
	"fmt"
	"time"
)

// =============================================================================
// Common Structured Errors
// =============================================================================

// ValidationError indicates a field has failed input validation rules.
type ValidationError struct {
	Component string            `json:"component"`          // e.g. "config", "task", "request"
	Field     string            `json:"field"`              // e.g. "max_concurrent_tasks", "timeout"
	Reason    string            `json:"reason"`             // e.g. "must be positive", "required"
	Metadata  map[string]string `json:"metadata,omitempty"` // Additional context details
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] validation failed on field %q: %s", e.Component, e.Field, e.Reason)
}

// NewValidationError constructs a new ValidationError.
func NewValidationError(component, field, reason string) error {
	return &ValidationError{
		Component: component,
		Field:     field,
		Reason:    reason,
	}
}

// NotFoundError indicates a required resource could not be found.
type NotFoundError struct {
	Resource string `json:"resource"` // e.g. "agent", "provider", "task"
	Name     string `json:"name"`     // e.g. "reviewer-agent", "gemini-flash"
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource %s %q not found", e.Resource, e.Name)
}

// NewNotFoundError constructs a new NotFoundError.
func NewNotFoundError(resource, name string) error {
	return &NotFoundError{
		Resource: resource,
		Name:     name,
	}
}

// TimeoutError represents a timeout condition with duration details.
type TimeoutError struct {
	Operation string        `json:"operation"` // e.g. "execute_task", "provider_call"
	Duration  time.Duration `json:"duration"`  // The configured timeout duration
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation %q timed out after %s", e.Operation, e.Duration)
}

// NewTimeoutError constructs a new TimeoutError.
func NewTimeoutError(operation string, duration time.Duration) error {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
	}
}

// ConflictError represents a state conflict, e.g. duplicate registrations.
type ConflictError struct {
	Resource string `json:"resource"`
	Key      string `json:"key"`
	Message  string `json:"message"`
}

// Error implements the error interface.
func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict on %s %q: %s", e.Resource, e.Key, e.Message)
}

// NewConflictError constructs a new ConflictError.
func NewConflictError(resource, key, message string) error {
	return &ConflictError{
		Resource: resource,
		Key:      key,
		Message:  message,
	}
}

// PermissionError indicates a security violation.
type PermissionError struct {
	Actor    string `json:"actor"`    // e.g. "agent-coder"
	Action   string `json:"action"`   // e.g. "run_command"
	Resource string `json:"resource"` // e.g. "rm -rf /"
}

// Error implements the error interface.
func (e *PermissionError) Error() string {
	return fmt.Sprintf("actor %q denied permission to perform action %q on resource %q", e.Actor, e.Action, e.Resource)
}

// NewPermissionError constructs a new PermissionError.
func NewPermissionError(actor, action, resource string) error {
	return &PermissionError{
		Actor:    actor,
		Action:   action,
		Resource: resource,
	}
}

// =============================================================================
// Resilience & Retry Errors
// =============================================================================

// RetryableError wraps another error to explicitly signal that retrying
// the operation is safe and recommended.
type RetryableError struct {
	Err        error         // The underlying transient error
	RetryAfter time.Duration // Recommended backoff wait
}

// Error implements the error interface.
func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v (retry after %s)", e.Err, e.RetryAfter)
}

// Unwrap returns the underlying error for errors.Unwrap support.
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError wraps an error as retryable.
func NewRetryableError(err error, retryAfter time.Duration) error {
	if err == nil {
		return nil
	}
	return &RetryableError{
		Err:        err,
		RetryAfter: retryAfter,
	}
}

// IsRetryable checks if an error has been marked as retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var rErr *RetryableError
	return errors.As(err, &rErr)
}

// =============================================================================
// Helper Wrappers for standard errors package (convenience)
// =============================================================================

// Is is a proxy to errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As is a proxy to errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Unwrap is a proxy to errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
```

## Rules
1. **Unwrap Support**: Any custom error wrapper (like `RetryableError`) must implement the `Unwrap() error` interface method to enable recursive checking via `errors.Is` and `errors.As`.
2. **Interface Nil Pollution Prevention**: Always return a raw `nil` interface value when validation succeeds, instead of returning typed pointers set to nil.
3. **Proxy Wrappers**: Exported helpers like `Is`, `As`, and `Unwrap` act as direct proxies to the standard Go library `errors` package, preventing package naming conflicts when imported inside user files.

## Pitfalls

### Pitfall 1: Returning concrete type nil pointers through interface values
```go
// WRONG:
func CheckConfig(c *Config) error {
    var valErr *ValidationError = nil
    if c.Path == "" {
        valErr = &ValidationError{Field: "path", Reason: "empty"}
    }
    return valErr // If c.Path is valid, returns a non-nil interface error value! (type is ValidationError, value is nil).
}

// CORRECT:
func CheckConfig(c *Config) error {
    if c.Path == "" {
        return &ValidationError{Field: "path", Reason: "empty"}
    }
    return nil // Returns clean, type-less nil interface.
}
```
Always return literal `nil` when returning successful outcomes from functions returning Go `error` interfaces.

### Pitfall 2: Omiting `Unwrap()` in wrapper errors
If a custom error type wraps an underlying error (e.g. `RetryableError`) but fails to implement `Unwrap()`, functions like `errors.Is(err, ErrProviderTimeout)` will evaluate to false because the compiler cannot crawl down the error chain.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/errors.go` is updated successfully
- [ ] Package: `contracts`
- [ ] `ValidationError` and `NewValidationError` helper are defined
- [ ] `NotFoundError` and `NewNotFoundError` helper are defined
- [ ] `TimeoutError` and `NewTimeoutError` helper are defined
- [ ] `ConflictError` and `NewConflictError` helper are defined
- [ ] `PermissionError` and `NewPermissionError` helper are defined
- [ ] `RetryableError` is declared and implements `Unwrap() error` method
- [ ] `IsRetryable` checks error wrap status recursively using `As`
- [ ] Proxy helper functions `Is`, `As`, and `Unwrap` compile cleanly
- [ ] `go build ./contracts/...` passes
