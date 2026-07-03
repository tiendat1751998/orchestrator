# Micro-Task 1.39: Update Validation for Task and Request (Input Hardening)

## Info
- **File**: 
  - `contracts/agent/task.go` (Add Validate method)
  - `contracts/provider/request.go` (Update Validate utilizing global contracts.ValidationError)
- **Package**: `agent`, `provider`
- **Depends on**: 1.18 (task.go), 1.09 (request.go), 1.37 (structured errors.go)
- **Time**: 20 min
- **Verify**: `go build ./contracts/...`

## Purpose
Upgrades input validation models. The `Validate()` methods for both agents and providers will directly return structured global `contracts.ValidationError` records instead of localized sentinel error strings, blocking invalid payloads at component boundaries.

## EXACT code to create

### Part 1: Update `contracts/agent/task.go`

Add `Validate()` method to `Task` struct inside [contracts/agent/task.go](file:///d:/project/orchestrator/contracts/agent/task.go):

```go
// Validate checks if the task has all required fields and valid dependency mappings.
//
// Validation rules:
//   - ID cannot be empty.
//   - Name cannot be empty.
//   - Type cannot be empty.
//   - Timeout must be >= 0.
//   - Dependencies must not contain duplicates and cannot self-reference the task.
func (t *Task) Validate() error {
	if t.ID.IsEmpty() {
		return contracts.NewValidationError("task", "id", "required")
	}
	if t.Name == "" {
		return contracts.NewValidationError("task", "name", "required")
	}
	if t.Type == "" {
		return contracts.NewValidationError("task", "type", "required")
	}
	if t.Timeout < 0 {
		return contracts.NewValidationError("task", "timeout", "must be >= 0")
	}

	// Validate dependencies
	seen := make(map[contracts.TaskID]bool)
	for _, depID := range t.Dependencies {
		if depID.IsEmpty() {
			return contracts.NewValidationError("task", "dependencies", "contains empty dependency ID")
		}
		if depID == t.ID {
			return contracts.NewValidationError("task", "dependencies", "self-dependency is not allowed: "+string(depID))
		}
		if seen[depID] {
			return contracts.NewValidationError("task", "dependencies", "duplicate dependency: "+string(depID))
		}
		seen[depID] = true
	}

	return nil
}
```

---

### Part 2: Update `contracts/provider/request.go`

Completely replace the `Request Validation` section at the end of [contracts/provider/request.go](file:///d:/project/orchestrator/contracts/provider/request.go) with validation code returning `contracts.ValidationError`:

```go
import (
	"github.com/tiendat1751998/orchestrator/contracts"
)

// Validate checks if the request has minimum required fields and valid parameter bounds.
// Returns nil if valid, or a structured *contracts.ValidationError.
func (r *Request) Validate() error {
	if len(r.Messages) == 0 {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "messages",
			Reason:    "at least one message is required",
		}
	}

	for i, msg := range r.Messages {
		if !msg.Role.IsValid() {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].role", i),
				Reason:    "invalid role: " + string(msg.Role),
			}
		}
		// System and user messages must contain content.
		if (msg.Role == RoleSystem || msg.Role == RoleUser) && msg.Content == "" {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].content", i),
				Reason:    "content is required for system/user messages",
			}
		}
		if msg.Role == RoleTool && msg.ToolCallID == "" {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].tool_call_id", i),
				Reason:    "tool_call_id is required for tool result messages",
			}
		}
	}

	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "temperature",
			Reason:    "must be between 0.0 and 2.0",
		}
	}

	if r.MaxTokens != nil && *r.MaxTokens < 1 {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "max_tokens",
			Reason:    "must be >= 1",
		}
	}

	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "top_p",
			Reason:    "must be between 0.0 and 1.0",
		}
	}

	return nil
}
```

Ensure the legacy local `ValidationError` struct definition inside [contracts/provider/request.go](file:///d:/project/orchestrator/contracts/provider/request.go) is removed.

## Rules
1. **No Duplicate Declarations**: The old `ValidationError` struct defined locally inside `provider/request.go` must be deleted.
2. **Duplicate Dependency Rejections**: Tasks must reject duplicate dependencies, empty dependency IDs, or self-dependencies in `Validate()`.
3. **Empty Message Safety Checks**: System and user messages must contain non-empty content strings. Tool responses must declare non-empty `ToolCallID` links.

## ⚠️ Pitfalls

### Pitfall 1: Type conflicts during compilation from duplicate struct declarations
If you forget to delete the legacy `ValidationError` struct inside `request.go`, the compiler will fail with a redeclaration error. Always delete local validations types in favor of the shared `contracts.ValidationError`.

### Pitfall 2: Permitting circular dependencies in task setups
If `Task.Validate()` overlooks self-referential dependencies, the scheduler scheduler will enter infinite loops or crash with stack overflow panics. Reject self-dependencies early.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] `Validate()` method added to `contracts/agent/task.go`
- [ ] Task validator catches duplicate dependencies and self-references
- [ ] Deleted local `ValidationError` from `contracts/provider/request.go`
- [ ] `Request.Validate()` returns global `contracts.ValidationError` types
- [ ] Validator checks for non-empty content on user/system roles
- [ ] Validator checks for non-empty `tool_call_id` on tool roles
- [ ] `go build ./contracts/...` passes
