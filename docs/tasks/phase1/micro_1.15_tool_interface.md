# Micro-Task 1.15: Tạo contracts/tool/tool.go

## Thông tin
- **File tạo**: `contracts/tool/tool.go`
- **Package**: `tool`
- **Dependencies trước**: 1.14
- **Thời gian**: 10 phút

## Nội dung CHÍNH XÁC cần tạo

```go
package tool

import (
	"context"
	"encoding/json"
)

// Tool is the interface that all tools must implement.
//
// Tools are capabilities that agents can use to interact with the
// outside world (filesystem, git, terminal, browser, etc.).
//
// The AI provider sees the tool's Name, Description, and Schema.
// When it decides to use a tool, it sends a ToolCall with arguments.
// The orchestrator then calls Execute() with those arguments.
type Tool interface {
	// Name returns the unique identifier for this tool.
	// Convention: snake_case (e.g., "read_file", "git_commit", "run_command")
	Name() string

	// Description returns a human-readable description of what the tool does.
	// This is sent to the AI so it knows WHEN to use the tool.
	//
	// Guidelines for good descriptions:
	//   - Be specific: "Read the contents of a file at the given absolute path" (good)
	//   - Not vague: "Read a file" (bad — which file? relative or absolute?)
	//   - Include constraints: "Maximum file size: 1MB" (helps AI avoid errors)
	//   - Max length: 200 characters
	Description() string

	// Schema returns the JSON Schema for the tool's input parameters.
	// This tells the AI what arguments to pass and their types.
	Schema() *Schema

	// Execute runs the tool with the given arguments.
	//
	// Parameters:
	//   - ctx: For cancellation and timeout. Tools MUST respect ctx.Done().
	//   - args: Raw JSON arguments from the AI's tool call.
	//           Each tool should unmarshal into its own typed struct:
	//             var myArgs struct {
	//                 Path string `json:"path"`
	//             }
	//             json.Unmarshal(args, &myArgs)
	//
	// Returns:
	//   - *Result: The output of the tool execution.
	//   - error: System-level errors (file not found, permission denied).
	//           Tool "failures" (e.g., command exited with code 1) are NOT errors —
	//           they are valid Results with ExitCode != 0.
	Execute(ctx context.Context, args json.RawMessage) (*Result, error)
}

// Result represents the output of a tool execution.
type Result struct {
	// Output is the text output of the tool.
	// This is sent back to the AI as the tool's response.
	Output string `json:"output"`

	// Error is a human-readable error message (if the tool failed).
	// Populated when ExitCode != 0.
	Error string `json:"error,omitempty"`

	// ExitCode is the exit code of the tool execution.
	//   0 = success
	//   1 = general error
	//   2 = invalid arguments
	// For non-shell tools, use 0 for success and 1 for failure.
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
1. **Execute nhận `json.RawMessage`**: Mỗi tool tự unmarshal vào struct riêng. Type-safe hơn `map[string]any`.
2. **Error vs Result.Error**: `error` return = system error (panic, impossible state). `Result.Error` = tool-level error (file not found, command failed). AI cần biết tool errors để retry hoặc dùng tool khác.
3. **ExitCode 0 = success**: Giống Unix convention. Cho phép AI biết command thành công hay không.

## Checklist
- [ ] Tool interface với 4 methods (Name, Description, Schema, Execute)
- [ ] Execute nhận `json.RawMessage`, KHÔNG `map[string]any`
- [ ] Result struct với 3 fields (Output, Error, ExitCode)
- [ ] `IsSuccess()` helper
- [ ] `String()` method
- [ ] Godoc comments chi tiết
- [ ] `go build ./contracts/...` không lỗi
