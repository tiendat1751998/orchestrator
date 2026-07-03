# Micro-Task 1.32: Create contracts/security/security.go

## Info
- **File**: `contracts/security/security.go`
- **Package**: `security`
- **Depends on**: 1.06
- **Time**: 10 min
- **Verify**: `go build ./contracts/security/...`

## Purpose
Declares structural contracts (`PermissionManager`, `AuditLogger`, `AuditEntry`, `AuditFilter`) for sandboxing agent tool execution, checking path access rules, and recording append-only audit traces.

## EXACT code to create

```go
// Package security defines contracts for access control and auditing.
// Security ensures agents can only perform actions they're allowed to.
package security

import "context"

// PermissionManager controls what agents can do.
//
// Default policy: DENY ALL. Agents must be explicitly granted permissions.
//
// The kernel calls PermissionManager before every tool execution:
//   1. Agent requests tool "run_command" with args "rm -rf /"
//   2. Kernel calls CanRunCommand("backend", "rm -rf /")
//   3. PermissionManager returns false
//   4. Kernel blocks the tool call and returns ErrPermissionDenied
type PermissionManager interface {
	// CanUseTool checks if an agent is allowed to use a specific tool.
	// Returns true if the tool is in the agent's allowed tools list.
	CanUseTool(agentName, toolName string) bool

	// CanAccessPath checks if an agent is allowed to read/write a file path.
	// Checks against allowed paths and blocked paths.
	// Blocked paths take precedence over allowed paths.
	//
	// Path matching rules:
	//   - "/src" matches "/src", "/src/main.go", "/src/pkg/util.go"
	//   - "/etc" blocked → "/etc/passwd" also blocked
	//   - Must handle both "/" (Unix) and "\" (Windows) separators
	CanAccessPath(agentName, path string) bool

	// CanRunCommand checks if a command is allowed.
	// Checks against a blocklist of dangerous commands.
	//
	// Command matching:
	//   - "rm -rf" blocked → "rm -rf /" also blocked
	//   - "rm" alone may be allowed (just delete a file)
	//   - Must normalize whitespace before matching
	CanRunCommand(agentName, command string) bool
}

// AuditLogger records all agent actions for security review.
//
// Every action (allowed or denied) is logged. This creates a complete
// audit trail for debugging, security review, and compliance.
//
// The audit log is append-only and must NOT be modified or deleted
// by the application. Use external log rotation if needed.
type AuditLogger interface {
	// Log records an audit entry.
	// Must be fast (< 1ms) — should not block the calling goroutine.
	// Implementations should buffer and flush asynchronously.
	Log(ctx context.Context, entry AuditEntry) error

	// Query retrieves audit entries matching the filter.
	// Results are sorted by timestamp (newest first).
	Query(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// AuditEntry is a single audit log record.
type AuditEntry struct {
	// Timestamp in RFC 3339 format.
	Timestamp string `json:"timestamp"`

	// Agent is the name of the agent that performed the action.
	Agent string `json:"agent"`

	// Action categorizes what happened.
	// Values: "tool_call", "file_read", "file_write", "command_exec"
	Action string `json:"action"`

	// Target is what the action was performed on.
	// For "tool_call": tool name. For "file_read": file path. For "command_exec": command.
	Target string `json:"target"`

	// Allowed indicates whether the action was permitted.
	Allowed bool `json:"allowed"`

	// Details provides additional context (optional).
	// For denied actions: the reason for denial.
	Details string `json:"details,omitempty"`
}

// AuditFilter for querying audit logs.
type AuditFilter struct {
	// Agent filters by agent name. Empty = all agents.
	Agent string `json:"agent,omitempty"`

	// Action filters by action type. Empty = all actions.
	Action string `json:"action,omitempty"`

	// Since filters entries after this timestamp (RFC 3339).
	Since string `json:"since,omitempty"`

	// Limit caps the number of results. 0 = no limit.
	Limit int `json:"limit,omitempty"`
}
```

## Rules
1. **Default DENY Policy**: If an agent requests access to a tool, path, or command, and no rule explicitly matches, the manager must deny the request.
2. **Blocked Path Precedence**: If a file path matches both allowed and blocked lists, the blocked rule always takes precedence.
3. **Audit Log Performance**: `Log` calls must process rapidly without executing blocking disk IO. Buffer entries and write them asynchronously.

## ⚠️ Pitfalls

### Pitfall 1: Vulnerability to whitespace command bypassing
```go
// Normalizing whitespace (replacing multi-spaces with single space) before evaluating policy rules.
```
Always normalize the spacing of command strings before validating them against permission policies.

### Pitfall 2: Confusing Windows and Unix path separators
On Windows systems, paths are represented using `\` (e.g. `\etc\passwd`), whereas permission blocklists often specify `/etc`. Convert all backward slashes to forward slashes before applying path filter matches.

## Verify
```bash
go build ./contracts/security/...
```

## Checklist
- [ ] File `contracts/security/security.go` exists
- [ ] Package: `security`
- [ ] `PermissionManager` interface declares CanUseTool, CanAccessPath, and CanRunCommand methods
- [ ] `AuditLogger` interface declares Log and Query methods
- [ ] `AuditEntry` struct contains Timestamp, Agent, Action, Target, Allowed, and Details fields
- [ ] `AuditFilter` contains Agent, Action, Since, and Limit fields
- [ ] `go build ./contracts/security/...` passes
