# Micro-Task 1.05: Create contracts/errors.go

## Info
- **File**: `contracts/errors.go`
- **Package**: `contracts`
- **Depends on**: 1.01
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Defines standard sentinel errors shared across all orchestrator components. This establishes a uniform error contract to enable clean error handling, retry classification, and logging.

## EXACT code to create

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

## ⚠️ Pitfalls

### Pitfall 1: Checking error string equality in consumer
```go
if errors.Is(err, contracts.ErrProviderTimeout) {
    // Retry logic...
}
```
Direct string comparisons fail if the error gets wrapped. Always use `errors.Is(err, Target)` to check sentinel error types.

### Pitfall 2: Using `%v` when wrapping sentinel errors
```go
return fmt.Errorf("failed to fetch response: %w", contracts.ErrProviderTimeout)
```
Always use `%w` in formatting statements to wrap error chains properly.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/errors.go` exists
- [ ] Package: `contracts`
- [ ] Error names start with `Err` prefix
- [ ] Defined sentinel variables use `errors.New` rather than formatting functions
- [ ] `go build ./contracts/...` passes
