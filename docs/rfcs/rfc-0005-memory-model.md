# RFC-0005: Memory Model

- **Status**: PROPOSED
- **Priority**: P1 — Core
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (State Machine), RFC-0001 (Kernel Architecture), RFC-0003 (Knowledge Engine)

## Summary

This RFC defines the memory taxonomy and persistence strategies for the AI Engineering Operating System (AEOS). It separates the system's memory into four distinct layers (Working Memory, Knowledge Graph, Artifact Store, and Event Timeline) with explicit scopes, lifetimes, and storage adapters.

## Motivation

Conflating memory, knowledge, outputs, and history leads to architectural decay. Working memory (such as short-lived chat history) should not pollute long-term pattern libraries, and generated source code (artifacts) should not be mixed with audit records (history). Defining clean boundaries ensures high performance, deterministic recovery, and simplifies search and retrieval.

## Design

### Memory Taxonomy & Lifetimes

```
 ┌────────────────────────────────────────────────────────┐
 │                      AEOS Kernel                       │
 └──────────────────────────┬─────────────────────────────┘
                            │
       ┌────────────────────┼────────────────────┐
       ▼                    ▼                    ▼
 Working Memory       Knowledge Graph     Artifact Store
 (Mission ID Scope)   (Global Scope)      (Mission ID Scope)
 [Ephemeral/RAM]      [Permanent/SQL]     [Permanent/Disk]
       │                    │                    │
       └───────────┬────────┴────────────┬───────┘
                   │                     │
                   ▼                     ▼
             History Timeline      Storage Layer
             (Global Scope)        (Adapters)
             [Immutable/SQL]       [SQLite, Local FS]
```

### 1. Working Memory (`contracts/memory/`)

**Scope**: Mission ID  
**Lifetime**: Ephemeral (starts when Mission is scheduled, cleared when Mission is completed/failed).  
**Purpose**: Stores task execution contexts, ongoing chat buffers, tool output caches, and execution variables.

```go
package memory

import "context"

// WorkingMemory provides read-write ephemeral storage for active missions.
type WorkingMemory interface {
	// Set stores a key-value pair under the active mission context.
	Set(ctx context.Context, key string, value any) error
	// Get retrieves a value. Returns (nil, nil) if not found.
	Get(ctx context.Context, key string) (any, error)
	// Delete removes a key.
	Delete(ctx context.Context, key string) error
	// Snapshot returns all current key-value pairs. Used for checkpoints.
	Snapshot(ctx context.Context) (map[string]any, error)
	// Restore populates the memory from a snapshot (recovery).
	Restore(ctx context.Context, snapshot map[string]any) error
	// Clear deletes all variables for the active mission.
	Clear(ctx context.Context) error
	// Keys returns all active keys in the memory.
	Keys(ctx context.Context) ([]string, error)
}
```

*Implementation*: Thread-safe in-memory map backed by a lock. During execution, checkpoints are written to the database in case of sudden crashes.

---

### 2. Knowledge Graph (`contracts/knowledge/`)

**Scope**: Global  
**Lifetime**: Permanent  
**Purpose**: Stores high-level, generalized patterns, plan templates, facts, rules, and semantic relationships between agents, tools, and past success profiles.

```go
package knowledge

import (
	"context"
	"time"
)

// NodeType classifies nodes within the Knowledge Graph.
type NodeType string

const (
	NodePattern  NodeType = "pattern"  // Reusable abstract design pattern
	NodeTemplate NodeType = "template" // Executable Plan Template
	NodeFact     NodeType = "fact"     // Static fact (e.g. system properties)
	NodeDecision NodeType = "decision" // Past routing decision and outcome
	NodeRule     NodeType = "rule"     // Decision engine rule
)

// Node represents a vertex in the Knowledge Graph.
type Node struct {
	ID           string    `json:"id"`
	Type         NodeType  `json:"type"`
	Title        string    `json:"title"`
	Content      any       `json:"content"` // JSON-serializable payload
	Tags         []string  `json:"tags,omitempty"`
	Score        float64   `json:"score"` // Confidence score (0.0 to 1.0)
	UsedCount    int       `json:"used_count"`
	SuccessCount int       `json:"success_count"`
	Embeddings   []float32 `json:"embeddings,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Edge represents a directed relation between nodes.
type Edge struct {
	FromID   string  `json:"from_id"`
	ToID     string  `json:"to_id"`
	Relation string  `json:"relation"` // e.g. "contains", "depends_on", "similar_to"
	Weight   float64 `json:"weight"`   // Connection strength (0.0 to 1.0)
}

type NodeStore interface {
	AddNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (*Node, error)
	UpdateNode(ctx context.Context, node Node) error
	RemoveNode(ctx context.Context, id string) error
}

type EdgeStore interface {
	AddEdge(ctx context.Context, edge Edge) error
	RemoveEdge(ctx context.Context, fromID, toID, relation string) error
	GetEdgesFrom(ctx context.Context, nodeID string) ([]Edge, error)
}

type GraphQuerier interface {
	Query(ctx context.Context, q Query) ([]Node, error)
	FindRelated(ctx context.Context, nodeID string, relation string) ([]Node, error)
}

