# RFC-0051: Knowledge Decay & TTL

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0003 (Knowledge Engine), RFC-0039 (Evolution Engine)

## Summary

This RFC specifies the design of the **Knowledge Decay & TTL** module in AEOS. To prevent the Knowledge Graph from recommending obsolete patterns (e.g. Go 1.18 templates instead of Go 1.26), the engine applies contextual depreciation weights to semantic nodes based on validation outcomes and usage statistics.

## Motivation

Knowledge graphs grow stale over time if not pruned.
- However, simple time-based TTL decay leads to **System Amnesia**, erasing stable libraries that have not changed in years (e.g. database drivers).
- AEOS implements **Contextual Depreciation**. A node's confidence rating is only decremented when a compiler error, linter warning, or explicit test failure occurs, preserving stable knowledge.

## Design

### 1. Architectural Placement

The Knowledge Decay module is an offline service running inside the `Knowledge Engine` database layer.

```
  Execution Failures ──► [Knowledge Decay] ──► Decrement Confidence weights in DB
```

---

### 2. Contracts (`contracts/knowledge/decay.go`)

```go
package knowledge

import "context"

// DecayParams configures the depreciation rate.
type DecayParams struct {
	DecayFactor      float64 `json:"decay_factor"` // EMA decrement value (e.g. 0.05)
	MinConfidence    float64 `json:"min_confidence"`
}

// KnowledgeDecayer deprecates stale or incorrect nodes.
type KnowledgeDecayer interface {
	// DeprecateNode decrements the confidence rating of a node.
	DeprecateNode(ctx context.Context, nodeID string, params DecayParams) error
	
	// PruneStaleNodes removes or hides nodes with confidence ratings below the minimum.
	PruneStaleNodes(ctx context.Context, minConfidence float64) (int, error)
}
```

## Impact

- **Accurate Technology Selection**: Obsolete configurations are naturally penalized when they generate compilation errors on newer compilers, ensuring the Planner chooses modern alternatives.
- **Zero Amnesia**: Stable, long-standing utilities maintain their high confidence scores as long as they compile successfully.

## Open Questions

1. **How do we handle newly introduced compiler versions?**
   - When a new toolchain is detected (e.g., upgrading to Go 1.26), the system runs standard conformance tests. Any template that fails compiling under the new version is immediately penalized.
