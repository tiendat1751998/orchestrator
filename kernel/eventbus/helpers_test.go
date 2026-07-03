package eventbus

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

func TestHelpers(t *testing.T) {
	bus := New(nil)

	var wg sync.WaitGroup
	var mu sync.Mutex
	receivedEvents := make([]event.Event, 0)

	unsub, err := bus.Subscribe("*", func(evt event.Event) {
		mu.Lock()
		receivedEvents = append(receivedEvents, evt)
		mu.Unlock()
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer unsub()

	// 1. PublishTaskStarted
	wg.Add(1)
	PublishTaskStarted(bus, "task-1", "agent-backend")

	// 2. PublishTaskCompleted
	wg.Add(1)
	PublishTaskCompleted(bus, "task-2", "output-data")

	// 3. PublishTaskFailed (with error)
	wg.Add(1)
	PublishTaskFailed(bus, "task-3", errors.New("something went wrong"))

	// 4. PublishTaskFailed (with nil error)
	wg.Add(1)
	PublishTaskFailed(bus, "task-4", nil)

	// 5. PublishKernelStarted
	wg.Add(1)
	PublishKernelStarted(bus)

	// 6. PublishKernelStopped
	wg.Add(1)
	PublishKernelStopped(bus)

	// Wait for handler to run (or timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for events to be delivered")
	}

	mu.Lock()
	events := make([]event.Event, len(receivedEvents))
	copy(events, receivedEvents)
	mu.Unlock()

	if len(events) != 6 {
		t.Fatalf("Expected 6 events, got %d", len(events))
	}

	// Helper to find event by type
	findEvent := func(evtType string) (event.Event, bool) {
		for _, e := range events {
			if e.Type == evtType {
				return e, true
			}
		}
		return event.Event{}, false
	}

	// Verify EventTaskStarted
	if ev, ok := findEvent(event.EventTaskStarted); !ok {
		t.Error("Expected EventTaskStarted not found")
	} else {
		if ev.Source != "runtime" {
			t.Errorf("Expected source 'runtime', got %q", ev.Source)
		}
		payload, ok := ev.Payload.(map[string]string)
		if !ok {
			t.Fatalf("Expected payload map[string]string, got %T", ev.Payload)
		}
		if payload["task_id"] != "task-1" {
			t.Errorf("Expected task_id 'task-1', got %q", payload["task_id"])
		}
		if payload["agent"] != "agent-backend" {
			t.Errorf("Expected agent 'agent-backend', got %q", payload["agent"])
		}
		if ev.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	}

	// Verify EventTaskCompleted
	if ev, ok := findEvent(event.EventTaskCompleted); !ok {
		t.Error("Expected EventTaskCompleted not found")
	} else {
		if ev.Source != "runtime" {
			t.Errorf("Expected source 'runtime', got %q", ev.Source)
		}
		payload, ok := ev.Payload.(map[string]any)
		if !ok {
			t.Fatalf("Expected payload map[string]any, got %T", ev.Payload)
		}
		if payload["task_id"] != "task-2" {
			t.Errorf("Expected task_id 'task-2', got %q", payload["task_id"])
		}
		if payload["output"] != "output-data" {
			t.Errorf("Expected output 'output-data', got %v", payload["output"])
		}
		if ev.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	}

	// Verify EventTaskFailed (with error)
	foundFailedErr := false
	foundFailedNil := false
	for _, ev := range events {
		if ev.Type == event.EventTaskFailed {
			if ev.Source != "runtime" {
				t.Errorf("Expected source 'runtime', got %q", ev.Source)
			}
			payload, ok := ev.Payload.(map[string]string)
			if !ok {
				t.Fatalf("Expected payload map[string]string, got %T", ev.Payload)
			}
			if payload["task_id"] == "task-3" {
				foundFailedErr = true
				if payload["error"] != "something went wrong" {
					t.Errorf("Expected error message 'something went wrong', got %q", payload["error"])
				}
			} else if payload["task_id"] == "task-4" {
				foundFailedNil = true
				if payload["error"] != "" {
					t.Errorf("Expected error message '', got %q", payload["error"])
				}
			}
			if ev.Timestamp.IsZero() {
				t.Error("Expected non-zero timestamp")
			}
		}
	}
	if !foundFailedErr {
		t.Error("Expected EventTaskFailed for task-3 not found")
	}
	if !foundFailedNil {
		t.Error("Expected EventTaskFailed for task-4 not found")
	}

	// Verify EventKernelStarted
	if ev, ok := findEvent(event.EventKernelStarted); !ok {
		t.Error("Expected EventKernelStarted not found")
	} else {
		if ev.Source != "kernel" {
			t.Errorf("Expected source 'kernel', got %q", ev.Source)
		}
		if ev.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	}

	// Verify EventKernelStopped
	if ev, ok := findEvent(event.EventKernelStopped); !ok {
		t.Error("Expected EventKernelStopped not found")
	} else {
		if ev.Source != "kernel" {
			t.Errorf("Expected source 'kernel', got %q", ev.Source)
		}
		if ev.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	}
}
