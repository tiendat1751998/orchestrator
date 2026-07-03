package tools_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/plugins/tools/filesystem"
	"github.com/tiendat1751998/orchestrator/plugins/tools/terminal"
)

func TestFilesystemTools_ReadWriteAndList(t *testing.T) {
	// 1. Setup temporary sandbox directory
	tmpDir, err := os.MkdirTemp("", "orchestrator-tools-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Initialize tools
	writeTool, err := filesystem.NewWriteFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create write_file tool: %v", err)
	}

	readTool, err := filesystem.NewReadFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create read_file tool: %v", err)
	}

	listTool, err := filesystem.NewListDirTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create list_dir tool: %v", err)
	}

	ctx := context.Background()

	// 3. Test Write File
	writeArgs := json.RawMessage(`{"path": "hello.txt", "content": "Hello, Orchestrator!"}`)
	writeRes, err := writeTool.Execute(ctx, writeArgs)
	if err != nil {
		t.Fatalf("write_file tool execution failed: %v", err)
	}
	if writeRes.Error != "" {
		t.Fatalf("write_file returned error: %s", writeRes.Error)
	}

	// 4. Test Read File
	readArgs := json.RawMessage(`{"path": "hello.txt", "start_line": 1, "end_line": 1}`)
	readRes, err := readTool.Execute(ctx, readArgs)
	if err != nil {
		t.Fatalf("read_file tool execution failed: %v", err)
	}
	if readRes.Error != "" {
		t.Fatalf("read_file returned error: %s", readRes.Error)
	}
	if !strings.Contains(readRes.Output, "Hello, Orchestrator!") {
		t.Errorf("expected output to contain file contents, got %q", readRes.Output)
	}

	// 5. Test List Directory
	listArgs := json.RawMessage(`{"path": ""}`)
	listRes, err := listTool.Execute(ctx, listArgs)
	if err != nil {
		t.Fatalf("list_dir tool execution failed: %v", err)
	}
	if listRes.Error != "" {
		t.Fatalf("list_dir returned error: %s", listRes.Error)
	}
	if !strings.Contains(listRes.Output, "hello.txt") {
		t.Errorf("expected directory list to contain hello.txt, got %q", listRes.Output)
	}
}

func TestFilesystemTools_PathTraversalDefense(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "orchestrator-traversal-test-*")
	defer os.RemoveAll(tmpDir)

	readTool, _ := filesystem.NewReadFileTool(tmpDir)
	writeTool, _ := filesystem.NewWriteFileTool(tmpDir)

	ctx := context.Background()

	// Verify read traversal blocks
	readArgs := json.RawMessage(`{"path": "../secret.txt"}`)
	res, err := readTool.Execute(ctx, readArgs)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if !strings.Contains(res.Error, "path traversal attempt") {
		t.Errorf("expected path traversal error, got output: %q, error: %q", res.Output, res.Error)
	}

	// Verify write traversal blocks
	writeArgs := json.RawMessage(`{"path": "../secret.txt", "content": "hacked"}`)
	res, err = writeTool.Execute(ctx, writeArgs)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if !strings.Contains(res.Error, "path traversal attempt") {
		t.Errorf("expected path traversal error, got output: %q, error: %q", res.Output, res.Error)
	}
}

func TestTerminalTool_SecurityBlocklist(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "orchestrator-terminal-test-*")
	defer os.RemoveAll(tmpDir)

	termTool, _ := terminal.NewRunCommandTool(tmpDir)
	ctx := context.Background()

	// Verify dangerous commands are blocked
	args := json.RawMessage(`{"command": "rm -rf /"}`)
	res, err := termTool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if !strings.Contains(res.Error, "security error") {
		t.Errorf("expected command blocked error, got output: %q, error: %q", res.Output, res.Error)
	}
}
