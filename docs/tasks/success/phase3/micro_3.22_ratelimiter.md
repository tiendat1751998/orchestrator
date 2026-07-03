# Micro-Task 3.22: Create sdk/helpers/ratelimit.go

## Info
- **File**: `sdk/helpers/ratelimit.go`
- **Package**: `helpers`
- **Depends on**: 1.06
- **Time**: 15 min
- **Verify**: `go build ./sdk/helpers/...`

## Purpose
Implements a thread-safe token bucket rate limiter (`TokenBucket` and constructors) that regulates request frequencies to external AI providers, avoiding HTTP 429 rate limit responses.

## EXACT code to create

```go
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
		}
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)
	tb.lastRefillTime = now

	addedTokens := float64(elapsed.Nanoseconds()) * tb.refillRate
	tb.tokens = math.Min(tb.capacity, tb.tokens+addedTokens)
}
```

## Rules
1. **Mutex Release During Waits**: Release mutex locks before entering a sleep state during rate limit throttling. Do not block other threads while sleeping.
2. **Float Nanosecond Division**: Cast duration variables to float64 before divisions to prevent precision loss.
3. **Context Cancellation Support**: Respect context closures within sleep loops to abort immediately when context is cancelled.
4. **Timer Resource Leak Prevention**: Use `time.NewTimer` and explicitly invoke `timer.Stop()` when context triggers cancellations or the select loop completes. Avoid using `time.After()`, which leaves unexpired timers allocated in memory.

## Pitfalls

### Pitfall 1: Sleeping inside mutex locks
Always unlock the mutex before starting context-aware delay periods to avoid stalling other callers.

### Pitfall 2: Timer accumulation memory leaks
```go
// WRONG:
select {
case <-ctx.Done():
    return ctx.Err()
case <-time.After(waitDuration): // Creates a new channel/timer that leaks until waitDuration expires!
}
```
Using `time.After` in loop structures allocates resources that remain in memory for their full duration. Always instantiate explicit timers and call `Stop()`.

## Verify
```bash
go build ./sdk/helpers/...
```

## Checklist
- [ ] File `sdk/helpers/ratelimit.go` exists
- [ ] Package: `helpers`
- [ ] Mutex synchronization protects token bucket access
- [ ] `Allow` performs instant checks
- [ ] `Wait` unlocks mutexes before waiting for token refills
- [ ] `refill` caps token counts at max capacity
- [ ] Timer resources are closed cleanly using `Stop()`
- [ ] `go build ./sdk/helpers/...` passes
