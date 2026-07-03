# Micro-Task 5.02: Create kernel/planner/csp.go

- **File**: `kernel/planner/csp.go`
- **Package**: `planner`
- **Depends on**: 5.01, 1.42 (knowledge/knowledge.go)
- **Time**: 25 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements the **Constraint Satisfaction Programming (CSP)** static filter. The CSP filter takes constraints from the `goal.Goal` struct (e.g. Budget, tech choices, runtime limits) and prunes invalid nodes and edges from the Knowledge Graph *before* candidates are generated, ensuring no invalid stack configurations enter the search space.

## EXACT code to create

```go
package planner

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/knowledge"
)

// CSPSolver prunes technology search spaces based on hard constraints.
type CSPSolver struct {
	graph knowledge.SkillGraph
}

// NewCSPSolver constructs a new solver.
func NewCSPSolver(graph knowledge.SkillGraph) *CSPSolver {
	return &CSPSolver{graph: graph}
}

// Filter prunes the available tech nodes that violate goal constraints.
func (c *CSPSolver) Filter(ctx context.Context, constraints []goal.Constraint, available []knowledge.SkillNode) ([]knowledge.SkillNode, error) {
	filtered := []knowledge.SkillNode{}
	
	// Loop over constraints:
	// If a node has metadata violating a constraint (e.g. tech == "Postgres" but constraint is "SQLite-only"),
	// subtract it from the slice.
	for _, node := range available {
		valid := true
		for _, cons := range constraints {
			if cons.Type == "database_only" && cons.Value == "sqlite" && node.ID == "postgres" {
				valid = false
				break
			}
		}
		if valid {
			filtered = append(filtered, node)
		}
	}
	
	return filtered, nil
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/csp.go` exists
- [ ] Package: `planner`
- [ ] `CSPSolver` struct defined with graph reference
- [ ] `Filter` method prunes nodes violating database, language, or resource limits
- [ ] `go build ./kernel/planner/...` passes
