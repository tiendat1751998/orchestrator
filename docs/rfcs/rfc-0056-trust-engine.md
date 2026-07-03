# RFC-0056: Trust Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0010 (Cognitive Layer), RFC-0045 (Digital Workforce)

## Summary

This RFC specifies the design of the **Trust Engine** in AEOS. It is responsible for dynamically auditing and rating LLM provider models (e.g. Claude, Gemini, OpenAI) based on their real-world validation pass rates in the Truth Pipeline, preventing planning drift caused by outdated or failing APIs.

## Motivation

Static trust ratings for AI models (e.g., assuming Claude is always a 9.8 and Gemini is 9.2) get stale quickly as providers update models weekly.
- A model version update might introduce regressions, causing compile failures in Go code generation.
- By dynamically tracking verification pass rates, the Trust Engine adjusts weights in real-time, preventing the Planner from routing tasks to failing models.

## Design

### 1. Architectural Placement

The Trust Engine runs inside the `Brain Runtime`, feeding provider ratings directly to the Planner's scoring function.

```
  Execution Output ──► Truth Pipeline ──► [Trust Engine Audit] ──► Update Provider Weights
```

---

### 2. Contracts (`contracts/brain/trust.go`)

```go
package brain

import "context"

// ProviderTrust represents the audited reliability of a model.
type ProviderTrust struct {
	ModelName    string  `json:"model_name"` // "gemini-1.5-pro", "claude-3-5-sonnet"
	TaskType     string  `json:"task_type"`  // "code_generation", "code_review"
	SuccessRuns  int     `json:"success_runs"`
	TotalRuns    int     `json:"total_runs"`
	TrustRating  float64 `json:"trust_rating"` // Dynamic float [0.0, 1.0]
}

// TrustEngine updates and queries provider trust ratings.
type TrustEngine interface {
	// AuditRecord logs the outcome of a provider execution.
	AuditRecord(ctx context.Context, modelName string, taskType string, success bool) error
	
	// GetTrustRating retrieves the current trust rating of a model.
	GetTrustRating(ctx context.Context, modelName string, taskType string) (*ProviderTrust, error)
}
```

## Impact

- **Self-Healing Routing**: If Gemini's compilation pass rate drops due to an upstream API outage, the Trust Engine adjusts its score down, forcing the Planner to automatically route subsequent tasks to Claude until the outage is resolved.
- **Dynamic Optimization**: The system protects itself from model deprecations and provider updates automatically.

## Open Questions

1. **How do we handle newly registered models?**
   - New models start with a default trust rating of 0.80. The UCB-1 exploration factor (RFC-0039) forces the system to run trials on the new model to compile actual pass rate data.
