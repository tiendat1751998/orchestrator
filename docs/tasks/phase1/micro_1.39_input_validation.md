# Micro-Task 1.39: Cập nhật Validation cho Task và Request (Production Hardening)

## Thông tin
- **File cập nhật**: 
  - `contracts/agent/task.go` (Thêm Validate method)
  - `contracts/provider/request.go` (Cập nhật Validate sử dụng structured ValidationError)
- **Package**: `agent`, `provider`
- **Dependencies trước**: 1.18 (task.go), 1.09 (request.go), 1.37 (structured errors.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./contracts/...`

## Purpose
Nâng cấp cơ chế Validate của các struct đầu vào. Thay vì trả về sentinel errors chung chung hoặc string-based error tự chế, các phương thức `Validate()` sẽ trực tiếp sử dụng struct `contracts.ValidationError` được chuẩn hóa để trả về thông tin chi tiết về field bị lỗi và lý do, ngăn chặn dữ liệu không hợp lệ đi sâu vào kernel.

## EXACT code to create

### Phần 1: Cập nhật `contracts/agent/task.go`

Thêm phương thức `Validate()` cho struct `Task`:

```go
// Validate checks if the task has all required fields and valid dependency mappings.
//
// Validation rules:
//   - ID cannot be empty.
//   - Name cannot be empty and should be in snake_case.
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

### Phần 2: Cập nhật `contracts/provider/request.go`

Thay thế hoàn toàn phần `Request Validation` ở cuối file bằng code sử dụng `contracts.ValidationError` toàn cục:

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
		// Assistant messages can have empty content if tool calls are present.
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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Trộn lẫn kiểu ValidationError cũ và mới
Không định nghĩa struct `ValidationError` cục bộ trong `request.go` nữa. Xóa hoàn toàn struct `ValidationError` cũ trong `request.go` và chuyển sang sử dụng `contracts.ValidationError` toàn cục từ file `contracts/errors.go` để tránh xung đột kiểu (type conflict) khi biên dịch.

### Pitfall 2: Chấp nhận dependencies rỗng hoặc lặp
```go
// ❌ SAI:
// Bỏ qua không check dependencies trùng nhau -> scheduler lập lịch sẽ bị lỗi vòng lặp hoặc chạy thừa task.
```
Validation ở tầng contract là chốt chặn cuối cùng. `Task.Validate()` bắt buộc phải check trùng lặp dependency ID và tự tham chiếu.

## Checklist
- [ ] File `contracts/agent/task.go` chứa phương thức `Validate()`
- [ ] Phương thức `Task.Validate()` bắt trùng lặp dependency, dependency rỗng, tự tham chiếu
- [ ] File `contracts/provider/request.go` đã xóa struct `ValidationError` cũ của nó
- [ ] Phương thức `Request.Validate()` sử dụng `contracts.ValidationError` toàn cục
- [ ] Bắt lỗi message rỗng đối với `system` và `user` roles
- [ ] Bắt lỗi thiếu `tool_call_id` đối với `tool` role
- [ ] `go build ./contracts/...` không lỗi
