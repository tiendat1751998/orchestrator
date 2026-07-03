# Micro-Task 1.19: Create contracts/agent/result.go

## Info
- **File**: `contracts/agent/result.go`
- **Package**: `agent`
- **Depends on**: 1.07, 1.10
- **Time**: 15 min
- **Verify**: `go build ./contracts/agent/...`

## Purpose
Defines the `Result` and `Artifact` models representing outputs produced by agent executions.

## EXACT code to create

```go
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
```

## ⚠️ Pitfalls

### Pitfall 1: Attempting to serialize `error` interface in Result.Error
```go
Error string `json:"error"` // Plain string is serializable and portable across DBs, JSON, and CLI.
```
Always marshal execution errors to string variables rather than Go interface types.

### Pitfall 2: Double-allocating artifact content for large files
For large files (e.g. built binaries or large images), loading the content directly into `Artifact.Content` will exhaust process heap space. Always store file paths in `Artifact.Path` and read them from disk on-demand.

## Verify
```bash
go build ./contracts/agent/...
```

## Checklist
- [ ] File `contracts/agent/result.go` exists
- [ ] Package: `agent`
- [ ] `Result` and `Artifact` structs are declared
- [ ] `Usage` is represented as a pointer to `provider.Usage`
- [ ] `IsSuccess()` and `IsFailed()` helpers evaluate status correctly
- [ ] `SuccessResult` and `FailedResult` builder helper functions exist
- [ ] `go build ./contracts/agent/...` passes
