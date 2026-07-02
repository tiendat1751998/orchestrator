# Micro-Task 1.40: Tạo contracts/plugin/health.go và Cập nhật Plugin Interface (Health Check Depth)

## Thông tin
- **File tạo**: `contracts/plugin/health.go`
- **File cập nhật**: `contracts/plugin/plugin.go`
- **Package**: `plugin`
- **Dependencies trước**: 1.24 (plugin.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./contracts/plugin/...`

## Purpose
Nâng cấp cơ chế kiểm tra sức khỏe (Health Check) từ nông (shallow) sang sâu (structured health depth). Thay vì chỉ trả về `error` đơn thuần, mỗi plugin sẽ trả về một `HealthReport` chi tiết chứa thông tin về trạng thái (`ok`, `degraded`, `down`), thời gian thực thi check, tài nguyên chi tiết (details), và báo cáo con (children reports) của các thành phần phụ thuộc.

## EXACT code to create

### Phần 1: Tạo `contracts/plugin/health.go`

```go
package plugin

import (
	"time"
)

// HealthStatus represents the high-level health state of a plugin.
type HealthStatus string

const (
	// HealthOK indicates the plugin is fully healthy and operational.
	HealthOK HealthStatus = "ok"

	// HealthDegraded indicates the plugin is running but with limited performance or minor errors.
	// Example: Gemini provider is working, but experiencing high latency or rate limits.
	HealthDegraded HealthStatus = "degraded"

	// HealthDown indicates the plugin is non-functional.
	// Example: API key is invalid or external server is completely unreachable.
	HealthDown HealthStatus = "down"
)

// HealthReport provides a structured, hierarchical report of plugin health.
// Suitable for JSON serialization in API endpoints.
type HealthReport struct {
	// Status is the overall health status of this plugin.
	Status HealthStatus `json:"status"`

	// Message describes the reason for non-healthy status.
	Message string `json:"message,omitempty"`

	// Details contains plugin-specific metric indicators (e.g. queue depth, latency).
	Details map[string]any `json:"details,omitempty"`

	// Children contains reports from internal dependencies (e.g. sub-agents, DB connections).
	Children map[string]HealthReport `json:"children,omitempty"`

	// Timestamp is when the health check was performed.
	Timestamp time.Time `json:"timestamp"`

	// Duration measures how long the health check took to execute.
	// Useful for identifying slow check methods before they block the kernel.
	Duration time.Duration `json:"duration"`
}

// IsHealthy returns true if the status is OK or Degraded (still operational).
func (hr HealthReport) IsHealthy() bool {
	return hr.Status == HealthOK || hr.Status == HealthDegraded
}
```

---

### Phần 2: Cập nhật `contracts/plugin/plugin.go`

Cập nhật phương thức `Health` trong interface `Plugin`:

```go
// Plugin is the lifecycle interface for all pluggable components.
type Plugin interface {
	Name() string
	Type() Type
	Version() string
	Init(ctx context.Context, config map[string]any) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Health checks if the plugin is functioning correctly and returns a detailed report.
	//
	// Parameters:
	//   - ctx: for timeout enforcement. Checkers should abort and return error on cancellation.
	//
	// Returns:
	//   - HealthReport: structured health status.
	//   - error: system level failure during the health check itself (e.g. context timeout).
	//            If the plugin is simply unhealthy (down), return (HealthReport{Status: HealthDown}, nil)
	//            rather than a non-nil error.
	Health(ctx context.Context) (HealthReport, error)
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Trả về error hệ thống thay vì HealthDown trong HealthReport
```go
// ❌ SAI:
func (p *MyPlugin) Health(ctx context.Context) (HealthReport, error) {
    if err := p.pingAPI(); err != nil {
        return HealthReport{}, err // Sai: API sập không phải lỗi hệ thống của checker
    }
}

// ✅ ĐÚNG:
func (p *MyPlugin) Health(ctx context.Context) (HealthReport, error) {
    if err := p.pingAPI(); err != nil {
        return HealthReport{
            Status:    HealthDown,
            Message:   "API unreachable: " + err.Error(),
            Timestamp: time.Now(),
        }, nil // Trả về err = nil vì việc kiểm tra đã thành công (phát hiện ra API sập)
    }
}
```
Lỗi trả về từ `Health` (parameter `error` thứ hai) chỉ dùng cho lỗi của bản thân việc chạy hàm check (ví dụ: Timeout của Context). Nếu phát hiện plugin lỗi/chết, hãy trả về `HealthReport{Status: HealthDown}` và `error = nil`. Nếu trả về `error != nil`, kernel sẽ coi đó là lỗi checker sập chứ không chỉ là plugin unhealthy.

### Pitfall 2: Bỏ qua Duration hoặc Timestamp
Báo cáo sức khỏe không có Timestamp và Duration rất khó trace trong các hệ thống phân tán. Luôn ghi nhận thời điểm kiểm tra (`time.Now()`) và đo đạc duration bằng `time.Since(startTime)`.

## Checklist
- [ ] File `contracts/plugin/health.go` tồn tại
- [ ] Định nghĩa `HealthStatus` với 3 hằng số: HealthOK, HealthDegraded, HealthDown
- [ ] Struct `HealthReport` chứa đủ các trường: Status, Message, Details, Children, Timestamp, Duration
- [ ] Phương thức `IsHealthy()` hỗ trợ cả OK và Degraded
- [ ] Interface `Plugin` trong `contracts/plugin/plugin.go` cập nhật method `Health` trả về `(HealthReport, error)`
- [ ] Method `Health` trong `plugin.go` có Godoc hướng dẫn cách trả về error đúng đắn
- [ ] `go build ./contracts/...` không lỗi
