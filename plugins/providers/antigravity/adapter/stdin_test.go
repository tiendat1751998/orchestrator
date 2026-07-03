package adapter

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

type mockWriteCloser struct {
	bytes.Buffer
}

func (m *mockWriteCloser) Close() error {
	return nil
}

func TestWritePrompt(t *testing.T) {
	t.Run("empty prompt", func(t *testing.T) {
		a := &CLIAdapter{}
		err := a.WritePrompt("")
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}
	})

	t.Run("process not running", func(t *testing.T) {
		a := &CLIAdapter{}
		err := a.WritePrompt("hello")
		if err == nil {
			t.Fatal("expected error when process is not running")
		}
	})

	t.Run("stdin pipe not open", func(t *testing.T) {
		a := &CLIAdapter{
			cmd: &exec.Cmd{
				Process: &os.Process{},
			},
		}
		err := a.WritePrompt("hello")
		if err == nil {
			t.Fatal("expected error when stdin is nil")
		}
	})

	t.Run("successful write and normalization", func(t *testing.T) {
		mockStdin := &mockWriteCloser{}
		a := &CLIAdapter{
			cmd: &exec.Cmd{
				Process: &os.Process{},
			},
			stdin: mockStdin,
		}

		// Prompt with mix of newlines
		prompt := "line1\nline2\r\nline3"
		err := a.WritePrompt(prompt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "line1\r\nline2\r\nline3\r\n"
		if mockStdin.String() != expected {
			t.Errorf("expected %q, got %q", expected, mockStdin.String())
		}
	})
}
