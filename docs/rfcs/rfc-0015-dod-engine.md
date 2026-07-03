# RFC-0015: Definition of Done (DoD) Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0014 (Quality Engine), RFC-0002 (Brain Architecture)

## Summary

This RFC specifies the design of the **Definition of Done (DoD) Engine** in AEOS. The DoD Engine is responsible for validating business and quality goals (e.g. compile pass, unit test coverage $\ge 85\%$, security scan clean) before declaring a mission complete. If criteria are unmet, it triggers replanning loops automatically.

## Motivation

AI agents often stop executing and declare a task complete when they finish generating text, regardless of whether the code compiles or meets requirements.
- By placing a formal DoD Engine at the exit boundary, AEOS guarantees that missions only terminate when defined business goals are fully verified.

## Design

### 1. Architectural Placement

The DoD Engine runs inside the `Brain Runtime` and acts as the final gate check for the mission lifecycle FSM.

```
  FSM: Running ──► [DoD Engine: Verify criteria] ──(Pass)──► Completed
                                 │
                               (Fail)
                                 ▼
                         Trigger Replan Loop
```

---

### 2. Contracts (`contracts/brain/dod.go`)

```go
package brain

import (
	"context"
)

// DoDCriteria represents configured goals.
type DoDCriteria struct {
	MinCoverage float64  `json:"min_coverage"` // e.g. 0.85
	RequireLint bool     `json:"require_lint"`
	CustomGates []string `json:"custom_gates,omitempty"`
}

// DoDEngine validates mission completion.
type DoDEngine interface {
	// IsDone checks if the workspace satisfies the DoD criteria.
	IsDone(ctx context.Context, criteria DoDCriteria) (bool, error)
}
```

## Impact

- **Reliability Guarantee**: Missions are never declared complete based on AI output assertions. Only deterministic Go validations satisfy the DoD Engine.
- **Auto-Correction**: Failed criteria automatically trigger the `Replan -> Execute -> Validate` loop.
