# RFC-0030: Goal Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0002 (Brain Architecture), RFC-0004 (Context Engine)

## Summary

This RFC specifies the design of the **Goal Engine** in the AI Engineering Operating System (AEOS). The Goal Engine is responsible for translating vague, high-level user prompts (e.g. "build Shopee app") into formal, structured target specifications: `Goal -> Objectives -> Constraints -> Acceptance -> Milestones -> Plan`. This prevents the Planner from reasoning on unstructured text.

## Motivation

AI agent planning on unstructured text is highly unstable. When a user input is complex or underspecified:
- The Planner can make incorrect assumptions and produce invalid DAG paths.
- It is difficult to validate whether business constraints (e.g. budget, offline-first) are violated during execution.
- We need a deterministic translation boundary at the entrance of the system to turn user intent into a typed data contract.

## Design

### 1. Architectural Placement

The Goal Engine resides in the `Brain Runtime` and runs as the first stage of the mission pipeline before invoking the Planner.

```
  User Prompt ──► [Goal Engine] ──► Goal (Typed Struct) ──► [Planner] ──► Plan DAG
```

---

### 2. Contracts (`contracts/goal/`)

```go
package goal

import "context"

// Objective represents a sub-goal in the mission.
type Objective struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// Constraint represents a budget, tech-stack, or security limit.
type Constraint struct {
	Type  string `json:"type"`  // "budget_usd", "offline_only", "language", "framework"
	Value string `json:"value"` // e.g. "30.0", "true", "go", "gin"
}

// Goal represents the formalized target contract.
type Goal struct {
	RawInput           string       `json:"raw_input"`
	Objectives         []Objective  `json:"objectives"`
	Constraints        []Constraint `json:"constraints"`
	AcceptanceCriteria []string     `json:"acceptance_criteria"`
	Milestones         []string     `json:"milestones"`
}

// GoalEngine translates raw user strings to structured Goals.
type GoalEngine interface {
	// Parse translates raw user input into the structured Goal model.
	Parse(ctx context.Context, input string) (*Goal, error)
}
```

## Impact

- **Deterministic Inputs**: The Planner no longer accepts raw user string inputs. It consumes only validated `Goal` structures.
- **Constraint Pruning**: The Planner's CSP filter reads the `Constraints` slice to dynamically prune invalid paths from the Knowledge Graph.

## Open Questions

1. **How do we handle dynamically updated goals mid-mission?**
   - If the user modifies a goal mid-execution, it triggers a new `Parse()` cycle. The resulting Goal struct is diffed, and the `ExecutionGraphManager` (RFC-0046) handles the DAG migration.
