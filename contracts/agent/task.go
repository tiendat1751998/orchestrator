package agent

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Task represents a unit of work assigned to an agent.
type Task struct {
	ID           contracts.TaskID   `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Type         string             `json:"type"`
	Input        map[string]any     `json:"input,omitempty"`
	Context      []ContextItem      `json:"context,omitempty"`
	Dependencies []contracts.TaskID `json:"dependencies,omitempty"`
	Priority     int                `json:"priority"`
	Timeout      time.Duration      `json:"timeout"`
	Metadata     map[string]string  `json:"metadata,omitempty"`
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

// Validate checks if the task has all required fields and valid dependency mappings.
//
// Validation rules:
//   - ID cannot be empty.
//   - Name cannot be empty.
//   - Type cannot be empty.
//   - Timeout must be >= 0.
//   - Dependencies must not contain duplicates and cannot self-reference the task.
func (t *Task) Validate() error {
	if t.ID.IsEmpty() {
		return contracts.NewValidationError("task", "id", "required")
	}
	if t.Name == "" {
		return contracts.NewValidationError("task", "name", "required")
	}
	if t.Type == "" {
		return contracts.NewValidationError("task", "type", "required")
	}
	if t.Timeout < 0 {
		return contracts.NewValidationError("task", "timeout", "must be >= 0")
	}

	// Validate dependencies
	seen := make(map[contracts.TaskID]bool)
	for _, depID := range t.Dependencies {
		if depID.IsEmpty() {
			return contracts.NewValidationError("task", "dependencies", "contains empty dependency ID")
		}
		if depID == t.ID {
			return contracts.NewValidationError("task", "dependencies", "self-dependency is not allowed: "+string(depID))
		}
		if seen[depID] {
			return contracts.NewValidationError("task", "dependencies", "duplicate dependency: "+string(depID))
		}
		seen[depID] = true
	}

	return nil
}
