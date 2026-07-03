package prompt

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

func TestBuildCLIPrompt_NilRequest(t *testing.T) {
	_, err := BuildCLIPrompt(nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
	if !strings.Contains(err.Error(), "request cannot be nil") {
		t.Errorf("expected error message to contain 'request cannot be nil', got %q", err.Error())
	}
}

func TestBuildCLIPrompt_Formatting(t *testing.T) {
	req := &provider.Request{
		Messages: []provider.Message{
			{
				Role:    provider.RoleSystem,
				Content: "You are a helpful assistant.",
			},
			{
				Role:    provider.RoleUser,
				Content: "Hello, what tools do you have?",
			},
			{
				Role:    provider.RoleAssistant,
				Content: "I have some tools.",
				ToolCalls: []provider.ToolCall{
					{
						ID:   "call-1",
						Name: "get_weather",
						Args: json.RawMessage(`{"location":"Paris"}`),
					},
				},
			},
			{
				Role:       provider.RoleTool,
				ToolCallID: "call-1",
				Content:    `{"temp": 15}`,
			},
		},
		Tools: []provider.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Gets the weather for a location",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"location":{"type":"string"}}}`),
			},
		},
	}

	result, err := BuildCLIPrompt(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Check tools JSON formatting
	if !strings.Contains(result, "System Instructions: You have access to the following tools.") {
		t.Error("missing tools system instructions")
	}
	if !strings.Contains(result, "get_weather") {
		t.Error("missing weather tool name in tools block")
	}

	// 2. Check system instruction formatting
	if !strings.Contains(result, "System Instruction: You are a helpful assistant.\n") {
		t.Error("missing or incorrect system message formatting")
	}

	// 3. Check user message formatting
	if !strings.Contains(result, "User: Hello, what tools do you have?\n") {
		t.Error("missing or incorrect user message formatting")
	}

	// 4. Check assistant content formatting
	if !strings.Contains(result, "Assistant: I have some tools.\n") {
		t.Error("missing or incorrect assistant message content formatting")
	}

	// 5. Check assistant tool calls formatting
	if !strings.Contains(result, "Assistant (Requested Tools):\n") {
		t.Error("missing assistant requested tools header")
	}
	if !strings.Contains(result, `- call tool "get_weather" with args: {"location":"Paris"}`) {
		t.Error("missing or incorrect tool call formatting")
	}

	// 6. Check tool response formatting
	if !strings.Contains(result, "Tool (ID: call-1) Output: {\"temp\": 15}\n") {
		t.Error("missing or incorrect tool response formatting")
	}

	// 7. Check sentinel delimiter
	if !strings.HasSuffix(result, "\n---END-OF-PROMPT---\n") {
		t.Error("missing or incorrect end of prompt sentinel")
	}
}
