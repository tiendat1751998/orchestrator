# Micro-Task 2.16: Tạo kernel/eventbus/helpers.go

## Thông tin
- **File tạo**: `kernel/eventbus/helpers.go`
- **Package**: `eventbus`
- **Dependencies trước**: 2.15 (bus.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package eventbus

import (
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// PublishTaskStarted emits a "task.started" event.
func PublishTaskStarted(bus event.Bus, taskID, agentName string) {
	bus.Publish(nil, event.Event{
		Type:      event.EventTaskStarted,
		Source:    "runtime",
		Payload:   map[string]string{"task_id": taskID, "agent": agentName},
		Timestamp: time.Now(),
	})
}

// PublishTaskCompleted emits a "task.completed" event.
func PublishTaskCompleted(bus event.Bus, taskID string, output any) {
	bus.Publish(nil, event.Event{
		Type:      event.EventTaskCompleted,
		Source:    "runtime",
		Payload:   map[string]any{"task_id": taskID, "output": output},
		Timestamp: time.Now(),
	})
}

// PublishTaskFailed emits a "task.failed" event.
func PublishTaskFailed(bus event.Bus, taskID string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	bus.Publish(nil, event.Event{
		Type:      event.EventTaskFailed,
		Source:    "runtime",
		Payload:   map[string]string{"task_id": taskID, "error": errMsg},
		Timestamp: time.Now(),
	})
}

// PublishKernelStarted emits a "kernel.started" event.
func PublishKernelStarted(bus event.Bus) {
	bus.Publish(nil, event.Event{
		Type:      event.EventKernelStarted,
		Source:    "kernel",
		Timestamp: time.Now(),
	})
}

// PublishKernelStopped emits a "kernel.stopped" event.
func PublishKernelStopped(bus event.Bus) {
	bus.Publish(nil, event.Event{
		Type:      event.EventKernelStopped,
		Source:    "kernel",
		Timestamp: time.Now(),
	})
}
```

## ⚠️ Pitfall: `bus.Publish(nil, ...)` — context=nil
Helpers pass `nil` context vì đây là fire-and-forget events.
Caller có thể dùng `Publish()` trực tiếp nếu cần cancellation.
Nếu bus.Publish kiểm tra nil ctx → phải xử lý (dùng context.Background()).

> **CẬP NHẬT cho bus.go**: Nếu bus.Publish nhận ctx=nil, thay bằng context.Background() internal:
> ```go
> if ctx == nil { ctx = context.Background() }
> ```

## Checklist
- [ ] File `kernel/eventbus/helpers.go` tồn tại
- [ ] 5 helper functions: TaskStarted, TaskCompleted, TaskFailed, KernelStarted, KernelStopped
- [ ] Dùng event constants từ contracts/event
- [ ] Timestamp = time.Now()
- [ ] `go build ./kernel/eventbus/...` không lỗi
