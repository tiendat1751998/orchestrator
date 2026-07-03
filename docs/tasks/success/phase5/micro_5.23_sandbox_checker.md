# Micro-Task 5.23: Create kernel/security/sandbox.go

## Info
- **File**: `kernel/security/sandbox.go`
- **Package**: `security`
- **Depends on**: 5.22
- **Time**: 15 min
- **Verify**: `go build ./kernel/security/...`

## Purpose
Implements the sandboxed execution validation check (`VerifyExecutionBounds`) to enforce policies during tool executions.

## EXACT code to create

```go
package security

import (
	"fmt"
)

// Sandbox enforces security constraints on execution tasks.
type Sandbox struct {
	pm *PermissionManager
}

// NewSandbox constructs a Sandbox.
func NewSandbox(pm *PermissionManager) (*Sandbox, error) {
	if pm == nil {
		return nil, fmt.Errorf("sandbox: PermissionManager cannot be nil")
	}
	return &Sandbox{pm: pm}, nil
}

// VerifyExecutionBounds validates tool execution commands against permission policies.
func (s *Sandbox) VerifyExecutionBounds(agentName, toolName, pathContext, commandContext string) error {
	// 1. Tool check
	if !s.pm.CanUseTool(agentName, toolName) {
		return fmt.Errorf("security: agent %q is not authorized to use tool %q", agentName, toolName)
	}

	// 2. Path check (if pathContext is provided)
	if pathContext != "" {
		if !s.pm.CanAccessPath(agentName, pathContext) {
			return fmt.Errorf("security: agent %q path access check failed for %q", agentName, pathContext)
		}
	}

	// 3. Command check (if commandContext is provided)
	if commandContext != "" {
		if !s.pm.CanRunCommand(agentName, commandContext) {
			return fmt.Errorf("security: agent %q command check blocked execution of %q", agentName, commandContext)
		}
	}

	return nil
}
```

## Pitfalls

### Pitfall 1: Bypassing checks when context fields are empty
If path or command parameters are omitted, skipping tool permission checks entirely creates a security vulnerability. Ensure the tool name is always verified.

### Pitfall 2: Silent failures in sandbox initializations
Allowing nil PermissionManager references leads to nil pointer dereferences during checks. Validate dependencies during constructor calls.

## Verify
```bash
go build ./kernel/security/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/security/sandbox.go`
- [ ] Package name is `security`
- [ ] All exported types have Godoc
- [ ] Sandbox constructors validate dependencies
- [ ] Tool permissions are checked for all operations
- [ ] Build command passes
