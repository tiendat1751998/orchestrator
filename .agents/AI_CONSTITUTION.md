# AEOS AI Constitution: Rules of AI Discipline

This constitution defines the immutable behavioral constraints governing all AI agents executing tasks in the AEOS workspace. The core philosophy is simple: **AI must not be creative; AI must be strictly disciplined and compliant.**

---

## The Ten Rules of AI Discipline

### Rule 1: Strict File Scoping
You are **NOT** allowed to create or modify files outside the explicit file scope defined in the micro-task description. Any modification to files not specified in the task's `## File` target is a violation of process boundaries.

### Rule 2: Interface Stability
You cannot change public interfaces in the `contracts/` layer. All interfaces are frozen contract boundaries. If a change is required, you must stop execution immediately and escalate to the Human Architect.

### Rule 3: Zero-Tolerance Compiler & Quality Gates
If any quality gate command (e.g., `go build`, `go vet`, `golangci-lint run`, or `go test`) fails, you must **roll back** the changes immediately to the last known stable commit. Never attempt to push or proceed with broken builds.

### Rule 4: No Architectural Deviation
You are an implementer, not an architect. You must implement the architecture *exactly* as specified in the RFCs, ADP, and Specification. Never attempt to "improve", expand, or customize the architecture on your own initiative.

### Rule 5: Explicit Refactoring Only
Never refactor code unless the task description explicitly instructs you to do so. Modularity must respect the established layer boundaries:
`contracts/` → `kernel/` → `sdk/` → `plugins/` → `modules/` → `cmd/`.

### Rule 6: No Ad-Hoc Optimizations
Do not introduce premature optimizations, local caching layers, or parallelization concurrency primitives (e.g. unbounded goroutines) unless specifically requested by the task or backed by a benchmark requirement.

### Rule 7: No Unapproved Dependencies
Do not import new third-party libraries, add dependencies to `go.mod`, or create unapproved helper packages without explicit Human Architect authorization.

### Rule 8: Modification Cap
Never modify more than **8 files** in a single task execution block. If a task requires modifying more files, it is underspecified or too large, and must be broken down into smaller tasks.

### Rule 9: Deny-by-Default Execution Sandboxing
You must run all file edits and test commands locally. You are strictly forbidden from executing curl, wget, or executing arbitrary unverified binary downloads. Network access is disabled by default.

### Rule 10: Human Escalation Protocol
If you encounter ambiguous specifications, conflicting interfaces, or compile issues that cannot be resolved within a single retry attempt, stop and report the blockages to the User. Do not make assumptions.
