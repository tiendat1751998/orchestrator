# RFC-0014: Quality Engine & Verification Pipelines

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design of the **Quality Engine & Verification Pipelines** in AEOS. The Quality Engine manages multi-stage automated validation gates (compiler, linter, tests, static analysis, security scan) to verify the correctness of agent-generated code before it is committed to the workspace.

## Motivation

AI code generation requires strict, automated guardrails to prevent introducing syntax errors or security vulnerabilities.
- We cannot rely on LLM self-evaluation to verify correctness.
- The verification pipeline must execute deterministically inside the Go kernel, utilizing local OS commands and sandbox environments.

## Design

### 1. Architectural Placement

The Quality Engine resides in the `Execution Runtime`, executing after code-modification tasks complete.

```
  Modified Workspace ──► [Quality Engine] ──► Compiler ──► Linter ──► Unit Tests ──► Pass/Fail
```

---

### 2. Contracts (`contracts/quality/verification.go`)

```go
package quality

import (
	"context"
)

// PipelineResult represents the outcome of a validation run.
type PipelineResult struct {
	Success bool     `json:"success"`
	Logs    []string `json:"logs,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// VerificationPipeline executes sequential validation checks.
type VerificationPipeline interface {
	// Verify runs the compiler, linter, and tests on the active workspace.
	Verify(ctx context.Context) (*PipelineResult, error)
}
```

## Impact

- **Automated Guardrails**: Generative tasks only complete when `Verify()` passes with zero errors, preventing broken code from accumulating.
- **Unified Testing**: Compilers, linters, and unit tests are wrapped in a single, predictable Go interface.
