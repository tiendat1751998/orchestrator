# Micro-Task 1.37: Tạo contracts/errors.go (Cập nhật Structured Error Types)

## Thông tin
- **File cập nhật**: `contracts/errors.go` (Ghi đè/cập nhật bản cũ để thêm structured errors)
- **Package**: `contracts`
- **Dependencies trước**: 1.05 (errors.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./contracts/...`

## Purpose
Nâng cấp hệ thống lỗi từ chuỗi thuần túy (sentinel errors) sang các kiểu lỗi có cấu trúc (Structured Error Types). Điều này cho phép caller (như HTTP API, scheduler, resilience policies) nhận diện chính xác nguyên nhân lỗi theo chương trình (programmatically) thay vì so sánh chuỗi thô.

## EXACT code to create

```go
package contracts

import (
	"fmt"
	"time"
)

// =============================================================================
// Common Structured Errors
// =============================================================================

// ValidationError indicates a field has failed input validation rules.
// Used at HTTP boundaries, config loading, and contract input checks.
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
// Triggers HTTP 404 at API level.
type NotFoundError struct {
	Resource string `json:"resource"` // e.g. "agent", "provider", "task"
	Name     string `json:"name"`     // e.g. "reviewer-agent", "gemini-flash"
}

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
	RetryAfter time.Duration // Recommended backoff wait (0 means immediate/default backoff)
}

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
// Safe to call with nil.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var rErr *RetryableError
	return As(err, &rErr)
}

// =============================================================================
// Helper Wrappers for standard errors package (convenience)
// =============================================================================

// Is is a proxy to errors.Is.
func Is(err, target error) bool {
	return is(err, target)
}

// As is a proxy to errors.As.
func As(err error, target any) bool {
	return as(err, target)
}

// Unwrap is a proxy to errors.Unwrap.
func Unwrap(err error) error {
	return unwrap(err)
}
```

> **Lưu ý triển khai**: Để tránh conflict tên hàm trùng với gói `errors` của standard library khi triển khai trong cùng một file, ta định nghĩa các unexported helpers tương tự proxy. Hoặc đơn giản là dùng trực tiếp `errors.Is` từ standard library. Ta định nghĩa proxy unexported như sau ở cuối file:

```go
import "errors"

func is(err, target error) bool { return errors.Is(err, target) }
func as(err error, target any) bool { return errors.As(err, target) }
func unwrap(err error) error     { return errors.Unwrap(err) }
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Trả về interface nil nhưng concrete type khác nil
```go
// ❌ SAI:
func ValidateConfig(cfg *Config) error {
    var err *ValidationError = nil
    if cfg.Tasks == nil {
        err = &ValidationError{...}
    }
    return err // Trả về interface error chứa concrete nil pointer -> err != nil kiểm tra sẽ bị TRUE!
}

// ✅ ĐÚNG:
func ValidateConfig(cfg *Config) error {
    if cfg.Tasks == nil {
        return &ValidationError{...}
    }
    return nil // Trả về interface nil thực sự
}
```
Trong Go, một interface chỉ `nil` khi cả type và value của nó đều `nil`. Trả về một con trỏ kiểu cụ thể có giá trị `nil` thông qua biến kiểu `error` sẽ làm kiểm tra `err != nil` trả về `true`.

### Pitfall 2: Bỏ quên Unwrap khi viết Custom Wrappers
```go
// ❌ SAI:
type RetryableError struct { Err error }
func (e *RetryableError) Error() string { return e.Err.Error() }
// Thiếu Unwrap() -> errors.Is() không thể duyệt sâu vào bên trong để tìm lỗi gốc

// ✅ ĐÚNG:
func (e *RetryableError) Unwrap() error { return e.Err }
```
Phương thức `Unwrap()` là bắt buộc để chain lỗi hoạt động với `errors.Is` và `errors.As`.

## Checklist
- [ ] File `contracts/errors.go` được tạo/cập nhật thành công
- [ ] Package name: `contracts`
- [ ] Định nghĩa `ValidationError` struct và constructor `NewValidationError`
- [ ] Định nghĩa `NotFoundError` struct và constructor `NewNotFoundError`
- [ ] Định nghĩa `TimeoutError` struct và constructor `NewTimeoutError`
- [ ] Định nghĩa `ConflictError` struct và constructor `NewConflictError`
- [ ] Định nghĩa `PermissionError` struct và constructor `NewPermissionError`
- [ ] Định nghĩa `RetryableError` struct và helper `IsRetryable`
- [ ] Mọi custom error struct đều có phương thức `Error()`
- [ ] Helper `IsRetryable` sử dụng `As()` để phân tích chuỗi lỗi lồng nhau
- [ ] `go build ./contracts/...` không lỗi
