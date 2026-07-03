# Micro-Task 1.18: Create contracts/agent/task.go

## Info
- **File**: `contracts/agent/task.go`
- **Package**: `agent`
- **Depends on**: 1.06 (contracts/types.go), 1.07 (contracts/status.go)
- **Time**: 15 min
- **Verify**: `go build ./contracts/agent/...`

## Purpose
Declares the `Task` and `ContextItem` schema structures used to delegate specific execution tasks to AI agents.

## EXACT code to create

```go
package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Task represents a unit of work assigned to an agent.
type Task struct {
	ID           contracts.TaskID  `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Type         string            `json:"type"`
	Input        map[string]any    `json:"input,omitempty"`
	Context      []ContextItem     `json:"context,omitempty"`
	Dependencies []contracts.TaskID `json:"dependencies,omitempty"`
	Priority     int               `json:"priority"`
	Timeout      time.Duration     `json:"timeout"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ContextItem provides additional context to an agent.
type ContextItem struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Source  string `json:"source,omitempty"`
}

// NewTask creates a new Task with a generated ID and default values.
func NewTask(name, description, taskType string) *Task {
	return &Task{
		ID:          contracts.NewTaskID(),
		Name:        name,
		Description: description,
		Type:        taskType,
		Priority:    5,
		Timeout:     5 * time.Minute,
	}
}

// AddDependency adds a task dependency. Returns self for method chaining.
func (t *Task) AddDependency(taskID contracts.TaskID) *Task {
	t.Dependencies = append(t.Dependencies, taskID)
	return t
}

// AddContext adds a context item. Returns self for method chaining.
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

## Pitfalls

### Pitfall 1: Referencing pointer tasks in dependencies collections
```go
// WRONG:
type Task struct {
    Dependencies []*Task // Creates circular reference maps: memory leaks on garbage collections!
}

// CORRECT:
type Task struct {
    Dependencies []contracts.TaskID // Serialized value types prevent GC leaks
}
```
Using task pointers to map dependency structures inside objects risks circular reference leaks. Standardise on unique string values.

### Pitfall 2: Overriding tasks input boundaries
Using loose inputs maps is flexible, but agent implementations should deserialize parameters to defined schemas inside executor packages.

## Verify
```bash
go build ./contracts/agent/...
```

## Checklist
- [ ] File exists at `contracts/agent/task.go`
- [ ] Package name is `agent`
- [ ] Task and ContextItem structures are defined
- [ ] Dependency collections use contracts.TaskID values
- [ ] Build command passes
