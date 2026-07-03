package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// Result represents the output of a task executed by an agent.
type Result struct {
	// TaskID links this result to the task that was executed.
	TaskID contracts.TaskID `json:"task_id"`

	// AgentName identifies which agent executed the task.
	AgentName string `json:"agent_name"`

	// Status indicates the outcome of the task.
	Status contracts.Status `json:"status"`

	// Output is the primary text output from the agent.
	Output string `json:"output"`

	// Artifacts are files, diffs, or other outputs produced by the agent.
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Error is a human-readable error description (populated on failure).
	Error string `json:"error,omitempty"`

	// Duration is how long the task took to execute.
	Duration time.Duration `json:"duration"`

	// Usage tracks token consumption from the AI provider.
	Usage *provider.Usage `json:"usage,omitempty"`

	// Metadata for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Artifact represents a file or data output produced by an agent.
type Artifact struct {
	// Name is a descriptive name (e.g. "main.go").
	Name string `json:"name"`

	// Type categorizes the artifact ("file", "diff", "log", "report", "image").
	Type string `json:"type"`

	// Path is the filesystem path (for "file" type artifacts).
	Path string `json:"path,omitempty"`

	// Content is inline content (for small artifacts).
	Content string `json:"content,omitempty"`
}

// IsSuccess returns true if the task completed successfully.
func (r *Result) IsSuccess() bool {
	return r.Status.IsSuccess()
}

// IsFailed returns true if the task failed.
func (r *Result) IsFailed() bool {
	return r.Status.IsFailed()
}

// SuccessResult creates a successful result.
func SuccessResult(taskID contracts.TaskID, agentName, output string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusSuccess,
		Output:    output,
	}
}

// FailedResult creates a failed result.
func FailedResult(taskID contracts.TaskID, agentName, errMsg string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusFailed,
		Error:     errMsg,
	}
}
