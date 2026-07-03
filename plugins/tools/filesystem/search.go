package filesystem

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// SearchTool performs substring queries on text files inside the workspace.
type SearchTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// SearchArgs maps JSON input parameters.
type SearchArgs struct {
	Query string `json:"query"`
}

// NewSearchTool constructs a SearchTool.
func NewSearchTool(workspaceDir string) (*SearchTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("search: invalid workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("query", contractstool.Property{
			Type:        "string",
			Description: "Substring query to search for inside workspace text files",
		}).
		AddRequired("query")

	baseTool, err := sdktool.NewBaseTool("search", "Searches for matching query text in workspace files", schema)
	if err != nil {
		return nil, err
	}

	return &SearchTool{
		BaseTool:     baseTool,
		workspaceDir: absWorkspace,
	}, nil
}

// MatchInfo holds details of a search match.
type MatchInfo struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// Execute scans all text files in the workspace up to the maximum match limit.
func (t *SearchTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args SearchArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(fmt.Sprintf("search: invalid arguments: %v", err)), nil
	}

	if args.Query == "" {
		return sdktool.Failure("search: query string cannot be empty"), nil
	}

	var matches []MatchInfo
	maxMatches := 50 // Limit total returned results to avoid overloading the context

	// Walk workspace tree
	err := filepath.Walk(t.workspaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with read permission errors
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Skip internal build folders and repositories to save time
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".gemini" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check text files under 2MB
		if info.Size() > 2*1024*1024 {
			return nil
		}

		// Open and search file
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		if isBinaryFile(f) {
			return nil
		}
		_, _ = f.Seek(0, 0)

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			lineText := scanner.Text()
			if strings.Contains(lineText, args.Query) {
				relPath, _ := filepath.Rel(t.workspaceDir, path)
				matches = append(matches, MatchInfo{
					File:    relPath,
					Line:    lineNum,
					Content: strings.TrimSpace(lineText),
				})

				if len(matches) >= maxMatches {
					return errors.New("limit reached") // Terminate Walk early
				}
			}
		}

		return nil
	})

	// "limit reached" represents normal early termination, not a fatal failure
	if err != nil && err.Error() != "limit reached" {
		return sdktool.Failure(fmt.Sprintf("search: search walk failed: %v", err)), nil
	}

	resultJSON, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		return sdktool.Failure(fmt.Sprintf("search: failed to serialize results: %v", err)), nil
	}

	return sdktool.Success(string(resultJSON)), nil
}
