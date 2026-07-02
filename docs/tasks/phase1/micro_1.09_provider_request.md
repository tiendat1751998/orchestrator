# Micro-Task 1.09: Tạo contracts/provider/request.go

## Thông tin
- **File tạo**: `contracts/provider/request.go`
- **Package**: `provider`
- **Dependencies trước**: 1.08 (message.go)
- **Thời gian**: 15 phút
- **Verify**: `go build ./contracts/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package provider

import "encoding/json"

// =============================================================================
// Request
// =============================================================================

// Request contains all parameters needed to send a prompt to an AI provider.
//
// Example usage:
//
//	req := &Request{
//	    Model:    "gemini-2.5-pro",
//	    Messages: []Message{
//	        NewSystemMessage("You are a Go developer."),
//	        NewUserMessage("Write a hello world program."),
//	    },
//	    Temperature: Float64Ptr(0.3),
//	    MaxTokens:   IntPtr(4096),
//	}
type Request struct {
	// Model specifies which AI model to use.
	// Example: "gemini-2.5-pro", "claude-sonnet-4-20250514", "llama-3.1-70b"
	// If empty, the provider's default model is used.
	Model string `json:"model,omitempty"`

	// Messages is the conversation history in chronological order.
	// Must contain at least one message.
	// First message is typically RoleSystem (system prompt).
	Messages []Message `json:"messages"`

	// Tools defines the tools/functions available for the AI to call.
	// If empty, the AI cannot call any tools.
	Tools []ToolDefinition `json:"tools,omitempty"`

	// Temperature controls randomness of output.
	// Range: 0.0 (deterministic) to 2.0 (very random).
	// Default varies by provider (typically 1.0).
	//
	// WHY pointer (*float64)?
	// → Distinguish between "user set 0.0" and "user didn't set".
	// → If value type: Temperature=0.0 and "not set" are both 0.0 → bug.
	// → If pointer: nil means "not set", &0.0 means "explicitly 0.0".
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits the maximum number of tokens in the response.
	// If nil, the provider's default limit is used.
	//
	// WHY pointer (*int)?
	// → Same reason as Temperature. 0 means "not set" in value type.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// TopP controls nucleus sampling (alternative to temperature).
	// Range: 0.0 to 1.0. Only set one of Temperature or TopP.
	TopP *float64 `json:"top_p,omitempty"`

	// StopSequences are strings that signal the AI to stop generating.
	// Example: ["\n\n", "```"]
	StopSequences []string `json:"stop,omitempty"`

	// Stream indicates whether to use streaming response.
	// When true, use Provider.Stream() instead of Provider.Send().
	Stream bool `json:"stream,omitempty"`

	// ResponseFormat constrains the output format.
	// Example: "json" forces the AI to output valid JSON.
	// Not all providers support this.
	ResponseFormat string `json:"response_format,omitempty"`
}

// =============================================================================
// Tool Definition (for function calling)
// =============================================================================

// ToolDefinition describes a tool that the AI can call.
// This is sent to the AI so it knows what tools are available.
//
// Format follows the OpenAI function calling schema, which is also
// supported by Gemini and Claude.
//
// Example:
//
//	ToolDefinition{
//	    Name:        "read_file",
//	    Description: "Read the contents of a file at the given path",
//	    Parameters:  json.RawMessage(`{
//	        "type": "object",
//	        "properties": {
//	            "path": {"type": "string", "description": "Absolute path to the file"},
//	            "start_line": {"type": "integer", "description": "Start line (1-indexed)"}
//	        },
//	        "required": ["path"]
//	    }`),
//	}
type ToolDefinition struct {
	// Name is the unique identifier for this tool.
	// The AI will use this name when making a tool call.
	// Convention: snake_case (e.g., "read_file", "git_commit")
	Name string `json:"name"`

	// Description explains what the tool does.
	// The AI reads this to decide WHEN to use the tool.
	// Write clearly and specifically — vague descriptions → wrong tool usage.
	// Max recommended length: 200 characters.
	Description string `json:"description"`

	// Parameters defines the input schema in JSON Schema format.
	// Must be a valid JSON Schema with "type": "object" at the top level.
	//
	// WHY json.RawMessage?
	// → JSON Schema can be complex (nested objects, arrays, enums).
	// → Representing it as a Go struct would be over-engineered.
	// → Raw JSON is flexible and directly compatible with all providers.
	Parameters json.RawMessage `json:"parameters"`
}

