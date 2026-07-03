package resilience_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

func TestRetry_SuccessImmediately(t *testing.T) {
	ctx := context.Background()
	cfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		MaxWait:     5 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}

	calls := 0
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	cfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		MaxWait:     5 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}

	calls := 0
	transientErr := contracts.NewRetryableError(errors.New("transient error"), 0)
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		if calls < 3 {
			return transientErr
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_FailureMaxAttempts(t *testing.T) {
	ctx := context.Background()
	cfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		MaxWait:     5 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}

	calls := 0
	transientErr := contracts.NewRetryableError(errors.New("transient error"), 0)
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		return transientErr
	})

	if !errors.Is(err, transientErr) {
		t.Fatalf("expected error %v, got %v", transientErr, err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		MaxWait:     5 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}

	calls := 0
	fatalErr := errors.New("fatal error") // Not a contracts.RetryableError
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		return fatalErr
	})

	if !errors.Is(err, fatalErr) {
		t.Fatalf("expected error %v, got %v", fatalErr, err)
	}
	if calls != 1 {
		t.Errorf("expected only 1 call for non-retryable error, got %d", calls)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := resilience.RetryConfig{
		MaxAttempts: 5,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     50 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}

	calls := 0
	transientErr := contracts.NewRetryableError(errors.New("transient error"), 0)
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		if calls == 2 {
			cancel()
		}
		return transientErr
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected retry to stop after cancellation (calls: %d)", calls)
	}
}

func TestRetry_Jitter(t *testing.T) {
	ctx := context.Background()
	cfg := resilience.RetryConfig{
		MaxAttempts: 2,
		InitialWait: 5 * time.Millisecond,
		MaxWait:     20 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      true,
	}

	calls := 0
	transientErr := contracts.NewRetryableError(errors.New("transient error"), 0)
	err := resilience.Retry(ctx, cfg, func() error {
		calls++
		if calls < 2 {
			return transientErr
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cfg := resilience.CBConfig{
		MaxFailures:   2,
		Cooldown:      20 * time.Millisecond,
		ProbeRequests: 2,
	}
	cb := resilience.NewCircuitBreaker(cfg)

	ctx := context.Background()

	// Initial State: Closed
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected StateClosed, got %s", cb.State())
	}

	// 1. Success does not change StateClosed
	err := cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected StateClosed, got %s", cb.State())
	}

	// 2. Non-retryable error does not change state or failure count
	nonRetryableErr := errors.New("non-retryable error")
	err = cb.Execute(ctx, func() error { return nonRetryableErr })
	if !errors.Is(err, nonRetryableErr) {
		t.Fatalf("expected nonRetryableErr, got %v", err)
	}
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected StateClosed after non-retryable error, got %s", cb.State())
	}

	// 3. Retryable failures up to MaxFailures transition state to Open
	retryableErr := contracts.NewRetryableError(errors.New("transient issue"), 0)
	for i := 0; i < 2; i++ {
		err = cb.Execute(ctx, func() error { return retryableErr })
		if !errors.Is(err, retryableErr) {
			t.Fatalf("expected retryableErr, got %v", err)
		}
	}
	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected StateOpen, got %s", cb.State())
	}

	// 4. While Open, Execute fails-fast immediately
	err = cb.Execute(ctx, func() error { return nil })
	if err == nil || err.Error() == "" {
		t.Fatalf("expected fail-fast error, got nil")
	}

	// 5. Cooldown expires, state transitions to Half-Open on next execution
	time.Sleep(25 * time.Millisecond)

	// Under Half-Open, first success should still keep it Half-Open (probeRequests = 2)
	err = cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cb.State() != resilience.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen after first success, got %s", cb.State())
	}

	// Second success should transition it back to Closed
	err = cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected StateClosed after probe successes, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	cfg := resilience.CBConfig{
		MaxFailures:   2,
		Cooldown:      10 * time.Millisecond,
		ProbeRequests: 2,
	}
	cb := resilience.NewCircuitBreaker(cfg)
	ctx := context.Background()

	retryableErr := contracts.NewRetryableError(errors.New("transient issue"), 0)

	// Trip the circuit breaker (Closed -> Open)
	_ = cb.Execute(ctx, func() error { return retryableErr })
	_ = cb.Execute(ctx, func() error { return retryableErr })
	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected StateOpen, got %s", cb.State())
	}

	// Cooldown wait
	time.Sleep(15 * time.Millisecond)

	// Next call should probe. Return retryable error during Half-Open
	err := cb.Execute(ctx, func() error { return retryableErr })
	if !errors.Is(err, retryableErr) {
		t.Fatalf("expected retryableErr, got %v", err)
	}

	// State must immediately transition back to Open
	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected StateOpen after Half-Open failure, got %s", cb.State())
	}
}

func TestCircuitBreaker_ThreadSafety(t *testing.T) {
	cfg := resilience.CBConfig{
		MaxFailures:   50,
		Cooldown:      5 * time.Second,
		ProbeRequests: 3,
	}
	cb := resilience.NewCircuitBreaker(cfg)
	ctx := context.Background()

	var wg sync.WaitGroup
	var successCount int64
	var failureCount int64

	workers := 10
	iterations := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				var callErr error
				if (workerID+j)%2 == 0 {
					callErr = nil
				} else {
					callErr = contracts.NewRetryableError(errors.New("transient"), 0)
				}

				err := cb.Execute(ctx, func() error {
					return callErr
				})

				if err == nil {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failureCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	state := cb.State()
	t.Logf("Finished parallel tests. Final state: %s, successes: %d, failures: %d", state, successCount, failureCount)
}
