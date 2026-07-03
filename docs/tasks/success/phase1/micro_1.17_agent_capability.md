# Micro-Task 1.17: Create contracts/agent/capability.go

## Info
- **File**: `contracts/agent/capability.go`
- **Package**: `agent`
- **Depends on**: 1.06
- **Time**: 10 min
- **Verify**: `go build ./contracts/agent/...`

## Purpose
Defines the `Capability` type indicating specific skill capabilities that agents can possess.

## EXACT code to create

```go
// Package agent defines the contract for AI agents.
package agent

// Capability represents a specific skill that an agent possesses.
type Capability string

const (
	// CapabilityCodeGeneration means the agent can generate new code.
	CapabilityCodeGeneration Capability = "code_generation"

	// CapabilityCodeReview means the agent can review and critique code.
	CapabilityCodeReview Capability = "code_review"

	// CapabilityArchitecture means the agent can design system architecture.
	CapabilityArchitecture Capability = "architecture"

	// CapabilityTesting means the agent can write or run tests.
	CapabilityTesting Capability = "testing"

	// CapabilityDocumentation means the agent can write documentation.
	CapabilityDocumentation Capability = "documentation"

	// CapabilityDeployment means the agent can handle deployment tasks.
	CapabilityDeployment Capability = "deployment"

	// CapabilityDebugging means the agent can debug and fix issues.
	CapabilityDebugging Capability = "debugging"

	// CapabilityRefactoring means the agent can refactor existing code.
	CapabilityRefactoring Capability = "refactoring"

	// CapabilityDataAnalysis means the agent can analyze data.
	CapabilityDataAnalysis Capability = "data_analysis"

	// CapabilityResearch means the agent can research topics and technologies.
	CapabilityResearch Capability = "research"
)

// String returns the string representation.
func (c Capability) String() string { return string(c) }

// IsValid checks if the capability is one of the defined constants.
func (c Capability) IsValid() bool {
	switch c {
	case CapabilityCodeGeneration, CapabilityCodeReview, CapabilityArchitecture, CapabilityTesting,
		CapabilityDocumentation, CapabilityDeployment, CapabilityDebugging, CapabilityRefactoring,
		CapabilityDataAnalysis, CapabilityResearch:
		return true
	default:
		return false
	}
}

// HasCapability checks if a capability exists in a slice.
func HasCapability(caps []Capability, target Capability) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}
```

## ⚠️ Pitfalls

### Pitfall 1: String type aliases vs Named types
```go
type Capability string // Named type -> prevents passing generic strings.
```
Use named type declarations to guarantee that parameter inputs to agent selector matches capability enums cleanly.

### Pitfall 2: Forgetting to update IsValid when adding capability constants
If you introduce a new skill constant, you must register it inside the `IsValid()` method switch statement, or validation check steps will reject it.

## Verify
```bash
go build ./contracts/agent/...
```

## Checklist
- [ ] File `contracts/agent/capability.go` exists
- [ ] Package: `agent`
- [ ] Capability constants prefixed with `Capability` exist (e.g. `CapabilityCodeGeneration`)
- [ ] `IsValid()` covers all 10 capability constants
- [ ] `HasCapability()` helper compiles without errors
- [ ] `go build ./contracts/agent/...` passes
