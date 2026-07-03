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
