package metrics_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/kernel/metrics"
)

func TestCounter(t *testing.T) {
	reg := metrics.NewRegistry()
	c := reg.Counter("test_counter")

	c.Inc()
	c.Add(4.5)
	c.Add(-2.0) // Should be ignored

	snap := reg.Snapshot()
	val, ok := snap["counter.test_counter"]
	if !ok {
		t.Fatal("counter.test_counter not found in snapshot")
	}
	fVal, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64 value, got %T", val)
	}
	if fVal != 5.5 {
		t.Errorf("expected counter value to be 5.5, got %f", fVal)
	}
}

func TestGauge(t *testing.T) {
	reg := metrics.NewRegistry()
	g := reg.Gauge("test_gauge")

	g.Set(10.0)
	g.Add(5.0)
	g.Add(-3.0)

	snap := reg.Snapshot()
	val, ok := snap["gauge.test_gauge"]
	if !ok {
		t.Fatal("gauge.test_gauge not found in snapshot")
	}
	fVal, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64 value, got %T", val)
	}
	if fVal != 12.0 {
		t.Errorf("expected gauge value to be 12.0, got %f", fVal)
	}
}

func TestHistogram(t *testing.T) {
	reg := metrics.NewRegistry()
	h := reg.Histogram("test_histogram")

	h.Observe(2.0)
	h.Observe(4.0)
	h.Observe(6.0)

	snap := reg.Snapshot()
	val, ok := snap["histogram.test_histogram"]
	if !ok {
		t.Fatal("histogram.test_histogram not found in snapshot")
	}
	hSnap, ok := val.(metrics.HistogramSnapshot)
	if !ok {
		t.Fatalf("expected HistogramSnapshot, got %T", val)
	}

	if hSnap.Count != 3 {
		t.Errorf("expected Count to be 3, got %d", hSnap.Count)
	}
	if hSnap.Sum != 12.0 {
		t.Errorf("expected Sum to be 12.0, got %f", hSnap.Sum)
	}
	if hSnap.Mean != 4.0 {
		t.Errorf("expected Mean to be 4.0, got %f", hSnap.Mean)
	}
	if hSnap.Min != 2.0 {
		t.Errorf("expected Min to be 2.0, got %f", hSnap.Min)
	}
	if hSnap.Max != 6.0 {
		t.Errorf("expected Max to be 6.0, got %f", hSnap.Max)
	}
}

func TestHistogramEmpty(t *testing.T) {
	reg := metrics.NewRegistry()
	_ = reg.Histogram("empty_histogram")

	snap := reg.Snapshot()
	val, ok := snap["histogram.empty_histogram"]
	if !ok {
		t.Fatal("histogram.empty_histogram not found in snapshot")
	}
	hSnap, ok := val.(metrics.HistogramSnapshot)
	if !ok {
		t.Fatalf("expected HistogramSnapshot, got %T", val)
	}

	if hSnap.Count != 0 {
		t.Errorf("expected Count to be 0, got %d", hSnap.Count)
	}
	if hSnap.Sum != 0.0 {
		t.Errorf("expected Sum to be 0.0, got %f", hSnap.Sum)
	}
	if hSnap.Mean != 0.0 {
		t.Errorf("expected Mean to be 0.0, got %f", hSnap.Mean)
	}
	if hSnap.Min != 0.0 {
		t.Errorf("expected Min to be 0.0, got %f", hSnap.Min)
	}
	if hSnap.Max != 0.0 {
		t.Errorf("expected Max to be 0.0, got %f", hSnap.Max)
	}
}

func TestRegistryConcurrencyAndHeavyLoad(t *testing.T) {
	reg := metrics.NewRegistry()
	const (
		goroutines = 50
		iterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(time.Now().UnixNano() + int64(id))))

			for j := 0; j < iterations; j++ {
				// Access registry metrics concurrently
				counterName := fmt.Sprintf("counter_%d", rng.Intn(10))
				gaugeName := fmt.Sprintf("gauge_%d", rng.Intn(10))
				histName := fmt.Sprintf("histogram_%d", rng.Intn(10))

				c := reg.Counter(counterName)
				g := reg.Gauge(gaugeName)
				h := reg.Histogram(histName)

				c.Inc()
				c.Add(rng.Float64() * 10)
				g.Set(rng.Float64() * 100)
				g.Add(rng.Float64() * 5)
				h.Observe(rng.Float64() * 10)

				// Periodically take snapshot
				if rng.Intn(100) == 0 {
					_ = reg.Snapshot()
				}
			}
		}(i)
	}

	wg.Wait()

	// Final snapshot
	snap := reg.Snapshot()
	if len(snap) == 0 {
		t.Error("expected snapshot to have metrics, got 0")
	}
}