// =============================================================================
// Pointer Helper Functions
// =============================================================================

// Float64Ptr returns a pointer to the given float64 value.
// Use for setting optional fields like Temperature and TopP.
//
// Example: req.Temperature = Float64Ptr(0.3)
func Float64Ptr(v float64) *float64 {
	return &v
}

// IntPtr returns a pointer to the given int value.
// Use for setting optional fields like MaxTokens.
//
// Example: req.MaxTokens = IntPtr(4096)
func IntPtr(v int) *int {
	return &v
}

// =============================================================================
// Request Validation
// =============================================================================

// Validate checks if the request has minimum required fields.
// Returns nil if valid, or an error describing what's missing.
func (r *Request) Validate() error {
	if len(r.Messages) == 0 {
		return errNoMessages
	}
	for i, msg := range r.Messages {
		if !msg.Role.IsValid() {
			return &ValidationError{
				Field:   "messages",
				Index:   i,
				Message: "invalid role: " + string(msg.Role),
			}
		}
	}
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return &ValidationError{
			Field:   "temperature",
			Message: "must be between 0.0 and 2.0",
		}
	}
	if r.MaxTokens != nil && *r.MaxTokens < 1 {
		return &ValidationError{
			Field:   "max_tokens",
			Message: "must be >= 1",
		}
	}
	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return &ValidationError{
			Field:   "top_p",
			Message: "must be between 0.0 and 1.0",
		}
	}
	return nil
}

// errNoMessages is an internal error for empty message list.
var errNoMessages = &ValidationError{
	Field:   "messages",
	Message: "at least one message is required",
}

// ValidationError provides structured validation error information.
type ValidationError struct {
	Field   string `json:"field"`
	Index   int    `json:"index,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Index > 0 {
		return "validation error: " + e.Field + "[" + intToStr(e.Index) + "]: " + e.Message
	}
	return "validation error: " + e.Field + ": " + e.Message
}

// intToStr converts int to string without importing strconv.
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
```

## Quy tắc
1. Pointer types (`*float64`, `*int`) cho optional numeric fields — phân biệt "không set" (nil) vs "set = 0"
2. `ToolDefinition.Parameters` dùng `json.RawMessage` — JSON Schema quá phức tạp để model bằng Go structs
3. Helper functions `Float64Ptr()`, `IntPtr()` — Go không hỗ trợ `&0.3` trực tiếp cho literals
4. `Validate()` kiểm tra minimum requirements — catch lỗi sớm trước khi gửi tới provider

## ⚠️ Pitfalls cần tránh
1. **Value type vs Pointer type**: `Temperature float64` → khi marshal JSON, `0.0` sẽ bị bỏ qua bởi `omitempty`. `Temperature *float64` → `nil` bị bỏ qua, `&0.0` được giữ lại. ĐÂY LÀ LÝ DO DÙNG POINTER.
2. **omitempty behavior**: JSON `omitempty` bỏ qua: `0`, `""`, `nil`, `false`, `[]`, `{}`. Nếu bạn MUỐN gửi `0` → PHẢI dùng pointer.
3. **ToolDefinition.Name convention**: snake_case, KHÔNG camelCase. Lý do: hầu hết AI APIs dùng snake_case cho tool names.
4. **intToStr helper**: Tránh import `strconv` chỉ cho 1 function nhỏ. Giữ contracts package minimal dependencies. Nếu cần nhiều conversions → import `strconv` thay.

## Checklist
- [ ] File `contracts/provider/request.go` tồn tại
- [ ] Request struct với 9 fields
- [ ] `Temperature` là `*float64` (pointer)
- [ ] `MaxTokens` là `*int` (pointer)
- [ ] `TopP` là `*float64` (pointer)
- [ ] ToolDefinition struct với 3 fields
- [ ] ToolDefinition.Parameters dùng `json.RawMessage`
- [ ] `Float64Ptr()` và `IntPtr()` helper functions
- [ ] `Validate()` method kiểm tra Messages, Temperature range, MaxTokens >= 1
- [ ] `ValidationError` struct implement `error` interface
- [ ] JSON tags đầy đủ với `omitempty` cho optional fields
- [ ] Godoc comments với examples
- [ ] `go build ./contracts/...` không lỗi
