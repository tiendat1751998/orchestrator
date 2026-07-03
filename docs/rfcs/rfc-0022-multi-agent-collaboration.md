# RFC-0022: Multi-Agent Collaboration & Event Routing

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0008 (Event Model), RFC-0011 (Scheduler)

## Summary

This RFC specifies the design of **Multi-Agent Collaboration & Event Routing** in AEOS. It defines the routing protocols that allow agents to delegate tasks to each other, raise events, and cooperate asynchronously by streaming messages via the central EventBus.

## Motivation

Monolithic agents that execute all tasks sequentially are slow and fragile.
- Complex software engineering tasks require collaboration (e.g. Coder writing code, QA writing tests, Reviewer auditing code).
- By routing tasks via events, agents can collaborate asynchronously without introducing tight coupling.

## Design

### 1. Architectural Placement

Collaboration is coordinated by the `Scheduler` using the `EventBus` to route tasks.

```
  Coder Agent ──(Emits event: CodeReady)──► [EventBus] ──► Reviewer Agent (Triggered)
```

---

### 2. Contracts (`contracts/brain/collaboration.go`)

```go
package brain

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// Message represents an inter-agent task communication.
type Message struct {
	FromAgent string `json:"from_agent"`
	ToAgent   string `json:"to_agent"`
	Payload   string `json:"payload"`
}

// EventRouter handles task routing between agents.
type EventRouter interface {
	// RouteTaskMessage sends a direct message payload to a target agent.
	RouteTaskMessage(ctx context.Context, msg Message) error
	
	// BroadcastTaskEvent registers a task event to all active subscribers.
	BroadcastTaskEvent(ctx context.Context, ev fsm.TransitionRecord) error
}
```

## Impact

- **Decoupled Pipelines**: Coder, QA, and Reviewer agents are coordinated by event triggers, enabling flexible work queues.
- **Auditable Collaboration**: All inter-agent communication is preserved in the Event Store for reproducibility.
