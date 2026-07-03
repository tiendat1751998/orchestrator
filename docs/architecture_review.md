# 🏗️ AI Engineering Operating System (AEOS) Architectural Review

> [!IMPORTANT]
> The detailed technical design and interface specifications of the system have been standardized across **30 RFCs (RFC-0000 to RFC-0029)**.
> Before implementing any code, please refer to the **[AEOS RFCs Catalog](rfcs/README.md)**.

## System Goal
> Build a system that serves as both a **task coordinator (Orchestrator)** and a **second brain** to automate complex tasks. AI providers (Gemini, Claude, Antigravity CLI...) **act solely as content-generation tools**, not the brain. The system makes decisions independently using deterministic Go kernel logic. The system is designed to run reliably for **10+ years**.

---

## 🧠 Core Philosophy: AI is a Tool, Not the Brain

> [!CAUTION]
> **Rule #1**: The system does NOT rely on AI models to make orchestration decisions.
> - If an AI provider fails: the system continues running (using templates, rules, and cache).
> - If an AI hallucinates: the system detects it and rejects the response (validation rules).
> - If an AI API changes: only the specific provider plugin is updated, leaving the core brain untouched.

```
┌─────────────────────────────────────────────────────┐
│             AEOS KERNEL (SECOND BRAIN)              │
│                                                     │
│  ┌────────────────┐  ┌──────────────────────────┐   │
│  │ DETERMINISTIC  │  │ KNOWLEDGE PLATFORM       │   │
│  │ BRAIN (Go)     │  │ (Memory & Learning)      │   │
│  │                │  │                          │   │
│  │ • Rule Engine  │  │ • SQLite Graph Store     │   │
│  │ • DAG Planner  │  │ • Pattern Miner          │   │
│  │ • Policy Eval  │  │ • Decision History       │   │
│  │ • DoD Engine   │  │ • Template Library       │   │
│  └────────┬───────┘  └──────────┬───────────────┘   │
│           │                     │                   │
│  ┌────────▼─────────────────────▼────────────────┐  │
│  │           ORCHESTRATION CORE                  │  │
│  │  • Task DAG Scheduler                         │  │
│  │  • Agent/CLI Coordinator                      │  │
│  │  • Result Aggregator & Truth Pipeline         │  │
│  └────────────────────┬──────────────────────────┘  │
│                       │                             │
│  ┌────────────────────▼──────────────────────────┐  │
│  │           TOOL LAYER (AI = 1 tool)             │  │
│  │                                                  │  │
│  │  ┌──────┐ ┌─────┐ ┌──────┐ ┌───────┐ ┌─────┐  │  │
│  │  │ AI   │ │ Git │ │ Shell│ │Docker │ │Files│  │  │
│  │  │Prov. │ │     │ │      │ │       │ │     │  │  │
│  │  └──────┘ └─────┘ └──────┘ └───────┘ └─────┘  │  │
│  └────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

| # | Principle | Meaning |
|---|---|---|
| 1 | **AI is a tool, not the brain** | The Planner and Orchestrator make decisions using Go logic. AI is only called when content generation (generating code, writing docs, reviewing text) is needed. |
| 2 | **Deterministic first, AI second** | All orchestration decisions (agent selection, task ordering, retries) must be based on rules + data, not raw LLM outputs. |
| 3 | **Self-learning Second Brain** | The system accumulates knowledge over time — patterns, templates, decision histories — to grow smarter without repeating mistakes. |
| 4 | **CLI/Agent Agnostic** | The system orchestrates any agent or CLI tool (Antigravity, Claude CLI, Codex, shell scripts, custom binaries) — they are all executors. |

---

## ✅ Strengths of Current Architecture

| Component | Evaluation |
|---|---|
| **`kernel/`** — Core separation | ✅ Excellent. The kernel-based design mirrors POSIX OS kernels, allowing external plugin replacements without mutating core orchestration logic. |
| **`contracts/`** — Interface-driven | ✅ Outstanding. Key to a 10-year stable lifecycle. Every agent, provider, and tool only interacts through static interfaces. |
| **`plugins/`** — Extensibility | ✅ Great. Allows adding or removing agents, providers, and tools dynamically. |
| **`sdk/`** — Developer Kit | ✅ Good. Empowers third-party developers or yourself to build new plugins with base helper structures. |
| **`modules/`** — Business domains | ✅ Good. Decouples business workspace logic from low-level infrastructure. |

---

## 🏛️ AEOS Architectural Specification

AEOS is organized into 4 independent runtimes and 3 core shared services. All system operations and data schemas are designed using Domain-Driven Design (DDD) with **Mission as the Aggregate Root**.

---

### 1. Kernel Decomposition & Runtime Responsibilities (RFC-0001, RFC-0013 & RFC-0016)

```
                              Kernel (FSM: Created→Booting→Running→Stopped)
                                │
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
    4 Runtimes              3 Shared Services       Infrastructure
         │                      │                      │
   ┌─────┼─────┬─────┐    ┌─────┼─────┼─────┐          ┌────┼────┐
   ▼     ▼     ▼     ▼    ▼     ▼     ▼     ▼          ▼    ▼    ▼
