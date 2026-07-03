// Package fsm defines the finite state machine contracts for long-lived system entities.
package fsm

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidTransition indicates a state machine transition was requested
	// but no transition is defined for the current state and event.
	ErrInvalidTransition = errors.New("fsm: invalid state transition")

	// ErrGuardRejected indicates a transition guard function returned false,
	// preventing the transition from proceeding.
	ErrGuardRejected = errors.New("fsm: transition guard rejected")
)

// State represents a named state in a state machine.
type State string

// TransitionRecord is an immutable record of a state transition.
type TransitionRecord struct {
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	From       State     `json:"from"`
	To         State     `json:"to"`
	Event      string    `json:"event"`
	Timestamp  time.Time `json:"timestamp"`
	Payload    any       `json:"payload,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// Transition defines a valid state change.
type Transition struct {
	From   State
	To     State
	Event  string
	Guard  func(ctx context.Context, payload any) bool
	Action func(ctx context.Context, payload any) error
}

// OnTransition is a callback invoked after every successful transition.
type OnTransition func(record TransitionRecord)

// Definition defines a state machine's structure.
type Definition struct {
	Name         string
	Initial      State
	Transitions  []Transition
	OnTransition OnTransition
	HistoryLimit int
}

// Machine is a finite state machine.
type Machine interface {
	Current() State
	Can(event string) bool
	AvailableEvents() []string
	Fire(ctx context.Context, event string, payload any) error
	History() []TransitionRecord
}

// DAGNode represents a node in a plan Directed Acyclic Graph (DAG).
type DAGNode struct {
	ID           string   `json:"id"`
	Dependencies []string `json:"dependencies,omitempty"`
	Status       State    `json:"status"`
}

// DAG represents a plan Directed Acyclic Graph.
type DAG struct {
	Nodes map[string]*DAGNode `json:"nodes"`
}
