package workflow

import (
	"sync"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/workflow"
)

func TestStateResolveValue(t *testing.T) {
	inputs := map[string]any{
		"env":     "prod",
		"version": 2,
		"debug":   false,
		"nil_val": nil,
	}

	state := NewState(inputs)

	state.SetStepResult("build", &workflow.StepResult{
		Status: contracts.StatusSuccess,
		Output: map[string]any{
			"artifacts": []any{"app.bin", "checksum.sha256"},
			"metadata": map[string]any{
				"author": "dev",
				"details": map[string]any{
					"commit": "abcdef",
				},
			},
		},
		Error: "",
	})

	state.SetStepResult("test", &workflow.StepResult{
		Status: contracts.StatusFailed,
		Output: "tests failed",
		Error:  "exit status 1",
	})

	t.Run("nil value", func(t *testing.T) {
		res, err := state.ResolveValue(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil, got %v", res)
		}
	})

	t.Run("plain string", func(t *testing.T) {
		res, err := state.ResolveValue("hello world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "hello world" {
			t.Errorf("expected 'hello world', got %v", res)
		}
	})

	t.Run("inputs lookup", func(t *testing.T) {
		res, err := state.ResolveValue("{{ inputs.env }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "prod" {
			t.Errorf("expected 'prod', got %v", res)
		}

		res, err = state.ResolveValue("{{ inputs.version }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != 2 {
			t.Errorf("expected 2, got %v", res)
		}
	})

	t.Run("inputs missing key", func(t *testing.T) {
		_, err := state.ResolveValue("{{ inputs.missing }}")
		if err == nil {
			t.Error("expected error for missing input key, got nil")
		}
	})

	t.Run("steps status and error", func(t *testing.T) {
		res, err := state.ResolveValue("{{ steps.build.status }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != string(contracts.StatusSuccess) {
			t.Errorf("expected %q, got %v", contracts.StatusSuccess, res)
		}

		res, err = state.ResolveValue("{{ steps.test.error }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "exit status 1" {
			t.Errorf("expected 'exit status 1', got %v", res)
		}
	})

	t.Run("steps output basic", func(t *testing.T) {
		res, err := state.ResolveValue("{{ steps.test.output }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "tests failed" {
			t.Errorf("expected 'tests failed', got %v", res)
		}
	})

	t.Run("steps output nested path", func(t *testing.T) {
		res, err := state.ResolveValue("{{ steps.build.output.metadata.details.commit }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "abcdef" {
			t.Errorf("expected 'abcdef', got %v", res)
		}
	})

	t.Run("steps output nested path not found", func(t *testing.T) {
		_, err := state.ResolveValue("{{ steps.build.output.metadata.missing }}")
		if err == nil {
			t.Error("expected error for missing nested property, got nil")
		}
	})

	t.Run("invalid expression format", func(t *testing.T) {
		_, err := state.ResolveValue("{{ inputs }}")
		if err == nil {
			t.Error("expected error for invalid expression structure, got nil")
		}

		_, err = state.ResolveValue("{{ steps.build }}")
		if err == nil {
			t.Error("expected error for incomplete steps expression, got nil")
		}
	})

	t.Run("unsupported source and property", func(t *testing.T) {
		_, err := state.ResolveValue("{{ config.port }}")
		if err == nil {
			t.Error("expected error for unsupported template source, got nil")
		}

		_, err = state.ResolveValue("{{ steps.build.duration }}")
		if err == nil {
			t.Error("expected error for unsupported step property, got nil")
		}
	})

	t.Run("recursive map", func(t *testing.T) {
		inputMap := map[string]any{
			"env":    "{{ inputs.env }}",
			"commit": "{{ steps.build.output.metadata.details.commit }}",
			"static": "value",
		}
		res, err := state.ResolveValue(inputMap)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resMap, ok := res.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", res)
		}
		if resMap["env"] != "prod" {
			t.Errorf("expected 'prod', got %v", resMap["env"])
		}
		if resMap["commit"] != "abcdef" {
			t.Errorf("expected 'abcdef', got %v", resMap["commit"])
		}
		if resMap["static"] != "value" {
			t.Errorf("expected 'value', got %v", resMap["static"])
		}
	})

	t.Run("recursive slice", func(t *testing.T) {
		inputSlice := []any{
			"{{ inputs.env }}",
			"{{ steps.build.output.metadata.details.commit }}",
			"static",
		}
		res, err := state.ResolveValue(inputSlice)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resSlice, ok := res.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", res)
		}
		if len(resSlice) != 3 {
			t.Fatalf("expected slice length 3, got %d", len(resSlice))
		}
		if resSlice[0] != "prod" {
			t.Errorf("expected 'prod', got %v", resSlice[0])
		}
		if resSlice[1] != "abcdef" {
			t.Errorf("expected 'abcdef', got %v", resSlice[1])
		}
		if resSlice[2] != "static" {
			t.Errorf("expected 'static', got %v", resSlice[2])
		}
	})
}

func TestStateConcurrency(t *testing.T) {
	state := NewState(map[string]any{
		"key": "val",
	})

	var wg sync.WaitGroup
	workers := 20
	iterations := 100

	wg.Add(workers * 2)

	// Writer goroutines
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				state.SetStepResult("step", &workflow.StepResult{
					Status: contracts.StatusSuccess,
					Output: j,
				})
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = state.ResolveValue("{{ inputs.key }}")
				_, _ = state.ResolveValue("{{ steps.step.status }}")
				_, _ = state.ResolveValue("{{ steps.step.output }}")
			}
		}()
	}

	wg.Wait()
}
