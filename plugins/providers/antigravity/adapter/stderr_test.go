package adapter

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type mockLoggerHandler struct {
	buf bytes.Buffer
}

func (h *mockLoggerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *mockLoggerHandler) Handle(ctx context.Context, record slog.Record) error {
	h.buf.WriteString(record.Message)
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "content" {
			h.buf.WriteString(":")
			h.buf.WriteString(attr.Value.String())
		}
		return true
	})
	return nil
}

func (h *mockLoggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *mockLoggerHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestMonitorStderr(t *testing.T) {
	t.Run("nil stderr pipe", func(t *testing.T) {
		a := &CLIAdapter{}
		err := a.MonitorStderr(nil)
		if err == nil {
			t.Fatal("expected error when stderr is nil")
		}
	})

	t.Run("drains stderr successfully", func(t *testing.T) {
		input := "error logs from stderr"
		mockStderr := &mockReadCloser{Reader: strings.NewReader(input)}
		a := &CLIAdapter{
			stderr: mockStderr,
		}

		handler := &mockLoggerHandler{}
		logger := slog.New(handler)

		err := a.MonitorStderr(logger)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Wait a bit for the goroutine to finish reading
		time.Sleep(50 * time.Millisecond)

		got := handler.buf.String()
		if !strings.Contains(got, "antigravity CLI stderr output") {
			t.Errorf("expected logs to contain message, got: %q", got)
		}
		if !strings.Contains(got, "error logs from stderr") {
			t.Errorf("expected logs to contain content, got: %q", got)
		}
	})
}
