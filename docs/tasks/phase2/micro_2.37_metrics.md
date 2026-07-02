# Micro-Task 2.37: Tạo kernel/metrics (Hệ thống Metrics & Observability)

## Thông tin
- **File tạo**: `kernel/metrics/metrics.go`
- **Package**: `metrics`
- **Dependencies trước**: 2.32 (kernel.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/metrics/...`

## Purpose
Triển khai hệ thống đo đạc (Metrics Collector) để cung cấp khả năng quan sát (Observability) cho orchestrator. Hệ thống định nghĩa các interfaces đo đạc chuẩn (Counter, Gauge, Histogram) và cung cấp một bộ thu thập in-memory hiệu năng cao, làm nền tảng để đẩy dữ liệu lên Prometheus/OpenTelemetry trong các pha sau.

## EXACT code to create

```go
// Package metrics implements an in-memory telemetry data collector.
// It tracks runtime statistics (counters, gauges, latency histograms)
// in a thread-safe, lock-free or low-lock manner.
package metrics

import (
	"fmt"
	"sync"
	"time"
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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Mutex Lock quá rộng trong Registry
```go
// ❌ SAI:
func (r *Registry) Snapshot() map[string]any {
    r.mu.Lock() // Writer lock trong khi chỉ đọc -> khóa tất cả tác vụ ghi metrics khác
    defer r.mu.Unlock()
    ...
}

// ✅ ĐÚNG:
func (r *Registry) Snapshot() map[string]any {
    r.mu.RLock() // Dùng Reader lock
    defer r.mu.RUnlock()
    ...
}
```
Việc ghi metrics xảy ra cực kỳ thường xuyên (mỗi bước chạy của agent/event bus). Dùng RLock cho hoạt động đọc dữ liệu snapshot giúp tránh thắt nút cổ chai (performance bottleneck).

### Pitfall 2: Metric name collisions
Đặt tên metric không nhất quán (CamelCase vs snake_case). 
Hệ thống này quy định chuẩn snake_case cho toàn bộ metric names, ví dụ: `active_workers`, `tasks_total`, `task_duration_seconds`.

## Checklist
- [ ] File `kernel/metrics/metrics.go` tồn tại
- [ ] Định nghĩa 3 interfaces đo đạc: Counter, Gauge, Histogram
- [ ] Triển khai `Registry` lưu trữ in-memory dùng mutex bảo vệ
- [ ] `memCounter` không cho giảm (`Add` giá trị âm bị bỏ qua)
- [ ] `memHistogram` hỗ trợ tính toán Snapshot (Mean, Min, Max, Count, Sum)
- [ ] RWMutex được sử dụng tối ưu (RLock cho thao tác đọc/Value/Snapshot)
- [ ] Hàm export `Snapshot()` trả về map chứa toàn bộ trạng thái hiện tại
- [ ] `go build ./kernel/metrics/...` không lỗi
