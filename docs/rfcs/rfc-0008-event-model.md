# RFC-0008: Event Model & Event Sourcing

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (State Machine), RFC-0003 (Knowledge Engine), RFC-0005 (Memory Model)

## Summary

This RFC specifies the Event Model, event schema serialization rules, database storage layouts, and Event Sourcing recovery pipeline for the AI Engineering Operating System (AEOS). All system transitions publish immutable events, enabling auditing, real-time reactive event loops, and full state reconstruction of active sessions on system crash.

## Motivation

Without a unified Event Model:
- **State Drift**: Database records might fall out of sync with active in-memory State Machine configurations.
- **Corrupt Reconstruct**: Resuming from system crash requires fragile ad-hoc logic per agent.
- **Tight Coupling**: Sub-components (like the UI dashboard or cognitive learning loop) must poll or hook directly into runtime execution blocks instead of listening to a clean event stream.

By using Event Sourcing, the system's state is defined as the *accumulation of past events*. State reconstruction is fully deterministic: we load the initial state and replay the historical event stream.

## Design

### 1. Unified Event Schema (`contracts/event/`)

All events flowing through the EventBus and written to the History Timeline must satisfy a single schema structure:

```go
// contracts/event/event.go — Expanded Event Model
package event

import (
	"context"
	"time"
)

// Aligned Event Store event type constants
const (
	TypeMissionCreated   = "mission.created"
	TypeMissionScheduled = "mission.scheduled"
	TypeMissionCompleted = "mission.completed"
	TypeMissionFailed    = "mission.failed"

	TypeTaskAssigned  = "task.assigned"
	TypeTaskStarted   = "task.started"
	TypeTaskCompleted = "task.completed"
	TypeTaskFailed    = "task.failed"
	TypeTaskRetried   = "task.retried"

	TypeAgentLeveledUp = "agent.leveled_up"
	TypeDecisionMade   = "decision.made"
	TypeSafetyAlerted  = "safety.alerted"
)

// Event represents an immutable occurrence inside the system.
type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	// Type represents the hierarchical dot-notation topic (e.g. "mission.state.running").
	Type      string    `json:"type"`
	// Source indicates origin identifier (e.g. "fsm:mission:m-102").
	Source    string    `json:"source"`
	// Entity specifies the category type (e.g. "mission", "task", "agent", "provider").
	Entity    string    `json:"entity"`
	// EntityID specifies the unique identifier for the targeted entity.
	EntityID  string    `json:"entity_id"`
	// Sequence is a monotonically increasing counter per EntityID to ensure order.
	Sequence  int64     `json:"sequence"`
	// Payload carries event-specific data fields.
	Payload   any       `json:"payload"`
}

// Handler handles received events.
type Handler func(ctx context.Context, ev Event)

// Bus provides real-time Publish/Subscribe capability.
type Bus interface {
	Publish(ctx context.Context, ev Event) error
	// Subscribe registers a subscriber. Supports dot-notation wildcards (e.g. "*.state.*").
	Subscribe(pattern string, h Handler) error
	Close() error
}
```

---

### 2. Database Schema (SQLite Layout)

The storage adapter persists the **History Timeline** (events) and the **Knowledge Graph** (nodes/edges) in a local SQLite file:

