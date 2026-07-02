# Micro-Task 3.25: Create sdk/workflow/workflow_test.go

## Info
- **File**: `sdk/workflow/workflow_test.go`
- **Package**: `workflow_test`
- **Depends on**: 3.13 (workflow.go), 3.24 (state.go)
- **Time**: 15 min
- **Verify**: `go test -v -race -count=1 ./sdk/workflow/...`

## Purpose
Triển khai bộ kiểm kiểm thử tự động (Unit Tests) cho gói `sdk/workflow`. Gói này kiểm nghiệm tính đúng đắn của thuật toán sắp xếp Topo (Topological Sort), cơ chế ngắt khi đồ thị tuần hoàn (Circular Dependency), và bộ phân giải tham số truyền tin động (`ResolveValue` của `WorkflowState`).

## EXACT code to create

```go
package workflow_test

import (
	"context"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
	contractsworkflow "github.com/tiendat1751998/orchestrator/contracts/workflow"
	sdkworkflow "github.com/tiendat1751998/orchestrator/sdk/workflow"
)

// =============================================================================
// Topological Sort Tests
// =============================================================================

func TestSortSteps_Success(t *testing.T) {
	steps := []contractsworkflow.Step{
		{Name: "deploy", DependsOn: []string{"test"}},
		{Name: "build", DependsOn: []string{}},
		{Name: "test", DependsOn: []string{"build"}},
	}

	sorted, err := sdkworkflow.SortSteps(steps)
	if err != nil {
		t.Fatalf("unexpected error sorting steps: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("expected 3 sorted steps, got %d", len(sorted))
	}

	// Verify order: build -> test -> deploy
	if sorted[0].Name != "build" {
		t.Errorf("expected first step to be 'build', got %q", sorted[0].Name)
	}
	if sorted[1].Name != "test" {
		t.Errorf("expected second step to be 'test', got %q", sorted[1].Name)
	}
	if sorted[2].Name != "deploy" {
		t.Errorf("expected third step to be 'deploy', got %q", sorted[2].Name)
	}
}

func TestSortSteps_CircularDependency(t *testing.T) {
	steps := []contractsworkflow.Step{
		{Name: "stepA", DependsOn: []string{"stepB"}},
		{Name: "stepB", DependsOn: []string{"stepA"}},
	}

	_, err := sdkworkflow.SortSteps(steps)
	if err == nil {
		t.Fatal("expected error sorting circular steps, got nil")
	}
}

func TestNewBaseWorkflow_DuplicateNames(t *testing.T) {
	steps := []contractsworkflow.Step{
		{Name: "stepA"},
		{Name: "stepA"},
	}

	_, err := sdkworkflow.NewBaseWorkflow("test-flow", steps)
	if err == nil {
		t.Fatal("expected error creating base workflow with duplicate step names, got nil")
	}
}

// =============================================================================
// Workflow State Parameter Resolution Tests
// =============================================================================

func TestWorkflowState_ResolveValue(t *testing.T) {
	inputs := map[string]any{
		"project_name": "orchestrator",
		"version":      "1.0.0",
		"debug_mode":   true,
	}

	state := sdkworkflow.NewState(inputs)

	// Save a mock step output
	stepResult := &contractsworkflow.StepResult{
		Status: contracts.StatusCompleted,
		Output: map[string]any{
			"binary_path": "/bin/app",
			"details": map[string]any{
				"size_bytes": 1048576,
			},
		},
		Error: "",
	}
	state.SetStepResult("compile_step", stepResult)

	// Case 1: Simple input resolution
	val, err := state.ResolveValue("{{ inputs.project_name }}")
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if val != "orchestrator" {
		t.Errorf("expected 'orchestrator', got %v", val)
	}

	// Case 2: Step output status resolution
	val, err = state.ResolveValue("{{ steps.compile_step.status }}")
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if val != "completed" {
		t.Errorf("expected 'completed', got %v", val)
	}

	// Case 3: Step output nested field resolution
	val, err = state.ResolveValue("{{ steps.compile_step.output.details.size_bytes }}")
	if err != nil {
		t.Fatalf("ResolveValue: %v", err)
	}
	if val != 1048576 {
		t.Errorf("expected 1048576, got %v", val)
	}

	// Case 4: Recursive map resolution
	inputMap := map[string]any{
		"name": "{{ inputs.project_name }}",
		"meta": map[string]any{
			"size": "{{ steps.compile_step.output.details.size_bytes }}",
		},
	}
	resolvedMap, err := state.ResolveValue(inputMap)
	if err != nil {
		t.Fatalf("ResolveValue map: %v", err)
	}

	rMap := resolvedMap.(map[string]any)
	if rMap["name"] != "orchestrator" {
		t.Errorf("expected name to be 'orchestrator', got %v", rMap["name"])
	}
	rMeta := rMap["meta"].(map[string]any)
	if rMeta["size"] != 1048576 {
		t.Errorf("expected size to be 1048576, got %v", rMeta["size"])
	}
}

func TestWorkflowState_ResolveValue_Errors(t *testing.T) {
	state := sdkworkflow.NewState(nil)

	// Missing input key
	_, err := state.ResolveValue("{{ inputs.non_existent }}")
	if err == nil {
		t.Error("expected error for missing input key")
	}

	// Missing step result
	_, err = state.ResolveValue("{{ steps.non_existent.output }}")
	if err == nil {
		t.Error("expected error for missing step result")
	}

	// Invalid template syntax
	_, err = state.ResolveValue("{{ invalid }}")
	if err == nil {
		t.Error("expected error for invalid expression syntax")
	}
}
```

## Verify
```bash
go test -v -race -count=1 ./sdk/workflow/...
```

## Checklist
- [ ] File `sdk/workflow/workflow_test.go` tồn tại
- [ ] Package name: `workflow_test`
- [ ] Test `TestSortSteps_Success` xác thực sắp xếp topological DAG đúng thứ tự
- [ ] Test `TestSortSteps_CircularDependency` phát hiện vòng lặp thành công
- [ ] Test `TestWorkflowState_ResolveValue` phân giải thành công các ports inputs và steps output
- [ ] Phân giải thành công đệ quy sâu với maps lồng nhau
- [ ] Bắt lỗi chính xác khi biểu thức trỏ tới khóa không tồn tại
- [ ] `go test -v -race ./sdk/workflow/...` trả về ALL PASS
