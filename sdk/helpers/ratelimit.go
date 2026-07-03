package helpers

import (
	"context"
	"math"
	"sync"
	"time"
)

// TokenBucket implements a thread-safe rate limiter using the Token Bucket algorithm.
type TokenBucket struct {
	mu             sync.Mutex
	capacity       float64
	tokens         float64
	refillRate     float64 // Tokens added per nanosecond
	lastRefillTime time.Time
}

// NewTokenBucket creates a new TokenBucket rate limiter.
func NewTokenBucket(capacity int, refillInterval time.Duration) *TokenBucket {
	if capacity <= 0 {
		capacity = 1
	}
	if refillInterval <= 0 {
		refillInterval = time.Second
	}

	refillRate := 1.0 / float64(refillInterval.Nanoseconds())

	return &TokenBucket{
		capacity:       float64(capacity),
		tokens:         float64(capacity),
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Allow reports whether a token is available immediately. Non-blocking.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// Wait blocks until a token becomes available or the context is cancelled.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		tb.mu.Lock()
		tb.refill()

		if tb.tokens >= 1.0 {
			tb.tokens -= 1.0
			tb.mu.Unlock()
			return nil
		}

		missingTokens := 1.0 - tb.tokens
		waitTimeNS := math.Ceil(missingTokens / tb.refillRate)
		waitDuration := time.Duration(waitTimeNS)

		tb.mu.Unlock()

		timer := time.NewTimer(waitDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			timer.Stop()
		}
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)
	if elapsed < 0 {
		// ponytail: clock drift, recover gracefully by setting the refilled time to now without decreasing tokens
		tb.lastRefillTime = now
		return
	}
	tb.lastRefillTime = now

	addedTokens := float64(elapsed.Nanoseconds()) * tb.refillRate
	tb.tokens = math.Min(tb.capacity, tb.tokens+addedTokens)
}
