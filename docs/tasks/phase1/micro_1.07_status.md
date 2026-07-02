# Micro-Task 1.07: Tạo contracts/status.go

## Thông tin
- **File tạo**: `contracts/status.go`
- **Package**: `contracts`
- **Dependencies trước**: 1.05
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package contracts

// Status represents the current state of a task, mission, or agent.
// It uses string constants for easy serialization to JSON/YAML
// and human-readable log output.
type Status string

const (
	// StatusPending indicates the item is queued but not yet started.
	StatusPending Status = "pending"

	// StatusRunning indicates the item is currently being executed.
	StatusRunning Status = "running"

	// StatusSuccess indicates the item completed successfully.
	StatusSuccess Status = "success"

	// StatusFailed indicates the item completed with an error.
	StatusFailed Status = "failed"

	// StatusCancelled indicates the item was cancelled before completion
	// (e.g., user pressed Ctrl+C, or parent context was cancelled).
	StatusCancelled Status = "cancelled"

	// StatusRetrying indicates the item failed and is being retried.
	StatusRetrying Status = "retrying"

	// StatusSkipped indicates the item was skipped
	// (e.g., a dependency failed and the item cannot run).
	StatusSkipped Status = "skipped"

	// StatusTimeout indicates the item exceeded its time limit.
	StatusTimeout Status = "timeout"
)

// IsTerminal returns true if the status represents a final state
// (no further transitions expected).
//
// Terminal states: success, failed, cancelled, skipped, timeout.
// Non-terminal states: pending, running, retrying.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled, StatusSkipped, StatusTimeout:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the status indicates successful completion.
func (s Status) IsSuccess() bool {
	return s == StatusSuccess
}

// IsFailed returns true if the status indicates a failure.
func (s Status) IsFailed() bool {
	return s == StatusFailed || s == StatusTimeout
}

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// ValidStatuses returns all valid status values.
// Useful for validation and documentation.
func ValidStatuses() []Status {
	return []Status{
		StatusPending,
		StatusRunning,
		StatusSuccess,
		StatusFailed,
		StatusCancelled,
		StatusRetrying,
		StatusSkipped,
		StatusTimeout,
	}
}

// IsValidStatus checks if a string is a valid Status value.
func IsValidStatus(s string) bool {
	for _, valid := range ValidStatuses() {
		if string(valid) == s {
			return true
		}
	}
	return false
}
```

## Quy tắc
1. Status dùng `string` constants — serialize/deserialize tự động với JSON/YAML
2. `IsTerminal()` quan trọng — scheduler dùng để biết task đã xong chưa
3. `IsValidStatus()` quan trọng — validate input từ config/API
4. `StatusTimeout` tách riêng khỏi `StatusFailed` — cho phép retry logic phân biệt timeout vs lỗi logic
5. `StatusSkipped` cho trường hợp dependency failed — task phụ thuộc không cần chạy nữa

## ⚠️ Pitfalls cần tránh
1. **KHÔNG dùng `iota`**: `iota` tạo giá trị `0, 1, 2...` → serialize thành số → khó đọc trong logs/JSON
2. **KHÔNG thêm status mới bừa bãi**: Mỗi status mới cần update `IsTerminal()`, `ValidStatuses()`, và tất cả switch statements trong codebase. Thêm khi CẦN, không phải "phòng ngừa"
3. **Status transitions**: Chưa enforce trong code lúc này, nhưng PHẢI tuân thủ:
   ```
   pending → running → success
                     → failed → retrying → running (loop)
                     → timeout
                     → cancelled
   pending → skipped (khi dependency failed)
   ```

## Checklist
- [ ] File `contracts/status.go` tồn tại
- [ ] 8 status constants
- [ ] `IsTerminal()` trả về đúng cho terminal states
- [ ] `IsSuccess()` và `IsFailed()` helper methods
- [ ] `IsValidStatus()` validation function
- [ ] `ValidStatuses()` trả về danh sách đầy đủ
- [ ] Godoc comment cho mỗi status constant
- [ ] `go build ./contracts/...` không lỗi
