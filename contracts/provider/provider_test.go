package provider_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// =============================================================================
// Test: Message JSON round-trip
// =============================================================================

func TestMessage_JSONRoundTrip(t *testing.T) {
	original := provider.Message{
		Role:    provider.RoleUser,
		Content: "Hello, world!",
		Name:    "test-agent",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded provider.Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Role != original.Role {
		t.Errorf("Role: got %q, want %q", decoded.Role, original.Role)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content: got %q, want %q", decoded.Content, original.Content)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
}

// =============================================================================
// Test: Message with ToolCalls JSON
// =============================================================================

func TestMessage_WithToolCalls_JSON(t *testing.T) {
	msg := provider.Message{
		Role:    provider.RoleAssistant,
		Content: "",
		ToolCalls: []provider.ToolCall{
			{
				ID:   "call_123",
				Name: "read_file",
				Args: json.RawMessage(`{"path":"/src/main.go"}`),
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded provider.Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.ToolCalls) != 1 {
		t.Fatalf("ToolCalls: got %d, want 1", len(decoded.ToolCalls))
	}
	if decoded.ToolCalls[0].Name != "read_file" {
		t.Errorf("ToolCall Name: got %q, want %q", decoded.ToolCalls[0].Name, "read_file")
	}
}

// =============================================================================
// Test: ToolCall Args preserves JSON precision
// =============================================================================

func TestToolCall_ArgsPreservesJSONPrecision(t *testing.T) {
	rawArgs := json.RawMessage(`{"count":9999999999999999}`)
	tc := provider.ToolCall{
		ID:   "call_456",
		Name: "test_tool",
		Args: rawArgs,
	}

	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded provider.ToolCall
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if string(decoded.Args) != `{"count":9999999999999999}` {
		t.Errorf("Args precision lost: got %s", string(decoded.Args))
	}
}

// =============================================================================
// Test: Request with pointer fields
// =============================================================================

func TestRequest_PointerFields_NilVsZero(t *testing.T) {
	// Case 1: nil Temperature → should omit from JSON
	req1 := provider.Request{
		Messages: []provider.Message{
			provider.NewUserMessage("hello"),
		},
	}
	data1, _ := json.Marshal(req1)
	jsonStr1 := string(data1)

	if strings.Contains(jsonStr1, "temperature") {
		t.Error("nil Temperature should be omitted from JSON")
	}

	// Case 2: Temperature = 0.0 → should be present in JSON
	req2 := provider.Request{
		Messages: []provider.Message{
			provider.NewUserMessage("hello"),
		},
		Temperature: provider.Float64Ptr(0.0),
	}
	data2, _ := json.Marshal(req2)
	jsonStr2 := string(data2)

	if !strings.Contains(jsonStr2, "temperature") {
		t.Error("explicit 0.0 Temperature should be present in JSON")
	}
}

// =============================================================================
// Test: Request Validation
// =============================================================================

func TestRequest_Validate_NoMessages(t *testing.T) {
	req := &provider.Request{}
	err := req.Validate()
	if err == nil {
		t.Error("expected validation error for empty messages")
	}
}

func TestRequest_Validate_InvalidRole(t *testing.T) {
	req := &provider.Request{
		Messages: []provider.Message{
			{Role: "invalid_role", Content: "hello"},
		},
	}
	err := req.Validate()
	if err == nil {
		t.Error("expected validation error for invalid role")
	}
}

func TestRequest_Validate_TemperatureRange(t *testing.T) {
	tests := []struct {
		name    string
		temp    float64
		wantErr bool
	}{
		{"valid 0.0", 0.0, false},
		{"valid 1.0", 1.0, false},
		{"valid 2.0", 2.0, false},
		{"invalid -0.1", -0.1, true},
		{"invalid 2.1", 2.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &provider.Request{
				Messages:    []provider.Message{provider.NewUserMessage("hi")},
				Temperature: provider.Float64Ptr(tt.temp),
			}
			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Test: Response helpers
// =============================================================================

func TestResponse_HasToolCalls(t *testing.T) {
	resp := &provider.Response{
		ToolCalls: []provider.ToolCall{
			{ID: "call_1", Name: "test"},
		},
	}
	if !resp.HasToolCalls() {
		t.Error("expected HasToolCalls() = true")
	}
}

func TestResponse_IsComplete(t *testing.T) {
	tests := []struct {
		reason string
		want   bool
	}{
		{"stop", true},
		{"end_turn", true},
		{"max_tokens", false},
		{"tool_calls", false},
	}
	for _, tt := range tests {
		resp := &provider.Response{FinishReason: tt.reason}
		if got := resp.IsComplete(); got != tt.want {
			t.Errorf("IsComplete(%q) = %v, want %v", tt.reason, got, tt.want)
		}
	}
}

func TestResponse_ToMessage(t *testing.T) {
	resp := &provider.Response{
		Content: "Here is the code.",
		ToolCalls: []provider.ToolCall{
			{ID: "call_1", Name: "write_file"},
		},
	}
	msg := resp.ToMessage()
	if msg.Role != provider.RoleAssistant {
		t.Errorf("Role: got %q, want %q", msg.Role, provider.RoleAssistant)
	}
	if msg.Content != resp.Content {
		t.Error("Content not preserved")
	}
}

// =============================================================================
// Test: Usage.Add
// =============================================================================

func TestUsage_Add(t *testing.T) {
	total := provider.Usage{}
	total.Add(provider.Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150})
	total.Add(provider.Usage{PromptTokens: 200, CompletionTokens: 100, TotalTokens: 300})

	if total.PromptTokens != 300 {
		t.Errorf("PromptTokens: got %d, want 300", total.PromptTokens)
	}
	if total.CompletionTokens != 150 {
		t.Errorf("CompletionTokens: got %d, want 150", total.CompletionTokens)
	}
	if total.TotalTokens != 450 {
		t.Errorf("TotalTokens: got %d, want 450", total.TotalTokens)
	}
}

// =============================================================================
// Test: Config helpers
// =============================================================================

func TestConfig_TimeoutOrDefault(t *testing.T) {
	cfg := &provider.Config{}
	if cfg.TimeoutOrDefault() != 120*time.Second {
		t.Errorf("expected default 120s, got %v", cfg.TimeoutOrDefault())
	}

	cfg.Timeout = 30 * time.Second
	if cfg.TimeoutOrDefault() != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.TimeoutOrDefault())
	}
}

func TestConfig_APIKeyNotInJSON(t *testing.T) {
	cfg := provider.Config{
		Name:   "test",
		APIKey: "super-secret-key",
	}
	data, _ := json.Marshal(cfg)
	jsonStr := string(data)
	if strings.Contains(jsonStr, "super-secret-key") {
		t.Error("APIKey should NOT appear in JSON output")
	}
}
