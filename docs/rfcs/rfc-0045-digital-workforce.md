# RFC-0045: Digital Workforce

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0035 (Capability Graph), RFC-0044 (Economic Engine)

## Summary

This RFC specifies the design of the **Digital Workforce** in AEOS. It models agents as **Virtual Employees** with distinct roles, competency graphs, virtual salaries, reliability ratings, and learning curves, where the Planner acts as the Virtual CTO directing and scheduling the workforce.

## Motivation

Treating LLM calls as simple API endpoints limits organization.
- For complex software engineering tasks, we need dedicated specialists: senior backend developer, frontend coder, security reviewer, QA engineer, and release manager.
- By modeling agents as structured, virtual employees, the system can assign, evaluate, and budget the workforce programmatically.

## Design

### 1. Architectural Placement

The Digital Workforce is managed inside the `Brain Runtime` and queried by the `Scheduler` to map task targets to available virtual agents.

```
  Plan DAG ──► [Scheduler] ──► Query [Digital Workforce] ──► Assign Agent Node
```

---

### 2. Contracts (`contracts/brain/workforce.go`)

```go
package brain

import "context"

// VirtualEmployee represents a configured agent in the system.
type VirtualEmployee struct {
	ID                 string             `json:"id"`
	Role               string             `json:"role"` // "Senior Go Developer", "QA Engineer", "Security Reviewer"
	ModelName          string             `json:"model_name"` // "gemini-1.5-pro", "claude-3-5-sonnet"
	VirtualSalaryPerHour float64          `json:"virtual_salary_per_hour"` // Estimated USD cost
	ReliabilityScore   float64            `json:"reliability_score"` // [0.0, 1.0] (EMA-updated)
	Capabilities       []string           `json:"capabilities"`
}

// WorkforceManager coordinates virtual employees.
type WorkforceManager interface {
	// RegisterEmployee registers a new agent in the workforce.
	RegisterEmployee(ctx context.Context, emp VirtualEmployee) error
	
	// GetBestFitAgent selects the optimal employee for a task.
	GetBestFitAgent(ctx context.Context, roleRequired string, budgetMax float64) (*VirtualEmployee, error)
	
	// UpdateReliability recalculates the agent's score after a task completes.
	UpdateReliability(ctx context.Context, employeeID string, success bool) error
}
```

## Impact

- **CTO Delegation**: The Planner schedules tasks at a high level. The Scheduler coordinates assignments among virtual specialists based on budget and capability constraints.
- **Dynamic Training**: Virtual employees accumulate reliability scores over time, allowing the system to deprecate or promote agents.

## Open Questions

1. **How is the virtual salary calculated?**
   - The virtual salary is derived from actual API token pricing (input/output counts) and the average latency of the underlying model.
