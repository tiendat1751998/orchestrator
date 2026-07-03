package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestPool_Basic(t *testing.T) {
	pool := NewPool(2, nil)
	if pool.MaxWorkers() != 2 {
		t.Errorf("expected max workers 2, got %d", pool.MaxWorkers())
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		err := pool.Submit(ctx, func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			t.Fatalf("failed to submit: %v", err)
		}
	}

	wg.Wait()
	pool.Wait()

	stats := pool.Stats()
	if stats.TotalSubmitted != 3 {
		t.Errorf("expected 3 submitted, got %d", stats.TotalSubmitted)
	}
	if stats.TotalCompleted != 3 {
		t.Errorf("expected 3 completed, got %d", stats.TotalCompleted)
	}
	if stats.ActiveWorkers != 0 {
		t.Errorf("expected 0 active workers, got %d", stats.ActiveWorkers)
	}
}

func TestPool_GuardMaxWorkers(t *testing.T) {
	pool := NewPool(0, nil)
	if pool.MaxWorkers() != 1 {
		t.Errorf("expected max workers to be guarded to 1, got %d", pool.MaxWorkers())
	}
}

func TestPool_ConcurrencyLimit(t *testing.T) {
	pool := NewPool(2, nil)
	ctx := context.Background()

	barrier := make(chan struct{})
	running := make(chan struct{}, 3)

	// Submit 3 tasks. Only 2 should run concurrently.
	for i := 0; i < 3; i++ {
		go func() {
			_ = pool.Submit(ctx, func(ctx context.Context) {
				running <- struct{}{}
				<-barrier
			})
		}()
	}

	// Give goroutines time to submit
	time.Sleep(10 * time.Millisecond)

	// Check stats / active workers
	// Since 3 were submitted, 2 should be active, and 1 should be queued (blocking on channel write)
	// We check active count directly
	active := pool.ActiveWorkers()
	if active > 2 {
		t.Errorf("expected at most 2 active workers, got %d", active)
	}

	// Let the workers finish
	close(barrier)
	pool.Wait()

	stats := pool.Stats()
	if stats.TotalSubmitted != 3 {
		t.Errorf("expected 3 submitted, got %d", stats.TotalSubmitted)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	pool := NewPool(1, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Fill the pool
	blocker := make(chan struct{})
	err := pool.Submit(ctx, func(ctx context.Context) {
		<-blocker
	})
	if err != nil {
		t.Fatalf("failed to submit first task: %v", err)
	}

	// Now try to submit another one, it should block.
	// We cancel context in a goroutine after a short delay.
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err = pool.Submit(ctx, func(ctx context.Context) {})
	if err == nil {
		t.Error("expected error on cancelled submit, got nil")
	} else if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Unblock first worker and wait
	close(blocker)
	pool.Wait()

	stats := pool.Stats()
	// 2 submitted (1 succeeded, 1 failed due to cancellation)
	if stats.TotalSubmitted != 2 {
		t.Errorf("expected 2 submitted, got %d", stats.TotalSubmitted)
	}
	if stats.TotalCompleted != 1 {
		t.Errorf("expected 1 completed, got %d", stats.TotalCompleted)
	}
}
