# Micro-Task 2.36: Create kernel/resilience (Retry and Circuit Breaker)

## Info
- **File created**:
  - `kernel/resilience/retry.go`
  - `kernel/resilience/circuitbreaker.go`
- **Package**: `resilience`
- **Depends on**: 2.32 (kernel.go), 1.37 (errors.go)
- **Time**: 25 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements self-healing resilience patterns: Exponential Backoff & Jitter retries, and a 3-state Circuit Breaker (Closed, Open, Half-Open). This guards AI provider integrations from transient network errors, rate limits, or service outages.

## EXACT code to create

### Part 1: Create `kernel/resilience/retry.go`

```go
package resilience

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// RetryConfig configures the retry policy.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of retry attempts (default: 3)
	InitialWait time.Duration // Initial delay duration (default: 1s)
	MaxWait     time.Duration // Maximum delay ceiling (default: 30s)
	Multiplier  float64       // Exponential growth multiplier (default: 2.0)
	Jitter      bool          // Enable randomized delay to prevent thundering herd (default: true)
}

// DefaultRetryConfig returns a pre-configured config.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}
}

// Retry executes a function. If it fails, retries it under the backoff policy.
// Only retries if the error satisfies contracts.IsRetryable(err).
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	_, err := RetryWithResult(ctx, cfg, func() (any, error) {
		return nil, fn()
	})
	return err
}

// RetryWithResult executes a function returning a result, with retry logic.
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error) {
	// Apply defaults if values are zero
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialWait <= 0 {
		cfg.InitialWait = 1 * time.Second
	}
	if cfg.MaxWait <= 0 {
		cfg.MaxWait = 30 * time.Second
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}

	var lastErr error
	delay := cfg.InitialWait

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context cancellation before executing
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Stop retrying if context is cancelled or the error is NOT retryable
		if !contracts.IsRetryable(err) || attempt == cfg.MaxAttempts {
			var zero T
			return zero, lastErr
		}

		// Calculate next delay
		actualDelay := delay
		if cfg.Jitter {
			actualDelay = applyJitter(delay)
		}

		// Sleep or context cancellation wait
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-time.After(actualDelay):
		}

		// Increase delay exponentially for next iteration
		nextDelay := float64(delay) * cfg.Multiplier
		if nextDelay > float64(cfg.MaxWait) {
			delay = cfg.MaxWait
		} else {
			delay = time.Duration(nextDelay)
		}
	}

	var zero T
	return zero, lastErr
}

// applyJitter adds random noise (±25%) to the backoff delay to prevent thundering herd.
func applyJitter(d time.Duration) time.Duration {
	jitterRange := d / 4
	if jitterRange <= 0 {
		return d
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(jitterRange*2)))
	var randVal int64
	if err == nil {
		randVal = n.Int64()
	}

	offset := randVal - int64(jitterRange)
	return d + time.Duration(offset)
}
```

---

### Part 2: Create `kernel/resilience/circuitbreaker.go`

```go
package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// CBState defines the states of the Circuit Breaker.
type CBState int

const (
	StateClosed CBState = iota // Normal operation: requests pass through.
	StateOpen                  // Failing state: requests fail-fast immediately.
	StateHalfOpen              // Probe state: testing if downstream has recovered.
)

func (s CBState) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateOpen:
		return "Open"
	case StateHalfOpen:
		return "Half-Open"
	default:
		return "Unknown"
	}
}

// CircuitBreaker implements the Circuit Breaker pattern.
// Thread-safe.
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CBState
	failureCount int
	successCount int
	lastStateMod time.Time

	// Configuration
	maxFailures   int           // Consecutive failures before tripping (default: 5)
	cooldown      time.Duration // Time to wait in Open state before testing recovery (default: 30s)
	probeRequests int           // Consecutive successes required in HalfOpen to close (default: 3)
}

// CBConfig configures the CircuitBreaker.
type CBConfig struct {
	MaxFailures   int
	Cooldown      time.Duration
	ProbeRequests int
}

// NewCircuitBreaker creates a new CircuitBreaker.
func NewCircuitBreaker(cfg CBConfig) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 30 * time.Second
	}
	if cfg.ProbeRequests <= 0 {
		cfg.ProbeRequests = 3
	}

	return &CircuitBreaker{
		state:         StateClosed,
		maxFailures:   cfg.MaxFailures,
		cooldown:      cfg.Cooldown,
		probeRequests: cfg.ProbeRequests,
		lastStateMod:  time.Now(),
	}
}

// Execute wraps a target call with Circuit Breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.beforeCall(); err != nil {
		return err
	}

	err := fn()

	cb.afterCall(err)
	return err
}

func (cb *CircuitBreaker) beforeCall() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// If Open state, check if cooldown time has expired.
	if cb.state == StateOpen {
		if time.Since(cb.lastStateMod) > cb.cooldown {
			cb.state = StateHalfOpen
			cb.lastStateMod = time.Now()
			cb.successCount = 0
			cb.failureCount = 0
		} else {
			return fmt.Errorf("circuit breaker is Open (cooldown remaining: %v)", cb.cooldown-time.Since(cb.lastStateMod))
		}
	}

	return nil
}

func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.handleSuccess()
		return
	}

	// We only trip on retryable (transient/network) errors.
	if contracts.IsRetryable(err) {
		cb.handleFailure()
	}
}

func (cb *CircuitBreaker) handleSuccess() {
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.probeRequests {
			cb.state = StateClosed
			cb.lastStateMod = time.Now()
			cb.failureCount = 0
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) handleFailure() {
	if cb.state == StateClosed {
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastStateMod = time.Now()
		}
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateMod = time.Now()
	}
}

// State returns the current state of the Circuit Breaker.
func (cb *CircuitBreaker) State() CBState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
```

## Rules
1. **Filter Non-Retryable Errors**: Avoid tripping Circuit Breakers on client-side errors like invalid inputs (`ValidationError`) or credentials errors. Only record failures when they satisfy `contracts.IsRetryable`.
2. **Context-Aware Backoff Waiting**: Enforce context monitoring inside retry delays. Do not use blocking `time.Sleep` statements which ignore cancellation requests. Use `select` with `ctx.Done()`.
3. **Cryptographically Safe Jitters**: Calculate random delay factors (±25%) using secure random generation APIs (`crypto/rand`) to prevent predictability.

## ⚠️ Pitfalls

### Pitfall 1: Tripping breakers on user validation errors
Failing to filter business/client errors causes the breaker to trip due to invalid API queries or incorrect configuration parameters, blocking all system tasks. Trip only on infrastructure or transient errors.

### Pitfall 2: Using blocking sleep calls within retries
Using `time.Sleep(delay)` leaves goroutines blocked even after tasks are canceled, leaking threads and wasting system resources. Always use context-aware delays.

## Verify
```bash
go build ./kernel/resilience/...
```

## Checklist
- [ ] File `kernel/resilience/retry.go` exists
- [ ] File `kernel/resilience/circuitbreaker.go` exists
- [ ] Retry implements exponential delays with ±25% random jitter using `crypto/rand`
- [ ] Retries are bypassed if `contracts.IsRetryable(err)` returns false
- [ ] Circuit Breaker handles Closed, Open, and Half-Open transitions
- [ ] Expirations in `Open` transition state automatically to `HalfOpen`
- [ ] Breaker updates fail states only for retryable errors
- [ ] Locks release cleanly via deferred statements
- [ ] `go build ./kernel/resilience/...` passes