type GraphLearner interface {
	RecordOutcome(ctx context.Context, outcome Outcome) error
}

type KnowledgeGraph interface {
	NodeStore
	EdgeStore
	GraphQuerier
	GraphLearner
	Stats(ctx context.Context) (*GraphStats, error)
}
```

*Implementation*: Persisted in SQLite using a `nodes` table and an `edges` table. Full-text indices are built on `title`, `tags`, and `content`.

---

### 3. Artifact Store (`contracts/artifact/`)

**Scope**: Mission ID  
**Lifetime**: Permanent (but cleanup policies can archive files after a specified period).  
**Purpose**: Stores physical files produced or read during a mission. This includes code files, logs, diagrams, and review reports.

```go
package artifact

import (
	"context"
	"io"
	"time"
)

type ArtifactType string

const (
	ArtifactCode     ArtifactType = "code"     // Generated source code files
	ArtifactPlan     ArtifactType = "plan"     // Execution plans
	ArtifactReview   ArtifactType = "review"   // Review & validation logs
	ArtifactDecision ArtifactType = "decision" // Brain decisions
	ArtifactReport   ArtifactType = "report"   // Mission summaries
)

type Artifact struct {
	ID        string       `json:"id"`
	MissionID string       `json:"mission_id"`
	TaskID    string       `json:"task_id,omitempty"`
	Type      ArtifactType `json:"type"`
	Name      string       `json:"name"`      // File name (e.g. main.go)
	Path      string       `json:"path"`      // Relative local path inside the workspace
	Size      int64        `json:"size"`
	Checksum  string       `json:"checksum"`  // SHA256 file hash
	CreatedAt time.Time    `json:"created_at"`
}

type ArtifactStore interface {
	// Store writes the artifact contents and saves metadata.
	Store(ctx context.Context, meta Artifact, reader io.Reader) error
	// Get retrieves artifact metadata.
	Get(ctx context.Context, id string) (*Artifact, error)
	// Read opens the artifact file for reading.
	Read(ctx context.Context, id string) (io.ReadCloser, error)
	// ListByMission lists all files belonging to a mission.
	ListByMission(ctx context.Context, missionID string) ([]Artifact, error)
	// Delete removes metadata and the physical file.
	Delete(ctx context.Context, id string) error
}
```

*Implementation*: Physical files are saved on the local filesystem under the Workspace root (e.g. `workspace/artifacts/{mission_id}/{filename}`). The metadata is stored in SQLite.

---

### 4. Event Timeline / History (`contracts/history/`)

**Scope**: Global  
**Lifetime**: Permanent (Immutable, append-only).  
**Purpose**: An event log capturing all state changes, agent activations, rule evaluations, and system events. This is the source of truth for Event Sourcing and Replays.

```go
// contracts/history/history.go — Event History (Timeline)
package history

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

type TimelineQuery struct {
	Entity   string    `json:"entity,omitempty"`
	EntityID string    `json:"entity_id,omitempty"`
	Event    string    `json:"event,omitempty"`
	From     time.Time `json:"from,omitempty"`
	To       time.Time `json:"to,omitempty"`
	Limit    int       `json:"limit,omitempty"`
}

type EntryIterator interface {
	Next() bool
	Entry() event.Event
	Err() error
	Close() error
}

type Timeline interface {
	// Append records an event.
	Append(ctx context.Context, ev event.Event) error
	// Query returns events matching criteria, in chronological order.
	Query(ctx context.Context, q TimelineQuery) ([]event.Event, error)
	// Replay returns an iterator to stream entries starting from a given time matching filters.
	Replay(ctx context.Context, fromTime time.Time, q TimelineQuery) (EntryIterator, error)
}
```

*Implementation*: Persisted to a dedicated SQLite table (`timeline`). Backed by a transaction log. On start, we can replay these events to reconstruct the exact State Machine configurations of running tasks.

## Storage Adapters (`kernel/knowledge/storage/`)

To keep the kernel independent of database engines, all persistence goes through generic Storage Adapters:

```go
// kernel/knowledge/storage/storage.go
package storage

import "context"

type KeyValueStore interface {
	Put(ctx context.Context, table, key string, val []byte) error
	Get(ctx context.Context, table, key string) ([]byte, error)
	Delete(ctx context.Context, table, key string) error
	Keys(ctx context.Context, table string) ([]string, error)
}

type SQLStore interface {
	Execute(ctx context.Context, query string, args ...any) error
	QueryRow(ctx context.Context, query string, args ...any) RowScanner
	QueryRows(ctx context.Context, query string, args ...any) (RowIterator, error)
}
```

The SQLite implementation of these interfaces acts as the single shared engine, but it is injected at startup. This decouples business logic from storage files.

## Impact

- **Decoupled Interfaces**: Working memory, Graph, Artifacts, and History are located in their respective `contracts/` folders.
- **Isolated Kernel Services**: `kernel/knowledge/` houses graph index, file adapters, search algorithms, and SQLite code.

## Open Questions

1. **Working Memory Checkpointing Frequency**:
   - Should we write working memory to disk on every `Set()` or periodically?
   - Recommendation: Periodic checkpointing (e.g. on every Task state transition or FSM transition) to keep I/O overhead minimal.
