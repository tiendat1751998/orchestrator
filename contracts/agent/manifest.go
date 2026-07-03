package agent

// Manifest describes an agent's configuration, loaded from a YAML file.
type Manifest struct {
	// Name is the unique identifier for this agent (e.g. "backend").
	Name string `yaml:"name" json:"name"`

	// Version follows semantic versioning (e.g., "0.1.0").
	Version string `yaml:"version" json:"version"`

	// Role is a human-readable role description (e.g., "Backend Developer").
	Role string `yaml:"role" json:"role"`

	// Description explains what this agent does.
	Description string `yaml:"description" json:"description"`

	// Capabilities lists what this agent can do.
	Capabilities []Capability `yaml:"capabilities" json:"capabilities"`

	// Provider is the name of the AI provider to use (e.g., "antigravity").
	Provider string `yaml:"provider" json:"provider"`

	// Model overrides the provider's default model (optional).
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// Tools lists the names of tools this agent can use (optional).
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`

	// SystemPrompt is the inline system prompt text (optional).
	SystemPrompt string `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`

	// PromptFile is the path to a file containing the system prompt (optional).
	PromptFile string `yaml:"prompt_file,omitempty" json:"prompt_file,omitempty"`

	// Temperature controls AI output randomness (0.0 to 2.0).
	Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`

	// MaxTokens limits the AI response length (optional).
	MaxTokens int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
}
