package eventbus

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

func TestSafeHandler(t *testing.T) {
	// Test normal handler execution
	called := false
	normalHandler := func(e event.Event) {
		called = true
	}
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	wrappedNormal := safeHandler(normalHandler, nil, logger)
	wrappedNormal(event.Event{Type: "test.event"})

	if !called {
		t.Error("expected normal handler to be called")
	}
	if buf.Len() > 0 {
		t.Errorf("expected no log message for normal execution, got: %s", buf.String())
	}

	// Test panic recovery with nil dlq
	panickingHandler := func(e event.Event) {
		panic("something went wrong")
	}
	wrappedPanic := safeHandler(panickingHandler, nil, logger)

	// This should not crash
	wrappedPanic(event.Event{Type: "panic.event", Source: "test"})

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected log output for panicked handler, but got empty")
	}

	// Test panic recovery with non-nil dlq
	dlq := NewDeadLetterQueue(10)
	wrappedPanicWithDLQ := safeHandler(panickingHandler, dlq, logger)
	wrappedPanicWithDLQ(event.Event{Type: "panic.event.dlq", Source: "test-dlq"})

	if dlq.Len() != 1 {
		t.Errorf("expected 1 entry in dlq, got %d", dlq.Len())
	}
	entries := dlq.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in dlq entries, got %d", len(entries))
	}
	if entries[0].Event.Type != "panic.event.dlq" {
		t.Errorf("expected event type panic.event.dlq, got %s", entries[0].Event.Type)
	}
	if entries[0].Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %s", entries[0].Error)
	}
}

func TestMakeUnsubscribe(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	sub := sm.add("test.pattern", handler)
	if sm.count() != 1 {
		t.Errorf("expected subscriber map to have 1 subscriber, got %d", sm.count())
	}

	unsub := makeUnsubscribe(sm, sub)

	// First call should deactivate and remove
	unsub()
	if sm.count() != 0 {
		t.Errorf("expected subscriber map to be empty, got %d", sm.count())
	}
	if sub.isActive() {
		t.Error("expected subscription to be deactivated")
	}

	// Second call should be a safe no-op (idempotency check)
	unsub()
	if sm.count() != 0 {
		t.Errorf("expected subscriber map to remain empty, got %d", sm.count())
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		pattern string
		isValid bool
	}{
		// Valid cases
		{"*", true},
		{"task.started", true},
		{"task.*", true},
		{"*.started", true},
		{"a.b.c", true},

		// Invalid cases
		{"", false},
		{".", false},
		{".started", false},
		{"task.", false},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			err := validatePattern(tc.pattern)
			if tc.isValid && err != nil {
				t.Errorf("expected pattern %q to be valid, got error: %v", tc.pattern, err)
			}
			if !tc.isValid && err == nil {
				t.Errorf("expected pattern %q to be invalid, but got no error", tc.pattern)
			}
		})
	}
}
