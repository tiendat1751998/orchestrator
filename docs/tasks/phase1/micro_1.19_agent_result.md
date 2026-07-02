# Micro-Task 1.19: Tạo contracts/agent/result.go

## Thông tin
- **File tạo**: `contracts/agent/result.go`
- **Package**: `agent`
- **Dependencies trước**: 1.07 (contracts/status.go), 1.10 (contracts/provider/response.go — cần Usage type)
- **Thời gian**: 15 phút
- **Verify**: `go build ./contracts/agent/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// Result represents the output of a task executed by an agent.
//
// Error handling convention:
//   - System errors (network, panic): Agent.Execute() returns (nil, error)
//   - Task failures (AI can't do it): returns (Result{Status: Failed}, nil)
//   - NEVER return (non-nil Result, non-nil error) simultaneously
type Result struct {
	// TaskID links this result to the task that was executed.
	TaskID contracts.TaskID `json:"task_id"`

	// AgentName identifies which agent executed the task.
	AgentName string `json:"agent_name"`

	// Status indicates the outcome of the task.
	// Use contracts.StatusSuccess, contracts.StatusFailed, etc.
	Status contracts.Status `json:"status"`

	// Output is the primary text output from the agent.
	// This is the "answer" to the task.
	Output string `json:"output"`

	// Artifacts are files, diffs, or other outputs produced by the agent.
	// Example: generated code files, test reports, architecture diagrams.
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Error is a human-readable error description.
	// Populated when Status is Failed or Timeout.
	//
	// WHY string instead of error interface?
	// → The error interface cannot be serialized to JSON.
	//   json.Marshal(error) produces {} because error has no exported fields.
	// → String is portable: JSON, gRPC, database, log files all handle strings.
	// → For programmatic error checking, use Status field instead.
	Error string `json:"error,omitempty"`

	// Duration is how long the task took to execute.
	Duration time.Duration `json:"duration"`

	// Usage tracks token consumption from the AI provider.
	// Nil if the task didn't use a provider (e.g., pure tool execution).
	//
	// WHY pointer (*provider.Usage)?
	// → Not all tasks use a provider. Tool-only tasks have no usage.
	// → Pointer allows nil to mean "no usage data".
	Usage *provider.Usage `json:"usage,omitempty"`

	// Metadata for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Artifact represents a file or data output produced by an agent.
type Artifact struct {
	// Name is a descriptive name.
	// Example: "main.go", "api_diff", "test_report"
	Name string `json:"name"`

	// Type categorizes the artifact.
	// Values: "file", "diff", "log", "report", "image"
	Type string `json:"type"`

	// Path is the filesystem path (for "file" type artifacts).
	// Use either Path or Content, not both.
	Path string `json:"path,omitempty"`

	// Content is inline content (for small artifacts).
	// For diffs, reports, short code snippets.
	// Use Path for large files to avoid memory issues.
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
//
// Parameters:
//   - taskID: the task that was executed
//   - agentName: the agent that executed it
//   - output: the text output
func SuccessResult(taskID contracts.TaskID, agentName, output string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusSuccess,
		Output:    output,
	}
}

// FailedResult creates a failed result.
//
// Parameters:
//   - taskID: the task that was executed
//   - agentName: the agent that executed it
//   - errMsg: human-readable error description
func FailedResult(taskID contracts.TaskID, agentName, errMsg string) *Result {
	return &Result{
		TaskID:    taskID,
		AgentName: agentName,
		Status:    contracts.StatusFailed,
		Error:     errMsg,
	}
}
```

## ⚠️ Pitfalls cần tránh
1. **Error là `string`, KHÔNG phải `error` interface**: `json.Marshal(error)` trả về `{}` vì error không có exported fields. Dùng string.
2. **Usage là `*provider.Usage` (pointer)**: Nil khi task không dùng provider. Value type (`provider.Usage{}`) không phân biệt được "không dùng" vs "dùng nhưng 0 tokens".
3. **Artifact: Path vs Content**: CHỌN 1. Path cho files lớn (tránh load toàn bộ vào memory). Content cho data nhỏ (<10KB).
4. **Import cycle check**: `agent` package import `provider` package. OK vì `provider` KHÔNG import `agent`. Nếu `provider` cần import `agent` → import cycle → build fail.
5. **Result constructor convention**: `SuccessResult()` và `FailedResult()` KHÔNG set Duration — caller set sau khi biết execution time.

## Checklist
- [ ] File `contracts/agent/result.go` tồn tại
- [ ] Package: `package agent`
- [ ] Import `contracts` (cho Status, TaskID) và `contracts/provider` (cho Usage)
- [ ] Result struct với 8 fields
- [ ] Artifact struct với 4 fields
- [ ] Error field là `string` (KHÔNG `error`)
- [ ] Usage field là `*provider.Usage` (pointer)
- [ ] `IsSuccess()` method
- [ ] `IsFailed()` method
- [ ] `SuccessResult()` constructor
- [ ] `FailedResult()` constructor
- [ ] JSON tags với `omitempty` cho optional fields
- [ ] Godoc comments
- [ ] `go build ./contracts/agent/...` không lỗi
- [ ] Không có import cycle
