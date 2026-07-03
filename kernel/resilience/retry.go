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
