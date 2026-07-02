# Micro-Task 2.11: Tạo kernel/logger/logger_test.go

## Thông tin
- **File tạo**: `kernel/logger/logger_test.go`
- **Package**: `logger_test`
- **Dependencies trước**: 2.07-2.10
- **Thời gian**: 20 phút
- **Verify**: `go test -v -race ./kernel/logger/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package logger_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/kernel/logger"
)

// =============================================================================
// Logger Core Tests
// =============================================================================

func TestNew_Defaults(t *testing.T) {
	// Empty options should not panic
	log := logger.New(logger.Options{})
	log.Info("test message") // Should not panic
}

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	log.Info("hello", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"hello"`) && !strings.Contains(output, `"msg": "hello"`) {
		t.Errorf("JSON output should contain msg field, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) && !strings.Contains(output, `"key": "value"`) {
		t.Errorf("JSON output should contain key field, got: %s", output)
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "info",
		Format: "text",
		Output: &buf,
	})

	log.Info("test message", "component", "kernel")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Text output should contain message, got: %s", output)
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "warn",
		Format: "text",
		Output: &buf,
	})

	log.Debug("debug msg")
	log.Info("info msg")
	log.Warn("warn msg")
	log.Error("error msg")

	output := buf.String()
	if strings.Contains(output, "debug msg") {
		t.Error("Debug should be filtered at warn level")
	}
	if strings.Contains(output, "info msg") {
		t.Error("Info should be filtered at warn level")
	}
	if !strings.Contains(output, "warn msg") {
		t.Error("Warn should be visible at warn level")
	}
	if !strings.Contains(output, "error msg") {
		t.Error("Error should be visible at warn level")
	}
}

func TestLogger_UnknownLevel_DefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "verbose", // Invalid level
		Format: "text",
		Output: &buf,
	})

	log.Debug("debug msg")
	log.Info("info msg")

	output := buf.String()
	if strings.Contains(output, "debug msg") {
		t.Error("Unknown level should default to Info → Debug filtered")
	}
	if !strings.Contains(output, "info msg") {
		t.Error("Unknown level should default to Info → Info visible")
	}
}

// =============================================================================
// Sub-logger Tests
// =============================================================================

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	subLog := log.With("request_id", "req-123")
	subLog.Info("handling request")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Errorf("Sub-logger should include persistent attrs, got: %s", output)
	}
}

func TestLogger_WithTask(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	taskLog := log.WithTask("tsk-a1b2c3d4")
	taskLog.Info("executing")

	output := buf.String()
	if !strings.Contains(output, "tsk-a1b2c3d4") {
		t.Errorf("WithTask should include task_id, got: %s", output)
	}
}

func TestLogger_WithComponent(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Options{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	kernelLog := log.WithComponent("kernel")
	kernelLog.Info("started")

	output := buf.String()
	if !strings.Contains(output, "kernel") {
		t.Errorf("WithComponent should include component, got: %s", output)
	}
}

func TestLogger_Slog(t *testing.T) {
	log := logger.New(logger.Options{})
	slogger := log.Slog()
	if slogger == nil {
		t.Error("Slog() should return non-nil *slog.Logger")
	}
}

// =============================================================================
// Redact Tests
// =============================================================================

func TestIsSensitiveField_ExactMatch(t *testing.T) {
	tests := []struct {
		field string
		want  bool
	}{
		{"api_key", true},
		{"password", true},
		{"token", true},
		{"secret", true},
		{"provider", false},
		{"name", false},
		{"model", false},
	}
	for _, tt := range tests {
		if got := logger.IsSensitiveField(tt.field); got != tt.want {
			t.Errorf("IsSensitiveField(%q) = %v, want %v", tt.field, got, tt.want)
		}
	}
}

func TestIsSensitiveField_CaseInsensitive(t *testing.T) {
	if !logger.IsSensitiveField("API_KEY") {
		t.Error("should be case-insensitive")
	}
	if !logger.IsSensitiveField("Password") {
		t.Error("should be case-insensitive")
	}
}

func TestIsSensitiveField_SuffixMatch(t *testing.T) {
	if !logger.IsSensitiveField("GEMINI_API_KEY") {
		t.Error("should match suffix api_key")
	}
	if !logger.IsSensitiveField("my_secret_token") {
		t.Error("should match suffix token")
	}
}

func TestRedact_SensitiveValue(t *testing.T) {
	result := logger.Redact("api_key", "sk-12345")
	if result != logger.RedactedValue {
		t.Errorf("expected [REDACTED], got %v", result)
	}
}

func TestRedact_NormalValue(t *testing.T) {
	result := logger.Redact("provider", "antigravity")
	if result != "antigravity" {
		t.Errorf("expected original value, got %v", result)
	}
}

func TestRedactString_LongString(t *testing.T) {
	result := logger.RedactString("sk-1234567890abcdef")
	if result != "sk-1****cdef" {
		t.Errorf("got %q, want %q", result, "sk-1****cdef")
	}
}

func TestRedactString_ShortString(t *testing.T) {
	result := logger.RedactString("short")
	if result != logger.RedactedValue {
		t.Errorf("short string should be fully redacted, got %q", result)
	}
}

func TestRedactMap_CopiesAndRedacts(t *testing.T) {
	original := map[string]string{
		"provider": "antigravity",
		"api_key":  "sk-secret",
		"model":    "gemini",
	}

	redacted := logger.RedactMap(original)

	// Check redacted map
	if redacted["provider"] != "antigravity" {
		t.Error("non-sensitive should be kept")
	}
	if redacted["api_key"] != logger.RedactedValue {
		t.Error("sensitive should be redacted")
	}
	if redacted["model"] != "gemini" {
		t.Error("non-sensitive should be kept")
	}

	// Check original is NOT modified
	if original["api_key"] != "sk-secret" {
		t.Error("original map should NOT be modified")
	}
}

// =============================================================================
// Formatter Tests
// =============================================================================

func TestFormatDuration_Milliseconds(t *testing.T) {
	result := logger.FormatDuration(150 * time.Millisecond)
	if result != "150ms" {
		t.Errorf("got %q, want %q", result, "150ms")
	}
}

func TestFormatDuration_Seconds(t *testing.T) {
	result := logger.FormatDuration(2 * time.Second)
	if result != "2.00s" {
		t.Errorf("got %q, want %q", result, "2.00s")
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	result := logger.FormatDuration(90 * time.Second)
	if result != "1m30s" {
		t.Errorf("got %q, want %q", result, "1m30s")
	}
}
```

## Lệnh verify
```bash
go test -v -race -count=1 ./kernel/logger/...
# Expected: ALL PASS, ≥ 20 test functions
```

## Checklist
- [ ] File `kernel/logger/logger_test.go` tồn tại
- [ ] Package: `logger_test`
- [ ] ≥ 20 test functions
- [ ] Tests for: New defaults, JSON format, Text format, level filtering
- [ ] Tests for: unknown level defaults to Info
- [ ] Tests for: With, WithTask, WithComponent, Slog
- [ ] Tests for: IsSensitiveField (exact, case-insensitive, suffix)
- [ ] Tests for: Redact, RedactString, RedactMap
- [ ] Tests for: FormatDuration (ms, s, min)
- [ ] RedactMap confirms original NOT modified
- [ ] `go test -v -race ./kernel/logger/...` ALL PASS
