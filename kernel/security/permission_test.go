package security

import (
	"path/filepath"
	"testing"
)

func TestPermissionManager(t *testing.T) {
	pm := NewPermissionManager()

	policy := &Policy{
		AllowedTools:    []string{"git", "grep"},
		BlockedCommands: []string{"rm -rf", "sudo"},
		AllowedPaths:    []string{"/tmp/allowed", "*"},
		BlockedPaths:    []string{"/tmp/blocked"},
	}

	pm.RegisterPolicy("test-agent", policy)

	// Test tool access
	if !pm.CanUseTool("test-agent", "git") {
		t.Error("expected access to git tool")
	}
	if pm.CanUseTool("test-agent", "curl") {
		t.Error("expected no access to curl tool")
	}
	if pm.CanUseTool("unknown-agent", "git") {
		t.Error("expected no access for unregistered agent")
	}

	// Test path access
	targetAllowed, _ := filepath.Abs("/tmp/allowed/file.txt")
	if !pm.CanAccessPath("test-agent", targetAllowed) {
		t.Errorf("expected access to allowed path: %s", targetAllowed)
	}

	targetBlocked, _ := filepath.Abs("/tmp/blocked/file.txt")
	if pm.CanAccessPath("test-agent", targetBlocked) {
		t.Errorf("expected no access to blocked path: %s", targetBlocked)
	}

	// Test command execution
	if !pm.CanRunCommand("test-agent", "git status") {
		t.Error("expected access to run git status")
	}
	if pm.CanRunCommand("test-agent", "sudo apt-get update") {
		t.Error("expected blocked command to fail")
	}
}
