# RFC-0027: Resilience, Circuit Breakers & Backoffs

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0007 (Provider & Runtime Separation), RFC-0001 (Kernel Architecture)

## Summary

This RFC specifies the design of **Resilience, Circuit Breakers & Backoffs** in AEOS. It defines the fault-tolerance models for external API calls, wrapping LLM provider client calls with automatic exponential backoffs and circuit breakers to handle rate-limiting and temporary provider outages safely.

## Motivation

External LLM API endpoints (Anthropic, Gemini, OpenAI) frequently suffer from rate limits (HTTP 429), timeouts, and transient errors.
- If these errors are not caught and handled, they trigger mission aborts.
- Implementing standard circuit breakers and backoffs in the Go provider layer ensures execution resilience.

## Design

### 1. Architectural Placement

Resilience filters wrap all outgoing HTTP/gRPC API client calls in the Provider layer.

```
  Provider request ──► [Circuit Breaker / Backoff wrapper] ──► LLM API endpoint
```

---

### 2. Contracts (`contracts/provider/resilience.go`)

```go
package provider

import (
	"context"
	"time"
)

// ResilienceConfig configures retry policies.
type ResilienceConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialBackoff time.Duration `json:"initial_backoff"`
	MaxBackoff     time.Duration `json:"max_backoff"`
	Timeout        time.Duration `json:"timeout"`
}

// ResilientClient wraps API requests.
type ResilientClient interface {
	// Call executes an API request with backoffs and circuit breaking.
	Call(ctx context.Context, action func() error, config ResilienceConfig) error
}
```

## Impact

- **Transient Error Safety**: Temporary connection drops or rate-limit blocks are resolved automatically via exponential backoffs.
- **Fail-Fast Outages**: Persistent provider failures trigger circuit breakers, alerting the Trust Engine (RFC-0056) to fallback to alternative models immediately.
