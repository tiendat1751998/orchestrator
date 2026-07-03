// Package filesystem implements filesystem query and write tools.
package filesystem

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// ReadFileTool reads content from files inside the workspace directory.
type ReadFileTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// ReadFileArgs maps JSON input parameters.
type ReadFileArgs struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// NewReadFileTool constructs a ReadFileTool.
func NewReadFileTool(workspaceDir string) (*ReadFileTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("read_file: invalid workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.Property{
			Type:        "string",
			Description: "Path to file, relative or absolute inside workspace",
		}).
		AddProperty("start_line", contractstool.Property{
			Type:        "integer",
			Description: "Start line (1-indexed)",
		}).
		AddProperty("end_line", contractstool.Property{
			Type:        "integer",
			Description: "End line (1-indexed, inclusive)",
		}).
		AddRequired("path")

	baseTool, err := sdktool.NewBaseTool("read_file", "Reads lines from a file within the workspace", schema)
	if err != nil {
		return nil, err
	}

	return &ReadFileTool{
		BaseTool:     baseTool,
		workspaceDir: absWorkspace,
	}, nil
}

// Execute performs file verification, path traversal checks, binary checks, and range reading.
func (t *ReadFileTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args ReadFileArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: invalid arguments: %v", err)), nil
	}

	absTarget, err := t.resolveAndValidatePath(args.Path)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: %v", err)), nil
	}

	f, err := os.Open(absTarget)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: failed to open file: %v", err)), nil
	}
	defer f.Close()

	if isBinaryFile(f) {
		return sdktool.Failure("read_file: cannot read binary file content"), nil
	}

	// Reset read head after binary check
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: failed to seek file: %v", err)), nil
	}

	content, err := readLines(f, args.StartLine, args.EndLine)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: error reading file lines: %v", err)), nil
	}

	return sdktool.Success(content), nil
}

// resolveAndValidatePath checks path validity, path traversal attempts, size limits, and existence.
func (t *ReadFileTool) resolveAndValidatePath(userInput string) (string, error) {
	targetPath := userInput
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(t.workspaceDir, targetPath)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	rel, err := filepath.Rel(t.workspaceDir, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("permission denied: path traversal attempt")
	}

	info, err := os.Stat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file %q not found", userInput)
		}
		return "", fmt.Errorf("failed to read file metadata: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory", userInput)
	}

	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("file exceeds maximum allowed size (10MB)")
	}

	return absTarget, nil
}

// readLines reads file lines filtering by line range.
func readLines(r io.Reader, startLine, endLine int) (string, error) {
	scanner := bufio.NewScanner(r)
	var output strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		// If start_line is set, skip preceding lines
		if startLine > 0 && lineNum < startLine {
			continue
		}
		// If end_line is set, break loop when reached
		if endLine > 0 && lineNum > endLine {
			break
		}

		output.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return output.String(), nil
}

// isBinaryFile reads the first 512 bytes and checks for null bytes to detect binary files.
func isBinaryFile(r io.Reader) bool {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}
	return bytes.Contains(buf[:n], []byte{0x00})
}
