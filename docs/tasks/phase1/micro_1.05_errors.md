# Micro-Task 1.05: Tạo contracts/errors.go

## Thông tin
- **File tạo**: `contracts/errors.go`
- **Package**: `contracts`
- **Dependencies trước**: 1.01 (go.mod)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/...`

## Nội dung CHÍNH XÁC cần tạo

```go
// Package contracts defines shared interfaces, types, and errors
// used across all orchestrator components.
package contracts

import "errors"

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
```

## Quy tắc
1. Mỗi error variable bắt đầu bằng `Err` — Go convention
2. Mỗi error có prefix `"orchestrator: "` — dễ identify nguồn error trong logs
3. Mỗi error có Godoc comment giải thích KHI NÀO lỗi này xảy ra
4. Comment chỉ rõ error có retryable hay không
5. Dùng `errors.New()` — KHÔNG dùng `fmt.Errorf()` cho sentinel errors

## ⚠️ Pitfalls cần tránh
1. **KHÔNG dùng string matching**: Consumer phải dùng `errors.Is(err, contracts.ErrProviderTimeout)`, KHÔNG dùng `err.Error() == "..."` 
2. **KHÔNG tạo custom error types lúc này**: Sentinel errors đủ cho Phase 1. Custom error types (với fields) thêm sau khi cần
3. **Wrap errors đúng cách**: Khi wrap, dùng `fmt.Errorf("context: %w", err)` với `%w` verb. KHÔNG dùng `%v` — sẽ mất khả năng `errors.Is()` và `errors.As()`

## Cách consumer sử dụng (tham khảo, KHÔNG code trong file này)
```go
// Ở nơi khác trong codebase:
result, err := provider.Send(ctx, req)
if errors.Is(err, contracts.ErrProviderTimeout) {
    // Retry
}
if errors.Is(err, contracts.ErrProviderAuthFailed) {
    // Don't retry, show error to user
}
```

## Checklist
- [ ] File `contracts/errors.go` tồn tại
- [ ] Package declaration: `package contracts`
- [ ] Có ít nhất 15 error variables
- [ ] Mỗi error bắt đầu bằng `Err`
- [ ] Mỗi error có Godoc comment
- [ ] Dùng `errors.New()` (không phải `fmt.Errorf()`)
- [ ] `go build ./contracts/...` không lỗi
- [ ] `go vet ./contracts/...` không warning
