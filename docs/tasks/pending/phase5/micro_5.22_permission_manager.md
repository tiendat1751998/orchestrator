# Micro-Task 5.22: Create kernel/security/permission.go

## Info
- **File**: `kernel/security/permission.go`
- **Package**: `security`
- **Depends on**: 5.21
- **Time**: 20 min
- **Verify**: `go build ./kernel/security/...`

## Purpose
Implements the core agent policies and permissions checker (`PermissionManager`, `Policy` and checks) to enforce strict sandboxing constraints on paths, tools, and commands.

## EXACT code to create

```go
// Package security implements agent permissions manager, audit logging, and env secrets loader.
package security

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
)

// Policy defines access control rules for a specific agent.
type Policy struct {
	AllowedTools    []string `yaml:"allowed_tools"`
	BlockedCommands []string `yaml:"blocked_commands"`
	AllowedPaths    []string `yaml:"allowed_paths"`
	BlockedPaths    []string `yaml:"blocked_paths"`
	MaxFileSize     int64    `yaml:"max_file_size"`
	MaxOutputSize   int64    `yaml:"max_output_size"`
}

// PermissionManager evaluates permission queries using registered policies.
// Thread-safe.
type PermissionManager struct {
	mu       sync.RWMutex
	policies map[string]*Policy
}

// NewPermissionManager constructs a NewPermissionManager.
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		policies: make(map[string]*Policy),
	}
}

// RegisterPolicy binds a security policy to an agent name.
func (pm *PermissionManager) RegisterPolicy(agentName string, policy *Policy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies[agentName] = policy
}

// CanUseTool evaluates if the agent is allowed to invoke the tool (Default Deny).
func (pm *PermissionManager) CanUseTool(agentName, toolName string) bool {
	pm.mu.RLock()
	policy, ok := pm.policies[agentName]
	pm.mu.RUnlock()

	if !ok || policy == nil {
		return false // Default Deny
	}

	for _, t := range policy.AllowedTools {
		if t == toolName || t == "*" {
			return true
		}
	}

	return false
}

// CanAccessPath evaluates path boundary constraints for the agent.
func (pm *PermissionManager) CanAccessPath(agentName, targetPath string) bool {
	pm.mu.RLock()
	policy, ok := pm.policies[agentName]
	pm.mu.RUnlock()

	if !ok || policy == nil {
		return false // Default Deny
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// 1. Check blocked paths first
	for _, p := range policy.BlockedPaths {
		absBlocked, err := filepath.Abs(p)
		if err == nil {
			if strings.HasPrefix(absTarget, absBlocked) {
				return false // Blocked path match
			}
		}
	}

	// 2. Check allowed paths
	for _, p := range policy.AllowedPaths {
		if p == "*" {
			return true
		}
		absAllowed, err := filepath.Abs(p)
		if err == nil {
			if strings.HasPrefix(absTarget, absAllowed) {
				return true
			}
		}
	}

	return false
}

// CanRunCommand checks the command query against blocklists.
func (pm *PermissionManager) CanRunCommand(agentName, command string) bool {
	pm.mu.RLock()
	policy, ok := pm.policies[agentName]
	pm.mu.RUnlock()

	if !ok || policy == nil {
		return false // Default Deny
	}

	normalized := strings.ToLower(strings.TrimSpace(command))

	for _, blocked := range policy.BlockedCommands {
		if strings.Contains(normalized, strings.ToLower(blocked)) {
			return false // Blocked command found
		}
	}

	return true
}
```

## Pitfalls

### Pitfall 1: Bypassing Default Deny architecture patterns
```go
// WRONG:
if policy == nil {
    return true // Default Allow if no policy registered: leaks permissions to untrusted agents!
}

// CORRECT:
if !ok || policy == nil {
    return false // Default Deny
}
```
If an agent has no policy defined, permitting access by default opens up the system to exploitation. Always block operations unless explicitly allowed.

### Pitfall 2: Path verification using relative strings
Relative paths (like `../../etc`) can bypass simple path checks. Always resolve paths to absolute paths (`filepath.Abs`) before running checks.

## Verify
```bash
go build ./kernel/security/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/security/permission.go`
- [ ] Package name is `security`
- [ ] All exported types have Godoc
- [ ] Policies enforce Default Deny access checks
- [ ] Paths are converted to absolute paths before checks
- [ ] Command strings are normalized for case-insensitive checks
- [ ] Build command passes
