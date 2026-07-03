// Package contracts defines shared interfaces, types, and errors
// used across all orchestrator components.
package contracts

import (
	"errors"
	"fmt"
	"time"
)

// =============================================================================
// Provider Errors
// =============================================================================

// ErrProviderUnavailable indicates the provider is not reachable or not configured.
// Example: Antigravity CLI binary not found, API endpoint down.
var ErrProviderUnavailable = errors.New("orchestrator: provider unavailable")

// ErrProviderTimeout indicates the provider did not respond within the configured timeout.
// This is a retryable error — the orchestrator should retry with exponential backoff.
var ErrProviderTimeout = errors.New("orchestrator: provider timeout")

// ErrProviderRateLimited indicates the provider returned a rate limit error (HTTP 429).
// This is a retryable error — the orchestrator should wait before retrying.
var ErrProviderRateLimited = errors.New("orchestrator: provider rate limited")

// ErrProviderAuthFailed indicates invalid API key or credentials.
// This is NOT retryable — requires user to fix configuration.
var ErrProviderAuthFailed = errors.New("orchestrator: provider authentication failed")

// =============================================================================
// Agent Errors
// =============================================================================

// ErrAgentBusy indicates the agent is currently executing another task.
var ErrAgentBusy = errors.New("orchestrator: agent is busy")

// ErrAgentNotFound indicates no agent with the given name is registered.
var ErrAgentNotFound = errors.New("orchestrator: agent not found")

// ErrNoAgentAvailable indicates no registered agent can handle the given task.
var ErrNoAgentAvailable = errors.New("orchestrator: no agent available for task")

// =============================================================================
// Task Errors
// =============================================================================

// ErrTaskCancelled indicates the task was cancelled via context cancellation.
var ErrTaskCancelled = errors.New("orchestrator: task cancelled")

// ErrTaskTimeout indicates the task exceeded its configured timeout.
var ErrTaskTimeout = errors.New("orchestrator: task timeout")

// ErrTaskFailed indicates the task completed but with a failure result.
// Check the task Result for details.
var ErrTaskFailed = errors.New("orchestrator: task failed")

// ErrTaskDependencyFailed indicates a dependency task failed,
// so this task cannot be executed.
var ErrTaskDependencyFailed = errors.New("orchestrator: task dependency failed")

// =============================================================================
// Config Errors
// =============================================================================

// ErrInvalidConfig indicates the configuration is invalid or incomplete.
var ErrInvalidConfig = errors.New("orchestrator: invalid configuration")

// ErrConfigNotFound indicates the configuration file was not found.
var ErrConfigNotFound = errors.New("orchestrator: configuration file not found")

// =============================================================================
// Plugin Errors
// =============================================================================

// ErrPluginNotFound indicates no plugin with the given name is registered.
var ErrPluginNotFound = errors.New("orchestrator: plugin not found")

// ErrPluginAlreadyRegistered indicates a plugin with the same name already exists.
var ErrPluginAlreadyRegistered = errors.New("orchestrator: plugin already registered")

// ErrPluginInitFailed indicates the plugin failed to initialize.
var ErrPluginInitFailed = errors.New("orchestrator: plugin initialization failed")

// =============================================================================
// Security Errors
// =============================================================================

// ErrPermissionDenied indicates the agent does not have permission for the action.
var ErrPermissionDenied = errors.New("orchestrator: permission denied")

// ErrBlockedCommand indicates the command is on the security blocklist.
var ErrBlockedCommand = errors.New("orchestrator: command blocked by security policy")

// =============================================================================
// Orchestrator Errors
// =============================================================================

// ErrMissionFailed indicates the mission could not be completed after all retries.
var ErrMissionFailed = errors.New("orchestrator: mission failed")

// ErrCircularDependency indicates tasks have circular dependencies.
var ErrCircularDependency = errors.New("orchestrator: circular dependency detected")

// ErrMaxRetriesExceeded indicates the maximum number of retry attempts was reached.
var ErrMaxRetriesExceeded = errors.New("orchestrator: max retries exceeded")

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
