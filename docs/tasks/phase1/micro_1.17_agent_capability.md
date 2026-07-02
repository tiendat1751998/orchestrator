# Micro-Task 1.17: Tạo contracts/agent/capability.go

## Thông tin
- **File tạo**: `contracts/agent/capability.go`
- **Package**: `agent`
- **Dependencies trước**: 1.06 (contracts/types.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/agent/...`

## Nội dung CHÍNH XÁC cần tạo

```go
// Package agent defines the contract for AI agents.
// An agent is a specialized AI persona that can execute tasks.
package agent

// Capability represents a specific skill that an agent possesses.
// The orchestrator uses capabilities to match tasks to agents.
//
// WHY string constants instead of iota?
// → Capabilities are stored in YAML manifests and JSON configs.
// → iota produces integers (0, 1, 2) which are unreadable in YAML.
// → String values ("code_generation") are self-documenting.
type Capability string

const (
	// CapCodeGeneration means the agent can generate new code.
	CapCodeGeneration Capability = "code_generation"

	// CapCodeReview means the agent can review and critique code.
	CapCodeReview Capability = "code_review"

	// CapArchitecture means the agent can design system architecture.
	CapArchitecture Capability = "architecture"

	// CapTesting means the agent can write or run tests.
	CapTesting Capability = "testing"

	// CapDocumentation means the agent can write documentation.
	CapDocumentation Capability = "documentation"

	// CapDeployment means the agent can handle deployment tasks.
	CapDeployment Capability = "deployment"

	// CapDebugging means the agent can debug and fix issues.
	CapDebugging Capability = "debugging"

	// CapRefactoring means the agent can refactor existing code.
	CapRefactoring Capability = "refactoring"

	// CapDataAnalysis means the agent can analyze data.
	CapDataAnalysis Capability = "data_analysis"

	// CapResearch means the agent can research topics and technologies.
	CapResearch Capability = "research"
)

// String returns the string representation.
func (c Capability) String() string { return string(c) }

// IsValid checks if the capability is one of the defined constants.
func (c Capability) IsValid() bool {
	switch c {
	case CapCodeGeneration, CapCodeReview, CapArchitecture, CapTesting,
		CapDocumentation, CapDeployment, CapDebugging, CapRefactoring,
		CapDataAnalysis, CapResearch:
		return true
	default:
		return false
	}
}

// HasCapability checks if a capability exists in a slice.
// Used by the orchestrator to check if an agent has a required capability.
//
// Example:
//
//	caps := []Capability{CapCodeGeneration, CapTesting}
//	HasCapability(caps, CapTesting) // true
//	HasCapability(caps, CapDeployment) // false
func HasCapability(caps []Capability, target Capability) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}
```

## ⚠️ Pitfalls cần tránh
1. **KHÔNG dùng `iota`**: `iota` tạo int (0, 1, 2...) → serialize vào YAML/JSON thành số → không đọc được. String constants tự mô tả.
2. **KHÔNG dùng type alias**: `type Capability = string` (alias) KHÔNG tạo type safety. PHẢI dùng `type Capability string` (named type).
3. **IsValid() phải update khi thêm capability mới**: Thêm constant mới → PHẢI thêm vào switch case trong IsValid().

## Checklist
- [ ] File `contracts/agent/capability.go` tồn tại
- [ ] Package declaration: `package agent`
- [ ] 10 capability constants
- [ ] Mỗi constant có Godoc comment
- [ ] `String()` method
- [ ] `IsValid()` method covers tất cả 10 constants
- [ ] `HasCapability()` function
- [ ] `go build ./contracts/agent/...` không lỗi
