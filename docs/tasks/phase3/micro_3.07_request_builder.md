# Micro-Task 3.07: Create sdk/provider/request.go

## Info
- **File**: `sdk/provider/request.go`
- **Package**: `provider`
- **Depends on**: 1.09 (request.go contract), 1.39 (input validation contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/provider/...`

## Purpose
Triển khai mẫu thiết kế Builder (`RequestBuilder`) để xây dựng đối tượng `provider.Request` một cách an toàn và trôi chảy (fluent API). Builder được thiết kế theo dạng bất biến (immutable pattern) nhằm ngăn ngừa tình trạng tranh chấp tài nguyên (race conditions) khi sử dụng chung cấu trúc mẫu request qua nhiều luồng (goroutines).

## EXACT code to create

```go
package provider

import (
	"fmt"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// RequestBuilder builds a provider.Request using a fluent, immutable API.
// Every method returns a new builder instance, making it completely safe for concurrent use.
type RequestBuilder struct {
	req provider.Request
}

// NewRequestBuilder initializes a new RequestBuilder with a target model.
func NewRequestBuilder(model string) *RequestBuilder {
	return &RequestBuilder{
		req: provider.Request{
			Model: model,
		},
	}
}

// clone creates a deep copy of the builder's internal request object.
func (b *RequestBuilder) clone() *RequestBuilder {
	// Deep copy messages slice
	var copiedMsgs []provider.Message
	if len(b.req.Messages) > 0 {
		copiedMsgs = make([]provider.Message, len(b.req.Messages))
		copy(copiedMsgs, b.req.Messages)
	}

	// Deep copy tools slice
	var copiedTools []provider.ToolDefinition
	if len(b.req.Tools) > 0 {
		copiedTools = make([]provider.ToolDefinition, len(b.req.Tools))
		copy(copiedTools, b.req.Tools)
	}

	// Deep copy stop sequences slice
	var copiedStop []string
	if len(b.req.StopSequences) > 0 {
		copiedStop = make([]string, len(b.req.StopSequences))
		copy(copiedStop, b.req.StopSequences)
	}

	newReq := provider.Request{
		Model:          b.req.Model,
		Messages:       copiedMsgs,
		Tools:          copiedTools,
		StopSequences:  copiedStop,
		Stream:         b.req.Stream,
		ResponseFormat: b.req.ResponseFormat,
	}

	// Copy pointer values
	if b.req.Temperature != nil {
		t := *b.req.Temperature
		newReq.Temperature = &t
	}
	if b.req.MaxTokens != nil {
		m := *b.req.MaxTokens
		newReq.MaxTokens = &m
	}
	if b.req.TopP != nil {
		tp := *b.req.TopP
		newReq.TopP = &tp
	}

	return &RequestBuilder{req: newReq}
}

// WithMessages appends message entries to the request history.
func (b *RequestBuilder) WithMessages(msgs ...provider.Message) *RequestBuilder {
	nb := b.clone()
	nb.req.Messages = append(nb.req.Messages, msgs...)
	return nb
}

// AddUserMessage appends a new user text message.
func (b *RequestBuilder) AddUserMessage(content string) *RequestBuilder {
	return b.WithMessages(provider.NewUserMessage(content))
}

// AddSystemMessage appends a new system behavior prompt.
func (b *RequestBuilder) AddSystemMessage(content string) *RequestBuilder {
	return b.WithMessages(provider.NewSystemMessage(content))
}

// AddAssistantMessage appends a new assistant text response.
func (b *RequestBuilder) AddAssistantMessage(content string) *RequestBuilder {
	return b.WithMessages(provider.NewAssistantMessage(content))
}

// AddToolResultMessage appends a tool execution outcome message.
func (b *RequestBuilder) AddToolResultMessage(toolCallID, content string) *RequestBuilder {
	return b.WithMessages(provider.NewToolResultMessage(toolCallID, content))
}

// WithTools replaces the list of tools available for AI call.
func (b *RequestBuilder) WithTools(tools ...provider.ToolDefinition) *RequestBuilder {
	nb := b.clone()
	nb.req.Tools = tools
	return nb
}

// WithTemperature sets the temperature control parameter.
func (b *RequestBuilder) WithTemperature(temp float64) *RequestBuilder {
	nb := b.clone()
	nb.req.Temperature = &temp
	return nb
}

// WithMaxTokens sets the response length limit.
func (b *RequestBuilder) WithMaxTokens(tokens int) *RequestBuilder {
	nb := b.clone()
	nb.req.MaxTokens = &tokens
	return nb
}

// WithTopP sets the nucleus sampling parameter.
func (b *RequestBuilder) WithTopP(topP float64) *RequestBuilder {
	nb := b.clone()
	nb.req.TopP = &topP
	return nb
}

// WithStopSequences sets custom stop word triggers.
func (b *RequestBuilder) WithStopSequences(seqs ...string) *RequestBuilder {
	nb := b.clone()
	nb.req.StopSequences = seqs
	return nb
}

// WithStream enables or disables response streaming.
func (b *RequestBuilder) WithStream(stream bool) *RequestBuilder {
	nb := b.clone()
	nb.req.Stream = stream
	return nb
}

// WithResponseFormat constrains model response format (e.g. "json").
func (b *RequestBuilder) WithResponseFormat(format string) *RequestBuilder {
	nb := b.clone()
	nb.req.ResponseFormat = format
	return nb
}

// Build compiles, validates and returns the final provider.Request.
// It executes Request.Validate() to ensure structural validity before returning.
func (b *RequestBuilder) Build() (*provider.Request, error) {
	// Deep copy to prevent caller mutating internal state of builder
	finalB := b.clone()
	req := &finalB.req

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("sdk/provider: built request is invalid: %w", err)
	}

	return req, nil
}
```

## ⚠️ Pitfalls

### Pitfall 1: Mutating state inside builder method (Making it mutable)
```go
// ❌ WRONG:
func (b *RequestBuilder) WithTemperature(t float64) *RequestBuilder {
    b.req.Temperature = &t
    return b // Mutates original builder instance -> not thread safe if template builder is reused.
}

// ✅ CORRECT:
func (b *RequestBuilder) WithTemperature(t float64) *RequestBuilder {
    nb := b.clone() // Create a copy of request state
    nb.req.Temperature = &t
    return nb // Safe for concurrent pipelines
}
```
If your builder modifies its internal pointer directly, two parallel tasks attempting to clone and customize temperature of a template request (e.g. system prompt template) will corrupt each other's data (data race).

### Pitfall 2: Reusing pointer values across clones without copying
When copying numeric pointers like `*float64`, copying the pointer address (`nb.req.Temperature = b.req.Temperature`) shares the memory. If the value pointed to is altered, it updates all templates. Always allocate new memory and copy the values (as done in `clone()`: `t := *b.req.Temperature; newReq.Temperature = &t`).

## Verify
```bash
go build ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/request.go` exists
- [ ] Package: `provider`
- [ ] `RequestBuilder` implements immutable fluent API pattern
- [ ] Method `clone()` deep copies slices (`Messages`, `Tools`, `StopSequences`)
- [ ] Pointers to float/int are allocated onto new memory during clone
- [ ] Helper methods `AddUserMessage`, `AddSystemMessage`, `AddToolResultMessage` exist
- [ ] `Build()` validates final request using `Request.Validate()`
- [ ] `go build ./sdk/provider/...` passes