```sql
-- SQLite Schema layout for AEOS Memory Store

-- 1. History Timeline (Immutable Event Store)
CREATE TABLE IF NOT EXISTS timeline (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    event_type TEXT NOT NULL,
    source TEXT NOT NULL,
    entity TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    sequence INTEGER NOT NULL,
    payload TEXT NOT NULL, -- JSON string representation
    UNIQUE(entity_id, sequence) -- Enforce Event Sourcing sequence checks
);
CREATE INDEX IF NOT EXISTS idx_timeline_entity ON timeline(entity, entity_id);
CREATE INDEX IF NOT EXISTS idx_timeline_timestamp ON timeline(timestamp);

-- 2. Knowledge Graph: Nodes
CREATE TABLE IF NOT EXISTS knowledge_nodes (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL, -- JSON string
    tags TEXT NOT NULL,    -- JSON list of strings (e.g. '["go", "api"]')
    score REAL DEFAULT 0.0,
    used_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_knodes_type ON knowledge_nodes(type);

-- 3. Knowledge Graph: Edges
CREATE TABLE IF NOT EXISTS knowledge_edges (
    from_id TEXT NOT NULL,
    to_id TEXT NOT NULL,
    relation TEXT NOT NULL,
    weight REAL DEFAULT 1.0,
    PRIMARY KEY (from_id, to_id, relation),
    FOREIGN KEY(from_id) REFERENCES knowledge_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY(to_id) REFERENCES knowledge_nodes(id) ON DELETE CASCADE
);

-- 4. Artifact Metadata Store
CREATE TABLE IF NOT EXISTS artifact_metadata (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL,
    task_id TEXT,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_artifacts_mission ON artifact_metadata(mission_id);
```

---

### 3. Event Sourcing & State Reconstruction

When the system boots after a crash, the **Execution Runtime** reconstructs the State Machine configurations of running entities using the following sequence:

```
                  System Boot (Post Crash)
                             │
                             ▼
  ┌────────────────────────────────────────────────────────┐
  │ 1. Load active Mission IDs (query timeline for         │
  │    unfinished missions: no "mission.state.completed")  │
  └──────────────────────────┬─────────────────────────────┘
                             │
                             ▼
  ┌────────────────────────────────────────────────────────┐
  │ 2. For each active Mission ID:                         │
  │    timeline.Replay(ctx, startTime, Filter{ID})         │
  └──────────────────────────┬─────────────────────────────┘
                             │ Stream of Events
                             ▼
  ┌────────────────────────────────────────────────────────┐
  │ 3. Apply events sequentially to target state machines  │
  │    - Mission FSM receives transition records           │
  │    - Child Tasks receive task states                   │
  └──────────────────────────┬─────────────────────────────┘
                             │
                             ▼
  ┌────────────────────────────────────────────────────────┐
  │ 4. Reconstruct Working Memory from last cached         │
  │    snapshot event                                      │
  └──────────────────────────┬─────────────────────────────┘
                             │
                             ▼
                  Resume Execution Loop
```

```go
// kernel/execution/reconstructor.go
package execution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/history"
)

// StateReconstructor rebuilds entity state machines from event log history.
type StateReconstructor struct {
	timeline history.Timeline
}

func NewReconstructor(t history.Timeline) *StateReconstructor {
	return &StateReconstructor{timeline: t}
}

// ReconstructEntity replays events for a specific entity ID to restore its FSM to the exact pre-crash state.
func (r *StateReconstructor) ReconstructEntity(ctx context.Context, entityType string, entityID string, machine fsm.Machine) error {
	// Replay history from unix epoch
	iter, err := r.timeline.Replay(ctx, time.Unix(0, 0), history.TimelineQuery{
		Entity:   entityType,
		EntityID: entityID,
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.Next() {
		entry := iter.Entry()
		
		// Parse transition record from the raw data payload
		var record fsm.TransitionRecord
		byteData, err := json.Marshal(entry.Payload)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(byteData, &record); err != nil {
			return err
		}

		// Fire transition logic locally in memory without triggering FSM side-effect actions
		// (FSM implementation must support passive state updates during replays)
		if passiveMachine, ok := machine.(interface {
			SetStateDirectly(state fsm.State)
		}); ok {
			passiveMachine.SetStateDirectly(record.To)
		}
	}

	return iter.Err()
}
```

## Impact

- **Event Sourcing pipeline**: Guarantees recovery from crashes. State machines can be easily audited, replayed, or reproduced on mock environments.
- **Relational Integrity**: Enforces uniqueness using a composite index `(entity_id, sequence)` on the SQLite event store table.

## Open Questions

1. **Passive FSM transitions during replay**:
   - During reconstruction, `Fire()` should not trigger side-effect actions (like executing a tool or calling Gemini API again).
   - *Resolution*: The FSM kernel code implements `SetStateDirectly` or supports a `ReplayMode` flag that disables side-effect actions during event loading.
