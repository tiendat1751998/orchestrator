# RFC-0032: Skill Graph

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0003 (Knowledge Engine), RFC-0010 (Cognitive Layer)

## Summary

This RFC specifies the design of the **Skill Graph** in AEOS. Upgrading from a simple tree, the Skill Graph models technologies and their dependencies as a network graph (e.g. `Build REST` requires `Go` which requires `Gin` which requires `HTTP` which requires `JSON`). The Skill Graph is stored in the SQLite Knowledge Graph database.

## Motivation

Technology stacks are not hierarchical trees; they are complex graphs where multiple frameworks share identical underlying dependencies (e.g., both REST and WebSockets depend on TCP and HTTP).
- Representing skills as a graph allows the Planner to perform recursive graph searches to identify all pre-requisite competencies before routing a task.
- It prevents redundant evaluations and helps target training and active exploration loops to missing core nodes.

## Design

### 1. Architectural Placement

The Skill Graph is implemented as an ontology extension inside the `Knowledge Engine`.

```
  (REST API Node) ──[requires]──► (Gin Node) ──[requires]──► (Go Node)
```

---

### 2. Contracts (`contracts/knowledge/skill.go`)

```go
package knowledge

import "context"

// SkillNode represents a specific technical competency.
type SkillNode struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"` // "language", "framework", "library", "protocol"
	Description string   `json:"description"`
}

// SkillEdge represents dependency relationship.
type SkillEdge struct {
	FromID   string `json:"from_id"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"` // "requires", "recommends", "conflicts_with"
}

// SkillGraph provides graph traversal queries over technology nodes.
type SkillGraph interface {
	// AddSkill registers a new technical node in the graph.
	AddSkill(ctx context.Context, skill SkillNode) error
	
	// LinkSkills creates a relationship edge between nodes.
	LinkSkills(ctx context.Context, edge SkillEdge) error
	
	// ResolveDependencies runs DFS/BFS to find all pre-requisite nodes for a skill.
	ResolveDependencies(ctx context.Context, skillID string) ([]SkillNode, error)
}
```

## Impact

- **Automated Stack Resolution**: When the Planner selects a template (e.g. `REST API`), it automatically traverses the Skill Graph to compile a list of all required competencies.
- **Competency Checking**: The Planner matches the required skills against the assigned Agent's capabilities before finalizing the DAG.

## Open Questions

1. **How do we handle conflicting skill requirements?**
   - The Skill Graph supports `conflicts_with` relationship edges. If the Planner compiles a path that includes conflicting nodes (e.g. both Gin and Fiber in the same target Go binary), the CSP filter invalidates the candidate plan immediately.