Exec-  Brain Plugin Obser-Knowl- History Work-         Config Logger EventBus
ution  Run-  Run-   vation edge  Timeline space
Run-   time  time   Run-   Engine        Engine
time                time
```

* **Execution Runtime (`kernel/execution/`)**: Handles task execution. Contains the `Executor`, thread-limited `Worker Pool`, `Task Dispatcher`, `Result Collector`, OS `Process Manager`, and `Resource Manager` (monitoring host metrics).
* **Brain Runtime (`kernel/brain/`)**: The decision-making core. Contains the `Decision Engine` (rule-based Go logic), `Planning Engine` (DAG planner), `Policy Engine` (safety evaluator), and `Context Engine` (prompt compression).
* **Plugin Runtime (`kernel/plugin/`)**: Manages the lifecycle of plugins (Init, Start, Stop) using a generic `Registry[T plugin.Plugin]`.
* **Observation Runtime (`kernel/observation/` - RFC-0016)**: **Isolates learning from execution**. Runs in the background to monitor AST changes, collect git commits, build logs, test results, benchmarks, and telemetry metrics to update the Knowledge Platform asynchronously.
* **Workspace Engine (`kernel/workspace/` - RFC-0013)**: Tracks physical codebase environments (analyzing Git, detecting languages/dependencies like `go.mod`, and triggering builds/cleans).

---

### 2. State Machines & Mission Aggregate Root (RFC-0000 & RFC-0008)

All long-lived entities are managed via Finite State Machines (FSMs):
* **Mission FSM**: `Created` → `Planning` → `Scheduled` → `Running` → `Reviewing` → `Learning` → `Completed/Failed/Cancelled`.
* **Task FSM**: `Pending` → `Assigned` → `Running` → `Completed/Failed/Skipped`.
* **Mission as the Aggregate Root**: To guarantee 100% deterministic replays and audit trails, all execution logs, artifacts, and decisions refer directly to a parent Mission ID:
  ```
  Mission ID (Aggregate Root)
     ├── Tasks (Sub-task specifications)
     ├── Events (Event Store logs of this mission)
     ├── Decisions (Decision Log - architectural choices)
     ├── Knowledge (Discovered semantic relationships)
     ├── Artifacts (Files and testing outputs generated)
     ├── Metrics (Cost $, latency, hardware usage)
     ├── Policies (Versioned policy rules enforced)
     └── Snapshots (Workspace Snapshot: Git Commit, dirty files, dependencies)
  ```
* **Event Store**: All FSM transitions trigger an `OnTransition` callback to log immutable entries in `history.Timeline` and broadcast on `event.Bus`.
* **Event Sourcing Recovery**: Upon crash, a `StateReconstructor` loads the `event.Event` chain from SQLite to restore passive FSM states without repeating side-effects.

---

### 3. Memory Compartmentalization & SQLite Layout (RFC-0003 & RFC-0005)

* **Working Memory (`contracts/memory/`)**: In-memory map (buffered to disk periodic snapshots) isolated per **Mission ID**. Destroyed once the mission terminates.
* **Knowledge Graph (`contracts/knowledge/`)**: Persistent semantic graph (Nodes & Edges). Supports Full-Text Search (FTS), tag matching, graph traversal, and offline vector search.
* **Artifact Store (`contracts/artifact/`)**: Manages physical output files (source code, test reports, logs) under the mission's directory.
* **History Timeline (`contracts/history/`)**: Immutable event timeline, retrieved sequentially via `EntryIterator`.

---

### 4. Decoupling Providers (API/CLI) from Runtime (RFC-0007)

* **Translation Only**: Providers only translate requests into API payloads or CLI strings and parse structured responses. They hold no execution state.
* **CLI Stream Parser**: Processes stdout/stderr chunks in real-time to match tools or detect errors.

---

### 5. Security & Boundary Sandboxing (RFC-0012)

* **Capability Guard**: All I/O, process spawns, and network operations are evaluated by `VerifyPermission` rules.
* **Path Traversal Protection**: Enforces clean absolute paths (`filepath.Clean`) and matches folder prefix bounds.
* **Docker Sandboxing**: Executes unverified test scripts or builds inside transient Docker containers.

---

### 6. 5-Tier Cognitive Architecture (RFC-0010)

```
  ┌────────────────────────────────────────────────────────┐
  │ 1. Perception Tier : AST, Git diffs, Execution logs    │
  ├────────────────────────────────────────────────────────┤
  │ 2. Memory Tier     : Working, Episodic, Semantic,      │
  │                      Procedural Memory                 │
  ├────────────────────────────────────────────────────────┤
  │ 3. Reasoning Tier  : Planner, Meta-Thinking (Self-Check)│
  ├────────────────────────────────────────────────────────┤
  │ 4. Learning Tier   : Reflection, Experience, Pattern    │
  ├────────────────────────────────────────────────────────┤
  │ 5. Action Tier     : Scheduler, Sandbox execution      │
  └────────────────────────────────────────────────────────┘
