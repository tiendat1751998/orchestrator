package orchestrator

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSupervisor_RegisterDeregister(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := NewSupervisor(logger)

	taskID := "task-1"
	s.Register(taskID, 10*time.Second)

	s.mu.RLock()
	active, ok := s.activeTasks[taskID]
	s.mu.RUnlock()

	if !ok {
		t.Fatalf("expected task %s to be registered", taskID)
	}
	if active.ID != taskID {
		t.Errorf("expected active task ID %s, got %s", taskID, active.ID)
	}
	if active.Timeout != 10*time.Second {
		t.Errorf("expected active task timeout 10s, got %v", active.Timeout)
	}

	s.Deregister(taskID)

	s.mu.RLock()
	_, ok = s.activeTasks[taskID]
	s.mu.RUnlock()

	if ok {
		t.Fatalf("expected task %s to be deregistered", taskID)
	}
}

func TestSupervisor_CheckTimeouts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := NewSupervisor(logger)

	// Register 3 tasks:
	// 1. Task that has already timed out
	s.Register("expired", 10*time.Millisecond)
	// Modify StartedAt to simulate time passage
	s.mu.Lock()
	s.activeTasks["expired"].StartedAt = time.Now().Add(-20 * time.Millisecond)
	s.mu.Unlock()

	// 2. Task that has not timed out yet
	s.Register("not-expired", 10*time.Second)

	// 3. Task with zero timeout (no timeout)
	s.Register("no-timeout", 0)

	s.checkTimeouts()

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.activeTasks["expired"]; ok {
		t.Error("expected expired task to be removed by checkTimeouts")
	}
	if _, ok := s.activeTasks["not-expired"]; !ok {
		t.Error("expected not-expired task to still be active")
	}
	if _, ok := s.activeTasks["no-timeout"]; !ok {
		t.Error("expected no-timeout task to still be active")
	}
}

func TestSupervisor_StartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := NewSupervisor(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Register("expired", 5*time.Millisecond)
	s.mu.Lock()
	s.activeTasks["expired"].StartedAt = time.Now().Add(-10 * time.Millisecond)
	s.mu.Unlock()

	// Start with short interval for testing
	s.Start(ctx, 10*time.Millisecond)

	// Wait for scanner to run
	time.Sleep(30 * time.Millisecond)

	s.mu.RLock()
	_, ok := s.activeTasks["expired"]
	s.mu.RUnlock()

	if ok {
		t.Error("expected task to be cleaned up by background scanner loop")
	}
}

func TestSupervisor_Concurrency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := NewSupervisor(logger)

	// Spin up goroutines writing/deleting/checking tasks concurrently to verify there are no race conditions.
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx, 5*time.Millisecond)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := string(rune(id))
			s.Register(taskID, 2*time.Millisecond)
			time.Sleep(3 * time.Millisecond)
			s.Deregister(taskID)
		}(i)
	}

	wg.Wait()
}
