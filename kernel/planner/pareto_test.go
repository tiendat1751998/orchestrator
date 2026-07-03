package planner

import (
	"context"
	"math"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

func TestScoreCandidate(t *testing.T) {
	weights := Weights{
		Quality:    0.3,
		Cost:       0.2,
		Time:       0.1,
		Confidence: 0.2,
		Risk:       0.2,
	}
	cFactor := 2.0
	scorer := NewScorer(weights, cFactor)
	ctx := context.Background()
	dag := fsm.DAG{}

	// Test case 1: usageCount == 0 (Exploration bonus should be 1.0)
	// pareto score calculation:
	// 0.3*0.8 + 0.2*0.9 + 0.2*(1.0-0.4) + 0.1*(1.0-0.5) - 0.2*0.1
	// = 0.24 + 0.18 + 0.2*0.6 + 0.1*0.5 - 0.02
	// = 0.24 + 0.18 + 0.12 + 0.05 - 0.02 = 0.57
	// Expected score: pareto + 1.0 = 1.57
	score1 := scorer.ScoreCandidate(ctx, dag, 0.8, 0.4, 0.5, 0.9, 0.1, 10, 0)
	expected1 := 1.57
	if math.Abs(score1-expected1) > 1e-9 {
		t.Errorf("Expected score %f, got %f", expected1, score1)
	}

	// Test case 2: usageCount > 0, totalRuns > 0
	// totalRuns = 100, usageCount = 10
	// exploration bonus: 2.0 * sqrt(ln(100)/10)
	// ln(100) = 4.605170185988092
	// sqrt(4.605170185988092 / 10) = sqrt(0.4605170185988092) = 0.6786140424422584
	// exploration = 2.0 * 0.6786140424422584 = 1.3572280848845169
	// Pareto base: 0.57
	// Expected total score: 0.57 + 1.3572280848845169 = 1.9272280848845169
	score2 := scorer.ScoreCandidate(ctx, dag, 0.8, 0.4, 0.5, 0.9, 0.1, 100, 10)
	expected2 := 0.57 + 1.3572280848845169
	if math.Abs(score2-expected2) > 1e-9 {
		t.Errorf("Expected score %f, got %f", expected2, score2)
	}
}
