# Micro-Task 1.31: Create contracts/resilience/resilience.go

## Info
- **File**: `contracts/resilience/resilience.go`
- **Package**: `resilience`
- **Depends on**: 1.05 (contracts/errors.go)
- **Time**: 10 min
- **Verify**: `go build ./contracts/resilience/...`

## Purpose
Declares structural contracts (`CircuitBreaker`, `RetryPolicy`, `Fallback`) for fault-tolerance execution patterns when dealing with flaky external LLM model providers.

## EXACT code to create

```go
// Package resilience defines contracts for fault-tolerance patterns.
// Circuit breaker, retry, fallback — essential for reliable AI provider calls.
package resilience

import "context"

// CircuitBreaker prevents cascading failures by stopping requests
// to a failing service.
//
// States:
//   Closed   → Normal. Requests pass through. Failures are counted.
//   Open     → Blocked. Requests are rejected immediately (no network call).
//   HalfOpen → Testing. Allow ONE request through. If success → Closed. If fail → Open.
//
// State transitions:
//   Closed → Open: when failure count reaches threshold
//   Open → HalfOpen: after reset timeout expires
//   HalfOpen → Closed: when test request succeeds
//   HalfOpen → Open: when test request fails
type CircuitBreaker interface {
	// Execute runs the given function through the circuit breaker.
	//
	// If circuit is Closed or HalfOpen: executes fn()
	// If circuit is Open: returns error immediately without calling fn()
	//
	// Returns:
	//   - nil if fn() succeeded
	//   - fn()'s error if fn() failed
	//   - ErrCircuitOpen if circuit is open
	Execute(fn func() error) error

	// State returns the current circuit breaker state ("closed", "open", "half-open").
	State() string

	// Reset forces the circuit breaker back to Closed state.
	// Used for manual recovery or testing.
	Reset()
}

// RetryPolicy defines retry behavior for transient failures.
//
// Typical configuration:
//   MaxAttempts: 3
//   InitialDelay: 1s → 2s → 4s (exponential backoff)
//   Jitter: true (random variation to avoid thundering herd)
type RetryPolicy interface {
	// Execute runs the given function with retry logic.
	//
	// Retries on retryable errors (timeout, rate limit, 503).
	// Does NOT retry on non-retryable errors (401, 403, invalid input).
	//
	// Respects ctx cancellation — stops retrying if ctx is cancelled.
	Execute(ctx context.Context, fn func() error) error
}

// Fallback provides alternative execution paths.
//
// When the primary function fails, the fallback function is called.
// Example: Primary = call Antigravity CLI. Fallback = call Gemini API directly.
type Fallback interface {
	// Execute tries the primary function first, then fallback on failure.
	Execute(primary func() error, fallback func() error) error
}

// ErrCircuitOpen is returned by CircuitBreaker.Execute when the circuit is open.
// This is a sentinel error that callers can check with errors.Is().
//
// NOTE: This is defined here (not in contracts/errors.go) because it's
// specific to the resilience package. Only resilience-related code checks for it.
var ErrCircuitOpen = circuitOpenError{}

type circuitOpenError struct{}

func (circuitOpenError) Error() string { return "circuit breaker is open" }
```

## Rules
1. **Circuit Breaker State Machine**: The state transitions strictly follow: `Closed` -> `Open` -> `HalfOpen` -> `Closed` (or back to `Open` if test execution fails).
2. **Context Cancellation Propagation**: The `RetryPolicy` execution must monitor the context's state. If `ctx.Done()` fires, the policy must stop executing retries immediately.
3. **Fallback Sequence**: Fallback attempts must run sequentially: the primary function first, and only on its failure does the secondary fallback execute.

## ⚠️ Pitfalls

### Pitfall 1: Retrying on permanent authentication or validation errors
```go
// Ensure RetryPolicy filters errors and only retries on transient errors (like ErrProviderTimeout or ErrProviderRateLimited).
```
Filter the errors so that you only retry on transient failure signatures.

### Pitfall 2: Not checking context cancellation during backoff sleeps
If the retry logic sleeps (e.g. for 5 seconds between retries) using `time.Sleep` instead of checking context channels (`select { case <-ctx.Done(): ... case <-time.After(delay): ... }`), a cancelled request will hang until the sleep completes, delaying goroutine cleanup.

## Verify
```bash
go build ./contracts/resilience/...
```

## Checklist
- [ ] File `contracts/resilience/resilience.go` exists
- [ ] Package: `resilience`
- [ ] `CircuitBreaker` interface declares Execute, State, and Reset methods
- [ ] `RetryPolicy` interface declares Execute method receiving `context.Context`
- [ ] `Fallback` interface declares Execute method
- [ ] `ErrCircuitOpen` custom error is declared
- [ ] `go build ./contracts/resilience/...` passes
