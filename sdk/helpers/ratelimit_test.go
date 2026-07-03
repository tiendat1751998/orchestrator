package helpers

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(2, 50*time.Millisecond)

	// Consume 2 tokens
	if !tb.Allow() {
		t.Error("expected first token to be allowed")
	}
	if !tb.Allow() {
		t.Error("expected second token to be allowed")
	}
	// Third token should be denied
	if tb.Allow() {
		t.Error("expected third token to be denied")
	}

	// Wait for refill
	time.Sleep(60 * time.Millisecond)
	if !tb.Allow() {
		t.Error("expected token to be allowed after refill")
	}
}

func TestTokenBucket_Wait(t *testing.T) {
	tb := NewTokenBucket(1, 50*time.Millisecond)

	// Consume capacity
	if err := tb.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	start := time.Now()
	// Next wait should block until refilled
	if err := tb.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	duration := time.Since(start)

	if duration < 40*time.Millisecond {
		t.Errorf("expected wait to block for at least ~50ms, got %v", duration)
	}
}

func TestTokenBucket_WaitCancel(t *testing.T) {
	tb := NewTokenBucket(1, 100*time.Millisecond)

	// Consume capacity
	if err := tb.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("expected deadline exceeded or canceled error, got %v", err)
	}
}
