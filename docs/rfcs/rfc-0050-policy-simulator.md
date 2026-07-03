# RFC-0050: Policy Simulator

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0012 (Security & Capability Model), RFC-0036 (Mission Simulator)

## Summary

This RFC specifies the design of the **Policy Simulator** in AEOS. The Policy Simulator dry-runs successful historical plan templates against newly updated Policy versions (e.g. banning root Docker access) to identify breaking constraints and regressions before applying changes to production.

## Motivation

Security policies change frequently.
- If a security administrator updates a policy to "deny-all-network-access", existing, validated execution templates will break, causing production failures.
- By dry-running historical plans against the new policy version, AEOS can flag violations beforehand.

## Design

### 1. Architectural Placement

The Policy Simulator runs in the `Brain Runtime` and acts as an verification gate for security configuration changes.

```
  New Policy Rules ──► [Policy Simulator] ──► Parse Historical Plans ──► Flag Violations
```

---

### 2. Contracts (`contracts/security/simulator.go`)

```go
package security

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// PolicySimulationReport contains violation details.
type PolicySimulationReport struct {
	PolicyVersion   string   `json:"policy_version"`
	Valid           bool     `json:"valid"`
	ViolatedRules   []string `json:"violated_rules,omitempty"`
	BreakingPlans   []string `json:"breaking_plans,omitempty"`
}

// PolicySimulator simulates policy constraints on templates.
type PolicySimulator interface {
	// SimulatePolicy evaluates a new policy model against historical success logs.
	SimulatePolicy(ctx context.Context, rules string, history []fsm.TransitionRecord) (*PolicySimulationReport, error)
}
```

## Impact

- **Safe Policy Deployments**: New security profiles are validated against a suite of standard workspace scenarios, preventing system-wide execution blocks.
- **Accurate Auditing**: Security teams get immediate reports detailing which development workflows violate new organizational guidelines.

## Open Questions

1. **How do we handle false positives?**
   - The simulator runs sandbox integration tests (Axiom 16) with intentional violations to verify that the policy blocks them accurately.
