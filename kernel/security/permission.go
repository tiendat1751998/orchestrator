// Package security implements agent permissions manager, audit logging, and env secrets loader.
package security

import (
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
