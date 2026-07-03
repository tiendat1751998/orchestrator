# Micro-Task 5.26: Create kernel/security/security_test.go

## Info
- **File**: `kernel/security/security_test.go`
- **Package**: `security_test`
- **Depends on**: 5.25
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./kernel/security/...`

## Purpose
Implements complete unit test coverage for the security features, checking permission checking, audit logging, and env secrets loading.

## EXACT code to create

```go
package security_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/kernel/security"
)

func TestPermissionManager_DefaultDeny(t *testing.T) {
	pm := security.NewPermissionManager()

	// Default Deny: should return false if no policy is registered
	if pm.CanUseTool("unknown-agent", "read_file") {
		t.Error("expected default deny for unknown agent tool use, got true")
	}

	if pm.CanAccessPath("unknown-agent", "/etc/passwd") {
		t.Error("expected default deny for unknown agent path access, got true")
	}
}

func TestPermissionManager_Policies(t *testing.T) {
	pm := security.NewPermissionManager()

	policy := &security.Policy{
		AllowedTools:    []string{"read_file", "write_file"},
		BlockedCommands: []string{"rm -rf"},
		AllowedPaths:    []string{"/tmp/sandbox"},
		BlockedPaths:    []string{"/tmp/sandbox/secret"},
	}

	pm.RegisterPolicy("test-agent", policy)

	// 1. Tool check
	if !pm.CanUseTool("test-agent", "read_file") {
		t.Error("expected tool read_file to be allowed")
	}
	if pm.CanUseTool("test-agent", "run_command") {
		t.Error("expected tool run_command to be blocked")
	}

	// 2. Command check
	if !pm.CanRunCommand("test-agent", "ls -la") {
		t.Error("expected command 'ls -la' to be allowed")
	}
	if pm.CanRunCommand("test-agent", "rm -rf /") {
		t.Error("expected command 'rm -rf /' to be blocked")
	}

	// 3. Path check
	if !pm.CanAccessPath("test-agent", "/tmp/sandbox/hello.txt") {
		t.Error("expected path '/tmp/sandbox/hello.txt' to be allowed")
	}
	if pm.CanAccessPath("test-agent", "/tmp/sandbox/secret/pass.txt") {
		t.Error("expected path '/tmp/sandbox/secret/pass.txt' to be blocked")
	}
}

func TestAuditLogger_Writes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "audit.log")
	logger, err := security.NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("failed to construct logger: %v", err)
	}

	err = logger.Log("test-agent", "tool_call", "read_file", true, "success")
	if err != nil {
		t.Fatalf("failed to write audit log: %v", err)
	}

	_ = logger.Close()

	// Verify log file contents
	bytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(bytes)
	if !strings.Contains(content, `"agent":"test-agent"`) ||
		!strings.Contains(content, `"action":"tool_call"`) ||
		!strings.Contains(content, `"allowed":true`) {
		t.Errorf("incorrect log contents: %q", content)
	}
}

func TestSecrets_Redaction(t *testing.T) {
	os.Setenv("TEST_API_KEY", "super-secret-12345")
	defer os.Unsetenv("TEST_API_KEY")

	// Verify load
	val := security.LoadSecret("TEST_API_KEY")
	if val != "super-secret-12345" {
		t.Errorf("expected secret, got %q", val)
	}

	// Verify log redaction
	logMsg := "connecting to API with key super-secret-12345 and status OK"
	redacted := security.RedactSecrets(logMsg, []string{"TEST_API_KEY"})

	if strings.Contains(redacted, "super-secret-12345") {
		t.Error("expected secret to be redacted from log message")
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Error("expected redacted placeholder [REDACTED] to be present")
	}
}
```

## Pitfalls

### Pitfall 1: Leaking temporary testing directories on failures
If you call `t.Fatalf()` before registers cleanups are run, directories can remain on system drives, cluttering memory. Always register `defer os.RemoveAll` immediately after successful temporary directory creation.

### Pitfall 2: Environment variables pollution
Modifying environment variables during tests can affect subsequent tests. Clean up by calling `defer os.Unsetenv` on test exits.

## Verify
```bash
go test -v -race -count=1 ./kernel/security/...
# Expected: PASS
```

## Checklist
- [ ] File exists at `kernel/security/security_test.go`
- [ ] Package name is `security_test`
- [ ] Default Deny is verified for unknown agents
- [ ] Allowed and blocked tools, commands, and paths are validated
- [ ] Audit logs write JSON formatted strings to disk
- [ ] Environment secrets load and redact properly
- [ ] Build command passes
