package eventbus_test

import (
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
)

func TestDLQ_BasicAddAndRetrieve(t *testing.T) {
	q := eventbus.NewDeadLetterQueue(3)
	if q.Len() != 0 {
		t.Errorf("expected length 0, got %d", q.Len())
	}

	evt := event.Event{ID: "e1", Type: "test.event"}
	q.Add(evt, "panic error")

	if q.Len() != 1 {
		t.Errorf("expected length 1, got %d", q.Len())
	}

	entries := q.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Event.ID != "e1" {
		t.Errorf("expected event ID e1, got %s", entries[0].Event.ID)
	}
	if entries[0].Error != "panic error" {
		t.Errorf("expected error 'panic error', got %s", entries[0].Error)
	}
}

func TestDLQ_CircularBufferAndSorting(t *testing.T) {
	q := eventbus.NewDeadLetterQueue(3)

	q.Add(event.Event{ID: "e1"}, "err1")
	q.Add(event.Event{ID: "e2"}, "err2")
	q.Add(event.Event{ID: "e3"}, "err3")

	// At this point q is full: entries are e1, e2, e3
	entries := q.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Event.ID != "e1" || entries[1].Event.ID != "e2" || entries[2].Event.ID != "e3" {
		t.Errorf("wrong initial order: %v, %v, %v", entries[0].Event.ID, entries[1].Event.ID, entries[2].Event.ID)
	}

	// Add e4: should overwrite e1
	q.Add(event.Event{ID: "e4"}, "err4")

	if q.Len() != 3 {
		t.Errorf("expected length 3 after overflow, got %d", q.Len())
	}

	entries = q.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", len(entries))
	}

	// Oldest first: e2, e3, e4
	if entries[0].Event.ID != "e2" || entries[1].Event.ID != "e3" || entries[2].Event.ID != "e4" {
		t.Errorf("wrong chronological order: got %v, %v, %v; want e2, e3, e4", entries[0].Event.ID, entries[1].Event.ID, entries[2].Event.ID)
	}

	q.Clear()
	if q.Len() != 0 {
		t.Errorf("expected 0 after Clear, got %d", q.Len())
	}
}
