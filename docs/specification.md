# 📜 AEOS standard specification v1.0

This specification defines the formal standards, interface protocols, schema payloads, and lifecycle state transitions of the **AI Engineering Operating System (AEOS)**. It ensures language-independent compatibility across kernels, planners, and benchmark runners.

---

## 1. The Mission Lifecycle (FSM Protocol)

A `Mission` is the root execution context in AEOS, modeled as a deterministic Finite State Machine (FSM).

```
   [Created] ──► [Planning] ──► [Running] ──► [Validating] ──► [Completed]
                                    ▲              │
                                    └────── [Failed]
```

### State Transitions
1. **Created**: Mission is initialized with raw goals, environment variables, and target constraints.
2. **Planning**: The Planner parses goals, prunes the search space, and produces an execution DAG.
3. **Running**: The Execution Runtime schedules and executes task nodes sequentially or in parallel.
4. **Validating**: The DoD Engine runs compiler, linter, test, and security gates to grade quality.
5. **Completed**: Transitioned only when all DoD criteria are satisfied.
6. **Failed**: Triggered if a task fails or validation fails. Auto-initiates the Replanner to mutate the DAG.

---

## 2. Event Store Schema

Every transition and side-effect must be persisted in an append-only Event Store using the following schema payload format:

```json
{
  "$schema": "https://aeos.dev/schemas/event.json",
  "event_id": "evt_01j7h8v9a2",
  "mission_id": "mis_9f3a8b2d",
  "timestamp": "2026-07-03T11:48:58Z",
  "event_type": "TaskExecuted",
  "sequence_number": 42,
  "payload": {
    "task_id": "tsk_compile_backend",
    "executor_agent": "age_backend_go",
    "exit_code": 0,
    "stdout_hash": "sha256:e3b0c442...",
    "mutated_files": ["kernel/main.go", "kernel/main_test.go"]
  }
}
```

---

## 3. The Planner Interface Contract

Planners must implement the following 4 core execution methods. The Planner reads the Knowledge Graph and outputs a standard Plan DAG.

```go
type Planner interface {
    // Plan decomposes a Goal into objectives and generates candidate DAGs.
    Plan(ctx context.Context, g Goal) ([]DAG, error)
    
    // Score grades plans based on Quality, Cost, Time, and Risk metrics.
    Score(ctx context.Context, candidates []DAG) (DAG, error)
    
    // Explain outputs the contrastive mathematical scoring rationales.
    Explain(ctx context.Context, chosen DAG, candidates []DAG) (string, error)
    
    // Learn updates success rates and registers failure association edges.
    Learn(ctx context.Context, history TransitionRecord) error
}
```

---

## 4. Workspace Transactions & Provenance

### Transactions
- **Stage**: Prior to executing code mutations, the system stashes dirty files: `git stash --include-untracked` or creates branch `aeos-tx-[mission-id]`.
- **Commit**: On successful validation, the changes are committed to the target branch.
- **Rollback**: On validation failure, the branch is reset: `git reset --hard` to restore workspace integrity.

### Provenance Lineage
Every generated binary, container image, or deployment manifest must inject standard metadata:
```json
{
  "provenance": {
    "aeos_version": "1.0.0",
    "mission_id": "mis_9f3a8b2d",
    "git_commit": "sha1:9f3a8b2d",
    "dod_score": 0.98,
    "build_timestamp": "2026-07-03T11:48:58Z"
  }
}
```

---

## 5. Benchmark Manifest Schema

Every performance test run must output a standardized manifest for scientific reproducibility:

```yaml
manifest_version: "aeos/bench/v1"
experiment_id: "uuid-v4-hash"
environment:
  cpu_architecture: "x86_64"
  system_ram_gb: 32
  toolchain_versions:
    go: "1.26"
    docker: "25.0.3"
parameters:
  model_provider: "gemini-1.5-pro"
  temperature: 0.0
  seed: 42
  workspace_hash: "sha256:e3b0c44..."
results:
  compile_success_rate: 0.98
  dod_verification_score: 0.94
  latency_ms: 1250
```
