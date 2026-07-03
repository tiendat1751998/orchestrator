# RFC-0000: Everything is State Machine

- **Status**: PROPOSED → **REVISED**
- **Priority**: P0 — Foundation
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Revised**: 2026-07-03 (Review fixes: A3, C1, D1)

## Summary

Every long-lived entity in the system (Mission, Task, Agent, Provider, Session, Plugin) MUST be modeled as a Finite State Machine (FSM). State transitions are the ONLY way entities change state. All transitions are logged, enabling recovery, audit, replay, and debug.

## Motivation

Without state machines:
- **Recovery** is ad-hoc: after crash, we don't know where entities are
- **Audit** is incomplete: we log actions but not state transitions
- **Replay** is impossible: can't reconstruct system state from logs
- **Debug** is painful: "is this task running or stuck?"
- **Concurrency** is dangerous: multiple goroutines changing state without coordination

With state machines:
- **Recovery**: Read last persisted state → resume from there
- **Audit**: Every transition = event record (who, what, when, why)
- **Replay**: Re-apply transitions from event log → reconstruct any point in time
- **Debug**: `entity.Current()` → instant answer
- **Concurrency**: State machine enforces valid transitions atomically

## Design

### Core Abstraction

```go
// contracts/fsm/machine.go
package fsm

import (
    "context"
    "time"
)

// State represents a named state in a state machine.
type State string

// TransitionRecord is an immutable record of a state transition.
// Used for audit logging, replay, and debugging.
//
// TransitionRecords are automatically persisted to history.Timeline
// by the FSM's OnTransition callback (see Integration with History below).
type TransitionRecord struct {
    // EntityType identifies the state machine type (e.g., "mission", "task").
    EntityType string    `json:"entity_type"`
    // EntityID identifies the specific entity instance.
    EntityID   string    `json:"entity_id"`
    From       State     `json:"from"`
    To         State     `json:"to"`
    Event      string    `json:"event"`
    Timestamp  time.Time `json:"timestamp"`
    Payload    any       `json:"payload,omitempty"`
    Error      string    `json:"error,omitempty"`
}

// Transition defines a valid state change.
//
// Note on function fields (Guard, Action):
// These are NOT serializable — they are compile-time Go functions.
// Only TransitionRecord (the output) needs serialization.
// If hot-reloadable rules are needed in the future, switch to
// expression-based guards (e.g., CEL or Rego). For v1, function
// fields are pragmatic and fast.
type Transition struct {
    // From is the source state.
    From State
    // To is the target state.
    To State
    // Event is the trigger name (e.g., "start", "complete", "fail").
    Event string
    // Guard is an optional condition that must be true for the transition.
    // If nil, the transition is always allowed.
    // MUST be a pure function (no I/O, no side effects).
    Guard func(ctx context.Context, payload any) bool
    // Action is an optional side effect executed during the transition.
    // If it returns error, the transition is rolled back (state NOT changed).
    // May perform I/O (e.g., dispatch task, store artifact).
    Action func(ctx context.Context, payload any) error
}

// Machine is a finite state machine.
//
// Thread-safety: All methods MUST be safe for concurrent use.
// Implementations use sync.Mutex to serialize transitions.
//
// Usage:
//
//   m := fsm.New(def)
//   err := m.Fire(ctx, "start", nil)    // pending → running
//   err = m.Fire(ctx, "complete", nil)  // running → completed
//   state := m.Current()                // "completed"
//   history := m.History()              // [{pending→running}, {running→completed}]
type Machine interface {
    // Current returns the current state.
    Current() State

    // Can returns true if the given event can fire from the current state.
    Can(event string) bool

    // AvailableEvents returns all events that can fire from the current state.
    AvailableEvents() []string

    // Fire triggers a state transition.
    //
    // Sequence:
    //   1. Find matching transition (From == current, Event == event)
    //   2. If Guard exists, evaluate it
    //   3. If Action exists, execute it
    //   4. Update state to transition.To
    //   5. Append TransitionRecord to in-memory history
    //   6. Call OnTransition callback (which publishes to EventBus + History)
    //
    // Returns error if:
    //   - No matching transition exists (ErrInvalidTransition)
    //   - Guard returns false (ErrGuardRejected)
    //   - Action returns error (state is NOT changed)
    Fire(ctx context.Context, event string, payload any) error

    // History returns all transition records in chronological order.
    // This is the in-memory history (bounded). For full history, query
    // history.Timeline.
    History() []TransitionRecord
}

// OnTransition is a callback invoked after every successful transition.
// Used for:
//   1. Publishing to EventBus → other components react to state changes
//   2. Appending to history.Timeline → persistent audit trail
//   3. Logging and metrics
type OnTransition func(record TransitionRecord)

// Definition defines a state machine's structure.
type Definition struct {
    // Name identifies this state machine type (e.g., "mission", "task").
    Name string
    // Initial is the starting state.
    Initial State
    // Transitions are all valid state changes.
    Transitions []Transition
    // OnTransition is called after every successful transition (optional).
    // Multiple callbacks can be composed by the caller.
    OnTransition OnTransition
    // HistoryLimit caps in-memory history. Default: 100.
    // Full history is persisted via OnTransition → history.Timeline.
    HistoryLimit int
}
```

