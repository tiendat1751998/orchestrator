# Micro-Task 3.11: Create sdk/tool/result.go

## Info
- **File**: `sdk/tool/result.go`
- **Package**: `tool`
- **Depends on**: 1.15 (tool.go contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/tool/...`

## Purpose
Implements tool result builders (`Success`, `Failure`, `WithExitCode`, and `JSON` serializers) that format tool outputs into consistent `tool.Result` structures, resolving serializations of complex data.

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

## Rules
1. **Tool Execution Success vs. Failure**: Differentiate between system-level execution errors (returned as Go `error` parameters, causing agent loop stops) and command execution failures (returned as `tool.Result` with non-zero exit codes, which agents can correct).
2. **Strict Serializations Check**: Always capture serialization errors during `JSON` marshals, returning them as system-level Go errors to stop execution.
3. **Structured String representations**: Ensure outputs are formatted cleanly inside JSON envelopes to allow AI models to read parameter maps.

## ⚠️ Pitfalls

### Pitfall 1: Returning Go errors for command failures
```go
```
If a file is missing or a command exits with code 1, return `(tool.Failure("..."), nil)` to let the agent self-correct.

### Pitfall 2: Silently swallowing json marshal errors
Omitting err checks on struct serialization in `JSON` builders results in returning empty values like `""`, confusing the calling agent. Always bubble up marshal failures.

## Verify
```bash
go build ./sdk/tool/...
```

## Checklist
- [ ] File `sdk/tool/result.go` exists
- [ ] Package: `tool`
- [ ] `Success` returns structured results with exit code 0
- [ ] `Failure` populates error fields and sets exit code 1
- [ ] `WithExitCode` accepts custom status values
- [ ] `JSON` helper serializes struct states to JSON strings
- [ ] JSON marshalling failures are returned as Go system errors
- [ ] `go build ./sdk/tool/...` passes
