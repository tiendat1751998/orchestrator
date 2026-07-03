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
	var copiedMsgs []provider.Message
	if len(b.req.Messages) > 0 {
		copiedMsgs = make([]provider.Message, len(b.req.Messages))
		copy(copiedMsgs, b.req.Messages)
	}

	var copiedTools []provider.ToolDefinition
	if len(b.req.Tools) > 0 {
		copiedTools = make([]provider.ToolDefinition, len(b.req.Tools))
		copy(copiedTools, b.req.Tools)
	}

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

// WithResponseFormat constrains model response format.
func (b *RequestBuilder) WithResponseFormat(format string) *RequestBuilder {
	nb := b.clone()
	nb.req.ResponseFormat = format
	return nb
}

// Build compiles, validates and returns the final provider.Request.
func (b *RequestBuilder) Build() (*provider.Request, error) {
	finalB := b.clone()
	req := &finalB.req

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("sdk/provider: built request is invalid: %w", err)
	}

	return req, nil
}
