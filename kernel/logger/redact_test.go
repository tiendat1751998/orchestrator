package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		// Exact matches (case-insensitive)
		{"api_key lowercase", "api_key", true},
		{"api_key uppercase", "API_KEY", true},
		{"apikey mixed", "ApiKey", true},
		{"api-key", "api-key", true},
		{"secret", "secret", true},
		{"password", "password", true},
		{"token", "token", true},
		{"access_token", "access_token", true},
		{"refresh_token", "refresh_token", true},
		{"authorization", "authorization", true},
		{"private_key", "private_key", true},
		{"secret_key", "secret_key", true},
		{"credentials", "credentials", true},

		// Suffix matches
		{"gemini api key", "GEMINI_API_KEY", true},
		{"custom secret", "custom_secret", true},
		{"my token", "my_token", true},
		{"suffix api-key", "some-api-key", true},
		{"suffix credentials", "db_credentials", true},

		// Non-sensitive
		{"provider", "provider", false},
		{"model", "model", false},
		{"username", "username", false},
		{"email", "email", false},
		{"safe_key", "safe_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSensitiveField(tt.field)
			if got != tt.expected {
				t.Errorf("IsSensitiveField(%q) = %v; want %v", tt.field, got, tt.expected)
			}
		})
	}
}

func TestRedact(t *testing.T) {
	// Sensitive field
	gotSensitive := Redact("api_key", "secret-value")
	if gotSensitive != RedactedValue {
		t.Errorf("Redact(api_key, secret-value) = %v; want %v", gotSensitive, RedactedValue)
	}

	// Non-sensitive field
	gotSafe := Redact("username", "john_doe")
	if gotSafe != "john_doe" {
		t.Errorf("Redact(username, john_doe) = %v; want john_doe", gotSafe)
	}
}

func TestRedactString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"long key", "sk-1234567890abcdef", "sk-1****cdef"},
		{"exactly 12 chars", "123456789012", "1234****9012"},
		{"11 chars", "12345678901", RedactedValue},
		{"short key", "short", RedactedValue},
		{"empty", "", RedactedValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactString(tt.input)
			if got != tt.expected {
				t.Errorf("RedactString(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactMap(t *testing.T) {
	original := map[string]string{
		"username": "john_doe",
		"api_key":  "sk-1234567890abcdef",
		"password": "supersecretpassword",
		"provider": "gemini",
	}

	redacted := RedactMap(original)

	// Check redacted values
	if redacted["username"] != "john_doe" {
		t.Errorf("expected username to remain unchanged, got: %s", redacted["username"])
	}
	if redacted["api_key"] != RedactedValue {
		t.Errorf("expected api_key to be redacted, got: %s", redacted["api_key"])
	}
	if redacted["password"] != RedactedValue {
		t.Errorf("expected password to be redacted, got: %s", redacted["password"])
	}
	if redacted["provider"] != "gemini" {
		t.Errorf("expected provider to remain unchanged, got: %s", redacted["provider"])
	}

	// Verify original map was not mutated (no in-place mutation)
	if original["api_key"] != "sk-1234567890abcdef" {
		t.Errorf("original map was mutated, api_key = %s", original["api_key"])
	}
}

func TestLoggerRedactionHook(t *testing.T) {
	var buf bytes.Buffer
	log := New(Options{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	log.Info("testing redaction", "username", "john_doe", "api_key", "super-secret-key")

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if parsed["username"] != "john_doe" {
		t.Errorf("expected username to be john_doe, got: %v", parsed["username"])
	}
	if parsed["api_key"] != RedactedValue {
		t.Errorf("expected api_key to be redacted to %q, got: %v", RedactedValue, parsed["api_key"])
	}
}
