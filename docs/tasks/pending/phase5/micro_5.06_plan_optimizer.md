# Micro-Task 5.06: Create kernel/planner/explain.go

- **File**: `kernel/planner/explain.go`
- **Package**: `planner`
- **Depends on**: 5.01, 5.04 (planner/pareto.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements the contrastive **Explainable Planning** reporter inside the Planner. It compares the score details of the chosen plan against alternative candidates and outputs a mathematical reasoning report detailing exactly why the plan was selected.

## EXACT code to create

```go
package planner

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// ExplainPlan compares chosen plan against runners-up.
func ExplainPlan(
	ctx context.Context,
	chosen fsm.DAG,
	chosenScore float64,
	candidates []fsm.DAG,
	scores []float64,
) (string, error) {
	var report strings.Builder
	report.WriteString("### Architectural Decision Rationale\n\n")
	report.WriteString(fmt.Sprintf("Selected Plan (Score: %.2f) as the optimal route.\n\n", chosenScore))
	report.WriteString("Comparison with candidates:\n")
	
	for i, cand := range candidates {
		delta := chosenScore - scores[i]
		report.WriteString(fmt.Sprintf("- Candidate %d: Score delta is %.2f\n", i, delta))
	}
	
	return report.String(), nil
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/explain.go` exists
- [ ] Package: `planner`
- [ ] `ExplainPlan` compiles contrastive score metrics
- [ ] Returns structured markdown report
- [ ] `go build ./kernel/planner/...` passes
