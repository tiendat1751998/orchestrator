# RFC-0042: Product Manager Runtime

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0030 (Goal Engine), RFC-0040 (Intent Engine)

## Summary

This RFC specifies the design of the **Product Manager (PM) Runtime** in AEOS. The PM Runtime manages a specialized PM Agent that sits at the head of the execution chain, translating high-level business goals into concrete User Acceptance Tests (UATs) and verifying them against the final code outputs before declaring a mission complete.

## Motivation

AI code generation often compiles and passes unit tests, but fails to meet the actual business requirements or user expectations.
- A unit test might verify that `AddUser()` compiles and saves to a mock database, but fail to verify whether the actual user sign-up page works under realistic flow conditions.
- By introducing a Product Manager agent at the validation gate, AEOS ensures that engineered software matches user intent.

## Design

### 1. Architectural Placement

The PM Runtime is a cognitive service running within the `Brain Runtime`, executing both at mission startup (requirement analysis) and mission validation (UAT checks).

```
  Goal ──► [PM Agent: Requirement analysis] ──► Task Spec ──► Code Gen ──► [PM Agent: UAT check]
```

---

### 2. Contracts (`contracts/brain/pm.go`)

```go
package brain

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/goal"
)

// UATSpec represents the compiled business verification tests.
type UATSpec struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	TargetState string   `json:"target_state"`
}

// PMAgent handles requirement translation and validation.
type PMAgent interface {
	// CompileUATs generates user acceptance criteria from the goals.
	CompileUATs(ctx context.Context, g goal.Goal) ([]UATSpec, error)
	
	// VerifyUATs runs checks against the generated system workspace.
	VerifyUATs(ctx context.Context, specs []UATSpec) (bool, []string, error)
}
```

## Impact

- **True Alignment**: Verification is not just compile/test checks. The PM Agent runs UAT validation to ensure business objectives are achieved.
- **Improved DoD Gates**: The final Definition of Done only passes when both technical gates (linter, unit tests) and PM UAT gates pass.

## Open Questions

1. **How does the PM Agent execute UAT checks?**
   - The PM Agent uses sandbox execution tools to run browser subagents or call REST endpoints, verifying actual system behavior.
