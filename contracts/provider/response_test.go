package provider_test

import (
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

func TestUsageAddAndIsZero(t *testing.T) {
	u1 := provider.Usage{
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
	}
	u2 := provider.Usage{
		PromptTokens:     5,
		CompletionTokens: 15,
		TotalTokens:      20,
	}

	if u1.IsZero() {
		t.Fatal("expected u1 not to be zero")
	}

	var u3 provider.Usage
	if !u3.IsZero() {
		t.Fatal("expected default Usage to be zero")
	}

	u1.Add(u2)
	if u1.PromptTokens != 15 {
		t.Errorf("expected PromptTokens to be 15, got %d", u1.PromptTokens)
	}
	if u1.CompletionTokens != 35 {
		t.Errorf("expected CompletionTokens to be 35, got %d", u1.CompletionTokens)
	}
	if u1.TotalTokens != 50 {
		t.Errorf("expected TotalTokens to be 50, got %d", u1.TotalTokens)
	}
}

func TestResponseHelpers(t *testing.T) {
	// HasToolCalls
	r := provider.Response{
		ID:        "resp-1",
		Content:   "Hello",
		CreatedAt: time.Now(),
	}
	if r.HasToolCalls() {
		t.Fatal("expected HasToolCalls to be false when empty")
	}

	r.ToolCalls = []provider.ToolCall{
		{ID: "call-1", Name: "test_tool"},
	}
	if !r.HasToolCalls() {
		t.Fatal("expected HasToolCalls to be true when not empty")
	}

	// IsComplete (handles OpenAI "stop" and Claude "end_turn")
	r.FinishReason = "stop"
	if !r.IsComplete() {
		t.Error("expected IsComplete to be true for stop")
	}
	r.FinishReason = "end_turn"
	if !r.IsComplete() {
		t.Error("expected IsComplete to be true for end_turn")
	}
	r.FinishReason = "other"
	if r.IsComplete() {
		t.Error("expected IsComplete to be false for other")
	}

	// IsTruncated (handles Claude "max_tokens" and OpenAI "length")
	r.FinishReason = "max_tokens"
	if !r.IsTruncated() {
		t.Error("expected IsTruncated to be true for max_tokens")
	}
	r.FinishReason = "length"
	if !r.IsTruncated() {
		t.Error("expected IsTruncated to be true for length")
	}
	r.FinishReason = "stop"
	if r.IsTruncated() {
		t.Error("expected IsTruncated to be false for stop")
	}

	// WantsToolCall (handles OpenAI "tool_calls" and Claude "tool_use")
	r.FinishReason = "tool_calls"
	if !r.WantsToolCall() {
		t.Error("expected WantsToolCall to be true for tool_calls")
	}
	r.FinishReason = "tool_use"
	if !r.WantsToolCall() {
		t.Error("expected WantsToolCall to be true for tool_use")
	}
	r.FinishReason = "stop"
	if r.WantsToolCall() {
		t.Error("expected WantsToolCall to be false for stop")
	}

	// ToMessage
	r.Content = "assistant response"
	msg := r.ToMessage()
	if msg.Role != provider.RoleAssistant {
		t.Errorf("expected role to be RoleAssistant, got %s", msg.Role)
	}
	if msg.Content != "assistant response" {
		t.Errorf("expected content to match, got %s", msg.Content)
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].ID != "call-1" {
		t.Errorf("expected tool calls to match, got %v", msg.ToolCalls)
	}
}
