# RFC-0053: Artifact Lineage

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0008 (Event Model), RFC-0013 (Workspace Engine)

## Summary

This RFC specifies the design of the **Artifact Lineage** module in AEOS. By leveraging the parent `MissionID` aggregate root, the engine injects cryptographic metadata (mission UUID, git commit hash, active prompt hashes, and policy versions) directly into compiled binaries, Docker images, and deployment manifests. This allows tracing production errors back to their exact prompt-agent origin.

## Motivation

When a microservice crashes in production, it is extremely difficult to trace the error back to the generative AI prompt, model, or policy that produced it.
- This results in debugging deadlocks.
- By injecting lineage metadata directly during compilation, developers can audit the exact generative context that led to the code change.

## Design

### 1. Architectural Placement

Lineage injection is handled inside the `Build Runner` stage of the Workspace Engine, reading metadata from the active Mission aggregate.

```
  Go Compilation ──► Inject Lineage Labels ──► Target Binary (with BuildInfo labels)
```

---

### 2. Contracts (`contracts/workspace/lineage.go`)

```go
package workspace

import "context"

// ProvenanceMetadata represents the injected lineage variables.
type ProvenanceMetadata struct {
	MissionID   string `json:"mission_id"`
	GitCommit   string `json:"git_commit"`
	PromptHash  string `json:"prompt_hash"`
	PolicyHash  string `json:"policy_hash"`
	BuilderID   string `json:"builder_id"`
}

// LineageEngine manages metadata injection.
type LineageEngine interface {
	// InjectProvenance writes lineage properties into the target binary or image.
	InjectProvenance(ctx context.Context, targetPath string, provenance ProvenanceMetadata) error
	
	// ReadProvenance extracts lineage properties from a compiled binary or image.
	ReadProvenance(ctx context.Context, targetPath string) (*ProvenanceMetadata, error)
}
```

## Impact

- **Cryptographic Audit Trails**: Production Docker containers can be traced back to the exact mission log in the Event Store.
- **Improved Policy Enforcement**: System administrators can verify if deployed code violates new policy standards by reading the binary's provenance metadata.

## Open Questions

1. **How do we inject metadata into compiled binaries?**
   - For Go binaries, we use `-ldflags "-X main.AEOSVersion=1.0 -X main.MissionID=..."` during compilation to embed properties directly.
