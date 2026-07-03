# RFC-0052: Distributed Mission

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0008 (Event Model), RFC-0001 (Kernel Architecture)

## Summary

This RFC specifies the design of the **Distributed Mission** module in AEOS. It enables streaming FSM transition events from the Event Store to remote execution nodes via lightweight WebSocket or gRPC channels, allowing distributed compilation, reviewing, and execution without requiring complex cluster consensus databases (like Raft or Paxos).

## Motivation

As teams and workloads scale, running everything on a single local development machine might become a bottleneck.
- We need a clean extension point that allows running code generation on Machine A, compilation on Machine B, and review on Machine C.
- By using our Event Sourcing model, the Event Store remains the single source of truth. Remote nodes simply subscribe to event streams, avoiding any distributed sync overhead.

## Design

### 1. Architectural Placement

Distributed Mission extends the `EventBus` to support remote gRPC/WebSocket client subscription endpoints.

```
  Kernel Event Store ──► [gRPC Stream] ──► Remote Worker ──► Execute ──► Event Return
```

---

### 2. Contracts (`contracts/event/distributed.go`)

```go
package event

import (
	"context"
)

// RemoteWorkerInfo represents a connected worker node.
type RemoteWorkerInfo struct {
	WorkerID     string   `json:"worker_id"`
	Address      string   `json:"address"`
	Capabilities []string `json:"capabilities"`
}

// DistributedEventBus streams event transitions to remote nodes.
type DistributedEventBus interface {
	// RegisterWorker registers a connected execution worker node.
	RegisterWorker(ctx context.Context, worker RemoteWorkerInfo) error
	
	// StreamEvents broadcasts EventStore records to registered subscribers.
	StreamEvents(ctx context.Context, missionID string) error
}
```

## Impact

- **Lightweight Scaling**: remote execution requires zero Paxos/Raft consensus setups. The central Kernel coordinates everything via deterministic event routing.
- **Local-First Core Preserved**: In local-first mode, the distributed bus operates in-process, keeping compilation simple and fast.

## Open Questions

1. **How do we handle network partition failures?**
   - If a remote worker disconnects mid-task, the Event Store registers a timeout event, triggering the Replanning Engine to re-assign the task to a local worker.
