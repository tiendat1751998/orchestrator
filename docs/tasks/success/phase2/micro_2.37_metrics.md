# Micro-Task 2.37: Create kernel/metrics (Telemetry Metrics & Observability)

## Info
- **File created**: `kernel/metrics/metrics.go`
- **Package**: `metrics`
- **Depends on**: 2.32 (kernel.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/metrics/...`

## Purpose
Implements a lightweight in-memory metrics and telemetry subsystem (`Counter`, `Gauge`, `Histogram`, `Registry`, and implementations) that collects orchestrator run indicators (such as execution durations, active task limits, and error counts) in a thread-safe manner.

## EXACT code to create

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

## Rules
1. **Lock Optimization**: Optimize locking for telemetry registrations and snapshot prints. Use mutual exclusion read locks (`RLock()`) for read operations (such as fetching values or taking snaps) to prevent performance bottlenecks.
2. **Negative Values Counter Guard**: Counters must only increase. Ignore any requests to add negative increments (`val < 0`) to counter structures.
3. **Structured snake_case naming**: Naming schemas for all recorded metric properties must follow `snake_case` guidelines (e.g. `active_workers`, `tasks_total`, `task_duration_seconds`).

## ⚠️ Pitfalls

### Pitfall 1: Using write locks during read or export routines
Acquiring exclusive write locks (`Lock()`) inside export wrappers like `Snapshot()` blocks concurrent telemetry writers. Always use read locks (`RLock()`) for read-only operations.

### Pitfall 2: Permitting counters to decrement
Failing to filter negative increments inside `Add()` methods violates standard telemetry design guidelines, corrupting the metrics dashboard history. Always guard `Add` methods against negative parameters.

## Verify
```bash
go build ./kernel/metrics/...
```

## Checklist
- [ ] File `kernel/metrics/metrics.go` exists
- [ ] Package: `metrics`
- [ ] Interfaces declare `Counter`, `Gauge`, and `Histogram` definitions
- [ ] `Registry` protects concurrent writes using mutex locks
- [ ] `memCounter` blocks negative inputs from modifying data
- [ ] `memHistogram` computes Min, Max, Mean, and Count summaries
- [ ] Read queries use `RLock` read locks
- [ ] `Snapshot` exports all registered metrics under structured keys
- [ ] `go build ./kernel/metrics/...` passes
