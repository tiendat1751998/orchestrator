# Micro-Task 3.11: Create sdk/tool/result.go

## Info
- **File**: `sdk/tool/result.go`
- **Package**: `tool`
- **Depends on**: 1.15 (tool.go contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/tool/...`

## Purpose
Triển khai bộ sinh kết quả nhanh cho công cụ (`ResultBuilders`). Cung cấp các hàm tiện ích (`Success`, `Failure`, `JSON`) để tạo đối tượng `tool.Result` chuẩn chỉnh, bao gồm tự động tuần tự hóa (serialization) các đối tượng dữ liệu phức tạp sang chuỗi JSON trước khi trả về cho AI.

## EXACT code to create

```go
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
```

## ⚠️ Pitfalls

### Pitfall 1: Confusing tool failure with system error (Go `error`)
```go
// ❌ WRONG:
func (t *MyTool) Execute(ctx, args) (*tool.Result, error) {
    if fileNotFound {
        return nil, errors.New("file not found") // Will interrupt the agent loop and fail the mission!
    }
}

// ✅ CORRECT:
func (t *MyTool) Execute(ctx, args) (*tool.Result, error) {
    if fileNotFound {
        return tool.Failure("file not found"), nil // Return result showing failure, allowing AI to handle it
    }
}
```
A Go `error` returned from `Execute` signifies a system-level breakdown (e.g. panic, memory limit). If the tool simply failed to find a file or command exited non-zero, this is a normal tool result. Return `(tool.Failure("..."), nil)` so the agent receives the error string and can correct its path.

### Pitfall 2: Silently ignoring serialization errors in JSON results
If `json.Marshal(data)` fails, returning a half-empty result hides the bug. Always catch the marshal error and return it as a system error `(nil, err)` so it fails the validation build gates early.

## Verify
```bash
go build ./sdk/tool/...
```

## Checklist
- [ ] File `sdk/tool/result.go` exists
- [ ] Package: `tool`
- [ ] `Success` constructor sets `ExitCode` = 0
- [ ] `Failure` constructor sets `ExitCode` = 1 and populates the `Error` field
- [ ] `WithExitCode` allows setting custom codes (e.g. 2 for invalid arguments)
- [ ] `JSON` helper marshals structs into JSON strings correctly
- [ ] `JSON` returns a system error if serialization fails
- [ ] `go build ./sdk/tool/...` passes
