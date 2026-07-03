# RFC-0040: Intent Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0030 (Goal Engine), RFC-0004 (Context Engine)

## Summary

This RFC specifies the design of the **Intent Engine** in AEOS. The Intent Engine parses high-level business goals, priorities, target constraints, and user expectations (e.g. "MVP", "Enterprise-scale", "Offline-first") to determine the overall plan strategy (e.g. choosing lightweight local code templates vs. complex distributed microservice setups).

## Motivation

A simple user command like "build a clone of Grab" can mean a local prototype (MVP) or an enterprise-grade high-scale system.
- If the system parses this goal blindly, the Planner might choose a microservices architecture that requires 20 servers, violating the developer's budget.
- By parsing business *intent* and constraints (budget, timeline, target platform), AEOS adjusts its planning scores to match reality.

## Design

### 1. Architectural Placement

The Intent Engine runs as a submodule of the `Goal Engine`, providing the Planner with structured intent metadata.

```
  Raw Input ──► [Intent Engine] ──► Intent Metadata (MVP, Target, Budget) ──► Planner
```

---

### 2. Contracts (`contracts/goal/intent.go`)

```go
package goal

import "context"

// IntentMetadata representsparsed business priorities.
type IntentMetadata struct {
	ScaleTarget string   `json:"scale_target"` // "MVP", "Departmental", "Enterprise"
	Priority    string   `json:"priority"`     // "speed", "quality", "cost"
	TargetOS    string   `json:"target_os"`     // "android", "ios", "web", "linux"
	Platforms   []string `json:"platforms"`
}

// IntentEngine parses business constraints.
type IntentEngine interface {
	// ParseIntent extracts business intent properties from the raw input.
	ParseIntent(ctx context.Context, input string) (*IntentMetadata, error)
}
```

## Impact

- **Strategic Planning**: If the intent is "MVP" and budget is low, the Planner automatically prunes expensive services and prioritizes simple, monolithic designs (e.g., Gin + SQLite).
- **Correct Scoping**: The generated Plan DAG matches the actual business context, preventing over-engineering.

## Open Questions

1. **How do we handle conflicts between constraints and intent?**
   - If the user requests an "Enterprise" scale Grab clone but sets a budget constraint of $10, the Intent Engine raises a validation error immediately before the mission starts.
