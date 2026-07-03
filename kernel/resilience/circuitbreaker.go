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
	StateClosed   CBState = iota // Normal operation: requests pass through.
	StateOpen                    // Failing state: requests fail-fast immediately.
	StateHalfOpen                // Probe state: testing if downstream has recovered.
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
