package provider

import (
	"encoding/json"
	"fmt"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Request contains all parameters needed to send a prompt to an AI provider.
type Request struct {
	// Model specifies which AI model to use.
	Model string `json:"model,omitempty"`

	// Messages is the conversation history in chronological order.
	Messages []Message `json:"messages"`

	// Tools defines the tools/functions available for the AI to call.
	Tools []ToolDefinition `json:"tools,omitempty"`

	// Temperature controls randomness of output.
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits the maximum number of tokens in the response.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// TopP controls nucleus sampling (alternative to temperature).
	TopP *float64 `json:"top_p,omitempty"`

	// StopSequences are strings that signal the AI to stop generating.
	StopSequences []string `json:"stop,omitempty"`

	// Stream indicates whether to use streaming response.
	Stream bool `json:"stream,omitempty"`

	// ResponseFormat constrains the output format.
	ResponseFormat string `json:"response_format,omitempty"`
}

// ToolDefinition describes a tool that the AI can call.
type ToolDefinition struct {
	// Name is the unique identifier for this tool.
	Name string `json:"name"`

	// Description explains what the tool does.
	Description string `json:"description"`

	// Parameters defines the input schema in JSON Schema format.
	Parameters json.RawMessage `json:"parameters"`
}

// Ptr returns a pointer to any value type (Go generic helper).
func Ptr[T any](v T) *T {
	return &v
}

// Float64Ptr returns a pointer to a float64 value.
// ponytail: helper required by specification tests.
func Float64Ptr(v float64) *float64 {
	return &v
}

// Validate checks if the request has minimum required fields and valid parameter bounds.
// Returns nil if valid, or a structured *contracts.ValidationError.
func (r *Request) Validate() error {
	if len(r.Messages) == 0 {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "messages",
			Reason:    "at least one message is required",
		}
	}

	for i, msg := range r.Messages {
		if !msg.Role.IsValid() {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].role", i),
				Reason:    "invalid role: " + string(msg.Role),
			}
		}
		// System and user messages must contain content.
		if (msg.Role == RoleSystem || msg.Role == RoleUser) && msg.Content == "" {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].content", i),
				Reason:    "content is required for system/user messages",
			}
		}
		if msg.Role == RoleTool && msg.ToolCallID == "" {
			return &contracts.ValidationError{
				Component: "request",
				Field:     fmt.Sprintf("messages[%d].tool_call_id", i),
				Reason:    "tool_call_id is required for tool result messages",
			}
		}
	}

	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "temperature",
			Reason:    "must be between 0.0 and 2.0",
		}
	}

	if r.MaxTokens != nil && *r.MaxTokens < 1 {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "max_tokens",
			Reason:    "must be >= 1",
		}
	}

	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return &contracts.ValidationError{
			Component: "request",
			Field:     "top_p",
			Reason:    "must be between 0.0 and 1.0",
		}
	}

	return nil
}
