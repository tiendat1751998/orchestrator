# Micro-Task 5.15: Create kernel/resilience/circuit_breaker.go

## Info
- **File**: `kernel/resilience/circuit_breaker.go`
- **Package**: `resilience`
- **Depends on**: 5.14
- **Time**: 20 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements the tri-state circuit breaker pattern (`CircuitBreaker` and state transition hooks) to shield downstream AI providers from spam calls during outage windows.

## EXACT code to create

```go
// Package resilience implements circuit breakers, retries, and fallback strategies.
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the circuit state (Closed, Open, Half-Open).
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit_breaker: call blocked, circuit is open")

// CircuitBreaker shields target service calls from cascading failures.
// Thread-safe.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failureThreshold int
	resetTimeout     time.Duration
	failures         int
	lastFailure      time.Time
}

// NewCircuitBreaker constructs a new CircuitBreaker.
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5 // Default failure count
	}
	if timeout <= 0 {
		timeout = 30 * time.Second // Default cool-down window
	}

	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: threshold,
		resetTimeout:     timeout,
	}
}

// Execute wraps target function execution inside breaker checks.
// Respects context cancellation before attempting the call.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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

	now := time.Now()

	// If Open and reset timeout elapsed, transition to HalfOpen
	if cb.state == StateOpen {
		if now.Sub(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.failures = 0
			return nil
		}
		return ErrCircuitOpen
	}

	return nil
}

func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
		}
		return
	}

	// Success: Reset failures and close circuit
	cb.failures = 0
	cb.state = StateClosed
}

// GetState returns the current state.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
```

## Pitfalls

### Pitfall 1: Race conditions from unsynchronized state updates
```go
// WRONG:
func (cb *CircuitBreaker) Execute(fn func() error) error {
    if cb.state == StateOpen { ... } // Reads state field without lock! Data race when multiple goroutines run Execute.
}

// CORRECT:
// Protect state reads and writes inside mutex-guarded helper methods.
```
Since the orchestrator schedules multiple task calls concurrently, modifying state variables without mutex locks will trigger race conditions.

### Pitfall 2: Permitting multiple requests in Half-Open state
If the breaker transitions to Half-Open and spams multiple test requests concurrently, a slow downstream failure will trigger multiple increments, extending Open states. Only allow one request through in Half-Open.

## Verify
```bash
go build ./kernel/resilience/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/resilience/circuit_breaker.go`
- [ ] Package name is `resilience`
- [ ] All exported types have Godoc
- [ ] Breaker updates are protected under mutex locks
- [ ] Open state triggers immediate return of ErrCircuitOpen
- [ ] Build command passes
