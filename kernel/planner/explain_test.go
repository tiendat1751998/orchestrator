package planner

import (
	"context"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

func TestExplainPlan(t *testing.T) {
	chosen := fsm.DAG{}
	chosenScore := 8.50
	candidates := []fsm.DAG{
		{Nodes: map[string]*fsm.DAGNode{}},
		{Nodes: map[string]*fsm.DAGNode{}},
	}

	t.Run("Success", func(t *testing.T) {
		scores := []float64{6.20, 9.10}
		report, err := ExplainPlan(context.Background(), chosen, chosenScore, candidates, scores)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedParts := []string{
			"### Architectural Decision Rationale",
			"Selected Plan (Score: 8.50) as the optimal route.",
			"Comparison with candidates:",
			"- Candidate 0: Score delta is 2.30",
			"- Candidate 1: Score delta is -0.60",
		}

		for _, p := range expectedParts {
			if !strings.Contains(report, p) {
				t.Errorf("expected report to contain %q, but got:\n%s", p, report)
			}
		}
	})

	t.Run("MismatchLength", func(t *testing.T) {
		scores := []float64{6.20}
		_, err := ExplainPlan(context.Background(), chosen, chosenScore, candidates, scores)
		if err == nil {
			t.Fatal("expected error due to slice length mismatch, got nil")
		}
		if !strings.Contains(err.Error(), "mismatch") {
			t.Errorf("expected error message to contain 'mismatch', got %q", err.Error())
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		scores := []float64{6.20, 9.10}
		_, err := ExplainPlan(ctx, chosen, chosenScore, candidates, scores)
		if err == nil {
			t.Fatal("expected error due to cancelled context, got nil")
		}
	})

	t.Run("NilContext", func(t *testing.T) {
		scores := []float64{6.20, 9.10}
		_, err := ExplainPlan(nil, chosen, chosenScore, candidates, scores) //nolint:staticcheck // testing boundary condition
		if err == nil {
			t.Fatal("expected error due to nil context, got nil")
		}
	})
}
