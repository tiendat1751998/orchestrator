# Micro-Task 1.15: Create contracts/tool/tool.go

## Info
- **File**: `contracts/tool/tool.go`
- **Package**: `tool`
- **Depends on**: 1.14
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Defines the `Tool` contract that all agent tools must implement.

## EXACT code to create

```go
package tool

import (
	"context"
	"encoding/json"
)

// Tool is the interface that all tools must implement.
type Tool interface {
	// Name returns the unique identifier for this tool (e.g. "read_file").
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Schema returns the JSON Schema for the tool's input parameters.
	Schema() *Schema

	// Execute runs the tool with the given arguments.
	Execute(ctx context.Context, args json.RawMessage) (*Result, error)
}

// Result represents the output of a tool execution.
type Result struct {
	// Output is the text output of the tool.
	Output string `json:"output"`

	// Error is a human-readable error message (if the tool failed).
	Error string `json:"error,omitempty"`

	// ExitCode is the exit code of the tool execution (0 = success).
	ExitCode int `json:"exit_code"`
}

// IsSuccess returns true if the tool executed successfully (exit code 0).
func (r *Result) IsSuccess() bool {
	return r.ExitCode == 0
}

// String returns a human-readable representation of the result.
func (r *Result) String() string {
	if r.IsSuccess() {
		return r.Output
	}
	if r.Error != "" {
		return "error: " + r.Error
	}
	return r.Output
}
```

## ⚠️ Pitfalls

### Pitfall 1: Returning Go errors for normal execution failures
```go
if fileNotFound {
    return &Result{ExitCode: 1, Error: "file not found"}, nil // Normal feedback to AI
}
```
Only return Go `error` when the execution system itself has broken down (e.g. panic or resource exhaustion). If a command exits with code 1, return it as `Result` with `ExitCode = 1` so the AI can correct its parameters and try again.

### Pitfall 2: Using unstructured maps for Execute inputs
Using `map[string]any` in interface arguments causes type coercion errors. `json.RawMessage` permits each tool to deserialize parameters into its own strongly typed local struct.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/tool/tool.go` exists
- [ ] Package: `tool`
- [ ] `Tool` interface contains Name, Description, Schema, and Execute methods
- [ ] `Execute` receives `json.RawMessage` arguments
- [ ] `Result` contains Output, Error, and ExitCode fields
- [ ] `IsSuccess()` and `String()` helper methods exist on `Result`
- [ ] `go build ./contracts/...` passes
