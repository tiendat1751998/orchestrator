# RFC-0025: Dependency Tree & AST Parser

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0013 (Workspace Engine)

## Summary

This RFC specifies the design of the **Dependency Tree & AST Parser** in AEOS. Located inside the Workspace Engine, it parses code source files into Abstract Syntax Trees (ASTs), mapping structures, functions, imports, and variables. This allows the system to analyze code dependencies and detect semantic violations before compilation.

## Motivation

AI models often write code that modifies function signatures without updating caller files, breaking compilation.
- Relying on compiler logs to detect signature changes is slow.
- By parsing the AST locally, the Workspace Engine can audit structural dependency trees and verify that all references are correct before running builds.

## Design

### 1. Architectural Placement

The AST Parser is a utility service inside the `Workspace Engine`.

```
  Source Files ──► [AST Parser] ──► Dependency Graph (Functions, Structures, Calls)
```

---

### 2. Contracts (`contracts/workspace/ast.go`)

```go
package workspace

import "context"

// NodeDefinition represents a code structure.
type NodeDefinition struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // "struct", "interface", "function"
	LineStart int      `json:"line_start"`
	LineEnd   int      `json:"line_end"`
	Deps      []string `json:"dependencies,omitempty"`
}

// ASTParser parses source files.
type ASTParser interface {
	// ParseFile extracts code structures and calls from a file.
	ParseFile(ctx context.Context, filePath string) ([]NodeDefinition, error)
	
	// GetCallGraph maps caller/callee relationships across the workspace.
	GetCallGraph(ctx context.Context) (map[string][]string, error)
}
```

## Impact

- **Pre-emptive Error Catching**: Catch mismatched function signatures or missing struct fields before invoking the compiler.
- **Accurate Code Rewriting**: Provides agent tools with precise coordinates (start/end lines) of targeted structs or functions.
