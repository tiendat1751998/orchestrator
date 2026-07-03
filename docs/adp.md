# ⚖️ AEOS Architecture Decision Principles (ADP)

This document serves as the **Constitution** of the AI Engineering Operating System (AEOS). All future RFCs, code implementations, and refactoring efforts must strictly comply with these 15 immutable design principles.

---

## 📜 The 15 Immutable Axioms

### 1. AI is Never the Source of Truth
- **Axiom**: LLMs and AI agents are content-generation utilities and execution agents. They do not define system correctness.
- **Implication**: Correctness is exclusively validated by deterministic software verification gates (compiler, linter, tests, policy engine) running in the Go kernel.

### 2. Event Store is the Only Immutable History
- **Axiom**: Every state transition, agent action, and scheduling decision must be recorded as an event in a append-only Event Store.
- **Implication**: The current state of the system is derived by replaying events. History cannot be modified, deleted, or bypassed.

### 3. Mission is the Only Aggregate Root
- **Axiom**: All tasks, events, decisions, snapshots, artifacts, and metrics must refer directly to a parent `MissionID` to form a single domain aggregate.
- **Implication**: Operations must never execute across missions without explicit context bridging, ensuring isolated, repeatable, and clean execution environments.

### 4. Zero Cyclic Dependencies
- **Axiom**: Package layers must maintain strict hexagonal boundaries: `contracts/` ──► `kernel/` ──► `sdk/` ──► `plugins/` ──► `cmd/`.
- **Implication**: Go packages must compile cleanly with zero circular imports. Inner layers (e.g. `contracts`) must never import outer layers (e.g. `kernel` or `plugins`).

### 5. Plugins Depend on Contracts, Never on Kernel Internals
- **Axiom**: Extension plugins (providers, tools, agents) communicate with the system exclusively through stable interfaces defined in `contracts/`.
- **Implication**: Plugin implementations must never access private struct fields or database models of the `kernel/` packages.

### 6. Everything is Replayable
- **Axiom**: Any execution step within a mission must be reproducible up to step N under the exact same starting snapshot conditions.
- **Implication**: Randomness, active time queries, and external network dependencies must be intercepted, sandboxed, or mocked during replays.

### 7. Every Side Effect Emits an Event
- **Axiom**: Actions that modify the external world (writing files, running bash commands, deploying code) must publish a corresponding event to the EventBus.
- **Implication**: The Event Store provides a complete cryptographic audit trail of all physical actions taken by agents.

### 8. Local-First Before Distributed
- **Axiom**: AEOS must run out-of-the-box on a single local development machine with zero external cluster dependencies.
- **Implication**: Remote worker node distribution (RFC-0052) is an optional extension. The core system compiles to a single local binary controlled by the `antigravity` CLI.

### 9. Git is the Workspace Transaction Layer
- **Axiom**: Native Git mechanisms handle all workspace mutations, staging, and rollbacks.
- **Implication**: Before executing generative tasks, AEOS creates a temporary Git checkpoint (stash/branch). If validation fails, AEOS performs `git reset --hard` to guarantee workspace integrity, avoiding complex virtual file mounts.

### 10. SQLite is the Default Persistence
- **Axiom**: SQLite handles all structural databases, timelines, and knowledge graphs.
- **Implication**: Do not introduce external Postgres, Redis, or Vector databases into the core local system. They must remain optional plugin modules.

### 11. Deterministic Planning
- **Axiom**: The Planner coordinates task dependency DAGs based on rigid constraints, capabilities, and templates.
- **Implication**: AI can suggest plans, but the final DAG scheduling and transition rules are validated deterministically by Go kernel logic.

### 12. Security is Deny-by-Default
- **Axiom**: Agents have zero initial capabilities. Permissions to read directories, execute commands, or call APIs must be granted explicitly.
- **Implication**: Sandboxes isolate processes and capability tokens authorize operations at runtime.

### 13. New Features Require Fitness Gains
- **Axiom**: Do not add new cognitive engines or features without a measurable improvement in an architectural fitness function (e.g., compile success, recovery time, token cost).
- **Implication**: Unmeasurable additions are rejected to prevent feature inflation.

### 14. Reality Overrides Theoretical Design
- **Axiom**: Feedback from executable integration benchmarks (the 4 vertical slices) overrides paper specifications.
- **Implication**: If a design spec makes local development or testing excessively complex, prune the spec to match implementation reality.

### 15. Simplicity Beats Cleverness
- **Axiom**: Code readability, stability, and compile-time validation are prioritized over complex meta-programming or dynamic runtime tricks.
- **Implication**: Keep the codebase boring. Use standard Go conventions, central runtimes, and clear interface layers.

### 16. Executable Empirical Evidence
- **Axiom**: Every architectural claim, optimization target, or planning hypothesis must eventually become an executable benchmark or integration test.
- **Implication**: Unmeasurable design claims are treated as unproven hypotheses. Architectural evolution must be driven by fitness metric regressions, not subjective design preferences.

### 17. Scientific Reproducibility
- **Axiom**: Every benchmark, planner evaluation, and architecture claim must be reproducible by an independent engineer. Each experiment must produce a complete execution manifest (software versions, configuration, seeds, inputs, environment, outputs), and benchmark suites must include hidden evaluation tasks to prevent overfitting.
- **Implication**: No experimental claim is valid without an attached Benchmark Manifest. Benchmark suites must maintain a strict split between visible training goals and hidden test goals to prevent benchmarking gaming or local model overfitting.