```

* **4 Memory Dimensions**:
  * **Working Memory**: Active variables and mission context.
  * **Episodic Memory**: Chronicles success/failure episodes (e.g. "Postgres Transaction Deadlock in Mission #25"). Next planning cycle alerts the planner to avoid matching failure patterns.
  * **Semantic Memory**: Graph relations (e.g., Gin uses HTTP Router).
  * **Procedural Memory (Skills)**: Codebase rules of thumb (e.g., Gin is preferred over GORM; Redis keys must match specific formats).
* **Experience Engine**: Accumulates technology stack scores (e.g., Gin + sqlc + Redis) to suggest optimal setups.
* **Pattern Engine**: Mines design patterns (Saga, Outbox, DDD Repository) from success directories.
* **Meta-Thinking**: The Planner checks generated plans for redundancies or deadlock cycles before scheduling.

---

## 🗺️ 10-Year Evolutionary Roadmap (57 RFCs Blueprint)

To guarantee architectural stability for the next 10 years, AEOS is planned across 57 RFCs:

### Phase 0 & 1: Foundation (RFC-0000 to RFC-0007)
* **RFC-0000**: Everything is State Machine.
* **RFC-0001**: Kernel Architecture (4 Runtimes).
* **RFC-0002**: Brain Architecture (Engines).
* **RFC-0003**: Knowledge Engine as Semantic Graph.
* **RFC-0004**: Context Engine & prompt compression.
* **RFC-0005**: 4 Memory Dimensions (Working, Episodic, Semantic, Procedural).
* **RFC-0006**: Generic Plugin Registry & lifecycle.
* **RFC-0007**: Raw Provider and Runtime Separation.

### Phase 2: Core Extensions & Security (RFC-0008 to RFC-0013)
* **RFC-0008**: Event Sourcing and Event Store.
* **RFC-0009**: Resource Manager (CPU/RAM metrics).
* **RFC-0010**: Cognitive Layer (EMA, Decision Log, Skill Tree).
* **RFC-0011**: Kubernetes-Style Scheduler.
* **RFC-0012**: Security Capability Guards.
* **RFC-0013**: Workspace Engine.

### Phase 3 & 4: Self-Improvement & Advanced Integration (Drafts, RFC-0014 to RFC-0056)
* **RFC-0014**: Quality Engine & Verification Pipelines.
* **RFC-0015**: Definition of Done (DoD) Engine.
* **RFC-0016**: Observation Runtime & Telemetry Collector.
* **RFC-0017**: Workspace Snapshots & Deterministic Replay.
* **RFC-0018**: Cost Engine & Planning Cache.
* **RFC-0019**: ADR as First-Class Object & Policy Versioning.
* **RFC-0020**: Agent Capabilities & Skill Tree Metrics.
* **RFC-0021**: Vector Search & Local Graph Embeddings.
* **RFC-0022**: Multi-Agent Collaboration & Event Routing.
* **RFC-0023**: Process Sandboxing & Container Isolation.
* **RFC-0024**: API Gateways (gRPC, REST, WebSockets).
* **RFC-0025**: Dependency Tree & AST Parser.
* **RFC-0026**: Benchmark & Runtime Profiling.
* **RFC-0027**: Resilience, Circuit Breakers & Backoffs.
* **RFC-0028**: Audit Trail & Cryptographic Event Logs.
* **RFC-0029**: Web UI & Mission Control Protocol.
* **RFC-0030**: Goal Engine (Converts high-level goals to concrete objectives, constraints, and milestones).
* **RFC-0031**: World Model (Enables the planner to reason over codebase objects rather than raw prompt strings).
* **RFC-0032**: Skill Graph (Models technology dependencies as a graph instead of a simple tree).
* **RFC-0033**: Confidence Engine (Spawns reviewers or helper agents if confidence score drops below threshold).
* **RFC-0034**: Advanced Quality Engine (Detailed multi-dimensional scorecards replacing simple pass/fail checks).
* **RFC-0035**: Capability Graph & Agent Competency Model (Separating skill technology from task capability).
* **RFC-0036**: Mission Simulation & Dry-run Planner (Dry-running FSM transitions to calculate cost and locks upfront).
* **RFC-0037**: Adaptive Recovery & Replanning Engine (Mutating plan DAGs on task failures rather than retrying).
* **RFC-0038**: Resource Planning & Budget Estimation (Budget estimation and plan splitting bounds).
* **RFC-0039**: Evolution Engine (Updating Planner parameters based on historical template successes).
* **RFC-0040**: Intent Engine (Analyzes business intent constraints like MVP vs. Enterprise budgets).
* **RFC-0041**: Product Memory (Saves business/product ontologies like Voucher or Search patterns).
* **RFC-0042**: PM Runtime (Requirement translation and business criteria DoD audits).
* **RFC-0043**: Release Intelligence (Gathers canary error logs and triggers automatic rollbacks).
* **RFC-0044**: Economic Engine (Optimizes planning for expected ROI and bug risk factors).
* **RFC-0045**: Digital Workforce (CTO-like orchestration of virtual employees with competency ratings).
* **RFC-0046**: Execution Graph Manager (Pausing, versioning, merging and migrating running DAGs).
* **RFC-0047**: Workspace Transaction Engine (Git-backed transaction branches and hard rollback stages).
* **RFC-0048**: Prompt Registry (Version control, hashes, diffing, and prompt A/B testing).
* **RFC-0049**: Benchmark Framework (Comparative cost, latency, compile, and verification success rates).
* **RFC-0050**: Policy Simulator (Dry-running historical plan templates against new Policy rules).
* **RFC-0051**: Knowledge Decay & TTL (Decay scoring formulas for obsolete semantic nodes).
* **RFC-0052**: Distributed Mission (gRPC/WebSocket Event Store streaming to remote nodes).
* **RFC-0053**: Artifact Lineage (Cryptographic metadata provenance tracing back to MissionID).
* **RFC-0054**: Time Travel Debugging (Read-only step playback and memory state inspection).
* **RFC-0055**: Multi-Workspace (Hierarchical submodules and project dependency maps).
* **RFC-0056**: Trust Engine (Dynamic provider trust ratings audited by Truth Pipeline passes).

---

## 🔄 The Closed-Loop Autonomous Software Factory

AEOS shifts the paradigm from simple **AI Code Generation** to a comprehensive **AI Workspace Operation** cycle. Instead of writing files in isolation, agents manipulate the Workspace directly:

```
Goal ──► Goal Engine ──► Planner ──► Execution ──► Quality Engine (DoD Verify)
 ^                                                        │
 │                                                        ▼
 └─────────────── Re-planning & Learning ◄────────── Observation
```

- **DoD Verification**: The loop only terminates when the DoD Engine confirms all business, architectural, and security constraints are satisfied.
- **Workspace-Centric**: Execution consists of a recursive `Read AST -> Modify Workspace -> Run Build -> Run Test -> Observe -> Repeat` cycle monitored by the Observation Runtime.

---

## ⚖️ System Constitution: Architecture Decision Principles (ADP)

To prevent feature inflation and maintain design integrity over the 10-year lifecycle, the core architecture is frozen at 57 RFCs. All future development, refactoring, and plugin integrations must comply with the 15 immutable design axioms defined in the system constitution:

👉 **[Architecture Decision Principles (ADP)](adp.md)**

---

## 💡 Conclusion

The **AEOS** architecture ensures the system remains stable, secure, and self-improving. AI models do not control the system; they are orchestrated as content-generation utilities under the Go Kernel. The 57-RFC blueprint and the Architecture Decision Principles (ADP) provide a solid, frozen constitution for engineering the system over the next 10 years, transitioning it from a theoretical design to an executable, local-first engineering platform.
