# RFC-0023: Process Sandboxing & Container Isolation

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0012 (Security & Capability Model), RFC-0001 (Kernel Architecture)

## Summary

This RFC specifies the design of **Process Sandboxing & Container Isolation** in AEOS. To protect the host system from executing malicious code, the kernel wraps command executions, compilers, and test suites in lightweight local process sandboxes or Docker container boundaries.

## Motivation

Executing untrusted, AI-generated code directly on the host machine presents severe security risks (data deletion, privilege escalation, network exploitation).
- All command executions (`go test`, `npm run build`) must be restricted to isolated runtimes with strict CPU, RAM, and directory access controls.

## Design

### 1. Architectural Placement

The Sandboxing Engine sits between the `Execution Runtime` and the host operating system, enforcing security boundaries.

```
  Execution Task ──► [Process Sandbox] ──(Authorized?)──► Spawn Isolated Process / Container
```

---

### 2. Contracts (`contracts/security/sandbox.go`)

```go
package security

import (
	"context"
	"io"
)

// SandboxConfig defines resource bounds.
type SandboxConfig struct {
	AllowedPaths []string `json:"allowed_paths"`
	AllowNetwork bool     `json:"allow_network"`
	MaxMemoryMB  int64    `json:"max_memory_mb"`
	TimeoutMs    int      `json:"timeout_ms"`
}

// SandboxProcess represents the spawned container/process.
type SandboxProcess interface {
	// Stdout returns the process output stream.
	Stdout() io.Reader
	
	// Wait blocks until the process terminates, returning the exit code.
	Wait() (int, error)
	
	// Kill terminates the running process.
	Kill() error
}

// SandboxProvider manages process containers.
type SandboxProvider interface {
	// CreateProcess spawns an isolated command shell.
	CreateProcess(ctx context.Context, cmd string, args []string, config SandboxConfig) (SandboxProcess, error)
}
```

## Impact

- **Host Safety**: Malicious actions or resource-hogging loops are blocked, protecting the host filesystem.
- **Resource Limits**: Configured CPU/RAM limits prevent runaway infinite loops from freezing the workstation.
