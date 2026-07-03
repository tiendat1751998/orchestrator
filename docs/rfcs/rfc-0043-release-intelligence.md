# RFC-0043: Release Intelligence

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0016 (Observation Runtime), RFC-0008 (Event Model)

## Summary

This RFC specifies the design of the **Release Intelligence** module in AEOS. It connects the system's learning loop to production deployment webhooks, canary metrics, alert channels, and rollbacks (`Code -> Build -> Deploy -> Canary -> Metrics -> Rollback -> Planner -> Learn`), allowing production outcomes to serve as active learning data.

## Motivation

The lifecycle of an application does not end when code is generated and passes local tests.
- A deployment might crash in production due to high concurrency, connection leaks, or scaling latency.
- By gathering telemetry and metrics from live canaries, AEOS can capture production failure logs and feed them back to the Planner, preventing similar deployment layouts in future missions.

## Design

### 1. Architectural Placement

Release Intelligence is part of the `Observation Runtime`, running in the background to gather metrics and trigger rollbacks on failure events.

```
  Canary Deploy ──► [Release Intelligence (Telemetry)] ──(Error Spikes?)──► Trigger Rollback
                                         │
                                         ▼
                            Feed Failure log to Replanner
```

---

### 2. Contracts (`contracts/release/`)

```go
package release

import "context"

// CanaryMetrics represents live performance telemetry.
type CanaryMetrics struct {
	DeployID     string  `json:"deploy_id"`
	ErrorRate5xx float64 `json:"error_rate_5xx"` // [0.0, 1.0]
	LatencyP99Ms int     `json:"latency_p99_ms"`
	CPUUsagePct  float64 `json:"cpu_usage_pct"`
}

// ReleaseEngine manages canary rollouts and production feedback.
type ReleaseEngine interface {
	// MonitorCanary reads telemetry data from a live deployment.
	MonitorCanary(ctx context.Context, deployID string) (*CanaryMetrics, error)
	
	// Rollback triggers an immediate deployment rollback on failure.
	Rollback(ctx context.Context, deployID string) error
}
```

## Impact

- **Production-Driven Learning**: Production failures are captured and saved in Episodic Memory, ensuring the Planner learns from real-world scaling issues.
- **Automated Operations**: The system handles canary gating, monitoring, and rollbacks without manual human intervention.

## Open Questions

1. **How does the system authenticate with production cloud providers?**
   - Credentials are isolated and managed securely by the Security Capability Model (RFC-0012) using read-only capability tokens.
