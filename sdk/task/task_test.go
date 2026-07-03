package task

import (
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// TestDefaultParameters verifies default parameters initialization.
func TestDefaultParameters(t *testing.T) {
	builder := NewTaskBuilder("test-task", "test-type")
	task, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}

	if task.ID == "" {
		t.Error("expected auto-generated task ID, got empty string")
	}
	if task.Name != "test-task" {
		t.Errorf("expected Name to be 'test-task', got '%s'", task.Name)
	}
	if task.Type != "test-type" {
		t.Errorf("expected Type to be 'test-type', got '%s'", task.Type)
	}
	if task.Priority != 5 {
		t.Errorf("expected default Priority to be 5, got %d", task.Priority)
	}
	if task.Timeout != 5*time.Minute {
		t.Errorf("expected default Timeout to be 5m, got %v", task.Timeout)
	}
	if task.Input == nil {
		t.Error("expected non-nil default Input map")
	}
	if task.Metadata == nil {
		t.Error("expected non-nil default Metadata map")
	}
	if len(task.Context) != 0 {
		t.Errorf("expected default Context slice to be empty, got length %d", len(task.Context))
	}
	if len(task.Dependencies) != 0 {
		t.Errorf("expected default Dependencies slice to be empty, got length %d", len(task.Dependencies))
	}
}

// TestBuilderImmutability verifies that modifying a derived builder does not mutate the parent builder.
func TestBuilderImmutability(t *testing.T) {
	base := NewTaskBuilder("base-task", "base-type")

	derived1 := base.WithDescription("derived description")
	derived2 := derived1.WithPriority(8)

	baseTask, err := base.Build()
	if err != nil {
		t.Fatalf("failed to build base task: %v", err)
	}
	if baseTask.Description != "" {
		t.Errorf("expected base task description to be empty, got '%s'", baseTask.Description)
	}
	if baseTask.Priority != 5 {
		t.Errorf("expected base task priority to be 5, got %d", baseTask.Priority)
	}

	d1Task, err := derived1.Build()
	if err != nil {
		t.Fatalf("failed to build derived1 task: %v", err)
	}
	if d1Task.Description != "derived description" {
		t.Errorf("expected derived1 task description to be 'derived description', got '%s'", d1Task.Description)
	}
	if d1Task.Priority != 5 {
		t.Errorf("expected derived1 task priority to be 5, got %d", d1Task.Priority)
	}

	d2Task, err := derived2.Build()
	if err != nil {
		t.Fatalf("failed to build derived2 task: %v", err)
	}
	if d2Task.Description != "derived description" {
		t.Errorf("expected derived2 task description to be 'derived description', got '%s'", d2Task.Description)
	}
	if d2Task.Priority != 8 {
		t.Errorf("expected derived2 task priority to be 8, got %d", d2Task.Priority)
	}
}

