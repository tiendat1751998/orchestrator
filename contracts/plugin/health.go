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
