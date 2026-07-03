# RFC-0036: Mission Simulation & Dry-run Planner

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (Everything is State Machine), RFC-0001 (Kernel Architecture)

## Summary

This RFC specifies the design of the **Mission Simulation & Dry-run Planner** in AEOS. The Dry-run Planner executes a mock FSM transition loop over a candidate plan DAG before scheduling it on remote workers, detecting circular dependencies, deadlocks, expected execution duration, token costs, and resource constraints.

## Motivation

Deploying complex plans directly can lead to expensive failures and time-wasting loops.
- A Plan DAG might contain circular task dependencies that deadlock the scheduler.
- By dry-running the plan transitions in a virtual local sandbox, AEOS detects scheduling faults and validates cost limits before running code edits.

## Design

### 1. Architectural Placement

The Simulator is located within the `Brain Runtime` and acts as a pre-flight validator for candidate plans.

```
  Plan DAG ──► [Mission Simulator] ──(Validation Pass?)──► Deploy to Execution
```

---

### 2. Contracts (`contracts/brain/simulation.go`)

```go
package brain

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// SimulationResult represents the pre-flight audit report.
type SimulationResult struct {
	Valid          bool     `json:"valid"`
	ExpectedCost   float64  `json:"expected_cost_usd"`
	ExpectedTimeS  int      `json:"expected_time_seconds"`
	Deadlocks      []string `json:"deadlocks,omitempty"`
	CircularDeps   []string `json:"circular_dependencies,omitempty"`
}

// MissionSimulator simulates FSM transitions on a plan DAG.
type MissionSimulator interface {
	// Simulate dry-runs the DAG transitions in a read-only local sandbox.
	Simulate(ctx context.Context, dag fsm.DAG) (*SimulationResult, error)
}
```

## Impact

- **Zero-Cost Deadlock Detection**: Circular dependencies are caught in Go memory before any APIs are invoked or containers are spawned.
- **Budget Protection**: If the simulation projects that token costs will exceed the mission's cap, the Planner is forced to regenerate a cheaper plan DAG.

## Open Questions

1. **How do we simulate dynamic task outputs?**
   - The simulator uses historical averages from the Knowledge Graph to mock task outputs (e.g. assuming a Go compilation task takes 2.5 seconds and passes with 95% probability).
