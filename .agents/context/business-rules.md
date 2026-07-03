# Business Rules

## 1. Micro-Task Execution Rules
- Each micro-task spec file is the **single source of truth** for what code to write.
- Implement EXACTLY what the spec describes — no additions, no modifications, no "improvements".
- Every task has a `## Verify` section — these commands MUST pass before marking the task as complete.
- Every task has a `## Checklist` — all items must be satisfied.

## 2. Dependency Rules
- Tasks form a Directed Acyclic Graph (DAG) defined in each phase's `index.md` Mermaid diagram.
- A task CANNOT be executed until ALL its predecessors are marked as SUCCESS.
- If a task FAILS after retry, all tasks that depend on it are marked as BLOCKED.

## 3. State Persistence
- Completed tasks are tracked in `.agents/state/context.md` and `.agents/tasks/success/`.
- In-progress tasks are tracked in `.agents/tasks/inprocess/`.
- State must survive session resets (context window overflow).

## 4. Quality Gates
Before marking any phase as complete, ALL of these must pass:
```bash
go fmt ./...
go vet ./...
golangci-lint run ./...
go build ./...
go test ./...
go test -race ./...
```
