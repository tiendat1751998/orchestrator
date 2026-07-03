# RFC-0055: Multi-Workspace

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0013 (Workspace Engine), RFC-0030 (Goal Engine)

## Summary

This RFC specifies the design of the **Multi-Workspace** engine in AEOS. To prevent flat folder context bloat (which confuses the Planner), AEOS treats multiple projects (frontend, backend, infra, mobile) as **Hierarchical Git Submodules / Projects**. The Planner reasons over the dependencies *between* workspaces using high-level API/event contracts, checking out only the active target workspace context during execution.

## Motivation

Loading all files from frontend, backend, and infra repositories simultaneously exceeds LLM context limits and leads to poor planning decisions.
- A change in the backend API does not require the Planner to read the entire mobile app source code.
- By treating workspaces as independent submodules separated by clear interfaces (e.g. OpenAPI / Swagger specs), the Planner handles huge multi-repo projects efficiently.

## Design

### 1. Architectural Placement

The Multi-Workspace engine is part of the `Workspace Engine` in the Kernel, managing the active paths of remote submodules.

```
  Multi-Workspace Root
     ├── Workspace A: Backend (Git Submodule) ◄── [Active Context]
     ├── Workspace B: Frontend (Git Submodule)
     └── Workspace C: Terraform (Git Submodule)
```

---

### 2. Contracts (`contracts/workspace/multi.go`)

```go
package workspace

import "context"

// SubWorkspace represents an independent sub-project.
type SubWorkspace struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	GitRepo   string `json:"git_repo"`
	Language  string `json:"language"`
	APIContract string `json:"api_contract,omitempty"` // OpenAPI or Protobuf file path
}

// MultiWorkspaceManager coordinates submodules.
type MultiWorkspaceManager interface {
	// RegisterSubWorkspace registers a new submodule.
	RegisterSubWorkspace(ctx context.Context, sub SubWorkspace) error
	
	// CheckoutWorkspace loads only the target project workspace into the active compiler context.
	CheckoutWorkspace(ctx context.Context, id string) error
	
	// GetDependencies maps the dependencies between sub-workspaces.
	GetDependencies(ctx context.Context) (map[string][]string, error)
}
```

## Impact

- **Capped Context Size**: The Planner only reads the target workspace files plus the API contracts of its dependencies, avoiding context overflow.
- **Microservices Support**: AEOS can compile and deploy complex multi-repo architectures by executing sequential, localized task DAGs.

## Open Questions

1. **How do we synchronize API contracts between workspaces?**
   - The Workspace Engine compiles API contracts (Protobuf/OpenAPI) and publishes them to a local shared registry, which dependent workspaces read during compilation.
