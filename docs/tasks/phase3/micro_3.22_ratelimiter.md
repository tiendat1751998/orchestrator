# Micro-Task 3.22: Create sdk/helpers/ratelimit.go

## Info
- **File**: `sdk/helpers/ratelimit.go`
- **Package**: `helpers`
- **Depends on**: 1.06
- **Time**: 15 min
- **Verify**: `go build ./sdk/helpers/...`

## Purpose
Triển khai bộ kiểm soát tốc độ gọi API (`TokenBucket` Rate Limiter). Đây là một helper đa luồng (thread-safe) sử dụng thuật toán giỏ chứa thẻ (Token Bucket) để giới hạn tần suất yêu cầu gửi tới nhà cung cấp mô hình AI, giúp tránh lỗi vượt ngưỡng hạn định cuộc gọi (HTTP 429 Rate Limits).

## EXACT code to create

```go
// Package helpers provides generic utilities like rate limiters and backoff helpers.
package helpers

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

// TokenBucket implements a thread-safe rate limiter using the Token Bucket algorithm.
type TokenBucket struct {
	mu            sync.Mutex
	capacity      float64
	tokens        float64
	refillRate    float64 // Tokens added per nanosecond
	lastRefillTime time.Time
}

// NewTokenBucket creates a new TokenBucket rate limiter.
//
// Parameters:
//   - capacity: maximum burst size (e.g. 5 tokens).
//   - refillInterval: duration between token increments (e.g. 1s for 1 token/sec).
func NewTokenBucket(capacity int, refillInterval time.Duration) *TokenBucket {
	if capacity <= 0 {
		capacity = 1
	}
	if refillInterval <= 0 {
		refillInterval = time.Second
	}

	// Refill rate = 1 token per refillInterval -> converted to tokens/nanosecond
	refillRate := 1.0 / float64(refillInterval.Nanoseconds())

	return &TokenBucket{
		capacity:       float64(capacity),
		tokens:         float64(capacity), // Start full
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
		// Check context cancellation first
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

		// Calculate wait time for the next token to arrive
		missingTokens := 1.0 - tb.tokens
		waitTimeNS := math.Ceil(missingTokens / tb.refillRate)
		waitDuration := time.Duration(waitTimeNS)

		tb.mu.Unlock()

		// Sleep for calculated wait time or abort on context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}
}

// refill adds tokens to the bucket based on the time elapsed since last check.
// Must be called with tb.mu lock held.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)
	tb.lastRefillTime = now

	addedTokens := float64(elapsed.Nanoseconds()) * tb.refillRate
	tb.tokens = math.Min(tb.capacity, tb.tokens+addedTokens)
}
```

## ⚠️ Pitfalls

### Pitfall 1: Holding lock during Wait sleep
```go
// ❌ WRONG:
tb.mu.Lock()
for tb.tokens < 1.0 {
    time.Sleep(waitDuration) // Holding lock while sleeping -> blocks all other goroutines from using the rate limiter!
}
tb.mu.Unlock()

// ✅ CORRECT:
tb.mu.Lock()
// calculate waitDuration...
tb.mu.Unlock() // Release lock before sleep!
select {
case <-ctx.Done(): return
case <-time.After(waitDuration):
}
```
Sleeping while holding a mutex locks up the entire system. Always release the lock before waiting, and re-acquire it on the next loop iteration.

### Pitfall 2: Integer division truncation in math calculations
Dividing nanoseconds via integers can drop fractional values, causing the refill rate to evaluate to zero for very small rates. Always cast nanoseconds to `float64` for precise calculations.

## Verify
```bash
go build ./sdk/helpers/...
```

## Checklist
- [ ] File `sdk/helpers/ratelimit.go` tồn tại
- [ ] Package: `helpers`
- [ ] Triển khai thuật toán Token Bucket an toàn đa luồng bằng Mutex
- [ ] Hàm `Allow` trả về kết quả ngay lập tức (non-blocking)
- [ ] Hàm `Wait` giải phóng lock trước khi chờ bằng `time.After`
- [ ] `refill` giới hạn số lượng token tối đa không vượt quá `capacity`
- [ ] `go build ./sdk/helpers/...` không lỗi
