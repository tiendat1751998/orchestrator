# Micro-Task 1.06: Tạo contracts/types.go

## Thông tin
- **File tạo**: `contracts/types.go`
- **Package**: `contracts`
- **Dependencies trước**: 1.05 (errors.go phải tồn tại trong cùng package)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package contracts

import (
	"crypto/rand"
	"fmt"
)

// =============================================================================
// ID Types — Named types cho type safety
// =============================================================================
//
// Tại sao dùng named types thay vì plain string?
// → Compiler bắt lỗi khi pass nhầm MissionID vào chỗ cần TaskID.
// Ví dụ:
//   func GetTask(id TaskID) → Gọi GetTask(missionID) sẽ compile error
//   func GetTask(id string)  → Gọi GetTask(missionID) sẽ compile OK → bug runtime

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

// NewID generates a new unique identifier.
// Format: 8 random hex characters (e.g., "a1b2c3d4").
// This is NOT a full UUID — it's shorter for readability in logs and CLI output.
// Collision probability is acceptable for this use case (~4 billion possible values).
func NewID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// crypto/rand.Read should never fail on modern systems.
		// If it does, fall back to a zero ID rather than panic.
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

// String returns the string representation of MissionID.
func (id MissionID) String() string { return string(id) }

// String returns the string representation of TaskID.
func (id TaskID) String() string { return string(id) }

// String returns the string representation of AgentID.
func (id AgentID) String() string { return string(id) }

// String returns the string representation of ProviderID.
func (id ProviderID) String() string { return string(id) }

// String returns the string representation of SessionID.
func (id SessionID) String() string { return string(id) }

// String returns the string representation of PluginID.
func (id PluginID) String() string { return string(id) }

// =============================================================================
// Validation
// =============================================================================

// IsEmpty returns true if the ID is empty or zero-value.
func (id MissionID) IsEmpty() bool { return id == "" }

// IsEmpty returns true if the ID is empty or zero-value.
func (id TaskID) IsEmpty() bool { return id == "" }

// IsEmpty returns true if the ID is empty or zero-value.
func (id AgentID) IsEmpty() bool { return id == "" }
```

## Quy tắc
1. Named types (`MissionID string`) thay vì type alias (`type MissionID = string`) — type alias KHÔNG tạo type safety
2. Prefix cho mỗi loại ID: `msn-`, `tsk-`, `agt-`, `prv-`, `ses-`, `plg-` — dễ identify loại ID trong logs
3. `NewID()` dùng `crypto/rand` — KHÔNG dùng `math/rand` (predictable, not secure)
4. `NewID()` KHÔNG panic khi `rand.Read()` fail — trả zero ID. Panic trong library = bad practice
5. Implement `String()` method — cho phép dùng `%s` format hoặc `fmt.Println(id)`
6. `IsEmpty()` helper — tránh `if id == ""` scattered khắp nơi

## ⚠️ Pitfalls cần tránh
1. **Type alias vs Named type**: 
   - `type MissionID = string` (alias) → Compiler KHÔNG bắt lỗi khi pass nhầm types → VÔ DỤNG
   - `type MissionID string` (named) → Compiler BẮT lỗi → ĐÂY LÀ CÁI CẦN DÙNG
2. **UUID length**: Full UUID v4 (`550e8400-e29b-41d4-a716-446655440000`) quá dài cho CLI output. 8 hex chars đủ cho hệ thống đơn lẻ.
3. **Conversion khi cần**: Khi cần string → `string(missionID)`. Khi cần MissionID từ string → `MissionID(str)`. Đây là Go convention cho named types.

## Checklist
- [ ] File `contracts/types.go` tồn tại
- [ ] 6 named ID types (MissionID, TaskID, AgentID, ProviderID, SessionID, PluginID)
- [ ] NewXxxID() functions cho mỗi type
- [ ] ID prefixes: msn-, tsk-, agt-, prv-, ses-, plg-
- [ ] String() method cho mỗi type
- [ ] IsEmpty() cho ít nhất MissionID, TaskID, AgentID
- [ ] Dùng `crypto/rand`, KHÔNG `math/rand`
- [ ] Không panic khi rand fail
- [ ] `go build ./contracts/...` không lỗi
