package provider

import (
	"time"
)

// Response represents a complete response from an AI provider.
type Response struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
	Usage        Usage      `json:"usage"`
	Model        string     `json:"model"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Usage tracks token consumption for cost monitoring and optimization.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Add merges another Usage into this one (for aggregation).
func (u *Usage) Add(other Usage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
}

// IsZero returns true if no tokens were used.
func (u *Usage) IsZero() bool {
	return u.TotalTokens == 0
}

// StreamChunk represents a single piece of a streaming response.
type StreamChunk struct {
	Delta        string     `json:"delta"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	Done         bool       `json:"done"`
	FinishReason string     `json:"finish_reason,omitempty"`
	Usage        *Usage     `json:"usage,omitempty"`
	Error        error      `json:"-"`
}

// HasToolCalls returns true if the response contains tool calls.
func (r *Response) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// IsComplete returns true if the AI finished naturally (not cut off).
func (r *Response) IsComplete() bool {
	return r.FinishReason == "stop" || r.FinishReason == "end_turn"
}

// IsTruncated returns true if the response was cut off due to token limit.
func (r *Response) IsTruncated() bool {
	return r.FinishReason == "max_tokens" || r.FinishReason == "length"
}

// WantsToolCall returns true if the AI stopped because it wants to call tools.
func (r *Response) WantsToolCall() bool {
	return r.FinishReason == "tool_calls" || r.FinishReason == "tool_use"
}

// ToMessage converts the response into a Message for conversation history.
func (r *Response) ToMessage() Message {
	return Message{
		Role:      RoleAssistant,
		Content:   r.Content,
		ToolCalls: r.ToolCalls,
	}
}
