# Micro-Task 4.17: Create plugins/tools/filesystem/read_file.go

## Info
- **File**: `plugins/tools/filesystem/read_file.go`
- **Package**: `filesystem`
- **Depends on**: 4.16
- **Time**: 20 min
- **Verify**: `go build ./plugins/tools/filesystem/...`

## Purpose
Implements the read file execution tool (`ReadFileTool` and schemas) satisfying `contracts/tool.Tool` to safely read file lines inside the workspace boundary, preventing path traversal attacks.

## EXACT code to create

```go
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

	// 1. Resolve and validate target path is inside workspace
	targetPath := args.Path
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(t.workspaceDir, targetPath)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: invalid path: %v", err)), nil
	}

	rel, err := filepath.Rel(t.workspaceDir, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return sdktool.Failure("read_file: permission denied: path traversal attempt"), nil
	}

	// 2. Check if file exists and is a regular file
	info, err := os.Stat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return sdktool.Failure(fmt.Sprintf("read_file: file %q not found", args.Path)), nil
		}
		return sdktool.Failure(fmt.Sprintf("read_file: failed to read file metadata: %v", err)), nil
	}
	if info.IsDir() {
		return sdktool.Failure(fmt.Sprintf("read_file: %q is a directory", args.Path)), nil
	}

	// Limit total size to prevent OOM
	if info.Size() > 10*1024*1024 {
		return sdktool.Failure("read_file: file exceeds maximum allowed size (10MB)"), nil
	}

	// 3. Open and check if file is binary
	f, err := os.Open(absTarget)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: failed to open file: %v", err)), nil
	}
	defer f.Close()

	if isBinaryFile(f) {
		return sdktool.Failure("read_file: cannot read binary file content"), nil
	}

	// Reset read head after binary check
	_, _ = f.Seek(0, 0)

	// 4. Read contents (handling line ranges)
	scanner := bufio.NewScanner(f)
	var output strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		// If start_line is set, skip preceding lines
		if args.StartLine > 0 && lineNum < args.StartLine {
			continue
		}
		// If end_line is set, break loop when reached
		if args.EndLine > 0 && lineNum > args.EndLine {
			break
		}

		output.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return sdktool.Failure(fmt.Sprintf("read_file: error reading file lines: %v", err)), nil
	}

	return sdktool.Success(output.String()), nil
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
```

## Pitfalls

### Pitfall 1: Permitting relative path traversal attacks
```go
// WRONG:
data, _ := os.ReadFile(args.Path) // If args.Path is "../../etc/passwd", it leaks system secrets.

// CORRECT:
rel, err := filepath.Rel(t.workspaceDir, absTarget)
if err != nil || strings.HasPrefix(rel, "..") {
    return sdktool.Failure("path traversal attempt")
}
```
Always verify absolute paths are inside the configured workspace boundary before executing file operations.

### Pitfall 2: Memory exhaustion from loading large files
Attempting to read large source files or logs (e.g. 50MB) entirely into memory will exhaust heap allocations. Set file size limits (like 10MB) and support reading line ranges.

## Verify
```bash
go build ./plugins/tools/filesystem/...
```

## Checklist
- [ ] File exists at `plugins/tools/filesystem/read_file.go`
- [ ] Package name is `filesystem`
- [ ] All exported types have Godoc
- [ ] Path traversal checks verify paths are within workspace boundaries
- [ ] Binary files are detected and rejected using null byte checks
- [ ] Build command passes
