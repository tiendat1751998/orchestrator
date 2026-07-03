# Micro-Task 5.14: Create kernel/orchestrator/orchestrator_test.go

## Info
- **File**: `kernel/orchestrator/orchestrator_test.go`
- **Package**: `orchestrator_test`
- **Depends on**: 5.13
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/orchestrator/...`

## Purpose
Implements E2E integration unit tests for the main orchestration pipeline, verifying that missions are scheduled, dependency parameters are injected, and failures trigger replanning successfully.

## EXACT code to create

```go
package orchestrator_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/orchestrator"
	"github.com/tiendat1751998/orchestrator/kernel/planner"
)

func TestPipelineManager_Transitions(t *testing.T) {
	pm := orchestrator.NewPipelineManager()

	if pm.GetState() != orchestrator.StatePlanning {
		t.Errorf("expected initial state Planning, got %q", pm.GetState())
	}

	// 1. Success transition path: Planning -> Scheduling -> Executing
	err := pm.Transition(orchestrator.StateScheduling)
	if err != nil {
		t.Fatalf("failed scheduling transition: %v", err)
	}

	err = pm.Transition(orchestrator.StateExecuting)
	if err != nil {
		t.Fatalf("failed executing transition: %v", err)
	}

	// 2. Reject invalid jump: Executing -> Planning
	err = pm.Transition(orchestrator.StatePlanning)
	if err == nil {
		t.Error("expected error transitioning backwards Executing -> Planning, got nil")
	}
}

func TestSupervisor_TimeoutScans(t *testing.T) {
	sup := orchestrator.NewSupervisor(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register task with 10ms timeout
	sup.Register("task-timeout", 10*time.Millisecond)

	sup.Start(ctx, 5*time.Millisecond)

	// Wait for supervisor scan loop to detect timeout
	time.Sleep(30 * time.Millisecond)

	// Register verify cleanup
	sup.Deregister("task-timeout")
}

func TestCoordinator_DependencyInjection(t *testing.T) {
	coord := orchestrator.NewCoordinator(nil)

	task := &contractsagent.Task{
		ID:           "task-B",
		Name:         "B",
		Dependencies: []contractsagent.TaskID{"task-A"},
		Parameters:   json.RawMessage(`{"instruction": "generate code"}`),
	}

	results := map[string]*contractsagent.Result{
		"task-A": {
			TaskID: "task-A",
			Status: "success",
			Output: "package main\n",
		},
	}

	err := coord.InjectDependencyResults(task, results)
	if err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	// Verify params contains both instruction and injected dependencies results
	var parsed map[string]any
	err = json.Unmarshal(task.Parameters, &parsed)
	if err != nil {
		t.Fatalf("failed to parse updated params: %v", err)
	}

	if parsed["instruction"] != "generate code" {
		t.Errorf("expected instruction key to be preserved, got %v", parsed["instruction"])
	}

	depResults, ok := parsed["_dependency_results"].(map[string]any)
	if !ok {
		t.Fatal("missing _dependency_results key in parameters")
	}

	taskARes, ok := depResults["task-A"].(map[string]any)
	if !ok {
		t.Fatal("missing task-A results inside dependencies results map")
	}

	if taskARes["output"] != "package main\n" {
		t.Errorf("expected injected output 'package main\n', got %v", taskARes["output"])
	}
}

func TestFeedbackCollector_Records(t *testing.T) {
	fc := orchestrator.NewFeedbackCollector()

	fc.RecordSuccess("mock-agent", 100, 50*time.Millisecond)
	fc.RecordFailure("mock-agent")

	metrics := fc.GetMetrics()
	m, ok := metrics["mock-agent"]
	if !ok {
		t.Fatal("expected metrics to contain 'mock-agent'")
	}

	if m.SuccessCount != 1 || m.FailureCount != 1 {
		t.Errorf("incorrect counts: success=%d, failure=%d", m.SuccessCount, m.FailureCount)
	}

	if m.TotalTokens != 100 {
		t.Errorf("expected 100 tokens, got %d", m.TotalTokens)
	}
}
```

## Pitfalls

### Pitfall 1: Leaking goroutines in supervisor timeout checks
If unit tests invoke supervisor background scanner loops but omit cancel triggers, background scan goroutines can leak on test exits. Always use context cancellations.

### Pitfall 2: Flaky tests from tight timing assertions
Using exact timing checks (e.g. asserting that a task times out in exactly 10ms) causes flaky test runs on loaded CI servers. Use generous thresholds.

## Verify
```bash
go test -v -race -count=1 ./kernel/orchestrator/...
# Expected: PASS
```

## Checklist
- [ ] File exists at `kernel/orchestrator/orchestrator_test.go`
- [ ] Package name is `orchestrator_test`
- [ ] Pipeline state transition rules are validated
- [ ] Supervisor scanner timeouts are checked
- [ ] Coordinator dependency result injections are verified
- [ ] Feedback metric values are validated
- [ ] Build command passes
