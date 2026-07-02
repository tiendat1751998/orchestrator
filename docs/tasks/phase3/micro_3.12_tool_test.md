# Micro-Task 3.12: Create sdk/tool/tool_test.go

## Info
- **File**: `sdk/tool/tool_test.go`
- **Package**: `tool_test`
- **Depends on**: 3.10 (tool.go), 3.11 (result.go)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/tool/...`

## Purpose
Triển khai bộ kiểm thử tự động (Unit Tests) cho Tool SDK. Kiểm tra tính chính xác của trình kiểm định tham số JSON Schema (`ValidateArguments`) với các biên kiểu dữ liệu (integer, boolean, enum) và định dạng dữ liệu lỗi, cùng các helper khởi tạo kết quả (`Success`, `Failure`, `JSON`).

## EXACT code to create

```go
package tool_test

import (
	"context"
	"encoding/json"
	"errors"
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
	// Success
	res := sdktool.Success("ok")
	if res.ExitCode != 0 || res.Output != "ok" || res.Error != "" {
		t.Errorf("Success: %v", res)
	}

	// Failure
	res = sdktool.Failure("error message")
	if res.ExitCode != 1 || res.Error != "error message" || res.Output != "" {
		t.Errorf("Failure: %v", res)
	}

	// WithExitCode
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

	// Nil check
	res, _ = sdktool.JSON(nil)
	if res.Output != "{}" {
		t.Errorf("Nil JSON: got %q", res.Output)
	}
}

// Unmarshallable type to trigger marshal error
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

## Verify
```bash
go test -v -race -count=1 ./sdk/tool/...
```

## Checklist
- [ ] File `sdk/tool/tool_test.go` tồn tại
- [ ] Package name: `tool_test`
- [ ] Test `TestBaseTool_ValidateArguments_Required` bắt lỗi thiếu tham số bắt buộc
- [ ] Test `TestBaseTool_ValidateArguments_Types` kiểm tra ép kiểu boolean, float64 integer chính xác
- [ ] Test `TestBaseTool_ValidateArguments_Enum` chặn giá trị ngoài enum
- [ ] Test `TestBaseTool_ValidateArguments_InvalidJSON` chặn chuỗi JSON lỗi
- [ ] Test `TestResultBuilders` kiểm tra code của các helpers khởi tạo
- [ ] Test `TestResultJSON` kiểm tra tuần tự hóa dữ liệu ra JSON string
- [ ] Test `TestResultJSON_Error` bắt lỗi serialization thành công
- [ ] `go test -v -race ./sdk/tool/...` trả về ALL PASS
