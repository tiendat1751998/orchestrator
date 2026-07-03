package session

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/adapter"
)

func TestSessionManager(t *testing.T) {
	binary := "echo"
	if runtime.GOOS == "windows" {
		binary = "cmd"
	}

	sm := NewSessionManager(binary, 2, 100*time.Millisecond)
	defer sm.Stop()

	ctx := context.Background()

	// 1. Create a session
	s1, err := sm.GetOrCreate(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to get or create session-1: %v", err)
	}
	if s1.ID != "session-1" {
		t.Errorf("expected session ID 'session-1', got '%s'", s1.ID)
	}

	// 2. Retrieve the same session
	s1Cached, err := sm.GetOrCreate(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to retrieve cached session-1: %v", err)
	}
	if s1Cached != s1 {
		t.Error("expected to retrieve the same session instance")
	}

	// 3. Create a second session
	s2, err := sm.GetOrCreate(ctx, "session-2")
	if err != nil {
		t.Fatalf("failed to get or create session-2: %v", err)
	}
	if s2.ID != "session-2" {
		t.Errorf("expected session ID 'session-2', got '%s'", s2.ID)
	}

	// 4. Try to create a third session (limit is 2)
	_, err = sm.GetOrCreate(ctx, "session-3")
	if err == nil {
		t.Error("expected error when exceeding maxSessions, got nil")
	}

	// 5. Close session-1 and check if we can create a new session
	err = sm.Close("session-1")
	if err != nil {
		t.Fatalf("failed to close session-1: %v", err)
	}

	s3, err := sm.GetOrCreate(ctx, "session-3")
	if err != nil {
		t.Fatalf("failed to create session-3 after closing session-1: %v", err)
	}
	if s3.ID != "session-3" {
		t.Errorf("expected session ID 'session-3', got '%s'", s3.ID)
	}
}

func TestSessionManager_EmptySessionID(t *testing.T) {
	binary := "echo"
	if runtime.GOOS == "windows" {
		binary = "cmd"
	}

	sm := NewSessionManager(binary, 2, 5*time.Minute)
	defer sm.Stop()

	_, err := sm.GetOrCreate(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty session ID, got nil")
	}
}

func TestStartHeartbeat_Restart(t *testing.T) {
	binary := "echo"
	if runtime.GOOS == "windows" {
		binary = "cmd"
	}

	sm := NewSessionManager(binary, 2, 5*time.Minute)
	defer sm.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := sm.GetOrCreate(ctx, "session-h")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Start heartbeat with a very short interval
	sm.StartHeartbeat(ctx, "session-h", 10*time.Millisecond, nil)

	// Wait a bit to ensure it ticks
	time.Sleep(30 * time.Millisecond)

	// Verify session still exists
	sm.mu.RLock()
	_, exists := sm.sessions["session-h"]
	sm.mu.RUnlock()
	if !exists {
		t.Fatal("session should exist")
	}

	// Force stop the adapter to simulate a crash (without removing from manager)
	err = s.Adapter.Stop()
	if err != nil {
		t.Fatalf("failed to stop adapter: %v", err)
	}

	// Heartbeat should detect the crash and restart the process successfully
	// since binary is valid. Let's wait for ticker.
	time.Sleep(50 * time.Millisecond)

	// Verify session still exists and has been restarted (Adapter has running cmd)
	sm.mu.RLock()
	sRestored, exists := sm.sessions["session-h"]
	sm.mu.RUnlock()
	if !exists {
		t.Fatal("session should still exist after successful restart")
	}

	// Verify the process is running again
	_, _, _, err = sRestored.Adapter.Pipes()
	if err != nil {
		t.Errorf("adapter should be running after restart: %v", err)
	}
}

func TestStartHeartbeat_RestartFailure(t *testing.T) {
	// Create manager with invalid binary path to force restart failure
	sm := NewSessionManager("invalid-binary-path-xxxx", 2, 5*time.Minute)
	defer sm.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Since starting invalid-binary-path-xxxx will fail, let's mock/setup the session manually
	cliAdapter := adapter.NewCLIAdapter("invalid-binary-path-xxxx")
	s := &Session{
		ID:       "session-bad",
		Adapter:  cliAdapter,
		Created:  time.Now(),
		LastUsed: time.Now(),
	}

	sm.mu.Lock()
	sm.sessions["session-bad"] = s
	sm.mu.Unlock()

	// Start heartbeat with a very short interval
	sm.StartHeartbeat(ctx, "session-bad", 10*time.Millisecond, nil)

	// Wait for ticker to run and attempt restart
	time.Sleep(50 * time.Millisecond)

	// Verify session has been removed from the map after restart failure
	sm.mu.RLock()
	_, exists := sm.sessions["session-bad"]
	sm.mu.RUnlock()
	if exists {
		t.Error("session should have been removed from the map after restart failure")
	}
}
