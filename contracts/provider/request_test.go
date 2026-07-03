package provider_test

import (
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

func TestRequestValidate(t *testing.T) {
	// Valid request
	req := provider.Request{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: "Hello"},
		},
		Temperature: provider.Ptr(0.7),
		MaxTokens:   provider.Ptr(100),
		TopP:        provider.Ptr(0.9),
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request, got err: %v", err)
	}

	// Invalid role
	invalidRoleReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.Role("invalid"), Content: "Hello"},
		},
	}
	err := invalidRoleReq.Validate()
	if err == nil {
		t.Fatal("expected error for invalid role, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "messages[0].role" || vErr.Reason != "invalid role: invalid" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Empty messages
	emptyMsgReq := provider.Request{
		Messages: []provider.Message{},
	}
	err = emptyMsgReq.Validate()
	if err == nil {
		t.Fatal("expected error for empty messages, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "messages" || vErr.Reason != "at least one message is required" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Empty content for user message
	emptyUserMsgReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: ""},
		},
	}
	err = emptyUserMsgReq.Validate()
	if err == nil {
		t.Fatal("expected error for empty user message content, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "messages[0].content" || vErr.Reason != "content is required for system/user messages" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Empty content for system message
	emptySystemMsgReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: ""},
		},
	}
	err = emptySystemMsgReq.Validate()
	if err == nil {
		t.Fatal("expected error for empty system message content, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "messages[0].content" || vErr.Reason != "content is required for system/user messages" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Empty tool_call_id for tool message
	emptyToolMsgReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleTool, Content: "result", ToolCallID: ""},
		},
	}
	err = emptyToolMsgReq.Validate()
	if err == nil {
		t.Fatal("expected error for empty tool call ID, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "messages[0].tool_call_id" || vErr.Reason != "tool_call_id is required for tool result messages" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Valid tool message
	validToolMsgReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleTool, Content: "result", ToolCallID: "call-123"},
		},
	}
	if err := validToolMsgReq.Validate(); err != nil {
		t.Fatalf("expected valid tool message request, got err: %v", err)
	}

	// Invalid temperature
	invalidTempReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: "Hello"},
		},
		Temperature: provider.Ptr(2.5),
	}
	err = invalidTempReq.Validate()
	if err == nil {
		t.Fatal("expected error for temperature > 2.0, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "temperature" || vErr.Reason != "must be between 0.0 and 2.0" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Invalid max tokens
	invalidTokensReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: "Hello"},
		},
		MaxTokens: provider.Ptr(0),
	}
	err = invalidTokensReq.Validate()
	if err == nil {
		t.Fatal("expected error for max_tokens < 1, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "max_tokens" || vErr.Reason != "must be >= 1" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}

	// Invalid top_p
	invalidTopPReq := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: "Hello"},
		},
		TopP: provider.Ptr(1.5),
	}
	err = invalidTopPReq.Validate()
	if err == nil {
		t.Fatal("expected error for top_p > 1.0, got nil")
	}
	if vErr, ok := err.(*contracts.ValidationError); !ok {
		t.Errorf("expected *contracts.ValidationError, got %T", err)
	} else {
		if vErr.Component != "request" || vErr.Field != "top_p" || vErr.Reason != "must be between 0.0 and 1.0" {
			t.Errorf("unexpected validation error: %+v", vErr)
		}
	}
}
