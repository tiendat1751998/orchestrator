# Workspace Rules: Orchestrator

This document defines project-scoped rules and style guidelines that Antigravity agents MUST follow when working in this workspace.

## 1. Autonomous Execution Protocol

**You are a Meta-Orchestrator. You do NOT wait for user prompts between tasks.**

When the user gives a high-level instruction (e.g. "Implement Phase 1"), you MUST:
1. Read the phase index file (`docs/tasks/phase{N}/index.md`) to discover all micro-tasks and their dependency DAG.
2. Check `.agents/state/context.md` to find the current progress (which tasks are already completed).
3. Pick the next pending task in dependency order.
4. Read the micro-task spec file completely.
5. Execute the task by implementing EXACTLY what the spec describes.
6. Run the task's verification commands (listed in the spec's `## Verify` section).
7. If verification passes: log success to `.agents/tasks/success/{task_id}.md`, update `.agents/state/context.md`, and immediately proceed to the next task.
8. If verification fails: retry once with corrections. If it fails again, log the failure and move to the next independent task (skip dependents).
9. **Repeat steps 3-8 until all tasks in the phase are complete.**
10. Only stop and report to the user when the entire phase is done or all remaining tasks are blocked.

### Platform Dispatch (Future)
When external coding platform CLIs are installed, dispatch complex tasks to them:
- `claude` CLI → complex multi-file refactoring, test generation
- `codex` CLI → scaffolding, boilerplate
- `antigravity` CLI → Go systems code
Currently, execute all tasks directly until external platforms are available.

---

## 2. Architectural Boundaries

All code modifications must strictly adhere to the kernel-based plugin layers:
`contracts/` → `kernel/` → `sdk/` → `plugins/` → `modules/` → `cmd/`

- **Contracts Layer (`contracts/`)**: Interface definitions only. Absolutely zero external imports (no external packages, no imports from other project layers). Must remain 100% stable to ensure plugin backward compatibility.
- **Kernel Layer (`kernel/`)**: Core orchestration logic. May only import from `contracts/`. Do not import from `sdk/`, `plugins/`, or `modules/`.
- **SDK Layer (`sdk/`)**: Developer kit for writing plugins. May only import from `contracts/`.
- **Plugins Layer (`plugins/`)**: Concrete implementations of agents, providers, and tools. May only import from `contracts/` and `sdk/`. Plugins must never import other plugins directly.
- **Modules Layer (`modules/`)**: Business persistence and workspace structures. May import from `contracts/` and `kernel/`.
- **CLI/Command Layer (`cmd/`)**: Wires the system together. May import from all layers.

## 3. Coding Standards & Conventions

- **Language**: Go 1.26.
- **Concurrency**: Do not spawn unbounded raw goroutines. Use centralized runtimes (`kernel/runtime`) or bounded concurrency primitives (e.g. `errgroup.SetLimit`).
- **Context**: Propagate `context.Context` as the first parameter to all I/O-bound, tool-bound, and provider-bound operations. Set explicit context deadlines.
- **Logging**: Use `log/slog` for structured logging. Never use `zap` or third-party loggers directly. Never log secrets or credentials.
- **Structs**: Always use named field initialization. Positional struct initialization is strictly forbidden.
- **Mocks**: Put mocks only in `*_test.go` or within `sdk/testing/`. No mock code should exist in production execution paths.
- **Errors**: Wrap errors with context using `fmt.Errorf("...: %w", err)`. Use typed errors defined in `contracts/errors.go` where appropriate.

## 4. Database & Caching Rules
- The runtime uses in-memory data structures and simple file-based serialization (YAML/JSON).
- Local persistent storage (e.g. for feedback loop historical scores) uses SQLite via the standard Go library.
- There are NO external PostgreSQL, Redis, or NATS requirements. Do not introduce them.

## 5. Verification & Quality Gates
Before any work is completed, ensure the following commands run successfully and pass with zero warnings/errors:
- `go fmt ./...`
- `go vet ./...`
- `golangci-lint run ./...`
- `go build ./...`
- `go test ./...`
- `go test -race ./...`

## 6. AI Discipline & Guardrails
All agents executing tasks in this workspace MUST comply with the **[AI Constitution](file://d:/project/orchestrator/.agents/AI_CONSTITUTION.md)**.
- Do not create/modify files outside the task scope (Rule 1).
- Public interfaces in `contracts/` are strictly frozen (Rule 2).
- Zero tolerance for compiler/quality gate failures: roll back on error (Rule 3).
- Do not add ad-hoc optimizations, caches, or unapproved dependencies (Rule 6, Rule 7).
- Limit code modifications to a maximum of 8 files per task execution (Rule 8).

