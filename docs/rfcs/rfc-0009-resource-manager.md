# RFC-0009: Resource Manager

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0007 (Provider and Runtime Separation)

## Summary

This RFC specifies the architecture of the **Resource Manager** in the AI Engineering Operating System (AEOS). To ensure stability under heavy local workloads, the Resource Manager monitors CPU/RAM resource limits, tracks API token quotas, counts concurrent active provider sessions, and implements backpressure mechanics for the Scheduler.

## Motivation

Issue 7 from the architecture review noted the lack of a Resource Manager. AI agents are highly resource-intensive:
- Running multi-agent plans can spike local CPU/RAM (e.g. compiling code, running docker testing sandboxes).
- External LLM APIs enforce strict rate-limits (Requests-Per-Minute / RPM, Tokens-Per-Minute / TPM).
- Spawning too many concurrent tasks leads to rate-limiting errors or machine freezes.

Without a Resource Manager, the Scheduler executes tasks blindly, leading to cascading failures.

## Design

### 1. Resource Manager Structure

The Resource Manager runs inside the Execution Runtime and interacts with the Scheduler:

```
  ┌────────────────────────────────────────────────────────┐
  │                 Scheduler Queue                        │
  └──────────────────────────┬─────────────────────────────┘
                             │ Next Task Request
                             ▼
  ┌────────────────────────────────────────────────────────┐
  │ Resource Manager Check:                                │
  │   - CPU Usage < Threshold (e.g. 85%)?                  │
  │   - Memory Free > Threshold (e.g. 500MB)?              │
  │   - Provider API Token quota available?                │
  │   - Active sessions limit not exceeded?                │
  └──────────────────────────┬─────────────────────────────┘
                             │
            ┌────────────────┴────────────────┐
            ▼ Allowed                         ▼ Blocked / Backpressure
  ┌──────────────────┐               ┌──────────────────┐
  │ Dispatch Task    │               │ Delay / Reschedule│
  └──────────────────┘               └──────────────────┘
```

### 2. Contracts (`contracts/resource/`)

```go
// contracts/resource/resource.go
package resource

import (
	"context"
	"time"
)

// SystemUsage reports host hardware utilization metrics.
type SystemUsage struct {
	CPUPercent float64 `json:"cpu_percent"`
	RAMTotal   uint64  `json:"ram_total"` // in Bytes
	RAMFree    uint64  `json:"ram_free"`  // in Bytes
	Load1      float64 `json:"load_1"`
}

// QuotaState tracks API rate limits per Provider Model.
type QuotaState struct {
	ProviderName   string    `json:"provider_name"`
	ModelName      string    `json:"model_name"`
	RequestsLimit  int       `json:"requests_limit"`
	RequestsUsed   int       `json:"requests_used"`
	TokensLimit    int       `json:"tokens_limit"`
	TokensUsed     int       `json:"tokens_used"`
	CooldownUntil  time.Time `json:"cooldown_until"`
	ActiveSessions int       `json:"active_sessions"`
}

// ResourceRequest represents constraints required to execute a task.
type ResourceRequest struct {
	ProviderName    string  `json:"provider_name,omitempty"`
	ModelName       string  `json:"model_name,omitempty"`
	EstimatedTokens int     `json:"estimated_tokens,omitempty"`
	CPURequirement  float64 `json:"cpu_requirement,omitempty"` // Required CPU cores (e.g., 2.0)
	RAMRequirement  uint64  `json:"ram_requirement,omitempty"` // Required RAM in Bytes (e.g., 512 * 1024 * 1024)
}

// ResourceManager manages resource allocations and rate limit quotas.
type ResourceManager interface {
	// Allocate attempts to reserve resources for a task. 
	// Returns false if resource thresholds (CPU/RAM) or API quotas are exceeded.
	Allocate(ctx context.Context, req ResourceRequest) (bool, error)
	
	// Release frees resources and decreases active session counts.
	Release(ctx context.Context, req ResourceRequest) error
	
	// SystemMetrics returns current host hardware metrics.
	SystemMetrics(ctx context.Context) (SystemUsage, error)
	
	// ProviderQuotas returns list of current rate limit states.
	ProviderQuotas(ctx context.Context) ([]QuotaState, error)
}
```

---

### 3. Implementation Plan (`kernel/execution/resource/`)

#### A. Host Hardware Monitor
For Windows compatibility, System Metrics are calculated using standard libraries (`os` and fallback heuristics using `wmic` commands if external libraries are excluded from contracts, or via generic Go packages like `shirou/gopsutil` when imported in the implementation layer).

```go
// kernel/execution/resource/monitor.go
package resource

import (
	"context"
	"runtime"
	
	"github.com/tiendat1751998/orchestrator/contracts/resource"
)

type sysMonitor struct{}

func (m *sysMonitor) GetUsage(ctx context.Context) (resource.SystemUsage, error) {
	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)

	// memory metrics
	return resource.SystemUsage{
		CPUPercent: 0.0, // Calculated using platform-specific OS metrics
		RAMTotal:   mStats.Sys,
		RAMFree:    mStats.HeapIdle,
		Load1:      0.0,
	}, nil
}
```

#### B. Token Bucket Rate-Limiting Strategy
API token usage is tracked using a sliding window or Token Bucket algorithm:

```go
// kernel/execution/resource/quota.go
package resource

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillRate   float64 // tokens per second
	lastRefilled time.Time
}

func (tb *TokenBucket) Consume(amount float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefilled).Seconds()
	tb.lastRefilled = now

	tb.tokens = tb.tokens + (elapsed * tb.refillRate)
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= amount {
		tb.tokens -= amount
		return true
	}
	return false
}
```

If a Provider API returns an HTTP 429 Rate-Limit error:
1. The **Execution Runtime** catches the status code.
2. It sets a `CooldownUntil = time.Now().Add(retryAfterHeader)` inside the Resource Manager.
3. The **Scheduler** pauses dispatching tasks targeting that provider model until the cooldown timestamp has expired.

## Impact

- **Backpressure**: Prevents node overload and minimizes API rate-limit penalties.
- **Scheduler Integration**: Scheduler checks `Allocate()` before popping tasks from the queue.

## Open Questions

1. **How does the system calculate estimated tokens before calling API?**
   - The Context Engine provides the exact token size of the assembled prompt context. The Resource Manager adds a safety padding (e.g. 50% of the max output token limit) for token consumption checks.
