# Micro-Task 1.22: Create contracts/agent/agent_test.go

## Info
- **File**: `contracts/agent/agent_test.go`
- **Package**: `agent_test`
- **Depends on**: 1.17, 1.18, 1.19, 1.20, 1.21
- **Time**: 20 min
- **Verify**: `go test -v -race ./contracts/agent/...`

## Purpose
Implements unit tests verifying the correctness of JSON serialization, builder helpers, validation properties, and capability matchers in the Agent contracts package.

## EXACT code to create

```go
package agent_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// =============================================================================
// Capability Tests
// =============================================================================

func TestCapability_IsValid(t *testing.T) {
	tests := []struct {
		cap  agent.Capability
		want bool
	}{
		{agent.CapabilityCodeGeneration, true},
		{agent.CapabilityCodeReview, true},
		{agent.CapabilityTesting, true},
		{agent.Capability("invalid"), false},
		{agent.Capability(""), false},
	}
	for _, tt := range tests {
		if got := tt.cap.IsValid(); got != tt.want {
			t.Errorf("Capability(%q).IsValid() = %v, want %v", tt.cap, got, tt.want)
		}
	}
}

func TestHasCapability(t *testing.T) {
	caps := []agent.Capability{agent.CapabilityCodeGeneration, agent.CapabilityTesting}

	if !agent.HasCapability(caps, agent.CapabilityTesting) {
		t.Error("expected HasCapability to find CapabilityTesting")
	}
	if agent.HasCapability(caps, agent.CapabilityDeployment) {
		t.Error("expected HasCapability to NOT find CapabilityDeployment")
	}
}

// =============================================================================
// Task Tests
// =============================================================================

func TestNewTask_Defaults(t *testing.T) {
	task := agent.NewTask("test_task", "do something", "code_generation")

	if task.ID.IsEmpty() {
		t.Error("ID should be generated")
	}
	if task.Name != "test_task" {
		t.Errorf("Name: got %q, want %q", task.Name, "test_task")
	}
	if task.Priority != 5 {
		t.Errorf("Priority: got %d, want 5", task.Priority)
	}
	if task.Timeout != 5*time.Minute {
		t.Errorf("Timeout: got %v, want 5m", task.Timeout)
	}
}

func TestTask_AddDependency(t *testing.T) {
	task := agent.NewTask("t1", "desc", "code_generation")
	depID := contracts.NewTaskID()

	result := task.AddDependency(depID)

	if result != task {
		t.Error("AddDependency should return self")
	}
	if len(task.Dependencies) != 1 {
		t.Fatalf("Dependencies: got %d, want 1", len(task.Dependencies))
	}
	if task.Dependencies[0] != depID {
		t.Error("Dependency ID mismatch")
	}
}

func TestTask_AddContext(t *testing.T) {
	task := agent.NewTask("t1", "desc", "code_generation")
	task.AddContext("file", "package main", "/src/main.go")

	if len(task.Context) != 1 {
		t.Fatalf("Context: got %d, want 1", len(task.Context))
	}
	if task.Context[0].Type != "file" {
		t.Errorf("Context Type: got %q, want %q", task.Context[0].Type, "file")
	}
	if task.Context[0].Source != "/src/main.go" {
		t.Errorf("Context Source: got %q", task.Context[0].Source)
	}
}

func TestTask_HasDependencies(t *testing.T) {
	task := agent.NewTask("t1", "desc", "code_generation")
	if task.HasDependencies() {
		t.Error("new task should not have dependencies")
	}

	task.AddDependency(contracts.NewTaskID())
	if !task.HasDependencies() {
		t.Error("task with dependency should return true")
	}
}

func TestTask_JSONRoundTrip(t *testing.T) {
	task := agent.NewTask("test", "description", "testing")
	task.AddDependency(contracts.NewTaskID())
	task.Input = map[string]any{"language": "go"}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded agent.Task
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Name != task.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, task.Name)
	}
	if len(decoded.Dependencies) != 1 {
		t.Error("Dependencies not preserved")
	}
}

// =============================================================================
// Result Tests
// =============================================================================

func TestResult_IsSuccess(t *testing.T) {
	r := &agent.Result{Status: contracts.StatusSuccess}
	if !r.IsSuccess() {
		t.Error("expected IsSuccess() = true")
	}
}

func TestResult_IsFailed(t *testing.T) {
	r := &agent.Result{Status: contracts.StatusFailed}
	if !r.IsFailed() {
		t.Error("expected IsFailed() = true")
	}
}

func TestSuccessResult(t *testing.T) {
	taskID := contracts.NewTaskID()
	r := agent.SuccessResult(taskID, "backend", "code generated")

	if r.TaskID != taskID {
		t.Error("TaskID mismatch")
	}
	if r.AgentName != "backend" {
		t.Error("AgentName mismatch")
	}
	if !r.IsSuccess() {
		t.Error("should be success")
	}
	if r.Output != "code generated" {
		t.Error("Output mismatch")
	}
}

func TestFailedResult(t *testing.T) {
	taskID := contracts.NewTaskID()
	r := agent.FailedResult(taskID, "backend", "timeout")

	if !r.IsFailed() {
		t.Error("should be failed")
	}
	if r.Error != "timeout" {
		t.Error("Error mismatch")
	}
}

func TestResult_JSONRoundTrip(t *testing.T) {
	r := agent.SuccessResult(contracts.NewTaskID(), "backend", "done")
	r.Artifacts = []agent.Artifact{
		{Name: "main.go", Type: "file", Path: "/src/main.go"},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded agent.Result
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(decoded.Artifacts) != 1 {
		t.Error("Artifacts not preserved")
	}
	if decoded.Artifacts[0].Name != "main.go" {
		t.Error("Artifact name mismatch")
	}
}
```

## ⚠️ Pitfalls

### Pitfall 1: Incorrect package import paths in unit tests
```go
import "github.com/tiendat1751998/orchestrator/contracts/agent" // Fully qualified path.
```
Always use the canonical module prefix when referencing package boundaries in test targets.

### Pitfall 2: Neglecting slice boundary checks in round-trip assertions
Verify that array limits (`len(decoded.Dependencies) != 1`) match initial parameters exactly before checking specific index data to avoid index out of range runtime panics.

## Verify
```bash
go test -v -race ./contracts/agent/...
```

## Checklist
- [ ] File `contracts/agent/agent_test.go` exists
- [ ] Package: `agent_test`
- [ ] `TestCapability_IsValid` tests valid capabilities checks
- [ ] `TestNewTask_Defaults` verifies timeout initialization default properties
- [ ] `TestTask_AddDependency` verifies builder method chains return self pointers
- [ ] `TestResult_JSONRoundTrip` verifies artifact values unmarshal correctly
- [ ] `go test -v -race ./contracts/agent/...` passes
