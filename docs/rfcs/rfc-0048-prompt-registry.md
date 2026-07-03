# RFC-0048: Prompt Registry

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0004 (Context Engine), RFC-0005 (Memory Model)

## Summary

This RFC specifies the design of the **Prompt Registry** in AEOS. Prompts are treated as version-controlled code assets. The Prompt Registry stores, hashes, diffs, and manages prompt templates inside the local repository, embedding them into the Go binary using `go:embed`. This guarantees synchronicity between code versions and LLM templates.

## Motivation

Prompts are code. Storing prompts as external database entities separates them from the codebase version control system (Git).
- This leads to runtime mismatches (e.g. running Go binary v1.2 which expects input parameter $X$, while the database prompt has mutated to v1.5 and outputs $Y$).
- Versioning prompts in Git alongside code guarantees that rolling back Git automatically rolls back prompts to compatible versions.

## Design

### 1. Architectural Placement

Prompts are stored as text assets in `/assets/prompts/` and read via the `PromptRegistry` service in the Kernel.

```
  Go Binary (embedded /assets/prompts/) ──► [PromptRegistry] ──► Parse variables ──► LLM API
```

---

### 2. Contracts (`contracts/context/prompt.go`)

```go
package context

import (
	"context"
)

// PromptTemplate represents a versioned text prompt.
type PromptTemplate struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Content   string `json:"content"`
	SHA256    string `json:"sha256"`
}

// PromptRegistry manages prompt assets.
type PromptRegistry interface {
	// GetTemplate retrieves a template by ID.
	GetTemplate(ctx context.Context, templateID string) (*PromptTemplate, error)
	
	// Render compiles variables into the template string.
	Render(ctx context.Context, templateID string, vars map[string]interface{}) (string, error)
}
```

## Impact

- **Zero-Drift Execution**: Prompt templates are locked to the compiled binary version, preventing unexpected API parsing breaks in production.
- **Auditable Prompts**: The Event Store logs contain the SHA256 of the prompt template used, making execution logs 100% reproducible.

## Open Questions

1. **How do we support prompt A/B testing?**
   - The registry supports loading local override files (e.g. `assets/prompts/overrides.yaml`). If present, the system runs local experiments without modifying the core binary template assets.
