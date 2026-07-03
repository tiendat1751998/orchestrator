package tool

import (
	"encoding/json"
	"fmt"

	"github.com/tiendat1751998/orchestrator/contracts/tool"
)

// Success creates a successful tool.Result with exit code 0.
func Success(output string) *tool.Result {
	return &tool.Result{
		Output:   output,
		ExitCode: 0,
	}
}

// Failure creates a failed tool.Result with exit code 1.
func Failure(errorMessage string) *tool.Result {
	return &tool.Result{
		Error:    errorMessage,
		ExitCode: 1,
	}
}

// WithExitCode constructs a tool.Result with a custom exit status.
// Uses exit code 0 for success and non-zero for failures.
func WithExitCode(output string, errorMsg string, exitCode int) *tool.Result {
	return &tool.Result{
		Output:   output,
		Error:    errorMsg,
		ExitCode: exitCode,
	}
}

// JSON serializes the given object into a JSON string and returns it as a successful tool.Result.
// Useful when a tool needs to return structured data (like file metadata lists or process lists).
func JSON(data any) (*tool.Result, error) {
	if data == nil {
		return Success("{}"), nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("sdk/tool: failed to serialize result struct to JSON: %w", err)
	}

	return Success(string(bytes)), nil
}
