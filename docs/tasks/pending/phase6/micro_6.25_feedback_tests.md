# Micro-Task 6.25: Create kernel/feedback/feedback_test.go

## Info
- **File**: `kernel/feedback/feedback_test.go`
- **Package**: `feedback_test`
- **Depends on**: 6.21-6.24
- **Time**: 20 min
- **Verify**: `go test ./kernel/feedback/...`

## Purpose
Unit tests for the evaluator, scorer, learner, and ranking engine.

## EXACT code to create

```go
package feedback_test

import (
	"context"
	"testing"
	"time"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/feedback"
)

func TestHeuristicEvaluatorCompletedTask(t *testing.T) {
	eval := feedback.NewHeuristicEvaluator()

	task := &contractsagent.Task{Name: "test", Type: "code_generation"}
	result := &contractsagent.Result{
		Status: "completed",
		Output: "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
	}

	evaluation, err := eval.Evaluate(context.Background(), task, result)
	if err != nil {
		t.Fatal(err)
	}

	if evaluation.Score < 0.5 {
		t.Errorf("expected score >= 0.5 for completed task with code output, got %.2f", evaluation.Score)
	}
	if evaluation.Score > 1.0 {
		t.Errorf("score should not exceed 1.0, got %.2f", evaluation.Score)
	}
}

func TestHeuristicEvaluatorFailedTask(t *testing.T) {
	eval := feedback.NewHeuristicEvaluator()

	task := &contractsagent.Task{Name: "test"}
	result := &contractsagent.Result{
		Status: "failed",
		Error:  "timeout",
	}

	evaluation, _ := eval.Evaluate(context.Background(), task, result)
	if evaluation.Score != 0.0 {
		t.Errorf("expected 0.0 for failed task, got %.2f", evaluation.Score)
	}
}

func TestScorerRunningAverage(t *testing.T) {
	scorer := feedback.NewScorer()

	// Record 3 tasks
	scorer.Record("agent-a", &feedback.Evaluation{Score: 0.8}, 10*time.Second, 100)
	scorer.Record("agent-a", &feedback.Evaluation{Score: 0.6}, 20*time.Second, 200)
	scorer.Record("agent-a", &feedback.Evaluation{Score: 1.0}, 5*time.Second, 50)

	score := scorer.Get("agent-a")
	if score == nil {
		t.Fatal("expected score for agent-a")
	}

	if score.TotalTasks != 3 {
		t.Errorf("expected 3 total tasks, got %d", score.TotalTasks)
	}

	expectedAvg := (0.8 + 0.6 + 1.0) / 3.0
	if diff := score.AvgScore - expectedAvg; diff > 0.01 || diff < -0.01 {
		t.Errorf("expected avg score ~%.2f, got %.2f", expectedAvg, score.AvgScore)
	}
}

func TestRankingEngineBestAgent(t *testing.T) {
	scorer := feedback.NewScorer()
	scorer.Record("agent-a", &feedback.Evaluation{Score: 0.9}, 5*time.Second, 100)
	scorer.Record("agent-b", &feedback.Evaluation{Score: 0.3}, 30*time.Second, 500)

	learner, _ := feedback.NewLearner(t.TempDir())
	engine := feedback.NewRankingEngine(scorer, learner)

	best := engine.BestAgent("code_generation", []string{"agent-a", "agent-b"})
	if best != "agent-a" {
		t.Errorf("expected agent-a as best, got %q", best)
	}
}

func TestRankingEngineRoundRobinFallback(t *testing.T) {
	scorer := feedback.NewScorer()
	learner, _ := feedback.NewLearner(t.TempDir())
	engine := feedback.NewRankingEngine(scorer, learner)

	// No history → round-robin
	agents := []string{"a", "b", "c"}
	seen := make(map[string]bool)
	for i := 0; i < 6; i++ {
		best := engine.BestAgent("unknown", agents)
		seen[best] = true
	}
	if len(seen) < 2 {
		t.Error("round-robin should rotate through agents")
	}
}
```

## Verify
```bash
go test ./kernel/feedback/... -v
```

## Checklist
- [ ] Evaluator tests for completed and failed tasks
- [ ] Scorer tests for running average accuracy
- [ ] Ranking engine tests for best agent selection
- [ ] Round-robin fallback test
- [ ] All tests pass
