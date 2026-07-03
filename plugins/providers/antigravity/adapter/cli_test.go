package adapter

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestCLIAdapter_Lifecycle(t *testing.T) {
	binary := "echo"
	if runtime.GOOS == "windows" {
		// Use a shell command or executable that exists. cmd /c pause or similar.
		// For a simple test, we can use cmd.exe /c echo hello, but cmd.exe alone will wait for stdin.
		binary = "cmd"
	}

	adapter := NewCLIAdapter(binary)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := adapter.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start adapter: %v", err)
	}

	// Double start should fail
	err = adapter.Start(ctx)
	if err == nil {
		t.Error("expected second Start to fail, got nil")
	}

	stdin, stdout, stderr, err := adapter.Pipes()
	if err != nil {
		t.Fatalf("failed to get pipes: %v", err)
	}
	if stdin == nil || stdout == nil || stderr == nil {
		t.Error("expected pipes to be non-nil")
	}

	err = adapter.Stop()
	if err != nil {
		t.Fatalf("failed to stop adapter: %v", err)
	}

	// Pipes should error after stop
	_, _, _, err = adapter.Pipes()
	if err == nil {
		t.Error("expected Pipes to fail after Stop, got nil")
	}
}
