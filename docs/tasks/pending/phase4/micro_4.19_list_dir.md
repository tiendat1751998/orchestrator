# Micro-Task 4.19: Create plugins/tools/filesystem/list_dir.go

## Info
- **File**: `plugins/tools/filesystem/list_dir.go`
- **Package**: `filesystem`
- **Depends on**: 4.18
- **Time**: 15 min
- **Verify**: `go build ./plugins/tools/filesystem/...`

## Purpose
Implements the directory lister tool (`ListDirTool` and schemas) to list files and folders inside the workspace while enforcing path traversal bounds.

## EXACT code to create

```go
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

// ListDirTool scans directory paths inside the workspace.
type ListDirTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// ListDirArgs maps JSON input parameters.
type ListDirArgs struct {
	Path string `json:"path"`
}

// NewListDirTool constructs a ListDirTool.
func NewListDirTool(workspaceDir string) (*ListDirTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("list_dir: invalid workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.Property{
			Type:        "string",
			Description: "Path of directory to list, relative or absolute inside workspace",
		})

	baseTool, err := sdktool.NewBaseTool("list_dir", "Lists files and directories inside a workspace folder", schema)
	if err != nil {
		return nil, err
	}

	return &ListDirTool{
		BaseTool:     baseTool,
		workspaceDir: absWorkspace,
	}, nil
}

// Execute performs path validation, checks if target is directory, and lists entries.
func (t *ListDirTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args ListDirArgs
	if len(rawArgs) > 0 && string(rawArgs) != "null" && string(rawArgs) != "{}" {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return sdktool.Failure(fmt.Sprintf("list_dir: invalid arguments: %v", err)), nil
		}
	}

	// 1. Resolve target path (default to workspace root if empty)
	targetPath := args.Path
	if targetPath == "" {
		targetPath = t.workspaceDir
	} else if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(t.workspaceDir, targetPath)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("list_dir: invalid path: %v", err)), nil
	}

	// 2. Validate path is inside workspace
	rel, err := filepath.Rel(t.workspaceDir, absTarget)
	if err != nil || (strings.HasPrefix(rel, "..") && rel != "..") {
		return sdktool.Failure("list_dir: permission denied: path traversal attempt"), nil
	}

	// 3. Verify target is a directory
	info, err := os.Stat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return sdktool.Failure(fmt.Sprintf("list_dir: directory %q not found", args.Path)), nil
		}
		return sdktool.Failure(fmt.Sprintf("list_dir: failed to read path info: %v", err)), nil
	}
	if !info.IsDir() {
		return sdktool.Failure(fmt.Sprintf("list_dir: %q is a file, not a directory", args.Path)), nil
	}

	// 4. Scan entries (shallow read, non-recursive)
	entries, err := os.ReadDir(absTarget)
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("list_dir: failed to read directory: %v", err)), nil
	}

	type fileInfo struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
		Size  int64  `json:"size_bytes,omitempty"`
	}

	var list []fileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}
		list = append(list, fileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  size,
		})
	}

	// Format output to JSON representation
	resultJSON, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("list_dir: failed to serialize results: %v", err)), nil
	}

	return sdktool.Success(string(resultJSON)), nil
}
```

## Pitfalls

### Pitfall 1: Recursive directory scanning by default
Attempting to scan nested subdirectories recursively (e.g. scanning `node_modules` or `.git`) will cause severe latency or OOM crashes. Default to shallow listings and require explicit path selection for subfolders.

### Pitfall 2: Permitting listing of paths outside the workspace
Failing to validate that the target directory is within workspace boundaries allows agents to inspect system directories. Verify absolute paths.

## Verify
```bash
go build ./plugins/tools/filesystem/...
```

## Checklist
- [ ] File exists at `plugins/tools/filesystem/list_dir.go`
- [ ] Package name is `filesystem`
- [ ] All exported types have Godoc
- [ ] Empty paths default to the workspace root path
- [ ] Listings are shallow (non-recursive)
- [ ] Build command passes
