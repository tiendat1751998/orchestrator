// Package event defines the contract for the event system.
// The event bus allows components to communicate without knowing each other.
package event

import (
	"context"
	"time"
)

// Event represents something that happened in the system.
//
// Events are published by components (kernel, agents, providers)
// and consumed by subscribers (logger, metrics, orchestrator).
//
// Example:
//
//	Event{
//	    ID:        "evt-a1b2c3d4",
//	    Type:      "task.completed",
//	    Source:     "agent:backend",
//	    Payload:   result,
//	    Timestamp: time.Now(),
//	}
type Event struct {
	// ID uniquely identifies this event instance.
	ID string `json:"id"`

	// Type identifies what happened.
	// Convention: dot-separated namespace (e.g., "task.started", "agent.error").
	// Supports wildcard subscription: "task.*" matches all task events.
	Type string `json:"type"`

	// Source identifies who emitted this event.
	// Convention: "component:name" (e.g., "kernel", "agent:backend", "provider:antigravity").
	Source string `json:"source"`

	// Payload carries event-specific data.
	// Type depends on event Type:
	//   "task.started"   → *agent.Task
	//   "task.completed" → *agent.Result
	//   "agent.error"    → error message string
	Payload any `json:"payload"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// Bus provides asynchronous publish/subscribe functionality.
//
// The bus is the central nervous system of the orchestrator.
// Components publish events, and interested parties subscribe.
// This decouples components — the publisher doesn't know who subscribes.
//
// Thread-safety requirement: All methods must be safe for concurrent use.
type Bus interface {
	// Publish emits an event to all matching subscribers.
	//
	// Publishing is asynchronous — this method returns immediately.
	// If a subscriber is slow, it does NOT block the publisher.
	//
	// Context is used for cancellation only (not for deadline).
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for events matching the given pattern.
	//
	// Pattern matching rules:
	//   - Exact match: "task.started" matches only "task.started"
	//   - Wildcard: "task.*" matches "task.started", "task.completed", etc.
	//   - All events: "*" matches everything
	//
	// The handler function is called in a separate goroutine.
	// If the handler panics, the bus must recover and log the error.
	//
	// Returns an unsubscribe function that MUST be called when done.
	// Calling unsubscribe multiple times is safe (idempotent).
	//
	// WHY return unsubscribe function instead of an ID?
	// → Cleaner API: defer unsubscribe() at call site.
	// → No need to manage subscription IDs manually.
	Subscribe(pattern string, handler func(Event)) (unsubscribe func(), err error)
}

// Common event type constants.
// Using constants avoids typos in event type strings.
const (
	EventTaskStarted    = "task.started"
	EventTaskCompleted  = "task.completed"
	EventTaskFailed     = "task.failed"
	EventTaskCancelled  = "task.cancelled"
	EventMissionStarted = "mission.started"
	EventMissionDone    = "mission.completed"
	EventMissionFailed  = "mission.failed"
	EventAgentError     = "agent.error"
	EventKernelStarted  = "kernel.started"
	EventKernelStopped  = "kernel.stopped"
)

// NewEvent creates a new Event with a generated ID and current timestamp.
func NewEvent(eventType, source string, payload any) Event {
	return Event{
		ID:        "evt-" + generateShortID(),
		Type:      eventType,
		Source:    source,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// generateShortID generates a short random hex ID.
// Duplicated here to avoid importing contracts package (keep event package independent).
func generateShortID() string {
	// Simple implementation — crypto/rand version in contracts/types.go
	// Here we use time-based for simplicity since event IDs don't need to be cryptographic
	return time.Now().Format("20060102150405.000000")[8:] // HHMMSSmmm
}
