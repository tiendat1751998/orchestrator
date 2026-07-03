// Package feedback defines contracts for quality evaluation and agent scoring.
// The feedback system enables continuous improvement of agent performance.
package feedback

import "context"

// Evaluator assesses the quality of agent outputs.
//
// Evaluation methods:
//   - Heuristic: "go build" passes? "go test" passes? (fast, reliable)
//   - AI-as-judge: ask another AI to score the output (flexible, but costly)
type Evaluator interface {
	// Evaluate scores an agent's output.
	//
	// Returns a score (0-1) and feedback text.
	// Score interpretation:
	//   0.0 - 0.3: Poor quality
	//   0.3 - 0.6: Acceptable
	//   0.6 - 0.8: Good
	//   0.8 - 1.0: Excellent
	Evaluate(ctx context.Context, input EvalInput) (*EvalResult, error)
}

// EvalInput is the data to evaluate.
type EvalInput struct {
	// TaskType is the type of task that was executed.
	TaskType string `json:"task_type"`

	// AgentName identifies which agent produced the output.
	AgentName string `json:"agent_name"`

	// Output is the agent's text output to evaluate.
	Output string `json:"output"`

	// Expected is the expected output (if available).
	// Used for comparison-based evaluation.
	Expected string `json:"expected,omitempty"`
}

// EvalResult is the evaluation outcome.
type EvalResult struct {
	// Score is the quality score (0.0 to 1.0).
	Score float64 `json:"score"`

	// Feedback is a human-readable explanation of the score.
	Feedback string `json:"feedback"`

	// Pass indicates whether the output meets minimum quality standards.
	// Threshold is configured per evaluator (typically 0.6).
	Pass bool `json:"pass"`
}

// Scorer tracks agent performance over time.
//
// The scorer aggregates evaluation results and computes
// per-agent, per-task-type performance metrics.
//
// Used by the planner to select the best agent for a task.
type Scorer interface {
	// RecordResult records a task execution result for scoring.
	//
	// Parameters:
	//   - agentName: which agent executed the task
	//   - taskType: the type of task (e.g., "code_generation")
	//   - success: whether the task succeeded
	//   - duration: execution time in seconds
	RecordResult(agentName, taskType string, success bool, duration float64)

	// GetScore returns the performance score for an agent on a task type.
	// Returns 0.0 if no data is available.
	GetScore(agentName, taskType string) float64

	// GetRanking returns agents ranked by performance for a task type.
	// Best performers first.
	GetRanking(taskType string) []AgentScore
}

// AgentScore represents an agent's performance metrics.
type AgentScore struct {
	// AgentName identifies the agent.
	AgentName string `json:"agent_name"`

	// Score is the overall performance score (0-1).
	Score float64 `json:"score"`

	// TaskCount is the total number of tasks executed.
	TaskCount int `json:"task_count"`

	// SuccessRate is the percentage of successful tasks (0-100).
	SuccessRate float64 `json:"success_rate"`
}
