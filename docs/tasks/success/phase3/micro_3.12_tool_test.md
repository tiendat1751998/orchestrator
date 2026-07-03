# Micro-Task 3.12: Create sdk/tool/tool_test.go

## Info
- **File**: `sdk/tool/tool_test.go`
- **Package**: `tool_test`
- **Depends on**: 3.10 (tool.go), 3.11 (result.go)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/tool/...`

## Purpose
Implements integration unit tests for the Tool SDK, verifying the accuracy of the JSON Schema arguments validator, testing type boundaries, enum restrictions, and result builder helper mappings.

## EXACT code to create

```go
package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// =============================================================================
// Helper Mock Tool for testing execution
// =============================================================================

type concreteTestTool struct {
	*sdktool.BaseTool
}

func (t *concreteTestTool) Execute(ctx context.Context, args json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(args); err != nil {
		return sdktool.Failure(err.Error()), nil
	}
	return sdktool.Success("test success"), nil
}

// =============================================================================
// Argument Validation Tests
// =============================================================================

func TestBaseTool_ValidateArguments_Required(t *testing.T) {
	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.StringProperty("File path")).
		AddProperty("lines", contractstool.IntProperty("Lines count")).
		AddRequired("path")

	toolInst, _ := sdktool.NewBaseTool("test_tool", "Description", schema)

	// Case 1: Missing required field
	err := toolInst.ValidateArguments(json.RawMessage(`{"lines": 10}`))
	if err == nil {
		t.Error("expected error due to missing required field 'path'")
	}

	// Case 2: Present required field
	err = toolInst.ValidateArguments(json.RawMessage(`{"path": "main.go", "lines": 5}`))
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Case 3: Empty arguments when fields are required
	err = toolInst.ValidateArguments(json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error on empty args when fields are required")
	}
}

func TestBaseTool_ValidateArguments_Types(t *testing.T) {
	schema := contractstool.NewSchema().
		AddProperty("str_val", contractstool.StringProperty("string")).
		AddProperty("int_val", contractstool.IntProperty("integer")).
		AddProperty("bool_val", contractstool.BoolProperty("boolean"))

	toolInst, _ := sdktool.NewBaseTool("test_tool", "Description", schema)

	// Case 1: Type mismatch (string as boolean)
	err := toolInst.ValidateArguments(json.RawMessage(`{"bool_val": "true"}`))
	if err == nil {
		t.Error("expected error when passing string as boolean")
	}

	// Case 2: Type mismatch (float as integer)
	err = toolInst.ValidateArguments(json.RawMessage(`{"int_val": 12.34}`))
	if err == nil {
		t.Error("expected error when passing decimal float as integer")
	}

	// Case 3: Correct integer (whole number float)
	err = toolInst.ValidateArguments(json.RawMessage(`{"int_val": 42}`))
	if err != nil {
		t.Errorf("unexpected error on valid integer: %v", err)
	}
}

func TestBaseTool_ValidateArguments_Enum(t *testing.T) {
	schema := contractstool.NewSchema().
		AddProperty("format", contractstool.EnumProperty("Output format", "json", "yaml"))

	toolInst, _ := sdktool.NewBaseTool("test_tool", "Description", schema)

	// Case 1: Value inside enum
	err := toolInst.ValidateArguments(json.RawMessage(`{"format": "yaml"}`))
	if err != nil {
		t.Errorf("unexpected error on valid enum value: %v", err)
	}

	// Case 2: Value outside enum
	err = toolInst.ValidateArguments(json.RawMessage(`{"format": "toml"}`))
	if err == nil {
		t.Error("expected error for value outside enum")
	}
}

func TestBaseTool_ValidateArguments_InvalidJSON(t *testing.T) {
	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.StringProperty("path"))

	toolInst, _ := sdktool.NewBaseTool("test_tool", "Description", schema)

	err := toolInst.ValidateArguments(json.RawMessage(`{"path": "unclosed`))
	if err == nil {
		t.Error("expected error on malformed JSON input")
	}
}

// =============================================================================
// Result Builder Tests
// =============================================================================

func TestResultBuilders(t *testing.T) {
	res := sdktool.Success("ok")
	if res.ExitCode != 0 || res.Output != "ok" || res.Error != "" {
		t.Errorf("Success: %v", res)
	}

	res = sdktool.Failure("error message")
	if res.ExitCode != 1 || res.Error != "error message" || res.Output != "" {
		t.Errorf("Failure: %v", res)
	}

	res = sdktool.WithExitCode("output", "err", 2)
	if res.ExitCode != 2 || res.Output != "output" || res.Error != "err" {
		t.Errorf("WithExitCode: %v", res)
	}
}

func TestResultJSON(t *testing.T) {
	type dummy struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := dummy{Name: "test", Value: 42}
	res, err := sdktool.JSON(data)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	expectedJSON := `{"name":"test","value":42}`
	if res.Output != expectedJSON {
		t.Errorf("JSON output: got %q, want %q", res.Output, expectedJSON)
	}

	res, _ = sdktool.JSON(nil)
	if res.Output != "{}" {
		t.Errorf("Nil JSON: got %q", res.Output)
	}
}

type unmarshallable struct {
	Ch chan int
}

func TestResultJSON_Error(t *testing.T) {
	data := unmarshallable{Ch: make(chan int)}
	_, err := sdktool.JSON(data)
	if err == nil {
		t.Error("expected JSON serialization error for unmarshallable channel field")
	}
}
```

## Rules
1. **Whole Number Float Checks**: Test integer validations using whole float values (e.g. `42` or `42.00`) to confirm the checker accepts standard unmarshalled floats.
2. **Error Assertions**: Assert that JSON marshalling errors are caught and bubbled up.
3. **Empty Input Guards**: Ensure validations fail if empty payloads `{}` are supplied to schemas containing required fields.

## ⚠️ Pitfalls

### Pitfall 1: Expecting integers to decode as Go `int` types in tests
```go
```
Go decodes all JSON numbers to `float64` inside maps. Make sure test assertions take this into account.

### Pitfall 2: Bypassing validations on unclosed JSON inputs
Malformed JSON (like `{"path": "unclosed`) will cause unmarshalling failures, so ensure `ValidateArguments` handles parsing errors gracefully.

## Verify
```bash
go test -v -race -count=1 ./sdk/tool/...
```

## Checklist
- [ ] File `sdk/tool/tool_test.go` exists
- [ ] Package: `tool_test` (external testing package)
- [ ] Required parameters checks verify presence of declared keys
- [ ] Decimal float values fail integer validations
- [ ] Whole float values pass integer validations
- [ ] Values outside of enum lists are rejected
- [ ] Malformed JSON input strings return unmarshalling errors
- [ ] `Success`, `Failure`, and `WithExitCode` helpers populate fields correctly
- [ ] JSON serializer converts structures to JSON string outputs
- [ ] JSON serializer returns errors for unmarshallable types like channels
- [ ] `go test -v -race ./sdk/tool/...` passes
