# Micro-Task 1.18: Tạo contracts/agent/task.go

## Thông tin
- **File tạo**: `contracts/agent/task.go`
- **Package**: `agent`
- **Dependencies trước**: 1.06 (contracts/types.go), 1.07 (contracts/status.go)
- **Thời gian**: 15 phút
- **Verify**: `go build ./contracts/agent/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Task represents a unit of work assigned to an agent.
//
// A mission is decomposed into multiple tasks by the Planner.
// Each task is assigned to an agent for execution.
//
// Example:
//
//	task := NewTask("implement_handler", "Implement the user CRUD handler", "code_generation")
//	task.AddDependency(designTaskID)
//	task.AddContext("file", fileContent, "/src/models/user.go")
type Task struct {
	// ID uniquely identifies this task.
	ID contracts.TaskID `json:"id"`

	// Name is a short descriptive name for this task.
	// Convention: snake_case (e.g., "implement_user_handler")
	Name string `json:"name"`

	// Description explains what needs to be done in detail.
	// This text is included in the prompt sent to the AI agent.
	// Be specific: the more detail here, the better the AI output.
	Description string `json:"description"`

	// Type categorizes the task for agent matching.
	// Should correspond to a Capability value (e.g., "code_generation").
	// The orchestrator uses this to find agents with matching capabilities.
	Type string `json:"type"`

	// Input contains task-specific input data.
	// Each task type has different inputs:
	//   code_generation: {"language": "go", "framework": "gin"}
	//   code_review:     {"diff": "...", "focus": "security"}
	//   testing:         {"test_type": "unit", "coverage_target": 80}
	//
	// WHY map[string]any instead of a typed struct?
	// → Different task types require different inputs.
	// → A generic map avoids creating a struct for every task type.
	// → Agent implementations can unmarshal into typed structs as needed.
	Input map[string]any `json:"input,omitempty"`

	// Context provides additional information for the agent.
	// Examples: source files, previous task outputs, documentation.
	// The agent uses this as reference material when executing the task.
	Context []ContextItem `json:"context,omitempty"`

	// Dependencies lists task IDs that must complete before this task can start.
	// The scheduler uses this to determine execution order (DAG).
	//
	// WHY []contracts.TaskID instead of []*Task?
	// → Pointer references between tasks create circular dependency risk.
	// → Pointer references cause memory leaks (GC can't collect cycles).
	// → String IDs are serializable to JSON/YAML for persistence.
	// → The DAG is built separately from the task structs.
	Dependencies []contracts.TaskID `json:"dependencies,omitempty"`

	// Priority determines execution order when multiple tasks are ready.
	// Higher number = run first. Range: 1 (lowest) to 10 (highest).
	// Default: 5 (medium priority).
	Priority int `json:"priority"`

	// Timeout is the maximum time allowed for this task.
	// If the agent doesn't finish within this time, the task is cancelled.
	// Each task has its own timeout because complexity varies.
	// Default: 5 minutes.
	Timeout time.Duration `json:"timeout"`

	// Metadata contains arbitrary key-value pairs for extensibility.
	// Use for passing custom data that doesn't fit standard fields.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ContextItem provides additional context to an agent.
// Context items are included in the prompt sent to the AI.
type ContextItem struct {
	// Type identifies the kind of context.
	// Values: "file", "snippet", "url", "memory", "task_output"
	Type string `json:"type"`

	// Content is the actual context data.
	// For "file" type: the file contents.
	// For "snippet" type: a code snippet.
	// For "task_output" type: output from a previous task.
	Content string `json:"content"`

	// Source identifies where this context came from.
	// For "file": the file path (e.g., "/src/main.go")
	// For "task_output": the task ID (e.g., "tsk-a1b2c3d4")
	// For "url": the URL
	Source string `json:"source,omitempty"`
}

// NewTask creates a new Task with a generated ID and default values.
//
// Parameters:
//   - name: short descriptive name (snake_case)
//   - description: detailed explanation of what to do
//   - taskType: capability type (e.g., "code_generation")
//
// Defaults:
//   - Priority: 5 (medium)
//   - Timeout: 5 minutes
func NewTask(name, description, taskType string) *Task {
	return &Task{
		ID:          contracts.NewTaskID(),
		Name:        name,
		Description: description,
		Type:        taskType,
		Priority:    5,               // Default: medium priority
		Timeout:     5 * time.Minute, // Default: 5 minutes
	}
}

// AddDependency adds a task dependency. Returns self for method chaining.
//
// Example:
//
//	task.AddDependency(taskA.ID).AddDependency(taskB.ID)
func (t *Task) AddDependency(taskID contracts.TaskID) *Task {
	t.Dependencies = append(t.Dependencies, taskID)
	return t
}

// AddContext adds a context item. Returns self for method chaining.
//
// Parameters:
//   - ctxType: "file", "snippet", "url", "memory", "task_output"
//   - content: the actual data
//   - source: where it came from (file path, URL, task ID)
func (t *Task) AddContext(ctxType, content, source string) *Task {
	t.Context = append(t.Context, ContextItem{
		Type:    ctxType,
		Content: content,
		Source:  source,
	})
	return t
}

// HasDependencies returns true if this task depends on other tasks.
func (t *Task) HasDependencies() bool {
	return len(t.Dependencies) > 0
}
```

## ⚠️ Pitfalls cần tránh
1. **Dependencies dùng `[]contracts.TaskID`** (named type) → compiler bắt lỗi nếu pass nhầm MissionID vào
2. **KHÔNG dùng `[]*Task` cho dependencies** → circular references → memory leak → GC không thu hồi
3. **Timeout per-task**: Mỗi task có timeout riêng vì độ phức tạp khác nhau. Code review = 2 phút, code generation = 10 phút
4. **Input dùng `map[string]any`**: Flexible, nhưng agent implementation nên unmarshal vào typed struct:
   ```go
   type CodeGenInput struct {
       Language  string `json:"language"`
       Framework string `json:"framework"`
   }
   var input CodeGenInput
   // convert task.Input to JSON then unmarshal
   ```
5. **Builder pattern**: `NewTask()` + `AddDependency()` + `AddContext()` trả về `*Task` cho chaining

## Checklist
- [ ] File `contracts/agent/task.go` tồn tại
- [ ] Package declaration: `package agent`
- [ ] Import `contracts` package (cho TaskID type)
- [ ] Task struct với 9 fields
- [ ] ContextItem struct với 3 fields
- [ ] `NewTask()` constructor với defaults (Priority=5, Timeout=5min)
- [ ] `AddDependency()` method chaining
- [ ] `AddContext()` method chaining
- [ ] `HasDependencies()` helper
- [ ] Dependencies dùng `[]contracts.TaskID` (KHÔNG `[]string`)
- [ ] JSON tags đầy đủ với `omitempty` cho optional fields
- [ ] Godoc comments với examples
- [ ] `go build ./contracts/agent/...` không lỗi
