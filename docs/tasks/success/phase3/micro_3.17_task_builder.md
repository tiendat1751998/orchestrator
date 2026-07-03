# Micro-Task 3.17: Create sdk/task/task.go

## Info
- **File**: `sdk/task/task.go`
- **Package**: `task`
- **Depends on**: 1.18 (task.go contract), 1.39 (input validation contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/task/...`

## Purpose
Implements the task builder (`TaskBuilder` and constructors) that enables programmatically creating `agent.Task` instances with an immutable fluent configuration style, ensuring task validation checks are applied before compilation.

## EXACT code to create

```go
// Package task provides fluent builders for creating agent tasks.
package task

import (
	"fmt"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// TaskBuilder constructs an agent.Task in a fluent, immutable style.
type TaskBuilder struct {
	t agent.Task
}

// NewTaskBuilder initializes a new TaskBuilder.
// Sets default values: Priority = 5, Timeout = 5 minutes.
func NewTaskBuilder(name, taskType string) *TaskBuilder {
	return &TaskBuilder{
		t: agent.Task{
			ID:          contracts.NewTaskID(),
			Name:        name,
			Type:        taskType,
			Input:       make(map[string]any),
			Priority:    5,
			Timeout:     5 * time.Minute,
			Metadata:    make(map[string]string),
		},
	}
}

// clone creates a deep copy of the builder's internal task configuration.
func (b *TaskBuilder) clone() *TaskBuilder {
	copiedInput := make(map[string]any)
	for k, v := range b.t.Input {
		copiedInput[k] = v
	}

	var copiedCtx []agent.ContextItem
	if len(b.t.Context) > 0 {
		copiedCtx = make([]agent.ContextItem, len(b.t.Context))
		copy(copiedCtx, b.t.Context)
	}

	var copiedDeps []contracts.TaskID
	if len(b.t.Dependencies) > 0 {
		copiedDeps = make([]contracts.TaskID, len(b.t.Dependencies))
		copy(copiedDeps, b.t.Dependencies)
	}

	copiedMeta := make(map[string]string)
	for k, v := range b.t.Metadata {
		copiedMeta[k] = v
	}

	newT := agent.Task{
		ID:           b.t.ID,
		Name:         b.t.Name,
		Description:  b.t.Description,
		Type:         b.t.Type,
		Input:        copiedInput,
		Context:      copiedCtx,
		Dependencies: copiedDeps,
		Priority:     b.t.Priority,
		Timeout:      b.t.Timeout,
		Metadata:     copiedMeta,
	}

	return &TaskBuilder{t: newT}
}

// WithID overrides the automatically generated task ID.
func (b *TaskBuilder) WithID(id contracts.TaskID) *TaskBuilder {
	nb := b.clone()
	nb.t.ID = id
	return nb
}

// WithDescription sets the detailed task guidelines.
func (b *TaskBuilder) WithDescription(desc string) *TaskBuilder {
	nb := b.clone()
	nb.t.Description = desc
	return nb
}

// WithPriority sets task execution priority (1-10).
func (b *TaskBuilder) WithPriority(priority int) *TaskBuilder {
	nb := b.clone()
	nb.t.Priority = priority
	return nb
}

// WithTimeout sets task deadline timeout duration.
func (b *TaskBuilder) WithTimeout(timeout time.Duration) *TaskBuilder {
	nb := b.clone()
	nb.t.Timeout = timeout
	return nb
}

// WithInput sets the complete input parameters map.
func (b *TaskBuilder) WithInput(input map[string]any) *TaskBuilder {
	nb := b.clone()
	nb.t.Input = input
	return nb
}

// AddInput inserts a single key-value parameter into the inputs map.
func (b *TaskBuilder) AddInput(key string, val any) *TaskBuilder {
	nb := b.clone()
	if nb.t.Input == nil {
		nb.t.Input = make(map[string]any)
	}
	nb.t.Input[key] = val
	return nb
}

// AddDependency appends a task ID that this task depends on.
func (b *TaskBuilder) AddDependency(depID contracts.TaskID) *TaskBuilder {
	nb := b.clone()
	nb.t.Dependencies = append(nb.t.Dependencies, depID)
	return nb
}

// AddContextItem attaches a custom context item.
func (b *TaskBuilder) AddContextItem(item agent.ContextItem) *TaskBuilder {
	nb := b.clone()
	nb.t.Context = append(nb.t.Context, item)
	return nb
}

// AddFileContext is a shorthand to attach a file content context item.
func (b *TaskBuilder) AddFileContext(filePath, content string) *TaskBuilder {
	return b.AddContextItem(agent.ContextItem{
		Type:    "file",
		Content: content,
		Source:  filePath,
	})
}

// AddMetadata inserts a single key-value string metadata detail.
func (b *TaskBuilder) AddMetadata(key, val string) *TaskBuilder {
	nb := b.clone()
	if nb.t.Metadata == nil {
		nb.t.Metadata = make(map[string]string)
	}
	nb.t.Metadata[key] = val
	return nb
}

// Build validates and compiles the final Task instance.
func (b *TaskBuilder) Build() (*agent.Task, error) {
	finalB := b.clone()
	task := &finalB.t

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("sdk/task: failed to build valid task: %w", err)
	}

	return task, nil
}
```

## Rules
1. **Immutable Modifications**: Clones task configurations before any configuration changes to avoid reference leaking across threads.
2. **Deep Copies of Collections**: Ensure deep copies are made for maps (`Input`, `Metadata`) and slices (`Context`, `Dependencies`) during cloning.
3. **Task Validation**: Enforce `task.Validate()` check calls inside the final `Build()` phase to capture configuration bounds issues.

## ⚠️ Pitfalls

### Pitfall 1: Mutating maps or slices without copying
```go
```
Always deep copy collections during `clone()` transitions to preserve data boundaries.

### Pitfall 2: Bypassing structural validations during compilation
Skipping validation checks at the end of compilation can lead to submitting invalid tasks (such as missing type declarations) to execution schedulers, causing engine crashes. Always run validations.

## Verify
```bash
go build ./sdk/task/...
```

## Checklist
- [ ] File `sdk/task/task.go` exists
- [ ] Package: `task`
- [ ] `TaskBuilder` implements immutable fluent API methods
- [ ] `clone` deep copies maps (`Input`, `Metadata`)
- [ ] `clone` deep copies slices (`Context`, `Dependencies`)
- [ ] `NewTaskBuilder` initializes priority and timeout defaults
- [ ] `AddFileContext` creates file context inputs correctly
- [ ] `Build` validates tasks before outputting them
- [ ] `go build ./sdk/task/...` passes
