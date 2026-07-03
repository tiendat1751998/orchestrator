# Micro-Task 5.21: Create kernel/resilience/resilience_test.go

## Info
- **File**: `kernel/resilience/resilience_test.go`
- **Package**: `resilience_test`
- **Depends on**: 5.20
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./kernel/resilience/...`

## Purpose
Implements complete unit test coverage for the resilience features, checking circuit breaker state progressions, retries delays, and health checkpoints.

## EXACT code to create

```go
package resilience_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	// Breaker with 2 failure threshold, 50ms cool-down timeout
	cb := resilience.NewCircuitBreaker(2, 50*time.Millisecond)

	if cb.GetState() != resilience.StateClosed {
		t.Errorf("expected closed state, got %v", cb.GetState())
	}

	dummyErr := errors.New("dummy error")

	// 1. First failure
	_ = cb.Execute(func() error { return dummyErr })
	if cb.GetState() != resilience.StateClosed {
		t.Errorf("expected closed state after 1st failure, got %v", cb.GetState())
	}

	// 2. Second failure triggers open circuit state
	_ = cb.Execute(func() error { return dummyErr })
	if cb.GetState() != resilience.StateOpen {
		t.Errorf("expected open state after 2nd failure, got %v", cb.GetState())
	}

	// 3. Request should be rejected immediately with ErrCircuitOpen
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}

	// 4. Wait for cool-down window, request should transition state to HalfOpen
	time.Sleep(60 * time.Millisecond)
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("expected success in half-open trial request, got %v", err)
	}

	if cb.GetState() != resilience.StateClosed {
		t.Errorf("expected closed state after successful half-open trial, got %v", cb.GetState())
	}
}

func TestRetry_SuccessOnRetry(t *testing.T) {
	cfg := resilience.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 5 * time.Millisecond,
		MaxDelay:     20 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	calls := 0
	err := resilience.Retry(context.Background(), cfg, func() error {
		calls++
		if calls < 2 {
			return errors.New("temporary error") // Retryable error mock
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestWithFallback_Action(t *testing.T) {
	primaryErr := errors.New("primary failed")

	calls := 0
	err := resilience.WithFallback(
		func() error { return primaryErr },
		func() error {
			calls++
			return nil
		},
	)

	if err != nil {
		t.Errorf("expected fallback to succeed, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected fallback function to be executed once, got %d", calls)
	}
}

type mockCheckable struct {
	status cplugin.HealthStatus
}

func (m *mockCheckable) Health(ctx context.Context) (cplugin.HealthReport, error) {
	return cplugin.HealthReport{
		Status: m.status,
	}, nil
}

func TestHealthAggregator_Check(t *testing.T) {
	agg := resilience.NewHealthAggregator()
	agg.Register("comp-A", &mockCheckable{status: cplugin.HealthUp})
	agg.Register("comp-B", &mockCheckable{status: cplugin.HealthDown})

	results := agg.CheckAll(context.Background())

	if results["comp-A"].Status != cplugin.HealthUp {
		t.Errorf("expected comp-A up, got %v", results["comp-A"].Status)
	}
	if results["comp-B"].Status != cplugin.HealthDown {
		t.Errorf("expected comp-B down, got %v", results["comp-B"].Status)
	}
}

func TestCheckpointStore_SaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := resilience.NewCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to construct store: %v", err)
	}

	c := &resilience.Checkpoint{
		MissionID: "mission-123",
		State:     "running",
		Results:   map[string]string{"task-1": "success"},
	}

	err = store.Save(c)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	loaded, err := store.Load("mission-123")
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if loaded == nil || loaded.MissionID != "mission-123" || loaded.State != "running" {
		t.Errorf("loaded checkpoint mismatch: %+v", loaded)
	}
}
```

## Pitfalls

### Pitfall 1: Leaking temporary testing directories on failures
If you call `t.Fatalf()` before registers cleanups are run, directories can remain on system drives, cluttering memory. Always register `defer os.RemoveAll` immediately after successful temporary directory creation.

### Pitfall 2: Flaky unit test timings
Using short backoff delay periods (like 1ms) in busy CI runners can cause calls to interleave unpredictably. Use robust multipliers and jitter filters.

## Verify
```bash
go test -v -race -count=1 ./kernel/resilience/...
# Expected: PASS
```

## Checklist
- [ ] File exists at `kernel/resilience/resilience_test.go`
- [ ] Package name is `resilience_test`
- [ ] Circuit breaker transitions Closed -> Open -> HalfOpen -> Closed are tested
- [ ] Retry logic handles temporary mock errors successfully
- [ ] Fallback redirects to secondary handlers
- [ ] Checkpoint files save and load cleanly in temp directories
- [ ] Build command passes
