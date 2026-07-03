# Orchestrator Project Architecture

## Layer Boundaries (Import DAG)

```
contracts/ (lowest — zero imports from other project packages)
    ↓
kernel/   and   sdk/
    ↓
plugins/  and   modules/
    ↓
cmd/ (highest — entry point, wires everything together)
```

- **contracts/**: Interface definitions only. Must never import any other project package.
- **kernel/**: Core orchestration logic (eventbus, scheduler, planner, runtime). Imports only from `contracts/`.
- **sdk/**: Developer kit for writing plugins. Imports only from `contracts/`.
- **plugins/**: Concrete implementations (agent plugins, provider plugins, tool plugins). Imports from `contracts/` and `sdk/`.
- **modules/**: Business persistence (workspace, config). Imports from `contracts/` and `kernel/`.
- **cmd/**: CLI entry point. Imports from all layers. Wires dependency injection.

## Development Phases

| Phase | Scope | Spec Location |
|-------|-------|---------------|
| Phase 1 | Contracts & Foundation (40 micro-tasks) | `docs/tasks/phase1/` |
| Phase 2 | Kernel Core | `docs/tasks/phase2/` |
| Phase 3 | Plugin System & SDK | `docs/tasks/phase3/` |
| Phase 4 | Provider & Agent Plugins | `docs/tasks/phase4/` |
| Phase 5 | Orchestration Engine | `docs/tasks/phase5/` |
| Phase 6 | CLI & API Polish | `docs/tasks/phase6/` |

## Execution Model
The IDE Agent reads micro-task specs and implements them autonomously.
See `.agents/skills/orchestrator/SKILL.md` for the execution loop protocol.
