package middleware

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/kernel/metrics"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

type mockProvider struct {
	name   string
	sendFn func(ctx context.Context, req *provider.Request) (*provider.Response, error)
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Send(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	return m.sendFn(ctx, req)
}
func (m *mockProvider) Stream(ctx context.Context, req *provider.Request) (<-chan provider.StreamChunk, error) {
	return nil, nil
}
func (m *mockProvider) IsAvailable(ctx context.Context) bool { return true }
func (m *mockProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func TestChainProvider(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return &provider.Response{
				Model: req.Model,
				Usage: provider.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}, nil
		},
	}

	var called []string
	mw1 := func(p provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: p,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				called = append(called, "mw1")
				return p.Send(ctx, req)
			},
		}
	}
	mw2 := func(p provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: p,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				called = append(called, "mw2")
				return p.Send(ctx, req)
			},
		}
	}

	chained := ChainProvider(mock, mw1, mw2)
	req := &provider.Request{Model: "mock-model"}
	_, err := chained.Send(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(called) != 2 || called[0] != "mw1" || called[1] != "mw2" {
		t.Errorf("expected middlewares to run left-to-right (mw1 then mw2), got: %v", called)
	}
}

func TestProviderLogging(t *testing.T) {
	mockSuccess := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return &provider.Response{
				Model: req.Model,
				Usage: provider.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}, nil
		},
	}

	mockFailure := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return nil, errors.New("provider failure")
		},
	}

	logger := slog.Default()
	mw := ProviderLogging(logger)

	// Test success path
	wrappedSuccess := mw(mockSuccess)
	_, err := wrappedSuccess.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test failure path
	wrappedFailure := mw(mockFailure)
	_, err = wrappedFailure.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestProviderRetry(t *testing.T) {
	retryableErr := contracts.NewRetryableError(errors.New("transient error"), 1*time.Millisecond)
	attempts := 0
	mock := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			attempts++
			if attempts < 2 {
				return nil, retryableErr
			}
			return &provider.Response{
				Model: req.Model,
				Usage: provider.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}, nil
		},
	}

	cfg := resilience.RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Millisecond,
		Multiplier:  1.0,
		Jitter:      false,
	}

	mw := ProviderRetry(cfg)
	wrapped := mw(mock)

	resp, err := wrapped.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("expected total tokens to be 30, got: %d", resp.Usage.TotalTokens)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got: %d", attempts)
	}
}

func TestProviderCircuitBreaker(t *testing.T) {
	retryableErr := contracts.NewRetryableError(errors.New("transient error"), 1*time.Millisecond)
	mock := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return nil, retryableErr
		},
	}

	cb := resilience.NewCircuitBreaker(resilience.CBConfig{
		MaxFailures: 2,
		Cooldown:    10 * time.Second,
	})

	mw := ProviderCircuitBreaker(cb)
	wrapped := mw(mock)

	// Call 1: fails, increases failure count to 1
	_, err := wrapped.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err == nil {
		t.Fatal("expected error")
	}

	// Call 2: fails, increases failure count to 2, trips breaker to Open
	_, err = wrapped.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err == nil {
		t.Fatal("expected error")
	}

	// Call 3: should fail fast because breaker is Open
	_, err = wrapped.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err == nil {
		t.Fatal("expected error")
	}
	if cb.State() != resilience.StateOpen {
		t.Errorf("expected breaker to be Open, got: %s", cb.State())
	}
}

func TestProviderMetrics(t *testing.T) {
	mockSuccess := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return &provider.Response{
				Model: req.Model,
				Usage: provider.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}, nil
		},
	}

	mockFailure := &mockProvider{
		name: "mock",
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return nil, errors.New("provider failure")
		},
	}

	reg := metrics.NewRegistry()
	mw := ProviderMetrics(reg)

	// Test success path
	wrappedSuccess := mw(mockSuccess)
	_, err := wrappedSuccess.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test failure path
	wrappedFailure := mw(mockFailure)
	_, err = wrappedFailure.Send(context.Background(), &provider.Request{Model: "mock-model"})
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	snapshot := reg.Snapshot()
	if val := snapshot["counter.provider_mock_requests_total"]; val != 2.0 {
		t.Errorf("expected provider_mock_requests_total to be 2.0, got: %v", val)
	}
	if val := snapshot["counter.provider_mock_errors_total"]; val != 1.0 {
		t.Errorf("expected provider_mock_errors_total to be 1.0, got: %v", val)
	}
	if val := snapshot["counter.orchestrator_provider_prompt_tokens_total"]; val != 10.0 {
		t.Errorf("expected prompt tokens to be 10.0, got: %v", val)
	}
	if val := snapshot["counter.orchestrator_provider_completion_tokens_total"]; val != 20.0 {
		t.Errorf("expected completion tokens to be 20.0, got: %v", val)
	}
	if val := snapshot["counter.orchestrator_provider_tokens_total"]; val != 30.0 {
		t.Errorf("expected total tokens to be 30.0, got: %v", val)
	}
}
