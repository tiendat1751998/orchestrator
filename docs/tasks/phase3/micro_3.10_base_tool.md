# Micro-Task 3.10: Create sdk/tool/tool.go

## Info
- **File**: `sdk/tool/tool.go`
- **Package**: `tool`
- **Depends on**: 3.01 (base_plugin.md), 1.14 (tool schema contract), 1.15 (tool interface contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/tool/...`

## Purpose
Triển khai cấu trúc nền tảng `BaseTool` và bộ kiểm định tham số tự động (`ValidateArguments`). Lớp helper này giúp loại bỏ hoàn toàn mã kiểm tra tham số lặp lại bằng cách xác thực dữ liệu thô `json.RawMessage` đối chiếu với lược đồ JSON Schema (`tool.Schema`) của công cụ, bảo vệ logic thực thi khỏi các lỗi ép kiểu (type assertion panics).

## EXACT code to create

```go
// Package tool provides base structures and validation engines for agent tools.
package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

// BaseTool implements the metadata and schema accessors of contractstool.Tool interface.
// It integrates BasePlugin to support registry lifecycles.
type BaseTool struct {
	*sdkplugin.BasePlugin

	description string
	schema      *contractstool.Schema
}

// NewBaseTool constructs a BaseTool.
func NewBaseTool(name, description string, schema *contractstool.Schema) (*BaseTool, error) {
	if name == "" {
		return nil, errors.New("sdk/tool: tool name cannot be empty")
	}
	if description == "" {
		return nil, errors.New("sdk/tool: tool description cannot be empty")
	}
	if schema == nil {
		return nil, errors.New("sdk/tool: tool schema cannot be nil")
	}

	basePlugin, err := sdkplugin.NewBasePlugin(name, contractsplugin.TypeTool, "1.0.0")
	if err != nil {
		return nil, err
	}

	return &BaseTool{
		BasePlugin:  basePlugin,
		description: description,
		schema:      schema,
	}, nil
}

// Description returns the tool's purpose.
func (bt *BaseTool) Description() string {
	return bt.description
}

// Schema returns the parameters validation schema.
func (bt *BaseTool) Schema() *contractstool.Schema {
	return bt.schema
}

// ValidateArguments checks raw JSON arguments against the tool's JSON Schema constraints.
//
// Validation Rules:
//   - Checks for presence of all required keys.
//   - Validates value types (string, boolean, number, integer).
//   - Validates enum membership constraints.
func (bt *BaseTool) ValidateArguments(rawArgs json.RawMessage) error {
	if bt.schema == nil {
		return nil
	}

	// Handle empty/null inputs
	if len(rawArgs) == 0 || string(rawArgs) == "null" || string(rawArgs) == "{}" {
		if len(bt.schema.Required) > 0 {
			return fmt.Errorf("sdk/tool: missing required parameters: %v", bt.schema.Required)
		}
		return nil
	}

	var args map[string]any
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("sdk/tool: invalid JSON arguments: %w", err)
	}

	// 1. Verify required fields are present
	for _, reqField := range bt.schema.Required {
		val, ok := args[reqField]
		if !ok || val == nil {
			return fmt.Errorf("sdk/tool: missing required parameter %q", reqField)
		}
	}

	// 2. Validate types and value constraints
	for k, val := range args {
		prop, exists := bt.schema.Properties[k]
		if !exists {
			// Skip validation for undeclared args (flexibility)
			continue
		}
		if val == nil {
			continue // Checked by required rule above
		}
		if err := validateValueType(k, val, prop); err != nil {
			return fmt.Errorf("sdk/tool: %w", err)
		}
	}

	return nil
}

// validateValueType asserts native JSON type matching with property definitions.
func validateValueType(field string, val any, prop contractstool.Property) error {
	switch prop.Type {
	case "string":
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("parameter %q must be a string (got %T)", field, val)
		}
		// Validate enum constraint
		if len(prop.Enum) > 0 {
			valid := false
			for _, enumVal := range prop.Enum {
				if strVal == enumVal {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("parameter %q contains invalid value %q (allowed: %v)", field, strVal, prop.Enum)
			}
		}

	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("parameter %q must be a boolean (got %T)", field, val)
		}

	case "number":
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("parameter %q must be a number (got %T)", field, val)
		}

	case "integer":
		numVal, ok := val.(float64) // JSON unmarshals all numbers to float64
		if !ok {
			return fmt.Errorf("parameter %q must be an integer (got %T)", field, val)
		}
		if math.Mod(numVal, 1.0) != 0 {
			return fmt.Errorf("parameter %q must be a whole integer (got float %f)", field, numVal)
		}

	case "array":
		sliceVal, ok := val.([]any)
		if !ok {
			return fmt.Errorf("parameter %q must be an array (got %T)", field, val)
		}
		if prop.Items != nil {
			for i, item := range sliceVal {
				if err := validateValueType(fmt.Sprintf("%s[%d]", field, i), item, *prop.Items); err != nil {
					return err
				}
			}
		}

	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("parameter %q must be a nested object (got %T)", field, val)
		}
	}

	return nil
}
```

## ⚠️ Pitfalls

### Pitfall 1: Checking integers against Go `int` type in JSON unmarshal
```go
// ❌ WRONG:
if _, ok := val.(int); !ok {
    return error // All JSON unmarshalled numbers are float64 in map[string]any. This check always fails!
}

// ✅ CORRECT:
numVal, ok := val.(float64)
if !ok { return error }
if math.Mod(numVal, 1.0) != 0 { return error } // Check if float has decimal fraction
```
Go's standard `json.Unmarshal` decodes interface{} numbers into `float64`. Type-asserting directly to `int` is a common bug.

### Pitfall 2: Reversing validation path on empty parameter lists
If the schema requires parameters, but the AI submits an empty JSON payload (`null` or `{}`), parsing can succeed while silently skipping validations. Always check `schema.Required` when arguments are missing or empty.

## Verify
```bash
go build ./sdk/tool/...
```

## Checklist
- [ ] File `sdk/tool/tool.go` exists
- [ ] Package: `tool`
- [ ] `BaseTool` constructor validates name, description, and schema
- [ ] `ValidateArguments` checks for presence of required keys
- [ ] `validateValueType` handles float64-based integer checks safely
- [ ] Enum verification checks string values correctly
- [ ] Nested arrays and nested objects are validated recursively
- [ ] `go build ./sdk/tool/...` passes
