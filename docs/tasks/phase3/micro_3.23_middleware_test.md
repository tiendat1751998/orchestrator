# Micro-Task 3.23: Create sdk/middleware/middleware_test.go

## Info
- **File**: `sdk/middleware/middleware_test.go`
- **Package**: `middleware_test`
- **Depends on**: 3.20 (agent_middleware.md), 3.21 (provider_middleware.md), 3.22 (ratelimiter.md)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/middleware/... ./sdk/helpers/...`

## Purpose
Triển khai bộ kiểm thử tự động (Unit Tests) cho Agent & Provider Middlewares và Rate Limiter. Bộ kiểm thử xác thực thứ tự bọc chuỗi (chaining order), bắt và phục hồi panic (recovery), đo đếm metric thời gian và số lượng, tự ngắt mạch (breaker), và điều phối luồng thẻ (token bucket throttling).

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
	// Mock agent that panics
	ma := &sdktesting.MockAgent{
		NameVal: "panic-agent",
		ExecuteFn: func(ctx context.Context, task *contractsagent.Task) (*contractsagent.Result, error) {
			panic("something went critically wrong")
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	
	// Chain with Recovery
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

	// Verify metrics were incremented
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
				// Return transient retryable error for the first 2 calls
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

	// Consume 2 tokens immediately (capacity = 2)
	if !tb.Allow() {
		t.Error("expected first token to be allowed")
	}
	if !tb.Allow() {
		t.Error("expected second token to be allowed")
	}
	if tb.Allow() {
		t.Error("expected third token to be rate-limited (empty bucket)")
	}

	// Wait for refill
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err != nil {
		t.Errorf("expected wait to succeed after refill, got err: %v", err)
	}
}
```

## Verify
```bash
go test -v -race -count=1 ./sdk/middleware/...
```

## Checklist
- [ ] File `sdk/middleware/middleware_test.go` tồn tại
- [ ] Package name: `middleware_test`
- [ ] Test `TestAgentMiddleware_Recovery` xác thực chuyển panic sang error
- [ ] Test `TestAgentMiddleware_Metrics` kiểm tra thu thập metric
- [ ] Test `TestProviderMiddleware_Retry` kiểm tra gọi lại transient errors thành công
- [ ] Test `TestTokenBucket_RateLimiting` kiểm tra giới hạn giỏ thẻ Allow và Wait thành công
- [ ] `go test -v -race ./sdk/middleware/...` trả về ALL PASS
