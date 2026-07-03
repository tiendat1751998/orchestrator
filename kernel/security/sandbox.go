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
