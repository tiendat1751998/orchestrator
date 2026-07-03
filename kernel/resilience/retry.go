package resilience

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// RetryConfig configures retry policies.
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryConfig returns a standard retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// Retry executes the target function in a retry loop using exponential backoff.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// 1. Terminate early if error is not retryable
		if !IsRetryable(err) {
			return err
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// 2. Compute backoff delay
		delay := calculateDelay(attempt, cfg)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("retry: max attempts reached (%d). last error: %w", cfg.MaxAttempts, lastErr)
}

// IsRetryable determines if an error is transient and can be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 1. Standard contracts sentinels
	if errors.Is(err, contracts.ErrProviderTimeout) ||
		errors.Is(err, contracts.ErrProviderRateLimited) ||
		errors.Is(err, contracts.ErrProviderUnavailable) {
		return true
	}

	// 2. Network connection or timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// 3. Explicitly retryable error contract
	if contracts.IsRetryable(err) {
		return true
	}

	return false
}

func calculateDelay(attempt int, cfg RetryConfig) time.Duration {
	temp := float64(cfg.InitialDelay)
	for i := 0; i < attempt; i++ {
		temp *= cfg.Multiplier
		if temp >= float64(cfg.MaxDelay) {
			temp = float64(cfg.MaxDelay)
			break
		}
	}

	delay := time.Duration(temp)

	if cfg.Jitter {
		// Apply ±20% randomized jitter
		jitterFactor := 0.8 + rand.Float64()*0.4
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	return delay
}

// RetryWithResult executes a result-returning function in a retry loop using exponential backoff.
// Used by provider middlewares where the function returns (*Response, error).
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !IsRetryable(err) {
			return zero, err
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := calculateDelay(attempt, cfg)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}

	return zero, fmt.Errorf("retry: max attempts reached (%d). last error: %w", cfg.MaxAttempts, lastErr)
}
