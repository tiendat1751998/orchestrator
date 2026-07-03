package orchestrator

import (
	"sync"
	"testing"
	"time"
)

func TestFeedbackCollector_RecordAndGet(t *testing.T) {
	fc := NewFeedbackCollector()

	// Record success and failure
	fc.RecordSuccess("agent-a", 100, 500*time.Millisecond)
	fc.RecordSuccess("agent-a", 150, 450*time.Millisecond)
	fc.RecordFailure("agent-a")
	fc.RecordFailure("agent-b")

	metrics := fc.GetMetrics()

	// Check agent-a metrics
	ma, ok := metrics["agent-a"]
	if !ok {
		t.Fatal("expected metrics for agent-a")
	}
	if ma.SuccessCount != 2 {
		t.Errorf("expected agent-a success count to be 2, got %d", ma.SuccessCount)
	}
	if ma.FailureCount != 1 {
		t.Errorf("expected agent-a failure count to be 1, got %d", ma.FailureCount)
	}
	if ma.TotalTokens != 250 {
		t.Errorf("expected agent-a total tokens to be 250, got %d", ma.TotalTokens)
	}
	if ma.TotalDuration != 950*time.Millisecond {
		t.Errorf("expected agent-a total duration to be 950ms, got %s", ma.TotalDuration)
	}

	// Check agent-b metrics
	mb, ok := metrics["agent-b"]
	if !ok {
		t.Fatal("expected metrics for agent-b")
	}
	if mb.SuccessCount != 0 {
		t.Errorf("expected agent-b success count to be 0, got %d", mb.SuccessCount)
	}
	if mb.FailureCount != 1 {
		t.Errorf("expected agent-b failure count to be 1, got %d", mb.FailureCount)
	}
}

func TestFeedbackCollector_DeepCopy(t *testing.T) {
	fc := NewFeedbackCollector()
	fc.RecordSuccess("agent-a", 100, 100*time.Millisecond)

	m1 := fc.GetMetrics()
	fc.RecordSuccess("agent-a", 100, 100*time.Millisecond)
	m2 := fc.GetMetrics()

	if m1["agent-a"].SuccessCount != 1 {
		t.Errorf("expected m1 success count to be 1, got %d", m1["agent-a"].SuccessCount)
	}
	if m2["agent-a"].SuccessCount != 2 {
		t.Errorf("expected m2 success count to be 2, got %d", m2["agent-a"].SuccessCount)
	}
}

func TestFeedbackCollector_Concurrent(t *testing.T) {
	fc := NewFeedbackCollector()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fc.RecordSuccess("agent-a", 10, 1*time.Millisecond)
			fc.RecordFailure("agent-a")
			_ = fc.GetMetrics()
		}()
	}

	wg.Wait()
	metrics := fc.GetMetrics()
	ma := metrics["agent-a"]

	if ma.SuccessCount != 100 {
		t.Errorf("expected success count 100, got %d", ma.SuccessCount)
	}
	if ma.FailureCount != 100 {
		t.Errorf("expected failure count 100, got %d", ma.FailureCount)
	}
}
