# RFC-0021: Vector Search & Local Graph Embeddings

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0003 (Knowledge Engine), RFC-0005 (Memory Model)

## Summary

This RFC specifies the design of **Vector Search & Local Graph Embeddings** in AEOS. To enable semantic retrieval of historical plans and code snippets, the Knowledge Engine integrates local vector embeddings (using standard lightweight providers or in-memory vector indexing) directly within the SQLite database schema.

## Motivation

Simple text searches cannot match semantic intent (e.g. searching for "database connection pool" should match a template for "pgx connection pool configuration").
- By storing local text embeddings, the system can perform cosine similarity queries to retrieve relevant code snippets or plan templates from memory.

## Design

### 1. Architectural Placement

Vector search functions act as an indexing filter inside the `Knowledge Engine` and `Memory Model`.

```
  Semantic Query ──► [Vector Search] ──► Cosine Similarity Match ──► Retrieve Graph Nodes
```

---

### 2. Contracts (`contracts/knowledge/vector.go`)

```go
package knowledge

import "context"

// VectorRecord represents an embedded text item.
type VectorRecord struct {
	ID        string    `json:"id"`
	Payload   string    `json:"payload"`
	Embedding []float32 `json:"embedding"`
}

// VectorStore provides cosine similarity searches.
type VectorStore interface {
	// AddVector inserts a vector embedding into the store.
	AddVector(ctx context.Context, record VectorRecord) error
	
	// QuerySimilarity returns the top K similar records.
	QuerySimilarity(ctx context.Context, vector []float32, k int) ([]VectorRecord, error)
}
```

## Impact

- **Semantic Memory**: The Planner retrieves related plan DAG templates using semantic matching instead of strict tag lookups.
- **Fast Search**: In-memory vector indexing ensures queries complete in under 5ms.
