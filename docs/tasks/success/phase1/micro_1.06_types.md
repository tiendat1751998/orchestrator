# Micro-Task 1.06: Create contracts/types.go

## Info
- **File**: `contracts/types.go`
- **Package**: `contracts`
- **Depends on**: 1.05
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Defines strongly typed identifiers (e.g. `MissionID`, `TaskID`) to prevent compiler type confusion, and implements secure unique ID generation using crypto/rand to avoid log bloat.

## EXACT code to create

```go
package contracts

import (
	"crypto/rand"
	"fmt"
)

// =============================================================================
// ID Types — Type-safe wrappers around string
// =============================================================================

// MissionID identifies a unique mission (a user's request).
type MissionID string

// TaskID identifies a unique task within a mission.
type TaskID string

// AgentID identifies a unique agent instance.
type AgentID string

// ProviderID identifies a unique provider instance.
type ProviderID string

// SessionID identifies a unique interaction session.
type SessionID string

// PluginID identifies a unique plugin instance.
type PluginID string

// =============================================================================
// ID Generation
// =============================================================================

// NewID generates a 8-character random hex string.
// Short enough for log files and CLI output, with low collision risk.
func NewID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback value to avoid panic
		return "00000000"
	}
	return fmt.Sprintf("%08x", b)
}

// NewMissionID generates a new MissionID.
func NewMissionID() MissionID {
	return MissionID("msn-" + NewID())
}

// NewTaskID generates a new TaskID.
func NewTaskID() TaskID {
	return TaskID("tsk-" + NewID())
}

// NewAgentID generates a new AgentID.
func NewAgentID() AgentID {
	return AgentID("agt-" + NewID())
}

// NewProviderID generates a new ProviderID.
func NewProviderID() ProviderID {
	return ProviderID("prv-" + NewID())
}

// NewSessionID generates a new SessionID.
func NewSessionID() SessionID {
	return SessionID("ses-" + NewID())
}

// NewPluginID generates a new PluginID.
func NewPluginID() PluginID {
	return PluginID("plg-" + NewID())
}

// =============================================================================
// String conversions
// =============================================================================

func (id MissionID) String() string  { return string(id) }
func (id TaskID) String() string     { return string(id) }
func (id AgentID) String() string    { return string(id) }
func (id ProviderID) String() string { return string(id) }
func (id SessionID) String() string  { return string(id) }
func (id PluginID) String() string   { return string(id) }

// =============================================================================
// Validation helpers
// =============================================================================

func (id MissionID) IsEmpty() bool  { return id == "" }
func (id TaskID) IsEmpty() bool     { return id == "" }
func (id AgentID) IsEmpty() bool    { return id == "" }
func (id ProviderID) IsEmpty() bool { return id == "" }
func (id SessionID) IsEmpty() bool  { return id == "" }
func (id PluginID) IsEmpty() bool   { return id == "" }
```

## ⚠️ Pitfalls

### Pitfall 1: Type Alias vs Named Type
```go
type MissionID string // Named type -> compiler flags invalid argument type mismatches.
```
Type aliases (`=`) tell the Go compiler that the types are identical. Use named type declarations (without `=`) to enforce proper type checks.

### Pitfall 2: Using predictable pseudo-random generators (math/rand)
Using `math/rand` without seeding, or using it in concurrent loops, generates predictable IDs that are vulnerable to trace guessing. Always use `crypto/rand`.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/types.go` exists
- [ ] Package: `contracts`
- [ ] Defined 6 ID named types (not aliases)
- [ ] Prefixed generation helpers exist (`msn-`, `tsk-`, `agt-`, `prv-`, `ses-`, `plg-`)
- [ ] String() conversion methods exist for all types
- [ ] IsEmpty() checks exist for all types
- [ ] Secure `crypto/rand` is used to generate bytes
- [ ] `go build ./contracts/...` passes
