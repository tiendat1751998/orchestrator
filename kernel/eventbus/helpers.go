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
