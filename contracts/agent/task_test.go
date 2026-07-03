package agent

import (
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

func TestTask(t *testing.T) {
	task := NewTask("test-task", "Description of test task", "execute")

	if task.ID.IsEmpty() {
		t.Error("expected task ID to be non-empty")
	}
	if task.Name != "test-task" {
		t.Errorf("expected task.Name to be 'test-task', got %q", task.Name)
	}
	if task.Description != "Description of test task" {
		t.Errorf("expected task.Description to be 'Description of test task', got %q", task.Description)
	}
	if task.Type != "execute" {
		t.Errorf("expected task.Type to be 'execute', got %q", task.Type)
	}
	if task.Priority != 5 {
		t.Errorf("expected task.Priority to be 5, got %d", task.Priority)
	}
	if task.Timeout != 5*time.Minute {
		t.Errorf("expected task.Timeout to be 5 minutes, got %v", task.Timeout)
	}

	if task.HasDependencies() {
		t.Error("expected task to have no dependencies initially")
	}

	depID := contracts.NewTaskID()
	task.AddDependency(depID)

	if !task.HasDependencies() {
		t.Error("expected task to have dependencies after adding one")
	}
	if len(task.Dependencies) != 1 || task.Dependencies[0] != depID {
		t.Errorf("expected task.Dependencies to contain %q", depID)
	}

	task.AddContext("file", "content-data", "test-source")
	if len(task.Context) != 1 {
		t.Fatalf("expected 1 context item, got %d", len(task.Context))
	}
	ctx := task.Context[0]
	if ctx.Type != "file" || ctx.Content != "content-data" || ctx.Source != "test-source" {
		t.Errorf("unexpected context item content: %+v", ctx)
	}
}

func TestTaskValidate(t *testing.T) {
	// 1. Successful validation
	task := NewTask("test-task", "Description", "execute")
	if err := task.Validate(); err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}

	// 2. Empty ID
	invalidTask := NewTask("test-task", "Description", "execute")
	invalidTask.ID = ""
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for empty task ID, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "id" || vErr.Reason != "required" {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 3. Empty Name
	invalidTask = NewTask("", "Description", "execute")
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for empty task Name, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "name" || vErr.Reason != "required" {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 4. Empty Type
	invalidTask = NewTask("test-task", "Description", "")
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for empty task Type, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "type" || vErr.Reason != "required" {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 5. Negative Timeout
	invalidTask = NewTask("test-task", "Description", "execute")
	invalidTask.Timeout = -1
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for negative Timeout, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "timeout" || vErr.Reason != "must be >= 0" {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 6. Empty Dependency ID
	invalidTask = NewTask("test-task", "Description", "execute")
	invalidTask.Dependencies = []contracts.TaskID{""}
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for empty dependency ID, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "dependencies" || vErr.Reason != "contains empty dependency ID" {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 7. Self Dependency ID
	invalidTask = NewTask("test-task", "Description", "execute")
	invalidTask.Dependencies = []contracts.TaskID{invalidTask.ID}
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for self-dependency, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "dependencies" || vErr.Reason != "self-dependency is not allowed: "+string(invalidTask.ID) {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}

	// 8. Duplicate Dependency ID
	invalidTask = NewTask("test-task", "Description", "execute")
	depID := contracts.NewTaskID()
	invalidTask.Dependencies = []contracts.TaskID{depID, depID}
	if err := invalidTask.Validate(); err == nil {
		t.Error("expected error for duplicate dependency, got nil")
	} else {
		vErr, ok := err.(*contracts.ValidationError)
		if !ok || vErr.Component != "task" || vErr.Field != "dependencies" || vErr.Reason != "duplicate dependency: "+string(depID) {
			t.Errorf("unexpected validation error structure: %v", err)
		}
	}
}
