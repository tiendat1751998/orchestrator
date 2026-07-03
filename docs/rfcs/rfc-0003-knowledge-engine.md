# RFC-0003: Knowledge Engine — Not a Database

- **Status**: PROPOSED → **REVISED**
- **Priority**: P0 — Foundation
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Revised**: 2026-07-03 (Review fixes: B5, C4, C5, D2, D3)
- **Depends on**: RFC-0000 (State Machine), RFC-0001 (Kernel Architecture)

## Summary

Knowledge is NOT a database. Knowledge is a rich, multi-layered system of semantic graphs, patterns, templates, facts, and decision history. SQLite is merely a persistence adapter. This RFC defines the Knowledge Engine architecture and separates four distinct memory concerns: Working Memory, Knowledge, Artifacts, and History.

## Motivation

Issues 2 and 4 from architecture review:
- **Issue 2**: Knowledge was incorrectly modeled as "SQLite store with entries"
- **Issue 4**: Memory, Knowledge, Artifacts, and History were conflated into one

## Design

### Four Memory Concerns (Separated)

```
┌─────────────────────────────────────────────────────────┐
│                    Memory Hierarchy                      │
│                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │   Working    │  │  Knowledge  │  │   Artifacts     │ │
│  │   Memory     │  │  (Graph)    │  │   (Files)       │ │
│  │             │  │             │  │                 │ │
│  │ Mission-     │  │ Permanent   │  │ Per-mission     │ │
│  │ scoped       │  │ patterns,   │  │ outputs:        │ │
│  │ state:       │  │ templates,  │  │ code, plans,    │ │
│  │ conversation │  │ facts,      │  │ reviews,        │ │
│  │ buffer,      │  │ embeddings  │  │ decisions       │ │
│  │ current plan │  │             │  │                 │ │
│  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘ │
│         │                │                   │          │
│         │         ┌──────┴──────┐            │          │
│         │         │   History   │            │          │
│         │         │  (Timeline) │            │          │
│         │         │             │            │          │
│         │         │ Immutable   │            │          │
│         │         │ event log:  │            │          │
│         │         │ transitions,│            │          │
│         │         │ decisions,  │            │          │
│         │         │ outcomes    │            │          │
│         │         └──────┬──────┘            │          │
│         │                │                   │          │
│  ┌──────▼────────────────▼───────────────────▼────────┐ │
│  │                Storage Layer                        │ │
│  │     SQLite  /  Filesystem  /  Any adapter           │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

| Concern | What it stores | Lifetime | Scoped by | Example |
|---|---|---|---|---|
| **Working Memory** | Current execution state | Mission | Mission ID | Conversation buffer, active plan, tool outputs |
| **Knowledge** | Learned patterns & templates | Permanent | Global | "REST API → routing+handler+test" pattern |
| **Artifacts** | Mission outputs & files | Per-mission | Mission ID | Generated code, review reports, diagrams |
| **History** | Event timeline | Permanent | Global | "Task X failed at 3pm, agent Y retried at 3:05pm" |

### Contracts

#### Working Memory (Fix for Issue C4)

```go
// contracts/memory/memory.go — Working Memory
package memory

import "context"

// WorkingMemory provides mission-scoped ephemeral storage.
//
// Scoping (Fix for Issue C4):
// Working memory is scoped by Mission ID, extracted from context.
// When a mission completes, its working memory is cleared (or
// selectively persisted to Knowledge if valuable).
//
// Concurrency: safe for concurrent use. Multiple tasks within
// the same mission share working memory via the same mission ID.
type WorkingMemory interface {
    // Set stores a value for the current mission.
    // Mission ID is extracted from ctx.
    Set(ctx context.Context, key string, value any) error
    // Get retrieves a value. Returns nil, nil if not found.
    Get(ctx context.Context, key string) (any, error)
    // Delete removes a value.
    Delete(ctx context.Context, key string) error
    // Clear removes all values for the current mission.
    Clear(ctx context.Context) error
    // Keys returns all keys for the current mission.
    Keys(ctx context.Context) ([]string, error)
    // Snapshot returns all key-value pairs (for persistence on crash).
    Snapshot(ctx context.Context) (map[string]any, error)
}
```

#### Knowledge Graph (Fix for Issue C5 — composable interfaces)

```go
// contracts/knowledge/knowledge.go — Knowledge Graph
package knowledge