// TestDeepCopyIsolation verifies deep copy isolation for maps (Input, Metadata) and slices (Context, Dependencies).
func TestDeepCopyIsolation(t *testing.T) {
	// 1. Map (Input) deep copy isolation when using WithInput
	inputMap := map[string]any{"key1": "val1"}
	b1 := NewTaskBuilder("task-1", "type-1").WithInput(inputMap)

	// Mutate original map passed to WithInput
	inputMap["key1"] = "mutated-val1"
	inputMap["key2"] = "val2"

	t1, err := b1.Build()
	if err != nil {
		t.Fatalf("failed to build t1: %v", err)
	}
	if val, ok := t1.Input["key1"]; !ok || val != "val1" {
		t.Errorf("expected t1.Input['key1'] to be 'val1', got '%v'", val)
	}
	if _, ok := t1.Input["key2"]; ok {
		t.Error("expected t1.Input['key2'] to be absent due to isolation")
	}

	// 2. Map (Input) deep copy isolation during builder derivation
	b2 := b1.AddInput("key3", "val3")
	t2, err := b2.Build()
	if err != nil {
		t.Fatalf("failed to build t2: %v", err)
	}
	if _, ok := t1.Input["key3"]; ok {
		t.Error("expected t1.Input['key3'] to be absent in parent task")
	}
	if val, ok := t2.Input["key3"]; !ok || val != "val3" {
		t.Errorf("expected t2.Input['key3'] to be 'val3', got '%v'", val)
	}

	// 3. Map (Metadata) deep copy isolation
	b3 := b1.AddMetadata("meta1", "metaval1")
	b4 := b3.AddMetadata("meta2", "metaval2")

	t3, _ := b3.Build()
	t4, _ := b4.Build()
	if _, ok := t3.Metadata["meta2"]; ok {
		t.Error("expected t3.Metadata['meta2'] to be absent in parent task")
	}
	if val, ok := t4.Metadata["meta2"]; !ok || val != "metaval2" {
		t.Errorf("expected t4.Metadata['meta2'] to be 'metaval2', got '%v'", val)
	}

	// 4. Slice (Context) deep copy isolation
	b5 := b1.AddFileContext("path/to/file.txt", "file content")
	b6 := b5.AddContextItem(agent.ContextItem{Type: "custom", Content: "custom content"})

	t5, _ := b5.Build()
	t6, _ := b6.Build()

	if len(t5.Context) != 1 {
		t.Errorf("expected t5 context size 1, got %d", len(t5.Context))
	}
	if len(t6.Context) != 2 {
		t.Errorf("expected t6 context size 2, got %d", len(t6.Context))
	}
	if t5.Context[0].Source != "path/to/file.txt" {
		t.Errorf("expected t5.Context[0].Source to be 'path/to/file.txt', got '%s'", t5.Context[0].Source)
	}

	// 5. Slice (Dependencies) deep copy isolation
	depID1 := contracts.TaskID("dep-1")
	depID2 := contracts.TaskID("dep-2")
	b7 := b1.AddDependency(depID1)
	b8 := b7.AddDependency(depID2)

	t7, _ := b7.Build()
	t8, _ := b8.Build()

	if len(t7.Dependencies) != 1 {
		t.Errorf("expected t7 dependency size 1, got %d", len(t7.Dependencies))
	}
	if len(t8.Dependencies) != 2 {
		t.Errorf("expected t8 dependency size 2, got %d", len(t8.Dependencies))
	}
}

// TestValidationOnBuild verifies validation checks trigger on Build().
func TestValidationOnBuild(t *testing.T) {
	// Case 1: Empty Name
	bErr1 := NewTaskBuilder("", "type-1")
	_, err := bErr1.Build()
	if err == nil {
		t.Error("expected error building task with empty name, got nil")
	}

	// Case 2: Empty Type
	bErr2 := NewTaskBuilder("name-1", "")
	_, err = bErr2.Build()
	if err == nil {
		t.Error("expected error building task with empty type, got nil")
	}

	// Case 3: Negative Timeout
	bErr3 := NewTaskBuilder("name-1", "type-1").WithTimeout(-1 * time.Second)
	_, err = bErr3.Build()
	if err == nil {
		t.Error("expected error building task with negative timeout, got nil")
	}

	// Case 4: Self-dependency
	bErr4 := NewTaskBuilder("name-1", "type-1")
	taskWithID, err := bErr4.Build()
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}
	// Add dependency pointing to its own ID
	bErr4WithSelfDep := bErr4.AddDependency(taskWithID.ID)
	_, err = bErr4WithSelfDep.Build()
	if err == nil {
		t.Error("expected error building task with self-dependency, got nil")
	}

	// Case 5: Duplicate dependency
	bErr5 := NewTaskBuilder("name-1", "type-1").
		AddDependency(contracts.TaskID("dep-a")).
		AddDependency(contracts.TaskID("dep-a"))
	_, err = bErr5.Build()
	if err == nil {
		t.Error("expected error building task with duplicate dependency, got nil")
	}
}