### Minimal EventBus Interface (Unblocks FSM integration)

> [!NOTE]
> Full event model is defined in RFC-0008. This is the minimal interface needed for FSM integration.

```go
// contracts/event/event.go — Minimal EventBus for FSM integration
package event

import "context"

// Event is a message published to the event bus.
type Event struct {
    // Type identifies the event (e.g., "mission.state.running").
    // Convention: {entity}.state.{new_state}
    Type    string `json:"type"`
    // Source identifies the origin (e.g., "fsm:mission:mission-123").
    Source  string `json:"source"`
    // Payload carries event data (typically a TransitionRecord).
    Payload any    `json:"payload"`
}

// Bus is a publish-subscribe event bus.
// Thread-safe: all methods must be safe for concurrent use.
type Bus interface {
    // Publish sends an event to all subscribers matching the event type.
    Publish(ctx context.Context, event Event) error
    // Subscribe registers a handler for events matching the type pattern.
    // Pattern supports wildcard: "mission.state.*" matches all mission state events.
    Subscribe(eventType string, handler func(ctx context.Context, event Event)) error
}
```

### Integration: FSM → EventBus → History (Fix for Issue A3)

> [!IMPORTANT]
> Every FSM transition **automatically** publishes to EventBus AND persists to History.
> This is wired during kernel bootstrap — individual entities do NOT need to handle this.

```go
// kernel/kernel.go — during bootstrap
func wireStateHistory(bus event.Bus, timeline history.Timeline) fsm.OnTransition {
    return func(record fsm.TransitionRecord) {
        ctx := context.Background()
        
        // 1. Publish to EventBus (real-time notification)
        bus.Publish(ctx, event.Event{
            Type:    record.EntityType + ".state." + string(record.To),
            Source:  "fsm:" + record.EntityType + ":" + record.EntityID,
            Payload: record,
        })
        
        // 2. Append to History (persistent audit trail)
        timeline.Append(ctx, history.Entry{
            ID:        generateID(),
            Timestamp: record.Timestamp,
            Entity:    record.EntityType,
            EntityID:  record.EntityID,
            Event:     "state." + record.Event,
            Data:      record,
        })
    }
}
```

**Flow**: `entity.Fire()` → `OnTransition(record)` → `EventBus.Publish()` + `Timeline.Append()`

### State Machine Definitions

#### Mission States

```
Created → Planning → Scheduled → Running → Reviewing → Completed
    ↓        ↓                      ↓          ↓
  Failed   Failed                 Failed    Failed
                                    ↓
                                 Cancelled
```

| From | Event | To | Guard | Action |
|---|---|---|---|---|
| Created | plan | Planning | — | Call Brain.Plan() |
| Created | fail | Failed | — | Log invalid mission |
| Planning | schedule | Scheduled | Plan is valid | Submit to Scheduler |
| Planning | fail | Failed | — | Log planning error |
| Scheduled | start | Running | Workers available | Dispatch first tasks |
| Running | review | Reviewing | All tasks done | Aggregate results |
| Running | fail | Failed | Critical task failed | Cancel remaining |
| Running | cancel | Cancelled | — | Cancel all in-flight |
| Reviewing | complete | Completed | Quality check pass | Store artifacts |
| Reviewing | fail | Failed | Quality check fail | Log quality error |

