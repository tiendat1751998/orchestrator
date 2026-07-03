package eventbus

import (
	"sync"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

func TestSubscriptionLifecycle(t *testing.T) {
	handler := func(e event.Event) {}
	sub := newSubscription(42, "test.pattern", handler)

	if sub.id != 42 {
		t.Errorf("expected subscription ID 42, got %d", sub.id)
	}
	if sub.pattern != "test.pattern" {
		t.Errorf("expected pattern 'test.pattern', got %q", sub.pattern)
	}
	if !sub.isActive() {
		t.Error("expected new subscription to be active")
	}

	sub.deactivate()
	if sub.isActive() {
		t.Error("expected subscription to be inactive after deactivation")
	}
}

func TestSubscriberMapAddAndRemove(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	sub1 := sm.add("evt.1", handler)
	sub2 := sm.add("evt.2", handler)

	if sub1.id == 0 || sub2.id == 0 {
		t.Error("expected non-zero IDs assigned")
	}
	if sub1.id == sub2.id {
		t.Errorf("expected unique IDs, got identical ID: %d", sub1.id)
	}

	if sm.count() != 2 {
		t.Errorf("expected count to be 2, got %d", sm.count())
	}

	sm.remove(sub1.id)
	if sm.count() != 1 {
		t.Errorf("expected count to be 1 after remove, got %d", sm.count())
	}

	if sub1.isActive() {
		t.Error("expected removed subscription to be deactivated")
	}

	// Verify the remaining subscriber is sub2
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if len(sm.subs) != 1 || sm.subs[0].id != sub2.id {
		t.Errorf("expected only sub2 to remain in slice, got %v", sm.subs)
	}
}

func TestSubscriberMapMatching(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	sub1 := sm.add("evt.*", handler)
	_ = sm.add("other.*", handler)

	matched := sm.matching("evt.test")
	if len(matched) != 1 || matched[0].id != sub1.id {
		t.Errorf("expected to match sub1, got %d matched items", len(matched))
	}
}

func TestSubscriberMapConcurrency(t *testing.T) {
	sm := newSubscriberMap()
	handler := func(e event.Event) {}

	var wg sync.WaitGroup
	workers := 10
	addsPerWorker := 50

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var subs []*subscription
			for i := 0; i < addsPerWorker; i++ {
				sub := sm.add("test", handler)
				subs = append(subs, sub)
			}
			// Concurrently remove half of them
			for i := 0; i < addsPerWorker; i += 2 {
				sm.remove(subs[i].id)
			}
		}()
	}

	wg.Wait()

	expectedCount := workers * (addsPerWorker / 2)
	if sm.count() != expectedCount {
		t.Errorf("expected count %d, got %d", expectedCount, sm.count())
	}
}
