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
//
//	→ Transition validation requires read-then-check-then-write.
//	→ This is a compound operation that needs a mutex for atomicity.
//	→ atomic only handles single reads/writes, not compound ops.
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