#### Task States

```
Pending → Assigned → Running → Completed
                       ↓
                     Failed → Retrying → Running
                       ↓
                    Skipped
```

| From | Event | To | Guard | Action |
|---|---|---|---|---|
| Pending | assign | Assigned | Agent available | Reserve agent |
| Assigned | start | Running | — | Execute task |
| Running | complete | Completed | — | Collect result |
| Running | fail | Failed | — | Record error |
| Failed | retry | Retrying | retry_count < max | Increment retry |
| Failed | skip | Skipped | Task not critical | Mark skipped |
| Retrying | start | Running | — | Re-execute |

#### Agent States

```
Idle → Busy → Idle
  ↓      ↓
Stopped  Waiting → Busy
```

| From | Event | To |
|---|---|---|
| Idle | assign_task | Busy |
| Busy | task_complete | Idle |
| Busy | wait_dependency | Waiting |
| Waiting | dependency_ready | Busy |
| Idle | stop | Stopped |
| Busy | stop | Stopped |

#### Provider States

```
Healthy → Degraded → Offline → Healthy
```

| From | Event | To | Guard |
|---|---|---|---|
| Healthy | error | Degraded | error_rate > threshold |
| Degraded | recover | Healthy | error_rate < threshold |
| Degraded | circuit_open | Offline | consecutive_failures > limit |
| Offline | circuit_half_open | Degraded | cooldown_expired |

#### Session States

```
Created → Streaming → Paused → Resumed → Streaming
                        ↓                    ↓
                      Closed               Closed
```

#### Plugin States

```
Registered → Initializing → Ready → Running → Stopping → Stopped
                              ↓                   ↓
                            Failed              Failed
```

#### Kernel States (new — from Issue C3)

```
Created → Booting → Running → ShuttingDown → Stopped
             ↓                     ↓
           Failed                Failed
             ↓
          Recovering → Booting
```

### Implementation

```go
// kernel/fsm/machine.go — Generic FSM implementation
package fsm

import "sync"

type machine struct {
    mu          sync.Mutex
    definition  Definition
    current     State
    history     []TransitionRecord
}

func New(def Definition) Machine {
    limit := def.HistoryLimit
    if limit <= 0 {
        limit = 100
    }
    return &machine{
        definition: def,
        current:    def.Initial,
        history:    make([]TransitionRecord, 0, limit),
    }
}
```

The implementation uses `sync.Mutex` (not RWMutex):
- `Current()`, `Can()`, `AvailableEvents()` could use read lock, but simplicity wins
- `Fire()` takes exclusive lock (serializes transitions)
- `History()` returns a copy (no lock needed after copy)

### Persistence & Recovery

State machines are persisted by serializing:
1. `Current` state
2. Full history lives in `history.Timeline` (SQLite-backed)

On crash recovery:
1. Load last persisted state from storage
2. Reconstruct FSM with `current = persisted_state`
3. Resume execution from that state
4. History is reconstructed from `history.Timeline` if needed

## Impact

### New Packages
- `contracts/fsm/` — Machine, State, Transition, TransitionRecord
- `contracts/event/` — Minimal EventBus interface (expanded in RFC-0008)
- `kernel/fsm/` — Generic FSM implementation

### Modified Packages
- ALL entity structs (Mission, Task, Agent, Provider, Session, Plugin) embed an FSM
- Kernel bootstrap wires OnTransition → EventBus + History

### Layer Compliance
- `contracts/fsm/` imports only stdlib ✅
- `contracts/event/` imports only stdlib ✅
- `kernel/fsm/` imports only `contracts/fsm/` ✅

## Open Questions

1. ~~**History size limit**: Should history be bounded (ring buffer) or unbounded?~~ **RESOLVED**: In-memory bounded at 100 per entity (configurable via `Definition.HistoryLimit`). Full history persisted to `history.Timeline` via OnTransition callback.
2. **Distributed state**: For multi-node deployments, should FSM state be synchronized? Recommendation: local FSM + event log for replay. Defer distributed consensus to Phase 5+.
