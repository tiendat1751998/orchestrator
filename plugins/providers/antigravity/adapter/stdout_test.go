package adapter

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

type blockingReader struct {
	done chan struct{}
}

func (b *blockingReader) Read(p []byte) (n int, err error) {
	<-b.done
	return 0, io.EOF
}

func TestReadResponse(t *testing.T) {
	t.Run("nil stdout pipe", func(t *testing.T) {
		a := &CLIAdapter{}
		_, err := a.ReadResponse(context.Background(), "---END---")
		if err == nil {
			t.Fatal("expected error when stdout is nil")
		}
	})

	t.Run("delimiter detected", func(t *testing.T) {
		input := "hello world---END---extra data"
		mockStdout := &mockReadCloser{Reader: strings.NewReader(input)}
		a := &CLIAdapter{
			stdout: mockStdout,
		}

		res, err := a.ReadResponse(context.Background(), "---END---")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "hello world" {
			t.Errorf("expected %q, got %q", "hello world", res)
		}
	})

	t.Run("EOF reached without delimiter", func(t *testing.T) {
		input := "hello world unfinished"
		mockStdout := &mockReadCloser{Reader: strings.NewReader(input)}
		a := &CLIAdapter{
			stdout: mockStdout,
		}

		res, err := a.ReadResponse(context.Background(), "---END---")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "hello world unfinished" {
			t.Errorf("expected %q, got %q", "hello world unfinished", res)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		blocking := &blockingReader{done: make(chan struct{})}
		defer close(blocking.done)

		mockStdout := &mockReadCloser{Reader: blocking}
		a := &CLIAdapter{
			stdout: mockStdout,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		_, err := a.ReadResponse(ctx, "---END---")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		blocking := &blockingReader{done: make(chan struct{})}
		defer close(blocking.done)

		mockStdout := &mockReadCloser{Reader: blocking}
		a := &CLIAdapter{
			stdout: mockStdout,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := a.ReadResponse(ctx, "---END---")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got: %v", err)
		}
	})

	t.Run("default delimiter", func(t *testing.T) {
		input := "data payload---END---"
		mockStdout := &mockReadCloser{Reader: strings.NewReader(input)}
		a := &CLIAdapter{
			stdout: mockStdout,
		}

		res, err := a.ReadResponse(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != "data payload" {
			t.Errorf("expected %q, got %q", "data payload", res)
		}
	})
}
