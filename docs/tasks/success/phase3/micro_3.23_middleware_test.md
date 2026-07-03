# Micro-Task 3.23: Create sdk/middleware/middleware_test.go

## Info
- **File**: `sdk/middleware/middleware_test.go`
- **Package**: `middleware_test`
- **Depends on**: 3.20 (agent_middleware.md), 3.21 (provider_middleware.md), 3.22 (ratelimiter.md)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/middleware/... ./sdk/helpers/...`

## Purpose
Implements integration unit tests for the Agent & Provider Middlewares and Rate Limiter, verifying execution chains, recovery handlers, metrics aggregations, retries, and token bucket throttling.

## EXACT code to create

```go
package middleware_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	contractsprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/kernel/metrics"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
	"github.com/tiendat1751998/orchestrator/sdk/helpers"
	"github.com/tiendat1751998/orchestrator/sdk/middleware"
	sdktesting "github.com/tiendat1751998/orchestrator/sdk/testing"
)

// =============================================================================
// Agent Middleware Tests
// =============================================================================

func TestAgentMiddleware_Recovery(t *testing.T) {
	ma := &sdktesting.MockAgent{
		NameVal: "panic-agent",
		ExecuteFn: func(ctx context.Context, task *contractsagent.Task) (*contractsagent.Result, error) {
			panic("something went critically wrong")
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	recoveredAgent := middleware.ChainAgent(ma, middleware.AgentRecovery(logger))

	task := &contractsagent.Task{ID: "tsk-p"}
	res, err := recoveredAgent.Execute(context.Background(), task)

	if err == nil {
		t.Fatal("expected error from recovered panic, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestAgentMiddleware_Metrics(t *testing.T) {
	ma := &sdktesting.MockAgent{NameVal: "metrics-agent"}
	reg := metrics.NewRegistry()

	measuredAgent := middleware.ChainAgent(ma, middleware.AgentMetrics(reg))

	task := &contractsagent.Task{ID: "tsk-m", Type: "code_generation"}
	_, err := measuredAgent.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := reg.Snapshot()
	if snap["counter.orchestrator_tasks_total"] != 1.0 {
		t.Errorf("expected counter task total to be 1.0, got %v", snap["counter.orchestrator_tasks_total"])
	}
}

// =============================================================================
// Provider Middleware Tests
// =============================================================================

func TestProviderMiddleware_Retry(t *testing.T) {
	var callCount int32
	mp := &sdktesting.MockProvider{
		NameVal: "retry-provider",
		SendFn: func(ctx context.Context, req *contractsprovider.Request) (*contractsprovider.Response, error) {
			count := atomic.AddInt32(&callCount, 1)
			if count < 3 {
				return nil, resilience.NewRetryableError(errors.New("connection reset"), 1*time.Millisecond)
			}
			return &contractsprovider.Response{Content: "success"}, nil
		},
	}

	retryCfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		Jitter:      false,
	}

	retriedProvider := middleware.ChainProvider(mp, middleware.ProviderRetry(retryCfg))

	resp, err := retriedProvider.Send(context.Background(), &contractsprovider.Request{})
	if err != nil {
		t.Fatalf("expected success after retries, got err: %v", err)
	}

	if resp.Content != "success" {
		t.Errorf("got content: %q", resp.Content)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// =============================================================================
// Rate Limiter Helper Tests
// =============================================================================

func TestTokenBucket_RateLimiting(t *testing.T) {
	tb := helpers.NewTokenBucket(2, 50*time.Millisecond)

	if !tb.Allow() {
		t.Error("expected first token to be allowed")
	}
	if !tb.Allow() {
		t.Error("expected second token to be allowed")
	}
	if tb.Allow() {
		t.Error("expected third token to be rate-limited (empty bucket)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err != nil {
		t.Errorf("expected wait to succeed after refill, got err: %v", err)
	}
}
```

## Rules
1. **Thread-Safe Mock Counters**: Use atomic integers (`atomic.AddInt32` / `atomic.LoadInt32`) when checking retry counts to prevent race conditions during test executions.
2. **Context-Aware Rate Limit Tests**: Restrict rate limit tests using context timeouts (`context.WithTimeout`) to avoid test hangs.
3. **Structured Metrics Assertions**: Access registry snapshot statistics using structured namespace keys (e.g. `counter.<metric_name>`).

## ⚠️ Pitfalls

### Pitfall 1: Data races on mock call counters in multi-threaded tests
Using standard non-atomic integers (e.g. `count++`) inside parallel provider mocks triggers Go's race detector. Always use atomic synchronization primitives.

### Pitfall 2: Test blocks during rate limiter wait assertions
Using blocking wait operations in tests can cause hangs if refill loops contain bugs. Always set context deadlines to fail fast.

## Verify
```bash
go test -v -race -count=1 ./sdk/middleware/... ./sdk/helpers/...
```

## Checklist
- [ ] File `sdk/middleware/middleware_test.go` exists
- [ ] Package: `middleware_test` (external testing package)
- [ ] Recovery tests verify panic intercepting and error reporting
- [ ] Metrics tests verify task execution count reporting
- [ ] Provider retry tests verify transient failures retry limits
- [ ] Throttling tests verify Token Bucket bucket sizes
- [ ] Wait operations respect context timeouts during rate limit waits
- [ ] `go test -v -race ./sdk/middleware/... ./sdk/helpers/...` passes
