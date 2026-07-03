# Micro-Task 6.21: Create kernel/feedback/evaluator.go

## Info
- **File**: `kernel/feedback/evaluator.go`
- **Package**: `feedback`
- **Depends on**: 1.19 (agent.Result)
- **Time**: 20 min
- **Verify**: `go build ./kernel/feedback/...`

## Purpose
Evaluates task output quality using heuristic-based scoring (Phase 1). Extensible interface for future AI-as-judge integration.

## EXACT code to create

```go
// Package feedback implements task output evaluation, agent scoring, and learning.
package feedback

import (
	"context"
	"strings"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
)

// Evaluation holds the quality assessment of a task result.
type Evaluation struct {
	TaskID     string  `json:"task_id"`
	Score      float64 `json:"score"`      // 0.0 - 1.0
	Confidence float64 `json:"confidence"` // 0.0 - 1.0
	Reasons    []string `json:"reasons,omitempty"`
}

// Evaluator assesses the quality of task outputs.
type Evaluator interface {
	Evaluate(ctx context.Context, task *contractsagent.Task, result *contractsagent.Result) (*Evaluation, error)
}

// HeuristicEvaluator uses rule-based heuristics for evaluation.
// Phase 1 implementation — no AI dependency.
type HeuristicEvaluator struct{}

// NewHeuristicEvaluator constructs a heuristic-based evaluator.
func NewHeuristicEvaluator() *HeuristicEvaluator {
	return &HeuristicEvaluator{}
}

// Evaluate scores the result using simple heuristics.
func (e *HeuristicEvaluator) Evaluate(ctx context.Context, task *contractsagent.Task, result *contractsagent.Result) (*Evaluation, error) {
	eval := &Evaluation{
		TaskID:     string(task.ID),
		Score:      0.0,
		Confidence: 0.5, // Heuristic confidence is moderate
	}

	// 1. Status-based scoring
	switch result.Status {
	case "completed":
		eval.Score = 0.7
		eval.Reasons = append(eval.Reasons, "task completed successfully")
	case "failed":
		eval.Score = 0.0
		eval.Reasons = append(eval.Reasons, "task failed: "+result.Error)
		return eval, nil
	default:
		eval.Score = 0.3
		eval.Reasons = append(eval.Reasons, "task status unknown: "+result.Status)
	}

	// 2. Output quality heuristics
	output := result.Output
	if output == "" {
		eval.Score -= 0.2
		eval.Reasons = append(eval.Reasons, "empty output")
	} else {
		// Reward substantive output
		if len(output) > 100 {
			eval.Score += 0.1
			eval.Reasons = append(eval.Reasons, "substantive output length")
		}

		// Check for code output indicators
		if strings.Contains(output, "func ") || strings.Contains(output, "package ") {
			eval.Score += 0.1
			eval.Reasons = append(eval.Reasons, "contains Go code")
		}

		// Penalty for error patterns in output
		errorPatterns := []string{"error:", "panic:", "FAIL", "undefined:"}
		for _, pattern := range errorPatterns {
			if strings.Contains(strings.ToLower(output), strings.ToLower(pattern)) {
				eval.Score -= 0.1
				eval.Reasons = append(eval.Reasons, "output contains error pattern: "+pattern)
			}
		}
	}

	// Clamp score to [0, 1]
	if eval.Score < 0 {
		eval.Score = 0
	}
	if eval.Score > 1 {
		eval.Score = 1
	}

	return eval, nil
}
```

## Rules
1. **Interface First**: `Evaluator` is an interface. `HeuristicEvaluator` is Phase 1 implementation. AI-as-judge can replace it later.
2. **Bounded Scores**: Always clamp to [0.0, 1.0].
3. **Confidence Level**: Heuristic evaluation has moderate confidence (0.5). AI-based evaluation would report higher confidence.

## Verify
```bash
go build ./kernel/feedback/...
```

## Checklist
- [ ] Evaluator interface defined
- [ ] HeuristicEvaluator implements Evaluator
- [ ] Score clamped to [0, 1]
- [ ] `go build ./kernel/feedback/...` passes
