package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

type mockHealthCheckable struct {
	fn func(ctx context.Context) (cplugin.HealthReport, error)
}

func (m *mockHealthCheckable) Health(ctx context.Context) (cplugin.HealthReport, error) {
	return m.fn(ctx)
}

func TestHealthAggregator(t *testing.T) {
	agg := resilience.NewHealthAggregator()

	// Mock component 1: Success
	comp1 := &mockHealthCheckable{
		fn: func(ctx context.Context) (cplugin.HealthReport, error) {
			return cplugin.HealthReport{
				Status:    cplugin.HealthOK,
				Timestamp: time.Now(),
			}, nil
		},
	}

	// Mock component 2: Error
	comp2 := &mockHealthCheckable{
		fn: func(ctx context.Context) (cplugin.HealthReport, error) {
			return cplugin.HealthReport{}, errors.New("connection failed")
		},
	}

	agg.Register("comp1", comp1)
	agg.Register("comp2", comp2)

	ctx := context.Background()
	results := agg.CheckAll(ctx)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	r1, ok := results["comp1"]
	if !ok {
		t.Error("expected comp1 in results")
	} else if r1.Status != cplugin.HealthOK {
		t.Errorf("expected comp1 to be OK, got %v", r1.Status)
	}

	r2, ok := results["comp2"]
	if !ok {
		t.Error("expected comp2 in results")
	} else {
		if r2.Status != cplugin.HealthDown {
			t.Errorf("expected comp2 to be Down, got %v", r2.Status)
		}
		if r2.Message != "connection failed" {
			t.Errorf("expected comp2 error message to be 'connection failed', got %v", r2.Message)
		}
	}
}
