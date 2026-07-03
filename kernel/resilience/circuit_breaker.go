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
	// StateClosed indicates the circuit is closed and requests flow normally.
	StateClosed State = iota
	// StateOpen indicates the circuit is open and requests fail fast.
	StateOpen
	// StateHalfOpen indicates the circuit is probing downstream recovery.
	StateHalfOpen
)

// ErrCircuitOpen is returned when a call is blocked because the circuit breaker is open.
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
	halfOpenInFlight bool // ponytail: tracks in-flight probe to prevent concurrent spams in Half-Open
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
			cb.halfOpenInFlight = true
			return nil
		}
		return ErrCircuitOpen
	}

	// If already in HalfOpen, only allow one request to probe at a time.
	if cb.state == StateHalfOpen {
		if cb.halfOpenInFlight {
			return ErrCircuitOpen
		}
		cb.halfOpenInFlight = true
		return nil
	}

	return nil
}

func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.halfOpenInFlight = false
		if err != nil {
			// A single failure during Half-Open transitions immediately back to Open
			cb.failures = cb.failureThreshold
			cb.lastFailure = time.Now()
			cb.state = StateOpen
		} else {
			// Success closes the circuit breaker
			cb.failures = 0
			cb.state = StateClosed
		}
		return
	}

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
		}
		return
	}

	// Success: Reset failures
	cb.failures = 0
}

// GetState returns the current state.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
