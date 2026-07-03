package tool_test

import (
	"encoding/json"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/tool"
)

// =============================================================================
// Test: Schema JSON serialization
// =============================================================================

func TestSchema_JSONRoundTrip(t *testing.T) {
	s := tool.NewSchema().
		AddProperty("path", tool.StringProperty("File path")).
		AddProperty("count", tool.IntProperty("Iteration count")).
		AddProperty("mode", tool.EnumProperty("Run mode", "dry", "live")).
		AddRequired("path")

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded tool.Schema
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Type != "object" {
		t.Errorf("Type: got %q, want 'object'", decoded.Type)
	}
	if len(decoded.Properties) != 3 {
		t.Errorf("Properties length: got %d, want 3", len(decoded.Properties))
	}
	if len(decoded.Required) != 1 || decoded.Required[0] != "path" {
		t.Errorf("Required fields: got %v", decoded.Required)
	}
}

// =============================================================================
// Test: Property Generators
// =============================================================================

func TestProperty_Generators(t *testing.T) {
	strProp := tool.StringProperty("string desc")
	if strProp.Type != "string" || strProp.Description != "string desc" {
		t.Errorf("StringProperty: %v", strProp)
	}

	intProp := tool.IntProperty("int desc")
	if intProp.Type != "integer" || intProp.Description != "int desc" {
		t.Errorf("IntProperty: %v", intProp)
	}

	boolProp := tool.BoolProperty("bool desc")
	if boolProp.Type != "boolean" || boolProp.Description != "bool desc" {
		t.Errorf("BoolProperty: %v", boolProp)
	}

	enumProp := tool.EnumProperty("enum desc", "a", "b")
	if enumProp.Type != "string" || len(enumProp.Enum) != 2 || enumProp.Enum[0] != "a" {
		t.Errorf("EnumProperty: %v", enumProp)
	}
}

// =============================================================================
// Test: Result Helpers
// =============================================================================

func TestResult_Helpers(t *testing.T) {
	successRes := &tool.Result{
		Output:   "command success output",
		ExitCode: 0,
	}

	if !successRes.IsSuccess() {
		t.Error("expected IsSuccess() = true for exit code 0")
	}
	if successRes.String() != "command success output" {
		t.Errorf("String(): got %q", successRes.String())
	}

	failRes := &tool.Result{
		Error:    "missing target file",
		ExitCode: 1,
	}

	if failRes.IsSuccess() {
		t.Error("expected IsSuccess() = false for exit code 1")
	}
	if failRes.String() != "error: missing target file" {
		t.Errorf("String(): got %q", failRes.String())
	}
}
