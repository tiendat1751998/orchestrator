# RFC-0037: Adaptive Recovery & Replanning Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (Everything is State Machine), RFC-0010 (Cognitive Layer)

## Summary

This RFC specifies the design of the **Adaptive Recovery & Replanning Engine** in AEOS. When a task fails, instead of blindly retrying the same failed DAG path, the Replanning Engine intercepts the error, categorizes the failure pattern, and mutates the remaining plan DAG to inject corrective tasks.

## Motivation

Raw retries are highly ineffective when a task fails due to structural environment changes, library updates, or code logic errors.
- If a compilation fails because of a missing package dependency, retrying the compilation 3 times will yield the same failure.
- By dynamically replanning, AEOS can inject corrective tasks (e.g. `go get package`) into the execution queue before resuming compilation.

## Design

### 1. Architectural Placement

The Replanning Engine intercepts scheduler failures and mutates the running FSM state.

```
  Task Failure ──► [Replanning Engine] ──► Generate Corrective Sub-Graph ──► Merge DAG
```

---

### 2. Contracts (`contracts/brain/replanner.go`)

```go
package brain

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// FailureContext contains metadata about the task failure.
type FailureContext struct {
	TaskID    string `json:"task_id"`
	Category  string `json:"category"` // "infra", "compilation", "test_failure", "policy_violation"
	ErrorLog  string `json:"error_log"`
	Workspace string `json:"workspace_hash"`
}

// Replanner mutates plan DAGs on failure.
type Replanner interface {
	// Replan generates a corrective sub-graph to recover from a failure.
	Replan(ctx context.Context, fCtx FailureContext, activeDAG fsm.DAG) (fsm.DAG, error)
}
```

## Impact

- **Dynamic Recovery**: Failed build steps trigger automated debugging (e.g. running AST lookups or dependency checks) to fix code automatically.
- **Resilient Execution**: The orchestrator does not crash on task errors, preserving the mission lifecycle state.

## Open Questions

1. **How do we prevent infinite replanning loops?**
   - The FSM limits the maximum replanning cycles per mission (default: 5). If exceeded, the mission is marked as Failed, and control escalates to a Human-in-the-Loop review gate.
