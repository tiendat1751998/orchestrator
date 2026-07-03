# Telemetry Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a thread-safe telemetry metrics registry and collection structures in the Go kernel.

**Architecture:** Implements `Counter`, `Gauge`, and `Histogram` interfaces and concrete thread-safe implementations (`memCounter`, `memGauge`, `memHistogram`) utilizing read-write mutexes (`sync.RWMutex`) to minimize lock contention. A central `Registry` handles registration and exports snapshots.

**Tech Stack:** Go 1.26 (Standard Library only - `fmt`, `sync`).

## Global Constraints

- Implement in `kernel/metrics/metrics.go` under package `metrics`.
- Test in `kernel/metrics/metrics_test.go` under package `metrics_test`.
- Strictly adhere to layer boundaries (only stdlib allowed).
- All struct initializations must use named fields.
- Keep complexity low and respect all constraints in `.agents/rules/ai_rules.md`.

---

### Task 1: Implement Telemetry Metrics structures

**Files:**
- Create: `kernel/metrics/metrics.go`

**Interfaces:**
- Consumes: Stdlib only (`fmt`, `sync`)
- Produces: `Counter` (interface), `Gauge` (interface), `Histogram` (interface), `Registry` (struct), `NewRegistry` (function), `HistogramSnapshot` (struct)

- [ ] **Step 1: Write the metrics.go implementation**

Create `kernel/metrics/metrics.go` containing:
```go
// Package metrics implements an in-memory telemetry data collector.
// It tracks runtime statistics (counters, gauges, latency histograms)
// in a thread-safe, lock-free or low-lock manner.
package metrics

import (
	"fmt"
	"sync"
)

// =============================================================================
// Interfaces
// =============================================================================

// Counter tracks values that only increase (e.g. total tasks executed, errors).
type Counter interface {
	Inc()
	Add(val float64)
}

// Gauge tracks a snapshot value that can go up and down (e.g. active worker count).
type Gauge interface {
	Set(val float64)
	Add(val float64)
}

// Histogram measures the distribution of values, typically durations (latency).
type Histogram interface {
	Observe(val float64)
}

// =============================================================================
// In-Memory Collector Implementation
// =============================================================================

// Registry acts as the central storage for all collected metrics.
// Thread-safe.
type Registry struct {
	mu         sync.RWMutex
	counters   map[string]*memCounter
	gauges     map[string]*memGauge
	histograms map[string]*memHistogram
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		counters:   make(map[string]*memCounter),
		gauges:     make(map[string]*memGauge),
		histograms: make(map[string]*memHistogram),
	}
}

// Counter returns or registers a counter by name.
func (r *Registry) Counter(name string) Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c, exists := r.counters[name]; exists {
		return c
	}
	c := &memCounter{}
	r.counters[name] = c
	return c
}

// Gauge returns or registers a gauge by name.
func (r *Registry) Gauge(name string) Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	if g, exists := r.gauges[name]; exists {
		return g
	}
	g := &memGauge{}
	r.gauges[name] = g
	return g
}

// Histogram returns or registers a histogram by name.
func (r *Registry) Histogram(name string) Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	if h, exists := r.histograms[name]; exists {
		return h
	}
	h := &memHistogram{}
	r.histograms[name] = h
	return h
}

// =============================================================================
// Concrete Memory Implementations
// =============================================================================

type memCounter struct {
	mu  sync.RWMutex
	val float64
}

func (c *memCounter) Inc() {
	c.Add(1.0)
}

func (c *memCounter) Add(val float64) {
	if val < 0 {
		return // Counters can only increase
	}
	c.mu.Lock()
	c.val += val
	c.mu.Unlock()
}

func (c *memCounter) Value() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type memGauge struct {
	mu  sync.RWMutex
	val float64
}

func (g *memGauge) Set(val float64) {
	g.mu.Lock()
	g.val = val
	g.mu.Unlock()
}

func (g *memGauge) Add(val float64) {
	g.mu.Lock()
	g.val += val
	g.mu.Unlock()
}

func (g *memGauge) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.val
}

type memHistogram struct {
	mu     sync.RWMutex
	values []float64
}

func (h *memHistogram) Observe(val float64) {
	h.mu.Lock()
	h.values = append(h.values, val)
	h.mu.Unlock()
}

// Snapshot calculates statistical summary of the histogram data.
func (h *memHistogram) Snapshot() HistogramSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.values) == 0 {
		return HistogramSnapshot{}
	}

	var sum float64
	min := h.values[0]
	max := h.values[0]

	for _, v := range h.values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	return HistogramSnapshot{
		Count: len(h.values),
		Sum:   sum,
		Mean:  sum / float64(len(h.values)),
		Min:   min,
		Max:   max,
	}
}

// HistogramSnapshot contains computed statistical aggregations.
type HistogramSnapshot struct {
	Count int     `json:"count"`
	Sum   float64 `json:"sum"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

// =============================================================================
// Export Helpers
// =============================================================================

// Snapshot captures all metrics stored in the Registry.
func (r *Registry) Snapshot() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snap := make(map[string]any)

	for k, c := range r.counters {
		snap[fmt.Sprintf("counter.%s", k)] = c.Value()
	}
	for k, g := range r.gauges {
		snap[fmt.Sprintf("gauge.%s", k)] = g.Value()
	}
	for k, h := range r.histograms {
		snap[fmt.Sprintf("histogram.%s", k)] = h.Snapshot()
	}

	return snap
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./kernel/metrics/...`
Expected: Success with no errors.

---

### Task 2: Create unit tests for metrics

**Files:**
- Create: `kernel/metrics/metrics_test.go`

**Interfaces:**
- Consumes: `kernel/metrics`
- Produces: Test functions for `Counter`, `Gauge`, `Histogram`, `Registry`, and concurrent benchmarks/snapshots.

- [ ] **Step 1: Write test code**

Create `kernel/metrics/metrics_test.go` containing:
```go
package metrics_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"tiendat1751998/orchestrator/kernel/metrics"
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
```

- [ ] **Step 2: Run verification**

Run: `go test -v -race ./kernel/metrics/...`
Expected: PASS with zero warnings or errors.
