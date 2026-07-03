# RFC-0041: Product Memory

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0005 (Memory Model), RFC-0003 (Knowledge Engine)

## Summary

This RFC specifies the design of the **Product Memory** in AEOS. Product Memory stores **Business/Product Ontologies** (e.g. Voucher systems, Flash sales, Livestream checkout) as semantic graph relationships in the SQLite Knowledge Graph, enabling the Planner to reason over business patterns alongside code patterns.

## Motivation

Knowledge is not just engineering frameworks (Go, HTTP, Gin). Software engineering is driven by business domains and features.
- If the system only remembers technical details, the Planner cannot assist in decomposing feature requirements.
- By modeling business domains (e.g. knowing that a "SaaS Product" requires a "Subscription Engine" and a "Billing Gateway"), the Planner can automatically suggest complete functional structures.

## Design

### 1. Architectural Placement

Product Memory is represented as business-level nodes and edges inside the `Knowledge Engine`.

```
  (SaaS Product Node) ──[requires]──► (Billing Node) ──[requires]──► (Stripe API)
```

---

### 2. Contracts (`contracts/knowledge/product.go`)

```go
package knowledge

import "context"

// ProductPattern represents a business feature layout.
type ProductPattern struct {
	ID          string   `json:"id"`
	FeatureName string   `json:"feature_name"`
	Domain      string   `json:"domain"` // "e-commerce", "fintech", "saas"
	Description string   `json:"description"`
	Components  []string `json:"components"` // required tech blocks
}

// ProductMemory provides access to business patterns.
type ProductMemory interface {
	// AddPattern registers a new business feature template.
	AddPattern(ctx context.Context, pattern ProductPattern) error
	
	// MatchPattern searches for compatible business layouts matching the goals.
	MatchPattern(ctx context.Context, goalQuery string) ([]ProductPattern, error)
}
```

## Impact

- **Domain-Aware Planning**: When the user requests a subscription-based app, the Planner traverses Product Memory to identify required business sub-graphs, automatically adding billing tasks to the DAG.
- **Accurate Code Scaffolding**: Code is structured around clean business modules, enforcing DDD boundaries.

## Open Questions

1. **How do we populate Product Memory?**
   - The Knowledge Engine harvests business patterns from successfully executed historical projects in the workspace repository.
