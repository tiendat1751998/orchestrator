# AEOS Implementation Order & Vertical Slices

This document establishes the frozen sequence of implementation and defines the five core **Vertical Slices** that must be built and verified iteratively. Every task must proceed in this exact sequence to ensure empirical validation at each step (Axiom 16 & 17).

---

## 1. Definition of Frozen Checklist
Before beginning implementation, the following components are officially frozen. No alterations are allowed without Human Architect escalation:

| Component | Status | Verification Reference |
|---|---|---|
| **RFC Catalog** | ✅ Frozen | Mapped from RFC-0000 to RFC-0056 |
| **ADP Constitution** | ✅ Frozen | 17 Axioms registered in `docs/adp.md` |
| **Interfaces Layer** | ✅ Frozen | `contracts/` package models established |
| **Package Layout** | ✅ Frozen | Layer structure (`contracts`, `kernel`, `sdk`, `plugins`) |
| **Import Rules** | ✅ Frozen | Layered boundaries enforced in `docs/engineering/ai_rules.md` |
| **Naming Conventions** | ✅ Frozen | Go standard camelCase, snake_case JSON tags |
| **Error Convention** | ✅ Frozen | Package-level error definitions + context wrapping |
| **Logging Schema** | ✅ Frozen | `log/slog` structured logs with secret redactors |
| **Event Schemas** | ✅ Frozen | JSON-serialized event records matching `specification.md` |
| **Goal Schemas** | ✅ Frozen | `goal.Goal` input constraints and objective structures |
| **Config Schemas** | ✅ Frozen | `.orchestrator/settings.yaml` defaults |

---

## 2. Vertical Slice Roadmaps
Rather than building horizontally layer-by-layer, the project is structured into **five vertical slices** that prove end-to-end functionality at each checkpoint:

```
┌──────────────────────────────────────────────────────────┐
│ Slice 1: Local-First Core (FSM + Scheduler + Git + Tx)  │
└────────────────────────────┬─────────────────────────────┘
                             ▼
┌──────────────────────────────────────────────────────────┐
│ Slice 2: Cognitive Search (Planner + CSP + Pareto + UCB)  │
└────────────────────────────┬─────────────────────────────┘
                             ▼
┌──────────────────────────────────────────────────────────┐
│ Slice 3: Learning & Feedback Loop (Scorecard + EMA Decay) │
└────────────────────────────┬─────────────────────────────┘
                             ▼
┌──────────────────────────────────────────────────────────┐
│ Slice 4: Plugin Engine & WASM Sandbox (Wazero SDK)        │
└────────────────────────────┬─────────────────────────────┘
                             ▼
┌──────────────────────────────────────────────────────────┐
│ Slice 5: Distributed Mission Synchronization (Remote Bus) │
└──────────────────────────────────────────────────────────┘
```

### Slice 1: Local-First Core
* **Scope**: Goal input → FSM State Machine → Scheduler Queue → Workspace Git Transaction → Go build test compiler → Success.
* **Goal**: Prove that we can load a goal, build a dummy plan, lock the workspace, run a command, compile it, and commit it using Git without errors.
* **Required Tasks**: Phase 1 Contracts, Phase 2 Logger & Config, Phase 2 Scheduler, Phase 5 Coordinator.

### Slice 2: Cognitive Search Planner
* **Scope**: Integrate the CSP static filter, Beam search candidates generator, Pareto scoring, and UCB-1 exploration.
* **Goal**: Prove that given a set of technology stack constraints, the planner prunes the search graph, grades candidates mathematically, and selects the optimal path.
* **Required Tasks**: Phase 2 Event Store, Phase 5 Planner, Phase 5 Pareto Scorer.

### Slice 3: Learning & Feedback Loop
* **Scope**: DoD Quality scorecard, EMA weight calculations, and TTL decaying memory.
* **Goal**: Prove that after a task completes, the DoD engine scores it, and the planner learns from the outcome to update template weights.
* **Required Tasks**: Phase 5 DoD Validator, Phase 6 Scorer & Learner.

### Slice 4: Plugin Engine & Sandboxing
* **Scope**: Wazero WASM tool sandboxing, tool schema validation, external CLI process wrapping (Claude Code / Antigravity).
* **Goal**: Prove that external provider CLI processes run sandboxed under OS-level container/user boundaries, while filesystem and compiling tools run inside Wazero WASM with strictly capped CPU/memory limits.
* **Required Tasks**: Phase 3 SDK Plugin, Phase 4 Provider Plugins.

### Slice 5: Distributed Mission Synchronization
* **Scope**: Event Store SQLite synchronization, remote worker nodes heartbeats, distributed lease-locks.
* **Goal**: Prove that multiple runner nodes can pull tasks from the scheduler queue concurrently without lock collisions.
* **Required Tasks**: Phase 2 Distributed Bus, Phase 6 REST/WebSocket Gateways.
