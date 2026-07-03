package planner

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// ExplainPlan compares chosen plan against runners-up.
// It formats a contrastive mathematical report detailing the score differences.
func ExplainPlan(
	ctx context.Context,
	chosen fsm.DAG,
	chosenScore float64,
	candidates []fsm.DAG,
	scores []float64,
) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("planner: nil context")
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("planner: context error: %w", err)
	}

	if len(candidates) != len(scores) {
		return "", fmt.Errorf("planner: candidates count (%d) and scores count (%d) mismatch", len(candidates), len(scores))
	}

	var report strings.Builder
	report.WriteString("### Architectural Decision Rationale\n\n")
	report.WriteString(fmt.Sprintf("Selected Plan (Score: %.2f) as the optimal route.\n\n", chosenScore))
	report.WriteString("Comparison with candidates:\n")

	for i := range candidates {
		delta := chosenScore - scores[i]
		report.WriteString(fmt.Sprintf("- Candidate %d: Score delta is %.2f\n", i, delta))
	}

	return report.String(), nil
}
