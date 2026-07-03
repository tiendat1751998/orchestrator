# Phase 6: CLI & API Polish — Specifications

This phase implements the user-facing CLI interfaces and API Gateway endpoints to monitor active mission DAGs, stream events, and run reproducibility validations.

---

## Task 6.1: REST/WebSocket API Gateway (`kernel/gateway/`)

- **REST Endpoints**:
  - Implement `/api/v1/health` returning `{"data":{"status":"healthy"}}`.
  - Implement `/api/v1/missions` returning the list of active/completed missions, their DAG status, and trajectories loaded from episodic memory.
- **WebSocket Real-Time Stream**:
  - Implement a WebSocket endpoint `/ws/missions/{id}`.
  - Subscribe to the EventBus (using wildcard subscriptions like `mission.<id>.*`) and stream task state transition events to connected web dashboard clients.
- **Read-Only Time Travel Inspector (RFC-0054)**:
  - Implement frame-by-frame FSM history playback, serving read-only snapshots to the API Gateway to prevent write corruption during debugging.

---

## Task 6.2: Cobra CLI & Final Verification

- **CLI Subcommands**:
  - Implement `orchestrator --version`.
  - Implement `orchestrator config init` and `orchestrator config show` (redacting credentials).
  - Implement `orchestrator agents list` and `orchestrator providers list`.
- **Reproducibility Manifest Validation (RFC-0049)**:
  - Implement `orchestrator verify-reproducibility` to read a Benchmark Manifest and compare execution runs, enforcing Axiom 17 (Scientific Reproducibility).
- **End-to-End Verification**:
  - Run all quality gates: `go fmt`, `go vet`, `golangci-lint run`, and `go test -race ./...`.
  - Verify that the CLI executes correctly and the Web UI displays trajectories and DAG states in real time.
