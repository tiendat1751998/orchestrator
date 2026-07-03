package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/kernel/metrics"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

// ProviderMiddleware wraps a provider.Provider to extend its behavior.
type ProviderMiddleware func(provider.Provider) provider.Provider

// ChainProvider wraps a provider with a slice of middlewares.
// Middlewares are executed in the order they are passed (left-to-right).
func ChainProvider(p provider.Provider, mws ...ProviderMiddleware) provider.Provider {
	for i := len(mws) - 1; i >= 0; i-- {
		p = mws[i](p)
	}
	return p
}

type wrappedProvider struct {
	provider.Provider
	sendFn func(ctx context.Context, req *provider.Request) (*provider.Response, error)
}

func (w *wrappedProvider) Send(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	return w.sendFn(ctx, req)
}

// =============================================================================
// 1. Provider Logging Middleware
// =============================================================================

// ProviderLogging returns a middleware that logs outgoing provider request parameters and metrics.
func ProviderLogging(logger *slog.Logger) ProviderMiddleware {
	return func(next provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: next,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				startTime := time.Now()

				if logger != nil {
					logger.Debug("sending request to model provider",
						"provider", next.Name(),
						"model", req.Model,
						"messages", len(req.Messages),
					)
				}

				resp, err := next.Send(ctx, req)

				duration := time.Since(startTime)
				if err != nil {
					if logger != nil {
						logger.Error("provider request failed",
							"provider", next.Name(),
							"model", req.Model,
							"duration", duration.String(),
							"error", err.Error(),
						)
					}
				} else {
					if logger != nil {
						logger.Info("provider request succeeded",
							"provider", next.Name(),
							"model", resp.Model,
							"duration", duration.String(),
							"prompt_tokens", resp.Usage.PromptTokens,
							"completion_tokens", resp.Usage.CompletionTokens,
							"total_tokens", resp.Usage.TotalTokens,
						)
					}
				}

				return resp, err
			},
		}
	}
}

// =============================================================================
// 2. Provider Retry Middleware
// =============================================================================

// ProviderRetry returns a middleware that automatically retries failed transient calls.
func ProviderRetry(cfg resilience.RetryConfig) ProviderMiddleware {
	return func(next provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: next,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				return resilience.RetryWithResult(ctx, cfg, func() (*provider.Response, error) {
					return next.Send(ctx, req)
				})
			},
		}
	}
}

// =============================================================================
// 3. Provider Circuit Breaker Middleware
// =============================================================================

// ProviderCircuitBreaker returns a middleware that blocks API calls if the provider fails repeatedly.
func ProviderCircuitBreaker(cb *resilience.CircuitBreaker) ProviderMiddleware {
	return func(next provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: next,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				var resp *provider.Response
				err := cb.Execute(ctx, func() error {
					var sendErr error
					resp, sendErr = next.Send(ctx, req)
					return sendErr
				})
				return resp, err
			},
		}
	}
}

// =============================================================================
// 4. Provider Metrics Middleware
// =============================================================================

// ProviderMetrics returns a middleware that tracks token usage and request latency distribution.
func ProviderMetrics(registry *metrics.Registry) ProviderMiddleware {
	return func(next provider.Provider) provider.Provider {
		return &wrappedProvider{
			Provider: next,
			sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
				startTime := time.Now()

				resp, err := next.Send(ctx, req)

				duration := time.Since(startTime).Seconds()
				providerName := next.Name()

				// Log metrics to registry
				registry.Counter(fmt.Sprintf("provider_%s_requests_total", providerName)).Inc()
				registry.Histogram(fmt.Sprintf("provider_%s_latency_seconds", providerName)).Observe(duration)

				if err != nil {
					registry.Counter(fmt.Sprintf("provider_%s_errors_total", providerName)).Inc()
				}

				if resp != nil {
					registry.Counter("orchestrator_provider_prompt_tokens_total").Add(float64(resp.Usage.PromptTokens))
					registry.Counter("orchestrator_provider_completion_tokens_total").Add(float64(resp.Usage.CompletionTokens))
					registry.Counter("orchestrator_provider_tokens_total").Add(float64(resp.Usage.TotalTokens))
				}

				return resp, err
			},
		}
	}
}
