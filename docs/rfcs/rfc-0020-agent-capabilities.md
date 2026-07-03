# RFC-0020: Agent Capabilities & Skill Tree Metrics

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0010 (Cognitive Layer), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design of **Agent Capabilities & Skill Tree Metrics** in AEOS. It defines the schema and execution hooks to measure and score agent task output quality, compile rates, linter compliance, and runtime errors, and maps them to technical skill tree nodes.

## Motivation

To route tasks to the best agent, the kernel needs to score agent performance objectively.
- Static capabilities are insufficient because model capabilities change and agents learn from success.
- By connecting validation outcome metrics to the agent's competency ratings, the system can choose models and agents dynamically.

## Design

### 1. Architectural Placement

The Capability metrics are tracked inside the `Knowledge Engine` and updated by the `Evolution Engine`.

```
  Validation outputs ──► [Evolution Engine] ──► Update Agent Competency metrics in DB
```

---

### 2. Contracts (`contracts/brain/metrics.go`)

```go
package brain

import "context"

// SkillMetric represents an agent's evaluated performance on a technology.
type SkillMetric struct {
	AgentID      string  `json:"agent_id"`
	SkillNodeID  string  `json:"skill_node_id"` // e.g. "go", "gin", "docker"
	SuccessCount int     `json:"success_count"`
	TotalCount   int     `json:"total_count"`
	Score        float64 `json:"score"` // Success/Total ratio
}

// CompetencyRegistry tracks agent metrics.
type CompetencyRegistry interface {
	// LogTaskOutcome records the outcome of a task execution.
	LogTaskOutcome(ctx context.Context, agentID string, skillID string, success bool) error
	
	// GetSkillMetric retrieves the competency score of an agent.
	GetSkillMetric(ctx context.Context, agentID string, skillID string) (*SkillMetric, error)
}
```

## Impact

- **Metric-Based Routing**: The scheduler routes code refactoring tasks to agents with high refactoring competency ratings, avoiding model errors.
- **Auto-evolving Registry**: The system adjusts agent metrics over time, dynamically adapting to new models.
