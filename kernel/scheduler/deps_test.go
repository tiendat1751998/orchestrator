package scheduler

import (
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
)

func TestDependencyTracker_Basic(t *testing.T) {
	dt := NewDependencyTracker()

	taskA := contracts.TaskID("tsk-A")
	taskB := contracts.TaskID("tsk-B")

	// Task with no dependencies should be ready
	if !dt.IsReady(taskA) {
		t.Errorf("expected taskA to be ready initially")
	}

	// Add dependency: A depends on B
	err := dt.AddDependency(taskA, taskB)
	if err != nil {
		t.Fatalf("unexpected error adding dependency: %v", err)
	}

	// Task A should not be ready since B is not completed
	if dt.IsReady(taskA) {
		t.Errorf("expected taskA not to be ready")
	}

	pending := dt.PendingDependencies(taskA)
	if len(pending) != 1 || pending[0] != taskB {
		t.Errorf("expected pending dependencies to contain taskB, got %v", pending)
	}

	// Mark B as completed
	dt.MarkCompleted(taskB)

	// Task A should now be ready
	if !dt.IsReady(taskA) {
		t.Errorf("expected taskA to be ready after B completed")
	}

	if len(dt.PendingDependencies(taskA)) != 0 {
		t.Errorf("expected no pending dependencies, got %v", dt.PendingDependencies(taskA))
	}
}

func TestDependencyTracker_SelfDependency(t *testing.T) {
	dt := NewDependencyTracker()
	taskA := contracts.TaskID("tsk-A")

	err := dt.AddDependency(taskA, taskA)
	if err == nil {
		t.Errorf("expected error when adding self dependency, got nil")
	}
}

func TestDependencyTracker_Cycles(t *testing.T) {
	dt := NewDependencyTracker()

	taskA := contracts.TaskID("tsk-A")
	taskB := contracts.TaskID("tsk-B")
	taskC := contracts.TaskID("tsk-C")

	// A -> B
	if err := dt.AddDependency(taskA, taskB); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// B -> C
	if err := dt.AddDependency(taskB, taskC); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// C -> A should fail due to cycle
	if err := dt.AddDependency(taskC, taskA); err == nil {
		t.Errorf("expected circular dependency error, got nil")
	}

	// Verify that C -> A was rolled back and is not in dependencies
	dt.mu.Lock()
	if dt.dependencies[taskC] != nil && dt.dependencies[taskC][taskA] {
		dt.mu.Unlock()
		t.Errorf("expected dependency from C to A to be rolled back")
	}
	dt.mu.Unlock()
}

func TestDependencyTracker_Diamond(t *testing.T) {
	dt := NewDependencyTracker()

	taskA := contracts.TaskID("tsk-A")
	taskB := contracts.TaskID("tsk-B")
	taskC := contracts.TaskID("tsk-C")
	taskD := contracts.TaskID("tsk-D")

	// A depends on B and C
	// B depends on D
	// C depends on D
	// This diamond structure should not cause cycle detection errors or infinite loops.

	if err := dt.AddDependency(taskA, taskB); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := dt.AddDependency(taskA, taskC); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := dt.AddDependency(taskB, taskD); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := dt.AddDependency(taskC, taskD); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's not detected as cycle
	if !dt.IsReady(taskD) {
		t.Errorf("expected D to be ready")
	}

	if dt.IsReady(taskA) {
		t.Errorf("expected A not to be ready")
	}
}

func TestDependencyTracker_Reset(t *testing.T) {
	dt := NewDependencyTracker()

	taskA := contracts.TaskID("tsk-A")
	taskB := contracts.TaskID("tsk-B")

	_ = dt.AddDependency(taskA, taskB)
	dt.MarkCompleted(taskB)

	dt.Reset()

	// After reset, A should be ready (no dependencies) and B should not be completed (though A has no deps anyway)
	if !dt.IsReady(taskA) {
		t.Errorf("expected A to be ready after reset since dependencies should be cleared")
	}

	dt.mu.Lock()
	if len(dt.dependencies) != 0 || len(dt.completed) != 0 {
		t.Errorf("expected empty maps after Reset, got deps: %v, completed: %v", dt.dependencies, dt.completed)
	}
	dt.mu.Unlock()
}
