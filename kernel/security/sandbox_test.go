package security

import (
	"strings"
	"testing"
)

func TestSandbox(t *testing.T) {
	// 1. Constructor nil check
	_, err := NewSandbox(nil)
	if err == nil {
		t.Error("expected error when constructing Sandbox with nil PermissionManager")
	}

	pm := NewPermissionManager()
	policy := &Policy{
		AllowedTools:    []string{"git"},
		BlockedCommands: []string{"rm -rf"},
		AllowedPaths:    []string{"/tmp/allowed"},
	}
	pm.RegisterPolicy("test-agent", policy)

	sb, err := NewSandbox(pm)
	if err != nil {
		t.Fatalf("failed to construct Sandbox: %v", err)
	}

	// 2. Successful verification
	err = sb.VerifyExecutionBounds("test-agent", "git", "/tmp/allowed/file.txt", "git status")
	if err != nil {
		t.Errorf("expected no error for allowed action, got: %v", err)
	}

	// 3. Tool check failure
	err = sb.VerifyExecutionBounds("test-agent", "curl", "", "")
	if err == nil || !strings.Contains(err.Error(), "not authorized to use tool") {
		t.Errorf("expected unauthorized tool error, got: %v", err)
	}

	// 4. Path check failure
	err = sb.VerifyExecutionBounds("test-agent", "git", "/tmp/blocked/file.txt", "")
	if err == nil || !strings.Contains(err.Error(), "path access check failed") {
		t.Errorf("expected path access failure, got: %v", err)
	}

	// 5. Command check failure
	err = sb.VerifyExecutionBounds("test-agent", "git", "", "rm -rf /")
	if err == nil || !strings.Contains(err.Error(), "command check blocked execution") {
		t.Errorf("expected command blocked error, got: %v", err)
	}
}
