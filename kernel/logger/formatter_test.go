package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func stripANSI(str string) string {
	replacements := []string{
		colorReset, "",
		colorRed, "",
		colorGreen, "",
		colorYellow, "",
		colorBlue, "",
		colorCyan, "",
		colorGray, "",
		colorBold, "",
	}
	r := strings.NewReplacer(replacements...)
	return r.Replace(str)
}

func TestPrettyHandler_Format(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(h)
	logger.Info("hello world", "key", "val", "duration", 500*time.Millisecond)

	rawOutput := buf.String()
	output := stripANSI(rawOutput)

	// Should contain the message
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain message, got: %q", output)
	}

	// Should contain the formatted key-value pair
	if !strings.Contains(output, "key=val") {
		t.Errorf("expected output to contain 'key=val', got: %q", output)
	}

	// Should format duration using FormatDuration (500ms)
	if !strings.Contains(output, "duration=500ms") {
		t.Errorf("expected output to contain formatted duration 'duration=500ms', got: %q", output)
	}

	// Should contain the emoji checkmark for Info (not stripped by stripANSI)
	if !strings.Contains(rawOutput, "✅") {
		t.Errorf("expected raw output to contain '✅', got: %q", rawOutput)
	}
}

func TestPrettyHandler_ReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "secret" {
				return slog.String("secret", "[REDACTED]")
			}
			return a
		},
	})

	logger := slog.New(h)
	logger.Info("login", "user", "alice", "secret", "password123")

	output := stripANSI(buf.String())
	if strings.Contains(output, "password123") {
		t.Error("expected secret password to be redacted")
	}
	if !strings.Contains(output, "secret=[REDACTED]") {
		t.Errorf("expected redacted output, got: %q", output)
	}
}

func TestPrettyHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(h).With("pre1", "val1")
	logger.Info("msg", "key1", "val2")

	output := stripANSI(buf.String())
	if !strings.Contains(output, "pre1=val1") {
		t.Errorf("expected output to contain With attribute, got: %q", output)
	}
	if !strings.Contains(output, "key1=val2") {
		t.Errorf("expected output to contain record attribute, got: %q", output)
	}
}

func TestPrettyHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// With group and nested attribute
	logger := slog.New(h).WithGroup("mygroup").With("key", "val")
	logger.Info("msg")

	output := stripANSI(buf.String())
	if !strings.Contains(output, "mygroup.key=val") {
		t.Errorf("expected output to prefix group name 'mygroup.key=val', got: %q", output)
	}
}

func TestPrettyHandler_NestedGroupValues(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(h)
	logger.Info("nested", slog.Group("group1", slog.Group("group2", slog.String("k", "v"))))

	output := stripANSI(buf.String())
	if !strings.Contains(output, "group1.group2.k=v") {
		t.Errorf("expected output to flatten nested groups as 'group1.group2.k=v', got: %q", output)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{5 * time.Second, "5.00s"},
		{12500 * time.Millisecond, "12.50s"},
		{65 * time.Second, "1m5s"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %q, expected %q", tt.d, result, tt.expected)
		}
	}
}

func TestPrettyHandler_CustomLevels(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug - 4, // allow custom trace level below Debug
	})

	logger := slog.New(h)

	// Test Debug - 2 level
	logger.Log(context.Background(), slog.LevelDebug-2, "trace message")
	// Test Info + 2 level
	logger.Log(context.Background(), slog.LevelInfo+2, "notice message")
	// Test Warn + 2 level
	logger.Log(context.Background(), slog.LevelWarn+2, "critical message")
	// Test Error + 100 level
	logger.Log(context.Background(), slog.LevelError+100, "fatal message")

	output := buf.String()
	if !strings.Contains(output, "trace message") {
		t.Error("expected trace message to be logged")
	}
	if !strings.Contains(output, "notice message") {
		t.Error("expected notice message to be logged")
	}
	if !strings.Contains(output, "critical message") {
		t.Error("expected critical message to be logged")
	}
	if !strings.Contains(output, "fatal message") {
		t.Error("expected fatal message to be logged")
	}
}

func TestPrettyHandler_Enabled(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info level to be disabled when min level is Warn")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn level to be enabled")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Error level to be enabled")
	}
}

func TestPrettyHandler_ErrorValue(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(h)
	err := errors.New("database failure")
	logger.Error("failed to connect", "error", err)

	output := stripANSI(buf.String())
	if !strings.Contains(output, "error=database failure") {
		t.Errorf("expected error=database failure in output, got: %q", output)
	}
}
