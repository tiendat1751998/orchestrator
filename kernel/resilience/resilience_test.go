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
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
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
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
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
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
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
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
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
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
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
		MaxAttempts:  2,
		InitialDelay: 5 * time.Millisecond,
		MaxDelay:     20 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
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
	cb := resilience.NewCircuitBreaker(2, 20*time.Millisecond)

	ctx := context.Background()

	// Initial State: Closed
	if cb.GetState() != resilience.StateClosed {
		t.Fatalf("expected StateClosed, got %v", cb.GetState())
	}

	// 1. Success does not change StateClosed
	err := cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if cb.GetState() != resilience.StateClosed {
		t.Fatalf("expected StateClosed, got %v", cb.GetState())
	}

	// 2. Failures up to failureThreshold (2) transition state to Open
	dummyErr := errors.New("dummy error")
	for i := 0; i < 2; i++ {
		err = cb.Execute(ctx, func() error { return dummyErr })
		if !errors.Is(err, dummyErr) {
			t.Fatalf("expected dummyErr, got %v", err)
		}
	}
	if cb.GetState() != resilience.StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.GetState())
	}

	// 3. While Open, Execute fails-fast immediately
	err = cb.Execute(ctx, func() error { return nil })
	if !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}

	// 4. Cooldown expires, state transitions to Half-Open on next execution
	time.Sleep(25 * time.Millisecond)

	// Under Half-Open, a success should transition it back to Closed
	err = cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cb.GetState() != resilience.StateClosed {
		t.Fatalf("expected StateClosed after probe success, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker(2, 10*time.Millisecond)
	ctx := context.Background()

	dummyErr := errors.New("dummy error")

	// Trip the circuit breaker (Closed -> Open)
	_ = cb.Execute(ctx, func() error { return dummyErr })
	_ = cb.Execute(ctx, func() error { return dummyErr })
	if cb.GetState() != resilience.StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.GetState())
	}

	// Cooldown wait
	time.Sleep(15 * time.Millisecond)

	// Next call should probe. Return error during Half-Open
	err := cb.Execute(ctx, func() error { return dummyErr })
	if !errors.Is(err, dummyErr) {
		t.Fatalf("expected dummyErr, got %v", err)
	}

	// State must immediately transition back to Open
	if cb.GetState() != resilience.StateOpen {
		t.Fatalf("expected StateOpen after Half-Open failure, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_ThreadSafety(t *testing.T) {
	cb := resilience.NewCircuitBreaker(50, 5*time.Second)
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
					callErr = errors.New("error")
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

	state := cb.GetState()
	t.Logf("Finished parallel tests. Final state: %v, successes: %d, failures: %d", state, successCount, failureCount)
}

func TestWithFallback_PrimarySuccess(t *testing.T) {
	primaryCalled := false
	fallbackCalled := false

	primary := func() error {
		primaryCalled = true
		return nil
	}
	fallback := func() error {
		fallbackCalled = true
		return nil
	}

	err := resilience.WithFallback(primary, fallback)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !primaryCalled {
		t.Error("expected primary to be called")
	}
	if fallbackCalled {
		t.Error("expected fallback NOT to be called")
	}
}

func TestWithFallback_PrimaryFailFallbackSuccess(t *testing.T) {
	primaryCalled := false
	fallbackCalled := false
	primaryErr := errors.New("primary error")

	primary := func() error {
		primaryCalled = true
		return primaryErr
	}
	fallback := func() error {
		fallbackCalled = true
		return nil
	}

	err := resilience.WithFallback(primary, fallback)
	if err != nil {
		t.Fatalf("expected no error from fallback, got %v", err)
	}
	if !primaryCalled {
		t.Error("expected primary to be called")
	}
	if !fallbackCalled {
		t.Error("expected fallback to be called")
	}
}

func TestWithFallback_PrimaryFailFallbackFail(t *testing.T) {
	primaryErr := errors.New("primary error")
	fallbackErr := errors.New("fallback error")

	primary := func() error {
		return primaryErr
	}
	fallback := func() error {
		return fallbackErr
	}

	err := resilience.WithFallback(primary, fallback)
	if !errors.Is(err, fallbackErr) {
		t.Fatalf("expected fallback error %v, got %v", fallbackErr, err)
	}
}

func TestWithFallback_PrimaryFailFallbackNil(t *testing.T) {
	primaryErr := errors.New("primary error")

	primary := func() error {
		return primaryErr
	}

	err := resilience.WithFallback(primary, nil)
	if !errors.Is(err, primaryErr) {
		t.Fatalf("expected primary error %v, got %v", primaryErr, err)
	}
}

func TestWithFallback_PrimaryNil(t *testing.T) {
	err := resilience.WithFallback(nil, nil)
	if err == nil {
		t.Fatal("expected error when primary is nil")
	}
	if err.Error() != "fallback: primary call function cannot be nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestCascadingTimeoutContext(t *testing.T) {
	t.Run("NoDeadline", func(t *testing.T) {
		parent := context.Background()
		ctx, cancel := resilience.CascadingTimeoutContext(parent, 50*time.Millisecond)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected context to have a deadline")
		}
		remaining := time.Until(deadline)
		if remaining > 60*time.Millisecond || remaining < 40*time.Millisecond {
			t.Errorf("expected deadline to be approx 50ms, remaining was %v", remaining)
		}
	})

	t.Run("ParentDeadlineShorter", func(t *testing.T) {
		parent, cancelParent := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancelParent()

		ctx, cancel := resilience.CascadingTimeoutContext(parent, 100*time.Millisecond)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected context to have a deadline")
		}
		remaining := time.Until(deadline)
		if remaining > 30*time.Millisecond || remaining < 10*time.Millisecond {
			t.Errorf("expected deadline to be capped to parent's approx 20ms, remaining was %v", remaining)
		}
	})

	t.Run("ParentAlreadyExpired", func(t *testing.T) {
		parent, cancelParent := context.WithDeadline(context.Background(), time.Now().Add(-10*time.Millisecond))
		defer cancelParent()

		ctx, cancel := resilience.CascadingTimeoutContext(parent, 50*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			// expected
		default:
			t.Error("expected context to be cancelled immediately")
		}
	})
}
