# RFC-0035: Capability Graph & Agent Competency Model

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0032 (Skill Graph), RFC-0010 (Cognitive Layer)

## Summary

This RFC specifies the design of the **Capability Graph & Agent Competency Model** in AEOS. We separate **Skills** (declarative knowledge about technologies like Go, Gin, Redis) from **Capabilities** (actionable task competencies like Code Generation, Debugging, Reviewing, Optimization). This enables precise agent routing based on task actions.

## Motivation

An agent that knows Go (Skill) might be excellent at reviewing Go code, but poor at writing it from scratch (Capability).
- Flat skill-matching models cannot capture this distinction, leading to poor agent assignment (e.g. routing a complex refactoring task to an agent that only excels at boilerplate generation).
- Modeling capability weights separately allows the Planner to route tasks to the most optimal agent for that specific action-technology intersection.

## Design

### 1. Architectural Placement

The Capability Graph is mapped inside the `Knowledge Engine` and queried by the `Scheduler` and `Planner`.

```
  Agent Node ──[can_perform]──► Capability (Review) ──[in_domain]──► Skill (Go)
```

---

### 2. Contracts (`contracts/brain/capability.go`)

```go
package brain

import "context"

// CompetencyRating represents the capability weight.
type CompetencyRating struct {
	AgentID     string  `json:"agent_id"`
	Capability  string  `json:"capability"` // "generate", "review", "debug", "refactor"
	DomainSkill string  `json:"domain_skill"` // "go", "python", "kubernetes"
	Rating      float64 `json:"rating"`       // [0.0, 1.0]
}

// CapabilityGraph maps agent capabilities to skill nodes.
type CapabilityGraph interface {
	// AddCompetency registers a rating for an agent.
	AddCompetency(ctx context.Context, rating CompetencyRating) error
	
	// RouteTask identifies the best candidate agent for a capability-skill requirement.
	RouteTask(ctx context.Context, capability string, skill string) (string, float64, error)
}
```

## Impact

- **Precise Routing**: If Agent A is a Go reviewer (Rating: 0.99) and Agent B is a Go coder (Rating: 0.92), the system routes compile errors to Agent B for code modification, but routes the final commit review to Agent A.
- **Cost Reduction**: The scheduler avoids calling expensive models (like Claude) for low-complexity capabilities (like CRUD generation), reserving them for high-complexity refactoring.

## Open Questions

1. **How do we cold-start capability ratings for new agents?**
   - New agents start with a baseline rating (e.g., 0.50). The rating is dynamically adjusted by the Evolution Engine based on validation outcome metrics.
