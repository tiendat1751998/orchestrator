package plugin

import (
	"testing"
)

func TestIsHealthy(t *testing.T) {
	tests := []struct {
		status HealthStatus
		want   bool
	}{
		{HealthOK, true},
		{HealthDegraded, true},
		{HealthDown, false},
		{HealthStatus("unknown"), false},
	}

	for _, tt := range tests {
		hr := HealthReport{Status: tt.status}
		if got := hr.IsHealthy(); got != tt.want {
			t.Errorf("HealthReport{Status: %q}.IsHealthy() = %v, want %v", tt.status, got, tt.want)
		}
	}
}
