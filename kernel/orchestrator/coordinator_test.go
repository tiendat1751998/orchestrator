package orchestrator

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
)

func TestCoordinator_InjectDependencyResults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	c := NewCoordinator(logger)

	t.Run("nil task error", func(t *testing.T) {
		err := c.InjectDependencyResults(nil, nil)
		if err == nil {
			t.Error("expected error for nil task, got nil")
		}
	})

	t.Run("no dependencies does nothing", func(t *testing.T) {
		task := &contractsagent.Task{
			ID:    contracts.TaskID("tsk-1"),
			Input: map[string]any{"existing": "value"},
		}

		err := c.InjectDependencyResults(task, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, exists := task.Input["_dependency_results"]; exists {
			t.Error("expected _dependency_results to not be injected for task with no dependencies")
		}
		if val, ok := task.Input["existing"]; !ok || val != "value" {
			t.Errorf("expected existing input parameter to be preserved, got %v", val)
		}
	})

	t.Run("injects dependency results and preserves existing parameters", func(t *testing.T) {
		task := &contractsagent.Task{
			ID:           contracts.TaskID("tsk-2"),
			Dependencies: []contracts.TaskID{"dep-1", "dep-2"},
			Input:        map[string]any{"existing": "value"},
		}

		results := map[string]*contractsagent.Result{
			"dep-1": {
				TaskID: "dep-1",
				Status: contracts.StatusSuccess,
				Output: "result 1 output",
			},
			// dep-2 is not present in results (missing result)
		}

		err := c.InjectDependencyResults(task, results)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check preservation
		if val, ok := task.Input["existing"]; !ok || val != "value" {
			t.Errorf("expected existing parameter to be preserved, got %v", val)
		}

		// Check dependency results injection
		depDataVal, ok := task.Input["_dependency_results"]
		if !ok {
			t.Fatal("expected _dependency_results to be injected")
		}

		depData, ok := depDataVal.(map[string]any)
		if !ok {
			t.Fatalf("expected _dependency_results to be map[string]any, got %T", depDataVal)
		}

		// dep-1 should be injected
		dep1Val, ok := depData["dep-1"]
		if !ok {
			t.Fatal("expected dep-1 to be present in _dependency_results")
		}

		dep1Map, ok := dep1Val.(map[string]any)
		if !ok {
			t.Fatalf("expected dep-1 value to be map[string]any, got %T", dep1Val)
		}

		if dep1Map["status"] != contracts.StatusSuccess {
			t.Errorf("expected status to be %s, got %v", contracts.StatusSuccess, dep1Map["status"])
		}
		if dep1Map["output"] != "result 1 output" {
			t.Errorf("expected output to be 'result 1 output', got %v", dep1Map["output"])
		}
		if dep1Map["error"] != "" {
			t.Errorf("expected error to be empty, got %v", dep1Map["error"])
		}

		// dep-2 should not be injected since its result was missing
		if _, ok := depData["dep-2"]; ok {
			t.Error("expected dep-2 to not be present in _dependency_results")
		}
	})
}
