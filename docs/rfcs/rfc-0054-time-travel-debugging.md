# RFC-0054: Time Travel Debugging

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0008 (Event Model), RFC-0005 (Memory Model)

## Summary

This RFC specifies the design of the **Time Travel Debugging** module in AEOS. To protect state determinism under Event Sourcing, the system implements **Read-Only Time Travel Inspection**. The FSM can be replayed to any step $N$ by applying events $1$ to $N$ from the Event Store, loading a read-only memory snapshot, and presenting it to the Web UI for analysis without permitting active writes or mutations that would corrupt the historical replay log.

## Motivation

Debugging automated agent executions is extremely difficult.
- If a plan fails at step 34, simply reading logs does not tell the developer what was in the active memory or the Knowledge Graph.
- However, permitting active mutations during playback causes the historical path to diverge (replay drift), making the replay non-replayable.
- Read-only time-travel inspection resolves this by allowing developers to inspect variable states at any point in history safely.

## Design

### 1. Architectural Placement

The Debugger reads events from the Event Store and populates a read-only memory canvas displayed on the Web UI.

```
  Event Store (Events 1 to N) ──► Replay FSM ──► Read-Only Snapshot ──► Web UI Inspector
```

---

### 2. Contracts (`contracts/history/debug.go`)

```go
package history

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// DebugFrame represents the state of memory at a specific step.
type DebugFrame struct {
	StepIndex     int                    `json:"step_index"`
	FSMState      string                 `json:"fsm_state"`
	MemoryState   map[string]interface{} `json:"memory_state"`
	ActiveGraph   string                 `json:"active_graph"`
}

// TimeTravelDebugger replays and inspects FSM frames.
type TimeTravelDebugger interface {
	// GetFrame reconstructs the state of a mission at step index N.
	GetFrame(ctx context.Context, missionID string, stepIndex int) (*DebugFrame, error)
}
```

## Impact

- **Accurate Fault Inspection**: Developers can step through historical runs frame-by-frame on the Web UI, inspecting memory variables and planner states at the exact moment of failure.
- **Protected History**: The Event Store remains immutable, preventing replay drift.

## Open Questions

1. **How do we handle large workspace files during time-travel?**
   - The Workspace Engine creates lightweight Git commits/stashes at each step, allowing the Git history to be synced with the FSM step index for full source inspection.
