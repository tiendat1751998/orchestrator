package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

type mockShutdownable struct {
	mu         sync.Mutex
	stopCalled bool
	stopCtx    context.Context
	stopErr    error
}

func (m *mockShutdownable) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalled = true
	m.stopCtx = ctx
	return m.stopErr
}

func (m *mockShutdownable) IsStopCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopCalled
}

func (m *mockShutdownable) GetStopCtx() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopCtx
}

func TestWaitForShutdown_CancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	target := &mockShutdownable{}
	logger := slog.Default()

	// Cancel context to trigger shutdown
	cancel()

	// Call WaitForShutdown. It should return quickly because context is already cancelled.
	WaitForShutdown(ctx, target, 100*time.Millisecond, logger)

	if !target.IsStopCalled() {
		t.Fatal("expected Stop to be called")
	}

	stopCtx := target.GetStopCtx()
	if stopCtx == nil {
		t.Fatal("expected stopCtx to be set")
	}

	// The context passed to Stop should have a deadline configured
	_, ok := stopCtx.Deadline()
	if !ok {
		t.Error("expected context to have a deadline")
	}
}

func TestWaitForShutdown_Signal(t *testing.T) {
	// Use a short timeout context as a fallback on platforms where sending OS signals fails or is unsupported.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	target := &mockShutdownable{}
	logger := slog.Default()

	go func() {
		// Wait a tiny bit for signal.Notify to be registered
		time.Sleep(50 * time.Millisecond)
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			// This might return an error on Windows, but that is fine since the context timeout fallback will trigger.
			_ = p.Signal(os.Interrupt)
		}
	}()

	WaitForShutdown(ctx, target, 100*time.Millisecond, logger)

	if !target.IsStopCalled() {
		t.Fatal("expected Stop to be called")
	}
}

func TestWaitForShutdownWithDefaults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	target := &mockShutdownable{}
	logger := slog.Default()

	cancel()

	// WaitForShutdownWithDefaults has 30s timeout, but since context is cancelled, it should exit instantly.
	WaitForShutdownWithDefaults(ctx, target, logger)

	if !target.IsStopCalled() {
		t.Fatal("expected Stop to be called")
	}
}