import (
    "context"
    "time"
)

// NodeType categorizes knowledge nodes.
type NodeType string
const (
    NodePattern  NodeType = "pattern"   // Recognized pattern
    NodeTemplate NodeType = "template"  // Reusable plan template
    NodeFact     NodeType = "fact"      // Known fact
    NodeDecision NodeType = "decision"  // Recorded decision + outcome
    NodeRule     NodeType = "rule"      // Learned rule
)

// Node is a vertex in the knowledge graph.
//
// Content field (Issue D3): Using 'any' is intentional for v1 flexibility.
// As patterns emerge, typed content structs will be defined:
// PatternContent, TemplateContent, FactContent, etc.
// For now, Content is typically map[string]any serialized from JSON.
type Node struct {
    ID           string    `json:"id"`
    Type         NodeType  `json:"type"`
    Title        string    `json:"title"`
    Content      any       `json:"content"`
    Tags         []string  `json:"tags,omitempty"`
    Score        float64   `json:"score"`         // Confidence (0-1)
    UsedCount    int       `json:"used_count"`
    SuccessCount int       `json:"success_count"`
    Embeddings   []float32 `json:"embeddings,omitempty"` // Phase 3+: vector search
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// SuccessRate returns the ratio of successful uses to total uses.
func (n *Node) SuccessRate() float64 {
    if n.UsedCount == 0 {
        return 0
    }
    return float64(n.SuccessCount) / float64(n.UsedCount)
}

// Edge is a directed relationship between two nodes.
type Edge struct {
    FromID   string  `json:"from_id"`
    ToID     string  `json:"to_id"`
    Relation string  `json:"relation"` // e.g., "depends_on", "similar_to", "derived_from"
    Weight   float64 `json:"weight"`   // Strength of relationship (0-1)
}

// Subgraph is a subset of the knowledge graph returned by Traverse.
type Subgraph struct {
    Nodes []Node `json:"nodes"`
    Edges []Edge `json:"edges"`
}

// Query defines search criteria.
type Query struct {
    Type     NodeType `json:"type,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    MinScore float64  `json:"min_score,omitempty"`
    Text     string   `json:"text,omitempty"` // Full-text search
    Limit    int      `json:"limit,omitempty"`
    OrderBy  string   `json:"order_by,omitempty"` // "score", "used_count", "created_at"
}

// Outcome records execution results for learning.
type Outcome struct {
    NodeID    string        `json:"node_id,omitempty"`
    TaskID    string        `json:"task_id"`
    MissionID string        `json:"mission_id,omitempty"`
    Success   bool          `json:"success"`
    Duration  time.Duration `json:"duration"`
    Quality   float64       `json:"quality"` // 0-1
    Notes     string        `json:"notes,omitempty"`
}

// GraphStats provides aggregate statistics. (Fix for Issue B5)
type GraphStats struct {
    // TotalNodes is the number of nodes in the graph.
    TotalNodes int `json:"total_nodes"`
    // TotalEdges is the number of edges.
    TotalEdges int `json:"total_edges"`
    // NodesByType counts nodes per type.
    NodesByType map[NodeType]int `json:"nodes_by_type"`
    // TotalOutcomes is the number of recorded outcomes.
    TotalOutcomes int `json:"total_outcomes"`
    // AverageScore is the mean score across all nodes.
    AverageScore float64 `json:"average_score"`
    // HighConfidenceCount is nodes with Score >= 0.8.
    HighConfidenceCount int `json:"high_confidence_count"`
}

// --- Composable Interfaces (Fix for Issue C5) ---
// Go convention: prefer small, focused interfaces.
// KnowledgeGraph composes them all for convenience.

// NodeStore provides CRUD operations on knowledge nodes.
type NodeStore interface {
    AddNode(ctx context.Context, node Node) error
    GetNode(ctx context.Context, id string) (*Node, error)
    UpdateNode(ctx context.Context, node Node) error
    RemoveNode(ctx context.Context, id string) error
}

// EdgeStore provides operations on relationships.
type EdgeStore interface {
    AddEdge(ctx context.Context, edge Edge) error
    RemoveEdge(ctx context.Context, fromID, toID, relation string) error
}

// GraphQuerier provides search and traversal.
type GraphQuerier interface {
    Query(ctx context.Context, q Query) ([]Node, error)
    FindRelated(ctx context.Context, nodeID string, relation string) ([]Node, error)
    Traverse(ctx context.Context, startID string, depth int) (*Subgraph, error)
}

// GraphLearner records outcomes and updates node scores.
type GraphLearner interface {
    RecordOutcome(ctx context.Context, outcome Outcome) error
}

// KnowledgeGraph is the full knowledge graph interface.
// It composes all sub-interfaces for convenience.
// Callers that need only a subset should accept the sub-interface.
type KnowledgeGraph interface {
    NodeStore
    EdgeStore
    GraphQuerier
    GraphLearner
    Stats(ctx context.Context) (*GraphStats, error)
}
```

#### Artifact Store

```go
// contracts/artifact/artifact.go — Artifact Store
package artifact

import (
    "context"
    "io"
    "time"
)

// ArtifactType categorizes artifacts.
type ArtifactType string
const (
    ArtifactCode     ArtifactType = "code"
    ArtifactPlan     ArtifactType = "plan"
    ArtifactReview   ArtifactType = "review"
    ArtifactDecision ArtifactType = "decision"
    ArtifactReport   ArtifactType = "report"
    ArtifactConfig   ArtifactType = "config"
)

// Artifact is a generated file or output from a mission.
type Artifact struct {
    ID        string       `json:"id"`
    MissionID string       `json:"mission_id"`
    TaskID    string       `json:"task_id,omitempty"`
    Type      ArtifactType `json:"type"`
    Name      string       `json:"name"`
    Path      string       `json:"path"`
    Size      int64        `json:"size"`
    Checksum  string       `json:"checksum,omitempty"` // SHA256
    CreatedAt time.Time    `json:"created_at"`
}

// ArtifactStore manages mission outputs and generated files.
type ArtifactStore interface {
    Store(ctx context.Context, a Artifact, content io.Reader) error
    Get(ctx context.Context, id string) (*Artifact, error)
    Read(ctx context.Context, id string) (io.ReadCloser, error)
    ListByMission(ctx context.Context, missionID string) ([]Artifact, error)
    Delete(ctx context.Context, id string) error
}
```

#### History Timeline (Fix for Issue D2 — iterator pattern)

```go
// contracts/history/history.go — Event History (Timeline)
package history

import (
    "context"
    "time"
)

// Entry is an immutable event record.
type Entry struct {
    ID        string    `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Entity    string    `json:"entity"`     // "mission", "task", "agent", "provider"
    EntityID  string    `json:"entity_id"`
    Event     string    `json:"event"`      // "state.changed", "decision.made", etc.
    Data      any       `json:"data"`
}

// TimelineQuery defines search criteria for history.
type TimelineQuery struct {
    Entity   string    `json:"entity,omitempty"`
    EntityID string    `json:"entity_id,omitempty"`
    Event    string    `json:"event,omitempty"`
    From     time.Time `json:"from,omitempty"`
    To       time.Time `json:"to,omitempty"`
    Limit    int       `json:"limit,omitempty"`
}

// EntryIterator iterates over history entries.
// Preferred over channel-based API for better error handling.
//
// Usage:
//
//   iter, err := timeline.Replay(ctx, startTime)
//   if err != nil { ... }
//   defer iter.Close()
//   for iter.Next() {
//       entry := iter.Entry()
//       // process entry
//   }
//   if err := iter.Err(); err != nil {
//       // handle mid-stream error (e.g., SQLite read failure)
//   }
type EntryIterator interface {
    // Next advances to the next entry. Returns false when done or on error.
    Next() bool
    // Entry returns the current entry. Only valid after Next() returns true.
    Entry() Entry
    // Err returns any error encountered during iteration.
    Err() error
    // Close releases resources. Must be called when done.
    Close() error
}

// Timeline provides an immutable, append-only event log.
// Used for audit, replay, and debugging.
//
// Thread-safe: all methods must be safe for concurrent use.
type Timeline interface {
    // Append records an event. Never fails silently — errors are logged.
    Append(ctx context.Context, entry Entry) error
    // Query returns events matching criteria, in chronological order.
    // Returns empty slice (not nil) if nothing matches.
    Query(ctx context.Context, q TimelineQuery) ([]Entry, error)
    // Replay returns an iterator over events from a starting time.
    // Uses iterator pattern (not channels) for proper error handling.
    Replay(ctx context.Context, fromTime time.Time) (EntryIterator, error)
    // Count returns the number of entries matching the query.
    Count(ctx context.Context, q TimelineQuery) (int, error)
}
```

### Knowledge Engine Implementation

```
kernel/knowledge/
├── engine.go           # KnowledgeEngine — orchestrates sub-components
├── graph/
│   ├── graph.go        # In-memory knowledge graph implementation
│   ├── index.go        # Tag-based index for fast lookup
│   └── graph_test.go
├── search/
│   ├── engine.go       # Multi-strategy search engine
│   ├── text.go         # Full-text search (ripgrep-style)
│   ├── graph.go        # Graph traversal search
│   └── vector.go       # Vector similarity search (Phase 3+)
├── storage/
│   ├── sqlite.go       # SQLite persistence adapter
│   ├── filesystem.go   # File-based persistence (for artifacts)
│   └── storage.go      # Storage interface (adapter pattern)
└── engine_test.go
```

### Search Engine Architecture

```
Search Request
      ↓
┌──────────────────────┐
│   Search Engine      │
│                      │
│  ┌────────────────┐  │
│  │ Text Search    │  │  ← ripgrep-style full-text
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Graph Search   │  │  ← traverse knowledge graph relationships
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Tag Search     │  │  ← filter by tags (fast, indexed)
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Vector Search  │  │  ← semantic similarity (Phase 3+)
│  └────────────────┘  │
│                      │
│     ┌────────────┐   │
│     │ Ranker     │   │  ← combine results, rank by relevance
│     └────────────┘   │
└──────────────────────┘
      ↓
Ranked Results
```

## Impact

### New Contract Packages
- `contracts/memory/` — Working Memory (mission-scoped)
- `contracts/knowledge/` — Knowledge Graph (composable interfaces)
- `contracts/artifact/` — Artifact Store
- `contracts/history/` — Event Timeline (with EntryIterator)

### New Kernel Packages
- `kernel/knowledge/` — Knowledge Engine
- `kernel/knowledge/graph/` — Graph implementation
- `kernel/knowledge/search/` — Multi-strategy search
- `kernel/knowledge/storage/` — Persistence adapters
- `kernel/history/` — Timeline implementation (SQLite-backed)

### Removed/Replaced
- Old `contracts/knowledge/` (flat Entry store) → replaced by KnowledgeGraph
- Old `contracts/memory/` (generic store) → replaced by WorkingMemory (mission-scoped)
- `contracts/search/` → absorbed into knowledge/search/

### Layer Compliance
- All new contracts import only stdlib ✅
- Kernel packages import only contracts ✅
- Storage adapters injected, not hard-coded ✅

## Resolved Questions

1. ~~**Graph persistence**~~ **RESOLVED**: Direct persistence (nodes + edges in SQLite) for performance. FSM transitions persisted via History Timeline for audit.

2. ~~**Knowledge sharing**~~ **RESOLVED**: Single shared graph. Missions tag their entries for filtering.

3. ~~**Working Memory scoping**~~ **RESOLVED**: Scoped by Mission ID (extracted from context). Cleared on mission completion.

4. ~~**KnowledgeGraph too large**~~ **RESOLVED**: Composed from NodeStore, EdgeStore, GraphQuerier, GraphLearner sub-interfaces.

5. ~~**Replay error handling**~~ **RESOLVED**: Iterator pattern (EntryIterator) instead of channel for proper mid-stream error handling.

## Open Questions

1. **Embedding generation**: Vector search requires embeddings. Who generates them? Recommendation: Brain calls AI provider with "embed" request type. Defer to Phase 3+.

2. **Node.Content typing**: `any` is used for v1 flexibility. When should we introduce typed content structs (PatternContent, TemplateContent, etc.)? Recommendation: After Phase 2, when usage patterns are clear.
