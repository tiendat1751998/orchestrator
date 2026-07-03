# Micro-Task 1.23: Create contracts/event/event.go

## Info
- **File**: `contracts/event/event.go`
- **Package**: `event`
- **Depends on**: 1.06 (contracts/types.go)
- **Time**: 15 min
- **Verify**: `go build ./contracts/event/...`

## Purpose
Defines the contract for the system's Event Bus. The Event Bus is the central nervous system of the orchestrator, enabling decoupled, asynchronous communication between various components (such as kernel, agents, providers, logging systems, and metrics recorders) without direct dependencies.

## EXACT code to create

```go
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
		Source:     source,
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
```

## Rules
1. **Event Types**: Dot-separated namespace strings are standard.
2. **Payload Flexibility**: Payload uses the empty interface `any` to support various data payloads. Consumers must perform type-assertions.
3. **Event IDs**: Generated automatically via the `NewEvent` helper function using short timestamps to decouple from direct contracts packages dependency.
4. **Subscribe Return Pattern**: `Subscribe` must return an `unsubscribe` function rather than an ID. The subscriber simply invokes this function or defers it.

## ⚠️ Pitfalls

### Pitfall 1: Leaking subscribers (Goroutine/Memory leaks)
```go
unsub, err := bus.Subscribe("task.*", func(e Event) {
    // handler code...
})
if err == nil {
    defer unsub() // Triggers clean cleanup of channel routing allocations.
}
```
If a component subscribes but does not call the unsubscribe function upon teardown, the event bus keeps a reference to the handler, preventing garbage collection and leaking memory.

### Pitfall 2: Blocking publishers with slow handlers
If the Event Bus calls subscriber handlers synchronously inside `Publish`, a single slow subscriber (e.g. one doing slow DB writes or sending slack alerts) will block the publisher thread, slowing down the entire agent execution runner. The Bus implementation MUST run handlers asynchronously in separate goroutines.

### Pitfall 3: Not recovering from subscriber handler panics
If a subscriber handler panics during execution, it can crash the entire application process if the Bus does not wrap the invocation with `recover()`. The Event Bus must protect itself against subscriber crashes.

## Verify
```bash
go build ./contracts/event/...
```

## Checklist
- [ ] File `contracts/event/event.go` exists
- [ ] Package: `event`
- [ ] `Event` struct contains ID, Type, Source, Payload, and Timestamp fields
- [ ] `Bus` interface declares `Publish` and `Subscribe` methods
- [ ] `Subscribe` returns an `unsubscribe func()` callback
- [ ] Defined common event name constants (`EventTaskStarted`, etc.)
- [ ] `NewEvent` builder function exists
- [ ] `go build ./contracts/event/...` passes
