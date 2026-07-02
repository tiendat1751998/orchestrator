# Micro-Task 1.38: Tạo contracts/context/metadata.go (Bổ sung Context Propagation)

## Thông tin
- **File tạo**: `contracts/context/metadata.go`
- **Package**: `agentcontext`
- **Dependencies trước**: 1.28 (contracts/context/context.go)
- **Thời gian**: 15 phút
- **Verify**: `go build ./contracts/context/...`

## Purpose
Định nghĩa cơ chế truyền tải metadata (Context Propagation) trong context. Điều này cho phép lưu trữ và trích xuất request-scoped data (Trace ID, Task ID, Mission ID, Agent Name) xuyên suốt các lời gọi hàm I/O, provider calls, và hệ thống logs để phục vụ tracing và log correlation.

## EXACT code to create

```go
package agentcontext

import (
	"context"
)

// contextKey is a private type for context keys to prevent collisions.
//
// WHY private type?
// → Standard Go practice. If we use string as key:
//     ctx = context.WithValue(ctx, "trace_id", val)
//     Another package does the same -> keys collide and overwrite.
// → private contextKey type guarantees only this package can read/write these keys.
type contextKey string

const (
	keyTraceID   contextKey = "trace_id"
	keyTaskID    contextKey = "task_id"
	keyMissionID contextKey = "mission_id"
	keyAgentName contextKey = "agent_name"
)

// WithTraceID returns a new context with the given Trace ID.
// Used at request ingress or mission startup.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyTraceID, traceID)
}

// TraceIDFrom extracts the Trace ID from context.
// Returns empty string if not set.
func TraceIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyTraceID).(string); ok {
		return v
	}
	return ""
}

// WithTaskID returns a new context with the given Task ID.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	if taskID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyTaskID, taskID)
}

// TaskIDFrom extracts the Task ID from context.
func TaskIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyTaskID).(string); ok {
		return v
	}
	return ""
}

// WithMissionID returns a new context with the given Mission ID.
func WithMissionID(ctx context.Context, missionID string) context.Context {
	if missionID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyMissionID, missionID)
}

// MissionIDFrom extracts the Mission ID from context.
func MissionIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyMissionID).(string); ok {
		return v
	}
	return ""
}

// WithAgentName returns a new context with the given Agent Name.
func WithAgentName(ctx context.Context, name string) context.Context {
	if name == "" {
		return ctx
	}
	return context.WithValue(ctx, keyAgentName, name)
}

// AgentNameFrom extracts the Agent Name from context.
func AgentNameFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyAgentName).(string); ok {
		return v
	}
	return ""
}

// GetMetadata returns a map of all metadata present in the context.
// Useful for populating logger structured fields automatically.
func GetMetadata(ctx context.Context) map[string]string {
	meta := make(map[string]string)
	if ctx == nil {
		return meta
	}

	if traceID := TraceIDFrom(ctx); traceID != "" {
		meta["trace_id"] = traceID
	}
	if taskID := TaskIDFrom(ctx); taskID != "" {
		meta["task_id"] = taskID
	}
	if missionID := MissionIDFrom(ctx); missionID != "" {
		meta["mission_id"] = missionID
	}
	if agentName := AgentNameFrom(ctx); agentName != "" {
		meta["agent"] = agentName
	}

	return meta
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Dùng string literal làm context key
```go
// ❌ SAI:
ctx := context.WithValue(parent, "trace_id", val)

// ✅ ĐÚNG:
type contextKey string
const keyTraceID contextKey = "trace_id"
ctx := context.WithValue(parent, keyTraceID, val)
```
Chuỗi thô `"trace_id"` có thể bị đè dữ liệu bởi bất kỳ thư viện bên thứ ba nào dùng cùng chuỗi làm key. Định nghĩa kiểu dữ liệu private là bắt buộc.

### Pitfall 2: Không kiểm tra nil context
```go
// ❌ SAI:
func TraceIDFrom(ctx context.Context) string {
    return ctx.Value(keyTraceID).(string) // Sẽ panic nếu ctx = nil hoặc key không tồn tại/không phải string
}

// ✅ ĐÚNG:
func TraceIDFrom(ctx context.Context) string {
    if ctx == nil {
        return ""
    }
    if v, ok := ctx.Value(keyTraceID).(string); ok {
        return v
    }
    return ""
}
```
Thiếu kiểm tra nil context hoặc không dùng type assertion an toàn (`v, ok := ...`) sẽ dẫn tới panic lúc chạy (runtime panic) nếu context rỗng hoặc chứa giá trị sai kiểu dữ liệu.

## Checklist
- [ ] File `contracts/context/metadata.go` tồn tại
- [ ] Package name: `agentcontext`
- [ ] Định nghĩa private type `contextKey`
- [ ] Có đủ 4 keys: keyTraceID, keyTaskID, keyMissionID, keyAgentName
- [ ] Các hàm Setter: WithTraceID, WithTaskID, WithMissionID, WithAgentName
- [ ] Các hàm Getter: TraceIDFrom, TaskIDFrom, MissionIDFrom, AgentNameFrom
- [ ] Các hàm Getter kiểm tra an toàn `ctx != nil` và type assertion `ok`
- [ ] Hàm `GetMetadata` gộp các giá trị phi rỗng vào map
- [ ] `go build ./contracts/context/...` không lỗi
