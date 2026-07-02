# Micro-Task 1.20: Tạo contracts/agent/manifest.go

## Thông tin
- **File tạo**: `contracts/agent/manifest.go`
- **Package**: `agent`
- **Dependencies trước**: 1.17 (capability.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/agent/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package agent

// Manifest describes an agent's configuration, loaded from a YAML file.
//
// Each agent plugin has a manifest file (agent.yaml) that declares:
//   - What the agent can do (capabilities)
//   - Which provider to use (by name)
//   - What tools the agent has access to (by name)
//   - How to configure the AI (temperature, max tokens)
//   - The system prompt (inline or file reference)
//
// Example agent.yaml:
//
//	name: backend
//	version: "0.1.0"
//	role: "Backend Developer"
//	description: "Generates Go backend code, APIs, and database schemas"
//	capabilities:
//	  - code_generation
//	  - testing
//	  - debugging
//	  - refactoring
//	provider: antigravity
//	model: gemini-2.5-pro
//	tools:
//	  - read_file
//	  - write_file
//	  - list_dir
//	  - search
//	  - git_status
//	  - git_diff
//	  - git_add
//	  - git_commit
//	  - run_command
//	prompt_file: prompts/system.md
//	temperature: 0.3
//	max_tokens: 8192
type Manifest struct {
	// Name is the unique identifier for this agent (e.g., "backend", "reviewer").
	Name string `yaml:"name" json:"name"`

	// Version follows semantic versioning (e.g., "0.1.0").
	Version string `yaml:"version" json:"version"`

	// Role is a human-readable role description (e.g., "Backend Developer").
	Role string `yaml:"role" json:"role"`

	// Description explains what this agent does (for documentation and AI selection).
	Description string `yaml:"description" json:"description"`

	// Capabilities lists what this agent can do.
	// Used by the orchestrator to match tasks to agents.
	Capabilities []Capability `yaml:"capabilities" json:"capabilities"`

	// Provider is the name of the AI provider to use (e.g., "antigravity").
	// The registry resolves this name to a Provider instance at runtime.
	// NOT a direct reference to avoid tight coupling.
	Provider string `yaml:"provider" json:"provider"`

	// Model overrides the provider's default model (optional).
	// Example: "gemini-2.5-pro", "gemini-2.5-flash"
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// Tools lists the names of tools this agent can use (optional).
	// The registry resolves these names to Tool instances at runtime.
	// If empty, the agent cannot call any tools.
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`

	// SystemPrompt is the inline system prompt text (optional).
	// For short prompts (< 500 chars), inline is convenient.
	// For longer prompts, use PromptFile instead.
	// If both are set, PromptFile takes precedence.
	SystemPrompt string `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`

	// PromptFile is the path to a file containing the system prompt (optional).
	// Path is relative to the manifest file location.
	// Example: "prompts/system.md"
	PromptFile string `yaml:"prompt_file,omitempty" json:"prompt_file,omitempty"`

	// Temperature controls AI output randomness (optional).
	// 0.0 = deterministic, 2.0 = very random. Default varies by provider.
	Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`

	// MaxTokens limits the AI response length (optional).
	// If 0, the provider's default is used.
	MaxTokens int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
}
```

## ⚠️ Pitfalls cần tránh
1. **SystemPrompt vs PromptFile**: Hỗ trợ CẢ HAI. Short prompts → inline. Long prompts → file. Nếu cả hai set → PromptFile wins.
2. **PromptFile path resolution**: Path PHẢI relative to manifest file location, KHÔNG relative to working directory. SDK sẽ resolve trong Phase 3.
3. **Provider và Tools là string names**: Registry resolve name → instance lúc runtime. KHÔNG lưu instance pointer trong manifest.
4. **YAML tags VÀ JSON tags**: Cần cả hai. YAML cho config files, JSON cho API responses.

## Checklist
- [ ] File `contracts/agent/manifest.go` tồn tại
- [ ] Package: `package agent`
- [ ] Manifest struct với 12 fields
- [ ] Mỗi field có cả `yaml:` VÀ `json:` tags
- [ ] Optional fields có `omitempty`
- [ ] Provider là `string` (tên, không phải instance)
- [ ] Tools là `[]string` (tên, không phải instances)
- [ ] Godoc comments với YAML example
- [ ] `go build ./contracts/agent/...` không lỗi
