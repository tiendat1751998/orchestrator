# Micro-Task 1.09: Create contracts/provider/request.go

## Info
- **File**: `contracts/provider/request.go`
- **Package**: `provider`
- **Depends on**: 1.08 (message.go)
- **Time**: 15 min
- **Verify**: `go build ./contracts/...`

## Purpose
Declares the `Request` and `ToolDefinition` schemas used to construct conversation and tool invocation payloads sent to AI providers.

## EXACT code to create

```go
package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

// Validate checks if the request has minimum required fields.
func (r *Request) Validate() error {
	if len(r.Messages) == 0 {
		return errors.New("request: at least one message is required")
	}

	for i, msg := range r.Messages {
		if !msg.Role.IsValid() {
			return fmt.Errorf("request: messages[%d] has invalid role %q", i, msg.Role)
		}
	}

	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return errors.New("request: temperature must be between 0.0 and 2.0")
	}

	if r.MaxTokens != nil && *r.MaxTokens < 1 {
		return errors.New("request: max_tokens must be >= 1")
	}

	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return errors.New("request: top_p must be between 0.0 and 1.0")
	}

	return nil
}
```

## Pitfalls

### Pitfall 1: Omission of zero values in optional temperature overrides
If temperature is defined as a primitive `float64` float type, setting it to `0.0` will be stripped by the `omitempty` marshaller tag, forcing the model to fallback to provider defaults. Always use pointer fields.

### Pitfall 2: Custom string conversion routines
Avoid writing custom digit-by-digit ASCII converters to format indices. Use standard Go library packages (`strconv.Itoa`) to ensure correctness and maintainability.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File exists at `contracts/provider/request.go`
- [ ] Package name is `provider`
- [ ] All exported types have Godoc
- [ ] Optional fields use pointer structures
- [ ] Standard strconv packages format index logs
- [ ] Generic `Ptr` helper is defined
- [ ] Build command passes
