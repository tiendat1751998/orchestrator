# Micro-Task 2.36: Tạo kernel/resilience (Retry và Circuit Breaker)

## Thông tin
- **File tạo**:
  - `kernel/resilience/retry.go`
  - `kernel/resilience/circuitbreaker.go`
- **Package**: `resilience`
- **Dependencies trước**: 2.32 (kernel.go), 1.37 (errors.go)
- **Thời gian**: 25 phút
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Triển khai hệ thống tự phục hồi (resilience patterns) bao gồm: Cơ chế Retry với Exponential Backoff & Jitter (độ trễ ngẫu nhiên) và Circuit Breaker (Bộ ngắt mạch 3 trạng thái: Closed, Open, Half-Open). Giúp hệ thống không bị quá tải khi gọi các API AI (như rate limits, timeouts) và tự động ngắt mạch khi provider sập liên tục.

## EXACT code to create

### Phần 1: Tạo `kernel/resilience/retry.go`

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
	// Max jitter factor: 25%
	jitterRange := d / 4
	if jitterRange <= 0 {
		return d
	}

	// Generate safe random number
	n, err := rand.Int(rand.Reader, big.NewInt(int64(jitterRange*2)))
	var randVal int64
	if err == nil {
		randVal = n.Int64()
	}

	// shift random value to center around 0 (-jitterRange to +jitterRange)
	offset := randVal - int64(jitterRange)
	return d + time.Duration(offset)
}
```

---

### Phần 2: Tạo `kernel/resilience/circuitbreaker.go`

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
	// If yes, transition to Half-Open. If no, fail-fast.
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
	// ValidationError or AuthFailed should NOT trip the breaker.
	if contracts.IsRetryable(err) {
		cb.handleFailure()
	}
}

func (cb *CircuitBreaker) handleSuccess() {
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.probeRequests {
			// Downstream recovered
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
			// Trip the breaker
			cb.state = StateOpen
			cb.lastStateMod = time.Now()
		}
	} else if cb.state == StateHalfOpen {
		// Any failure in Half-Open trips it back to Open
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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Bắt tất cả lỗi làm ngắt mạch (Tripping on non-retryable errors)
```go
// ❌ SAI:
cb.handleFailure() // Ngắt mạch khi người dùng gửi API Key sai (AuthFailed) hoặc nhập Task không hợp lệ.
// Kết quả: Hệ thống tự sập liên tục mà không có cơ hội phục hồi.

// ✅ ĐÚNG:
if contracts.IsRetryable(err) {
    cb.handleFailure() // Chỉ ngắt mạch đối với lỗi kết nối mạng, Rate limit hoặc Timeout.
}
```
Lỗi nghiệp vụ (lỗi tham số cấu hình, auth, validation) là lỗi của người dùng, KHÔNG phải lỗi của hạ tầng. Breaker chỉ được ngắt khi hạ tầng hoặc nhà cung cấp (provider) sập.

### Pitfall 2: Time.Sleep chặn đứng Goroutine trong Retry
```go
// ❌ SAI:
time.Sleep(delay) // Block goroutine chạy task, không phản hồi khi context cancel.

// ✅ ĐÚNG:
select {
case <-ctx.Done():
    return ctx.Err()
case <-time.After(actualDelay):
}
```
`time.Sleep` không tôn trọng Context Cancellation. Nếu một mission bị cancel, Goroutine sẽ bị kẹt lại cho tới khi ngủ đủ giấc gây rò rỉ tài nguyên.

## Checklist
- [ ] File `kernel/resilience/retry.go` tồn tại
- [ ] File `kernel/resilience/circuitbreaker.go` tồn tại
- [ ] Retry sử dụng Exponential Backoff + Jitter an toàn bằng `crypto/rand`
- [ ] Chỉ retry những lỗi thỏa mãn `contracts.IsRetryable`
- [ ] Circuit Breaker hỗ trợ 3 trạng thái: Closed, Open, Half-Open
- [ ] Breaker tự động chuyển từ Open sang Half-Open sau khi hết Cooldown
- [ ] Breaker chỉ đếm lỗi hạ tầng (`IsRetryable` = true)
- [ ] Mọi tài nguyên lock được giải phóng bằng `defer`
- [ ] `go build ./kernel/resilience/...` không lỗi
