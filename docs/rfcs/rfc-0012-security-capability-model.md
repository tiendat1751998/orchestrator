# RFC-0012: Security & Capability Model

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0002 (Brain Architecture — Policy Engine)

## Summary

This RFC specifies the **Security & Capability Model** for the AI Engineering Operating System (AEOS). Since AI agents execute code, write files, and run terminal commands, they pose significant security risks. AEOS implements a strict Capability-based Security Model where Agents must be granted explicit permissions (Capabilities) to access system resources. Tool execution is confined within isolated sandbox runtimes (Process/Docker namespaces).

## Motivation

AI agents can be highly unpredictable:
- A code generation agent might accidentally run destructive commands (like `rm -rf /`).
- A compromised prompt could instruct an agent to read sensitive SSH keys or environment secrets.
- Hardcoded command checks (e.g. blocking `rm`) are easily bypassed by wrapping commands in shell variables or encoding them.

Implementing a robust security model ensures that the kernel restricts what an agent can touch, regardless of how the AI formats its requests.

## Design

### 1. Capability-Based Security

Every Agent and Tool has a designated manifest specifying its required and allowed capabilities:

```
  Agent execution requested
             │
             ▼
 ┌────────────────────────────────────────────────────────┐
 │ 1. Policy Engine check:                                │
 │    - Does Agent possess the "filesystem.write" capability?│
 │    - Does Target resource match allowed scopes?        │
 └──────────────────────────┬─────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼ Allowed                       ▼ Denied / Escalate
 ┌──────────────────┐              ┌──────────────────┐
 │ 2. Dispatch to   │              │ Abort task and   │
 │    Sandbox       │              │ raise Alert      │
 └──────────┬───────┘              └──────────────────┘
            │
            ▼
 ┌────────────────────────────────────────────────────────┐
 │ 3. Sandbox Constraints:                                │
 │    - Restricted read-write directory scopes            │
 │    - Bounded execution timeouts                        │
 │    - Docker namespace isolation (for dangerous code)   │
 └────────────────────────────────────────────────────────┘
```

---

### 2. Contracts (`contracts/security/`)

```go
// contracts/security/security.go
package security

import (
	"context"
)

type ResourceScope string

const (
	ScopeFilesystem ResourceScope = "filesystem"
	ScopeNetwork    ResourceScope = "network"
	ScopeProcess    ResourceScope = "process"
	ScopeSecret     ResourceScope = "secret"
)

// Capability defines a specific action permission.
type Capability struct {
	Scope      ResourceScope `json:"scope"`
	Action     string        `json:"action"`      // e.g. "read", "write", "execute", "listen"
	Constraint string        `json:"constraint"`  // e.g. glob directory patterns, URL domains
}

// Sandbox defines execution virtualization types.
type SandboxMode string

const (
	ModeLocalProcess SandboxMode = "local"
	ModeDockerContainer SandboxMode = "docker"
	ModeUnconfined      SandboxMode = "unconfined" // Warning! Only for admin processes
)

// SecurityManager evaluates policies and prepares sandboxes.
type SecurityManager interface {
	// VerifyPermission checks if an actor possesses the required capability.
	VerifyPermission(ctx context.Context, actorName string, cap Capability) (bool, string)
	
	// GetSandboxMode returns the required execution mode for a tool.
	GetSandboxMode(ctx context.Context, toolName string) SandboxMode
}
```

---

### 3. Confinement & Sandboxing (`kernel/execution/process/`)

#### Path Traversal & Directory Escape Prevention Rules (Fix for Issue A4)

> [!CAUTION]
> AI agents can bypass simple directory match checks using path traversal (`../`). To prevent escape:
> 1. All target paths requested by agents/tools MUST be resolved to absolute paths and cleaned via `filepath.Clean`.
> 2. The allowed constraint path (e.g. workspace directory) must also be cleaned and resolved to an absolute path.
> 3. Verify prefix using absolute path prefix matching. Strings contains check is strictly forbidden.
> 4. On Windows, paths are case-insensitive; comparison must lower-case both paths before matching.

```go
// Example path validation logic
func IsPathSafe(targetPath string, allowedDir string) bool {
	cleanTarget := filepath.Clean(targetPath)
	cleanAllowed := filepath.Clean(allowedDir)
	
	// Enforce prefix matching
	return strings.HasPrefix(strings.ToLower(cleanTarget), strings.ToLower(cleanAllowed))
}
```

The **Process Manager** inside the Execution Runtime uses these security policies to spawn processes under restricted environments:

#### A. Local Process Sandboxing (Windows/POSIX)
- On POSIX hosts, commands are run inside isolated user accounts (e.g. `nobody`) with namespace limitations.
- On Windows hosts, the Process Manager spawns commands under job objects, sets memory and CPU bounds, and restricts directory access using ACL policies.

```go
// kernel/execution/process/sandbox.go
package process

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/tiendat1751998/orchestrator/contracts/security"
)

type WindowsSandbox struct {
	AllowedDir string
}

func (s *WindowsSandbox) SecureCmd(ctx context.Context, program string, args []string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, program, args...)
	
	// Windows job objects or restriction flags (e.g. CREATE_BREAKAWAY_FROM_JOB)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	return cmd, nil
}
```

#### B. Docker Confinement
For highly dangerous tools (like running untrusted user code or test scripts), the Process Manager redirects the command to execute inside a transient Docker container:

```go
// kernel/execution/process/docker.go
package process

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

type DockerSandbox struct {
	ImageName string
}

func (ds *DockerSandbox) Execute(ctx context.Context, cmd *provider.CLICommand) (*provider.ProcessResult, error) {
	// 1. Convert CLICommand Program & Args to: docker run --rm -v localdir:/workspace image program args
	// 2. Call Docker API or local Docker CLI to spawn container
	// 3. Collect logs and return ProcessResult
	return &provider.ProcessResult{
		ExitCode: 0,
		Stdout:   []byte("Executed inside container sandbox successfully"),
		Stderr:   nil,
	}, nil
}
```

## Impact

- **Decoupled Security Policies**: Policy Engine decides *rules* (`AllowedDir`, `Scope`), and Sandbox Manager enforces the *jail* (Docker, OS permissions).
- **Security Logs**: All policy rejections publish alert events to the History Timeline.

## Open Questions

1. **How do we handle human-in-the-loop approvals?**
   - If an agent requests a capability it doesn't possess, or attempts a critical operation (like writing to a system file), the Policy Engine triggers `ActionEscalate` back to the FSM to wait for explicit human approval via the CLI/dashboard.
