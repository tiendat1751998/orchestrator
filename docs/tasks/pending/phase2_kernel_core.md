# Phase 2: Kernel Core — Specifications

This phase implements the core Orchestrator engine (`kernel/`) using the interfaces defined in Phase 1. All implementations must strictly adhere to the kernel-based plugin layers, the 17-Axiom ADP Constitution (`docs/adp.md`), and the standard specifications (`docs/specification.md`).

---

## Task 2.1: Config Loader & Structured Logger

- **YAML Config Loader (`kernel/config/`)**:
  - Load configurations from `.orchestrator/settings.yaml` and support environment overrides (`ORCHESTRATOR_LOG_LEVEL`, etc.).
  - Map configurations for local CLI platforms and tools.
- **Structured slog Logger (`kernel/logger/`)**:
  - Implement a standard structured logger using Go's `log/slog` (no zap or third-party loggers directly).
  - **Secret Redaction**: Automatically scan log records and redact sensitive fields (`api_key`, `secret`, `password`) with `[REDACTED]`. Never log secrets or credentials.

---

## Task 2.2: Distributed-Ready EventBus (`kernel/eventbus/`)

Implement the event bus port supporting wildcard subscriptions and event sourcing storage.
- **Wildcard Subscription Matching**:
  - Route topic notifications using dot-notation (e.g. `task.assigned` matched by `task.*`).
  - Use read-write mutexes (`sync.RWMutex`) to guarantee thread safety.
- **Append-Only Event Store Integration (RFC-0008)**:
  - Implement the Event Store adapter using local SQLite storage.
  - Implement **State Memento Snapshots** (every 100 transition events, serialize and snapshot FSM state to SQLite) to ensure replay time remains under 50ms (Axiom 17).
  - Implement **Cryptographic Hash Chaining** (RFC-0028) to sign event records, creating a tamper-proof audit trail.

---

## Task 2.3: Lease-Based Scheduler (`kernel/scheduler/`)

Implement task scheduling that supports concurrency bounds.
- **Worker Registry & Heartbeats**:
  - Track active worker nodes (`scheduler.WorkerInfo`).
  - Implement heartbeats to detect and recover from worker crashes.
- **Task Lease Locking & Preemption**:
  - Implement task leasing to prevent race conditions during parallel branch executions.
  - Allow higher-priority tasks to preempt leases of lower-priority tasks on lock contention.
  - Enforce concurrency bounds using centralized primitives (e.g., `errgroup.SetLimit`), avoiding unbounded raw goroutines (Axiom 3).
