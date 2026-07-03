# RFC-0019: ADR (Architecture Decision Record) & Policy Versioning

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0004 (Context Engine), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design of the **ADR (Architecture Decision Record) & Policy Versioning** module in AEOS. It manages version-controlled system policies and architecture records, ensuring that policy updates (e.g. strict linter checks) are applied cleanly without breaking existing, running mission execution paths.

## Motivation

Security policies and code guidelines evolve as teams grow.
- If a security policy is updated (e.g., banning root user containers) mid-mission, running tasks might fail unexpectedly.
- By versioning policies alongside code and binding them to the Mission context, we guarantee execution consistency.

## Design

### 1. Architectural Placement

Policies are loaded by the `Context Engine` at mission start, and validated by the `Security Capability Model` during task execution.

```
  Policy file (v1.2) ──► [Context Engine] ──► Bind to Mission ──► Security checks
```

---

### 2. Contracts (`contracts/context/policy.go`)

```go
package context

import "context"

// PolicyRule represents a single security constraint.
type PolicyRule struct {
	ID        string `json:"id"`
	Action    string `json:"action"` // "allow", "deny"
	Resource  string `json:"resource"` // "network", "file_read", "exec"
	Condition string `json:"condition,omitempty"`
}

// PolicySet represents a versioned configuration.
type PolicySet struct {
	Version string       `json:"version"`
	Rules   []PolicyRule `json:"rules"`
}

// PolicyManager coordinates policy loading.
type PolicyManager interface {
	// GetPolicy retrieves a versioned policy configuration.
	GetPolicy(ctx context.Context, version string) (*PolicySet, error)
}
```

## Impact

- **Consistent Execution**: Missions utilize the exact policy version they started with, protecting active pipelines from unexpected policy drift.
- **Traceable Decisions**: Policy versions and decisions are recorded in the Event Store for auditing.
