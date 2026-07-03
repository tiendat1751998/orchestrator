# RFC-0013: Workspace Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design and interface of the **Workspace Engine** in the AI Engineering Operating System (AEOS). The Workspace Engine is responsible for codebase contextual awareness. It detects project languages, tracks Git branch states, parses dependency configurations, triggers compilation toolchains (builds), and maps project modules. It provides the Perception tier of the Cognitive Core with raw data about the workspace environment.

## Motivation

Issue 5 from the ChatGPT architecture review identified the lack of a Workspace Engine. A project's codebase environment (filesystems, branches, package configurations, module trees, compile tools) is not a business domain detail — it is a system-level workspace.
Without a Workspace Engine:
- Agents must query Git or dependencies using ad-hoc shell tools, making it impossible for the kernel to trace.
- The Brain Planner has no built-in awareness of the project's language or dependencies (e.g., GORM vs sqlc) before creating plans, leading to invalid steps.
- Compilation and builds are triggered blindly as generic commands instead of standardized lifecycle gates.

## Design

### 1. Architectural Placement

The Workspace Engine is a shared service inside the AEOS Kernel. The Perception tier of the Brain Runtime queries the Workspace Engine to construct semantic and procedural memory contexts:

```
                  Brain Runtime (Perception Tier)
                                │
                                ▼
  ┌────────────────────────────────────────────────────────┐
  │                 Workspace Engine                       │
  │                                                        │
  │  ┌────────────────┐  ┌──────────────────────────────┐  │
  │  │ Git Analyzer   │  │ Language Detector            │  │
  │  └────────────────┘  └──────────────────────────────┘  │
  │  ┌────────────────┐  ┌──────────────────────────────┐  │
  │  │ Build Runner   │  │ Dependency Parser            │  │
  │  └────────────────┘  └──────────────────────────────┘  │
  └─────────────────────────────┬──────────────────────────┘
                                │
                                ▼
                    Physical Codebase Files
```

---

### 2. Contracts (`contracts/workspace/`)

```go
// contracts/workspace/workspace.go
package workspace

import (
	"context"
	"time"
)

// ProjectMetadata represents parsed codebase structures.
type ProjectMetadata struct {
	RootPath     string            `json:"root_path"`
	Language     string            `json:"language"` // "go", "python", "typescript"
	Framework    string            `json:"framework,omitempty"`
	Dependencies map[string]string `json:"dependencies"` // package -> version
	Modules      []string          `json:"modules,omitempty"`
}

// GitDetails represents active Git repository states.
type GitDetails struct {
	Branch      string   `json:"branch"`
	CommitSHA   string   `json:"commit_sha"`
	DirtyFiles  []string `json:"dirty_files,omitempty"`
	Ahead       int      `json:"ahead"`
	Behind      int      `json:"behind"`
}

// BuildResult represents compile outcome metrics.
type BuildResult struct {
	Success     bool          `json:"success"`
	Output      string        `json:"output,omitempty"`
	Errors      []string      `json:"errors,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// WorkspaceEngine provides unified access to codebase properties.
type WorkspaceEngine interface {
	// Detect analyzes the directory to map language and dependencies.
	Detect(ctx context.Context) (*ProjectMetadata, error)
	
	// GitStatus queries current git branch, dirty files, and diff details.
	GitStatus(ctx context.Context) (*GitDetails, error)
	
	// Build triggers compilation toolchains (e.g. "go build", "npm run build").
	Build(ctx context.Context) (*BuildResult, error)
	
	// Clean resets build caches or temporary files.
	Clean(ctx context.Context) error
}
```

---

### 3. Implementation Details (`kernel/workspace/`)

The Workspace Engine implements automatic detection by looking for signature files:

* **Go**: Parses `go.mod` to extract module paths and Go compiler targets.
* **NodeJS**: Parses `package.json` to extract frameworks (NextJS, React, Express) and lockfile dependencies.
* **Python**: Parses `requirements.txt`, `Pipfile`, or `pyproject.toml`.

```go
// kernel/workspace/detector.go
package workspace

import (
	"context"
	"os"
	"path/filepath"

	"github.com/tiendat1751998/orchestrator/contracts/workspace"
)

type engine struct {
	rootDir string
}

func NewWorkspaceEngine(rootDir string) workspace.WorkspaceEngine {
	return &engine{rootDir: rootDir}
}

func (e *engine) Detect(ctx context.Context) (*workspace.ProjectMetadata, error) {
	meta := &workspace.ProjectMetadata{
		RootPath:     e.rootDir,
		Dependencies: make(map[string]string),
	}

	// 1. Detect Go
	if _, err := os.Stat(filepath.Join(e.rootDir, "go.mod")); err == nil {
		meta.Language = "go"
		e.parseGoMod(meta)
	} else if _, err := os.Stat(filepath.Join(e.rootDir, "package.json")); err == nil {
		meta.Language = "typescript"
		e.parsePackageJSON(meta)
	}

	return meta, nil
}

func (e *engine) parseGoMod(meta *workspace.ProjectMetadata) {
	// Read and parse go.mod dependencies using standard strings manipulation
}

func (e *engine) parsePackageJSON(meta *workspace.ProjectMetadata) {
	// Unmarshal package.json dependencies
}

func (e *engine) GitStatus(ctx context.Context) (*workspace.GitDetails, error) {
	// Query local git status using git command execution adapters (CLIProvider fallback)
	return nil, nil
}

func (e *engine) Build(ctx context.Context) (*workspace.BuildResult, error) {
	// Automatically run corresponding compiler commands:
	// Go -> "go build ./..."
	// JS -> "npm run build"
	return nil, nil
}

func (e *engine) Clean(ctx context.Context) error {
	return nil
}
```

## Impact

- **Language-Aware Planning**: Before generating a task DAG, the Planner calls `WorkspaceEngine.Detect()` to query the codebase metadata. If it detects `go.mod`, it matches Go-specific handlers and templates.
- **Unified Build Gate**: Instead of agents running arbitrary compile commands, the Orchestrator invokes `WorkspaceEngine.Build()` as a standardized quality check.

## Open Questions

1. **How does Build handle complex custom build scripts?**
   - We will support a local `.aeos/workspace.yaml` configuration file where users can define custom build, test, and clean command overrides. If present, the Workspace Engine uses these definitions; otherwise, it falls back to language autodetect.
