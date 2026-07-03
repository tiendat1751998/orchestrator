package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// WriteFileTool writes content to files inside the workspace directory.
type WriteFileTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// WriteFileArgs maps JSON input parameters.
type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// NewWriteFileTool constructs a WriteFileTool.
func NewWriteFileTool(workspaceDir string) (*WriteFileTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("write_file: invalid workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.Property{
			Type:        "string",
			Description: "Path to write, relative or absolute inside workspace",
		}).
		AddProperty("content", contractstool.Property{
			Type:        "string",
			Description: "Full string content to write to the file",
		}).
		AddRequired("path").
		AddRequired("content")

	baseTool, err := sdktool.NewBaseTool("write_file", "Writes content to a file within the workspace", schema)
	if err != nil {
		return nil, err
	}

	return &WriteFileTool{
		BaseTool:     baseTool,
		workspaceDir: absWorkspace,
	}, nil
}

// Execute validates paths, creates directories, and performs atomic writes.
func (t *WriteFileTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args WriteFileArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: invalid arguments: %v", err)), nil
	}

	// 1. Resolve and validate target path
	targetPath := args.Path
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(t.workspaceDir, targetPath)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: invalid path: %v", err)), nil
	}

	rel, err := filepath.Rel(t.workspaceDir, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return sdktool.Failure("write_file: permission denied: path traversal attempt"), nil
	}

	// 2. Create parent directories if they don't exist
	dir := filepath.Dir(absTarget)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: failed to create directories: %v", err)), nil
	}

	// 3. Write atomically: Write to temp file in same directory, then rename
	tempFile, err := os.CreateTemp(dir, ".write_file_tmp_*")
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: failed to create temp file: %v", err)), nil
	}
	tempName := tempFile.Name()
	defer func() {
		// Clean up temp file if rename failed
		if _, err := os.Stat(tempName); err == nil {
			_ = os.Remove(tempName)
		}
	}()

	_, err = tempFile.WriteString(args.Content)
	_ = tempFile.Close() // Close file to flush buffers and release Windows locks
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: failed to write content: %v", err)), nil
	}

	// Rename temp file to target destination (atomic rename)
	if err := os.Rename(tempName, absTarget); err != nil {
		return sdktool.Failure(fmt.Sprintf("write_file: failed to save destination file: %v", err)), nil
	}

	return sdktool.Success(fmt.Sprintf("Successfully wrote %d bytes to %s", len(args.Content), args.Path)), nil
}
