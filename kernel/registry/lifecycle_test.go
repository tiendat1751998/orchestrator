package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

type mockPlugin struct {
	name      string
	pType     plugin.Type
	inits     int
	starts    int
	stops     int
	initErr   error
	startErr  error
	stopErr   error
	healthErr error
	healthRep plugin.HealthReport
	stopFunc  func(context.Context) error
}

func (m *mockPlugin) Name() string      { return m.name }
func (m *mockPlugin) Type() plugin.Type { return m.pType }
func (m *mockPlugin) Version() string   { return "1.0.0" }
func (m *mockPlugin) Init(ctx context.Context, config map[string]any) error {
	m.inits++
	return m.initErr
}
func (m *mockPlugin) Start(ctx context.Context) error {
	m.starts++
	return m.startErr
}
func (m *mockPlugin) Stop(ctx context.Context) error {
	m.stops++
	if m.stopFunc != nil {
		return m.stopFunc(ctx)
	}
	return m.stopErr
}
func (m *mockPlugin) Health(ctx context.Context) (plugin.HealthReport, error) {
	return m.healthRep, m.healthErr
}

func TestInitAll(t *testing.T) {
	r := New(nil)
	p1 := &mockPlugin{name: "p1", pType: plugin.TypeSearch}
	p2 := &mockPlugin{name: "p2", pType: plugin.TypeSearch}

	if err := r.Register(p1); err != nil {
		t.Fatalf("failed to register p1: %v", err)
	}
	if err := r.Register(p2); err != nil {
		t.Fatalf("failed to register p2: %v", err)
	}

	configs := map[string]map[string]any{
		"p1": {"key": "val1"},
	}

	ctx := context.Background()
	if err := r.InitAll(ctx, configs); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}

	if p1.inits != 1 || p2.inits != 1 {
		t.Errorf("expected each plugin to be initialized once, got p1=%d, p2=%d", p1.inits, p2.inits)
	}
}

func TestStartAll_Rollback(t *testing.T) {
	r := New(nil)
	p1 := &mockPlugin{name: "p1", pType: plugin.TypeSearch}
	p2 := &mockPlugin{name: "p2", pType: plugin.TypeSearch}
	p3 := &mockPlugin{name: "p3", pType: plugin.TypeSearch, startErr: errors.New("start failed")}

	_ = r.Register(p1)
	_ = r.Register(p2)
	_ = r.Register(p3)

	ctx := context.Background()
	err := r.StartAll(ctx)
	if err == nil {
		t.Fatal("expected StartAll to fail")
	}

	if p1.starts != 1 || p2.starts != 1 || p3.starts != 1 {
		t.Errorf("expected all 3 to try to start: p1=%d, p2=%d, p3=%d", p1.starts, p2.starts, p3.starts)
	}

	// Rollback: p1 and p2 should have been stopped (in reverse order). p3 should NOT be stopped.
	if p1.stops != 1 || p2.stops != 1 {
		t.Errorf("expected already-started plugins to be stopped, got p1.stops=%d, p2.stops=%d", p1.stops, p2.stops)
	}
	if p3.stops != 0 {
		t.Errorf("expected failed plugin p3 NOT to be stopped, got p3.stops=%d", p3.stops)
	}
}

func TestStopAll_Reverse(t *testing.T) {
	r := New(nil)
	var stopOrder []string

	p1 := &mockPlugin{name: "p1", pType: plugin.TypeSearch}
	p2 := &mockPlugin{name: "p2", pType: plugin.TypeSearch}

	_ = r.Register(p1)
	_ = r.Register(p2)

	// Set stopFunc to track order
	p1.stopFunc = func(ctx context.Context) error {
		stopOrder = append(stopOrder, "p1")
		return nil
	}
	p2.stopFunc = func(ctx context.Context) error {
		stopOrder = append(stopOrder, "p2")
		return nil
	}

	ctx := context.Background()
	if err := r.StopAll(ctx); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if len(stopOrder) != 2 || stopOrder[0] != "p2" || stopOrder[1] != "p1" {
		t.Errorf("expected reverse stop order [p2, p1], got %v", stopOrder)
	}
}

func TestHealthCheckAll(t *testing.T) {
	r := New(nil)
	p1 := &mockPlugin{
		name:  "p1",
		pType: plugin.TypeSearch,
		healthRep: plugin.HealthReport{
			Status: plugin.HealthOK,
		},
	}
	p2 := &mockPlugin{
		name:      "p2",
		pType:     plugin.TypeSearch,
		healthErr: errors.New("check error"),
	}

	_ = r.Register(p1)
	_ = r.Register(p2)

	ctx := context.Background()
	reports := r.HealthCheckAll(ctx)

	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	if reports["p1"].Status != plugin.HealthOK {
		t.Errorf("expected p1 to be HealthOK, got %v", reports["p1"].Status)
	}

	if reports["p2"].Status != plugin.HealthDown {
		t.Errorf("expected p2 to be HealthDown, got %v", reports["p2"].Status)
	}
}
