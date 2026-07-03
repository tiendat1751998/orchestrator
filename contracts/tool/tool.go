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
