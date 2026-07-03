# Micro-Task 1.40: Create contracts/plugin/health.go and Update Plugin Interface (Health Check Depth)

## Info
- **File to create**: `contracts/plugin/health.go`
- **File to update**: `contracts/plugin/plugin.go` (Ensure method signatures align)
- **Package**: `plugin`
- **Depends on**: 1.24 (plugin.go)
- **Time**: 20 min
- **Verify**: `go build ./contracts/plugin/...`

## Purpose
Upgrades health checking mechanisms from basic error strings to detailed structured reporting (`HealthReport`, `HealthStatus`). This permits plugins to return latency statistics, connection states, and dependent sub-component statuses (`Children`) to API query endpoints.

## EXACT code to create

### Part 1: Create `contracts/plugin/health.go`

```go
package plugin

import (
	"time"
)

// HealthStatus represents the high-level health state of a plugin.
type HealthStatus string

const (
	// HealthOK indicates the plugin is fully healthy and operational.
	HealthOK HealthStatus = "ok"

	// HealthDegraded indicates the plugin is running but with limited performance or minor errors.
	HealthDegraded HealthStatus = "degraded"

	// HealthDown indicates the plugin is non-functional.
	HealthDown HealthStatus = "down"
)

// HealthReport provides a structured, hierarchical report of plugin health.
// Suitable for JSON serialization in API endpoints.
type HealthReport struct {
	// Status is the overall health status of this plugin.
	Status HealthStatus `json:"status"`

	// Message describes the reason for non-healthy status.
	Message string `json:"message,omitempty"`

	// Details contains plugin-specific metric indicators (e.g. queue depth, latency).
	Details map[string]any `json:"details,omitempty"`

	// Children contains reports from internal dependencies.
	Children map[string]HealthReport `json:"children,omitempty"`

	// Timestamp is when the health check was performed.
	Timestamp time.Time `json:"timestamp"`

	// Duration measures how long the health check took to execute.
	Duration time.Duration `json:"duration"`
}

// IsHealthy returns true if the status is OK or Degraded (still operational).
func (hr HealthReport) IsHealthy() bool {
	return hr.Status == HealthOK || hr.Status == HealthDegraded
}
```

---

### Part 2: Update `contracts/plugin/plugin.go`

Ensure the `Health` method in [contracts/plugin/plugin.go](file:///d:/project/orchestrator/contracts/plugin/plugin.go) has the following signature:

```go
	// Health checks if the plugin is functioning correctly and returns a detailed report.
	//
	// Parameters:
	//   - ctx: for timeout enforcement. Checkers should abort and return error on cancellation.
	//
	// Returns:
	//   - HealthReport: structured health status.
	//   - error: system level failure during the health check itself (e.g. context timeout).
	//            If the plugin is simply unhealthy (down), return (HealthReport{Status: HealthDown}, nil)
	//            rather than a non-nil error.
	Health(ctx context.Context) (HealthReport, error)
```

## Rules
1. **Unhealthy vs System Error**: System `error` (the second returned value) is reserved for checker failures (such as context timeouts or memory faults). If a plugin's target API endpoint is down, it must return `(HealthReport{Status: HealthDown, Message: "offline"}, nil)`.
2. **Periodic Execution safety**: Health checker routines must respect the passed context deadline and execute fast (typically completing in under 5 seconds).
3. **Telemetry Tracking**: Reports must record the absolute start time and duration details (`time.Since(start)`) to identify slow running checks.

## ⚠️ Pitfalls

### Pitfall 1: Returning Go errors for simple component downtime
```go
func (p *MyPlugin) Health(ctx context.Context) (HealthReport, error) {
    if err := p.pingAPI(); err != nil {
        return HealthReport{
            Status:    HealthDown,
            Message:   "API offline: " + err.Error(),
            Timestamp: time.Now(),
        }, nil // Correct: checker completed successfully, reporting unhealthy state.
    }
}
```
Always return `error = nil` if the health checker successfully ran and identified a degraded or offline state. Only return Go error if the check execution itself fails (e.g. deadline exceeded).

### Pitfall 2: Omiting Timestamp or Duration metrics
Omitting checking timestamps and durations makes tracking transient latencies in telemetry dashboards difficult. Always initialize `Timestamp` with `time.Now()` and compute `Duration` accurately.

## Verify
```bash
go build ./contracts/plugin/...
```

## Checklist
- [ ] File `contracts/plugin/health.go` exists
- [ ] Package: `plugin`
- [ ] `HealthStatus` contains OK, Degraded, and Down status values
- [ ] `HealthReport` struct contains Status, Message, Details, Children, Timestamp, and Duration fields
- [ ] `IsHealthy()` matches both OK and Degraded states
- [ ] `Plugin` interface `Health` signature updated to `(HealthReport, error)`
- [ ] `go build ./contracts/plugin/...` passes
