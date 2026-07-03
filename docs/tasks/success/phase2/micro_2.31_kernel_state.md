# Micro-Task 2.31: Create kernel/state.go

## Info
- **File**: `kernel/state.go`
- **Package**: `kernel`
- **Depends on**: none (standalone)
- **Time**: 15 min
- **Verify**: `go build ./kernel/...`

## Purpose
Implements the kernel runtime state machine (`State`, `StateMachine`, and constants) that monitors lifecycle transitions (Created → Initializing → Running → ShuttingDown → Stopped), preventing illegal execution calls (such as double start attempts).

## EXACT code to create

```go
// Package kernel provides the orchestrator core runtime.
//
// The kernel wires together all components (config, logger, eventbus, registry,
// runtime, scheduler) and manages their lifecycle.
package kernel

import (
	"fmt"
	"sync"
)

// State represents the kernel's lifecycle state.
type State int

const (
	// StateCreated is the initial state after New().
	// Valid transitions: → Initializing
	StateCreated State = iota

	// StateInitializing is the state during Init().
	// Plugins are being initialized.
	// Valid transitions: → Running (success) or → Stopped (failure)
	StateInitializing

	// StateRunning is the normal operating state.
	// Tasks can be submitted and executed.
	// Valid transitions: → ShuttingDown
	StateRunning

	// StateShuttingDown is the state during Stop().
	// New tasks are rejected. In-flight tasks are allowed to complete.
	// Valid transitions: → Stopped
	StateShuttingDown

	// StateStopped is the final state.
	// All resources are released. No further transitions.
	StateStopped
)

// String returns the human-readable state name.
func (s State) String() string {
	switch s {
	case StateCreated:
		return "created"
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StateShuttingDown:
		return "shutting_down"
	case StateStopped:
		return "stopped"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// validTransitions defines the allowed state transitions.
//
// State machine diagram:
//
//	Created → Initializing → Running → ShuttingDown → Stopped
//	                       ↘ Stopped (init failure)
var validTransitions = map[State][]State{
	StateCreated:      {StateInitializing},
	StateInitializing: {StateRunning, StateStopped},
	StateRunning:      {StateShuttingDown},
	StateShuttingDown: {StateStopped},
	StateStopped:      {},
}

// StateMachine manages kernel state transitions.
//
// Thread-safety: all methods are safe for concurrent use.
//
// Why not sync/atomic?
//   → Transition validation requires read-then-check-then-write.
//   → This is a compound operation that needs a mutex for atomicity.
//   → atomic only handles single reads/writes, not compound ops.
type StateMachine struct {
	mu      sync.RWMutex
	current State
}

// NewStateMachine creates a state machine in the Created state.
func NewStateMachine() *StateMachine {
	return &StateMachine{
		current: StateCreated,
	}
}

// Current returns the current state.
func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// Transition attempts to move to a new state.
//
// Returns error if the transition is not valid.
// Valid transitions are defined by validTransitions.
//
// Thread-safety: acquires write lock.
func (sm *StateMachine) Transition(to State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	allowed, exists := validTransitions[sm.current]
	if !exists {
		return fmt.Errorf("kernel: unknown state %v", sm.current)
	}

	for _, valid := range allowed {
		if valid == to {
			sm.current = to
			return nil
		}
	}

	return fmt.Errorf("kernel: invalid state transition %v → %v", sm.current, to)
}

// Is checks if the current state matches the given state.
func (sm *StateMachine) Is(state State) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current == state
}

// IsRunning returns true if the kernel is in Running state.
func (sm *StateMachine) IsRunning() bool {
	return sm.Is(StateRunning)
}

// IsStopped returns true if the kernel is in Stopped state.
func (sm *StateMachine) IsStopped() bool {
	return sm.Is(StateStopped)
}
```

## Rules
1. **Compound State Transitions Locks**: Validating and setting state changes requires read-then-write checking. Use mutual exclusion locks (`sync.RWMutex`) rather than simple atomic types to guarantee safety across compound state changes.
2. **Terminal State Constraints**: The `Stopped` state is terminal. Transitions out of `Stopped` are rejected with errors.
3. **Init Failure routing**: If component initialization fails, route the state machine directly to `Stopped` rather than falling back to `Created`. This prevents reusing half-initialized configurations.

## ⚠️ Pitfalls

### Pitfall 1: Attempting to use atomic integers for complex transition rules
```go
```
Use read/write locks to check allowed next states before modifying variables.

### Pitfall 2: Re-starting stopped kernel instances
Once stopped, attempting to trigger transitions back to running states is blocked to avoid resource leaks. Implementations must discard stopped kernels and instantiate new ones.

## Verify
```bash
go build ./kernel/...
```

## Checklist
- [ ] File `kernel/state.go` exists
- [ ] Package: `kernel`
- [ ] `State` iota defines Created, Initializing, Running, ShuttingDown, and Stopped
- [ ] `validTransitions` map defines the state machine paths
- [ ] `StateMachine` struct isolates modifications behind RWMutex locks
- [ ] `Transition` validates requested state changes and sets values
- [ ] Read queries use `RLock` read locks
- [ ] Terminal states do not configure transition outputs
- [ ] `go build ./kernel/...` passes
