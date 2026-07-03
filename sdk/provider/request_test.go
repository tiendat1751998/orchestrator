package provider_test

import (
	"encoding/json"
	"testing"

	cprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

func TestRequestBuilder_ImmutabilityAndDeepCopies(t *testing.T) {
	// Initialize base builder
	baseBuilder := sdkprovider.NewRequestBuilder("gpt-4").
		AddUserMessage("Hello").
		WithTemperature(0.7).
		WithMaxTokens(100).
		WithTopP(0.9).
		WithStopSequences("STOP").
		WithTools(cprovider.ToolDefinition{
			Name:        "get_weather",
			Description: "Get weather",
			Parameters:  json.RawMessage(`{"type":"object"}`),
		}).
		WithStream(true).
		WithResponseFormat("json")

	// Build the first request
	req1, err := baseBuilder.Build()
	if err != nil {
		t.Fatalf("failed to build base request: %v", err)
	}

	// 1. Verify builder immutability: modifying a clone/derived builder does not mutate the parent/base builder
	derivedBuilder := baseBuilder.
		AddAssistantMessage("Response").
		WithTemperature(0.5).
		WithMaxTokens(50).
		WithTopP(0.8).
		WithStopSequences("STOP", "HALT").
		WithTools(cprovider.ToolDefinition{
			Name:        "get_time",
			Description: "Get current time",
			Parameters:  json.RawMessage(`{"type":"object"}`),
		})

	req2, err := derivedBuilder.Build()
	if err != nil {
		t.Fatalf("failed to build derived request: %v", err)
	}

	// Verify parent builder req1 is unchanged
	if req1.Model != "gpt-4" {
		t.Errorf("expected parent model to be gpt-4, got %s", req1.Model)
	}
	if len(req1.Messages) != 1 || req1.Messages[0].Content != "Hello" {
		t.Errorf("expected parent to have 1 message 'Hello', got: %v", req1.Messages)
	}
	if req1.Temperature == nil || *req1.Temperature != 0.7 {
		t.Errorf("expected parent temp to be 0.7, got %v", req1.Temperature)
	}
	if req1.MaxTokens == nil || *req1.MaxTokens != 100 {
		t.Errorf("expected parent max tokens to be 100, got %v", req1.MaxTokens)
	}
	if req1.TopP == nil || *req1.TopP != 0.9 {
		t.Errorf("expected parent top_p to be 0.9, got %v", req1.TopP)
	}
	if len(req1.StopSequences) != 1 || req1.StopSequences[0] != "STOP" {
		t.Errorf("expected parent stop sequences to be ['STOP'], got %v", req1.StopSequences)
	}
	if len(req1.Tools) != 1 || req1.Tools[0].Name != "get_weather" {
		t.Errorf("expected parent tools to contain only 'get_weather', got %v", req1.Tools)
	}

	// Verify derived builder req2 has the modifications
	if len(req2.Messages) != 2 || req2.Messages[1].Content != "Response" {
		t.Errorf("expected derived to have 2 messages, got: %v", req2.Messages)
	}
	if req2.Temperature == nil || *req2.Temperature != 0.5 {
		t.Errorf("expected derived temp to be 0.5, got %v", req2.Temperature)
	}
	if req2.MaxTokens == nil || *req2.MaxTokens != 50 {
		t.Errorf("expected derived max tokens to be 50, got %v", req2.MaxTokens)
	}
	if req2.TopP == nil || *req2.TopP != 0.8 {
		t.Errorf("expected derived top_p to be 0.8, got %v", req2.TopP)
	}
	if len(req2.StopSequences) != 2 || req2.StopSequences[1] != "HALT" {
		t.Errorf("expected derived stop sequences to have ['STOP', 'HALT'], got %v", req2.StopSequences)
	}
	if len(req2.Tools) != 1 || req2.Tools[0].Name != "get_time" {
		t.Errorf("expected derived tools to contain 'get_time', got %v", req2.Tools)
	}

	// 2. Slices (Messages, Tools, StopSequences) deep copies:
	// Let's modify the slice elements in req2's underlying slice and make sure it does not affect req1.
	req2.Messages[0].Content = "MUTATED"
	if req1.Messages[0].Content != "Hello" {
		t.Errorf("pointer slice sharing detected! Mutating req2 messages mutated req1 messages")
	}

	req2.StopSequences[0] = "MUTATED"
	if req1.StopSequences[0] != "STOP" {
		t.Errorf("pointer slice sharing detected! Mutating req2 stop sequences mutated req1 stop sequences")
	}

	req2.Tools[0].Name = "MUTATED"
	if req1.Tools[0].Name != "get_weather" {
		t.Errorf("pointer slice sharing detected! Mutating req2 tools mutated req1 tools")
	}

	// 3. Pointer fields (Temperature, MaxTokens, TopP) memory address reallocation.
	// Update values of req2 pointers and ensure req1 does not change.
	if req1.Temperature == req2.Temperature {
		t.Errorf("pointer address for Temperature is shared between builders")
	}
	if req1.MaxTokens == req2.MaxTokens {
		t.Errorf("pointer address for MaxTokens is shared between builders")
	}
	if req1.TopP == req2.TopP {
		t.Errorf("pointer address for TopP is shared between builders")
	}
}

func TestRequestBuilder_ValidationTrigger(t *testing.T) {
	// 4. Validation triggering in Build() using Request.Validate().
	// An invalid request has empty messages or invalid parameters.
	builder := sdkprovider.NewRequestBuilder("gpt-4") // No messages added, which violates validation rule.

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected Build() to fail validation for empty messages, but got nil error")
	}

	// Now add an invalid user message (empty content)
	builder2 := sdkprovider.NewRequestBuilder("gpt-4").AddUserMessage("")
	_, err = builder2.Build()
	if err == nil {
		t.Fatal("expected Build() to fail validation for empty user message, but got nil error")
	}

	// Now add a valid user message but invalid temperature
	builder3 := sdkprovider.NewRequestBuilder("gpt-4").
		AddUserMessage("Hello").
		WithTemperature(2.5) // Max temperature allowed is 2.0
	_, err = builder3.Build()
	if err == nil {
		t.Fatal("expected Build() to fail validation for temperature > 2.0, but got nil error")
	}
}
