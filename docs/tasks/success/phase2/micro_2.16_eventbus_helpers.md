# Micro-Task 2.16: Create kernel/eventbus/helpers.go

## Info
- **File**: `kernel/eventbus/helpers.go`
- **Package**: `eventbus`
- **Depends on**: 2.15 (bus.go)
- **Time**: 10 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Implements helper functions (`PublishTaskStarted`, `PublishTaskCompleted`, `PublishTaskFailed`, `PublishKernelStarted`, `PublishKernelStopped`) that wrap event creation and publication patterns for common execution state updates.

## EXACT code to create

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

## Rules
1. **Event Constant Matches**: Always map event types to the exported string constants declared in the `contracts/event` package (e.g. `event.EventTaskStarted`).
2. **Context Parameter Handling**: Since helper calls are fire-and-forget, they pass `nil` contexts. The receiving `Publish` method must internally handle `nil` inputs by substituting `context.Background()` to prevent panic crashes.

## ⚠️ Pitfalls

### Pitfall 1: Bypassing context checks inside Bus.Publish
Passing `nil` context values to functions that call `ctx.Done()` directly causes nil pointer dereference panics. Ensure `Bus.Publish` resolves `nil` variables to default backgrounds early.

### Pitfall 2: Typos in payload map key assignments
Mispelling payload keys (such as `task_id`) will prevent subscriber consumers from properly filtering event records. Constrain payload names strictly to the defined keys.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/helpers.go` exists
- [ ] Package: `eventbus`
- [ ] Helpers defined: PublishTaskStarted, PublishTaskCompleted, PublishTaskFailed, PublishKernelStarted, PublishKernelStopped
- [ ] Event structures use standard event type constants from contracts
- [ ] Event timestamps are set with `time.Now()`
- [ ] `go build ./kernel/eventbus/...` passes
