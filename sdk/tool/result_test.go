package tool_test

import (
	"testing"

	"github.com/tiendat1751998/orchestrator/sdk/tool"
)

func TestSuccess(t *testing.T) {
	res := tool.Success("hello")
	if res.Output != "hello" {
		t.Errorf("expected output 'hello', got %q", res.Output)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if res.Error != "" {
		t.Errorf("expected empty error, got %q", res.Error)
	}
}

func TestFailure(t *testing.T) {
	res := tool.Failure("error occurred")
	if res.Error != "error occurred" {
		t.Errorf("expected error 'error occurred', got %q", res.Error)
	}
	if res.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", res.ExitCode)
	}
	if res.Output != "" {
		t.Errorf("expected empty output, got %q", res.Output)
	}
}

func TestWithExitCode(t *testing.T) {
	res := tool.WithExitCode("out", "err", 42)
	if res.Output != "out" {
		t.Errorf("expected output 'out', got %q", res.Output)
	}
	if res.Error != "err" {
		t.Errorf("expected error 'err', got %q", res.Error)
	}
	if res.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestJSON(t *testing.T) {
	// Case 1: nil input
	res, err := tool.JSON(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "{}" {
		t.Errorf("expected output '{}' for nil, got %q", res.Output)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}

	// Case 2: struct input
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "Alice",
		Age:  30,
	}
	res, err = tool.JSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"name":"Alice","age":30}`
	if res.Output != expected {
		t.Errorf("expected output %q, got %q", expected, res.Output)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}

	// Case 3: error during marshaling (e.g. channel or function)
	_, err = tool.JSON(make(chan int))
	if err == nil {
		t.Error("expected error marshaling channel, got nil")
	}
}
