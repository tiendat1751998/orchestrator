# RFC Process

## How to read RFCs

Each RFC is a standalone document in `docs/rfcs/` that describes a foundational architectural decision. RFCs are numbered sequentially (RFC-0000, RFC-0001, ...) and must be approved before any related code is written.

## RFC Status

| Status | Meaning |
|---|---|
| `DRAFT` | Under development, open for feedback |
| `PROPOSED` | Ready for review |
| `ACCEPTED` | Approved — code can be written |
| `IMPLEMENTED` | Code written and verified |
| `SUPERSEDED` | Replaced by a newer RFC |

## RFC Template

See [RFC-TEMPLATE.md](RFC-TEMPLATE.md) for the standard format.

## Index

| RFC | Title | Status | Priority |
|---|---|---|---|
| [RFC-0000](rfc-0000-state-machine.md) | Everything is State Machine | PROPOSED | P0 |
| [RFC-0001](rfc-0001-kernel-architecture.md) | Kernel Architecture (3 Runtimes) | PROPOSED | P0 |
| [RFC-0002](rfc-0002-brain-architecture.md) | Brain Architecture (Engines) | PROPOSED | P0 |
| [RFC-0003](rfc-0003-knowledge-engine.md) | Knowledge Engine | PROPOSED | P0 |
| [RFC-0004](rfc-0004-context-engine.md) | Context Engine | PROPOSED | P1 |
| [RFC-0005](rfc-0005-memory-model.md) | Memory Model | PROPOSED | P1 |
| [RFC-0006](rfc-0006-plugin-sdk-registry.md) | Plugin SDK & Registry | PROPOSED | P1 |
| [RFC-0007](rfc-0007-provider-runtime-separation.md) | Provider ≠ Runtime | PROPOSED | P1 |
| [RFC-0008](rfc-0008-event-model.md) | Event Model & Event Sourcing | PROPOSED | P2 |
| [RFC-0009](rfc-0009-resource-manager.md) | Resource Manager | PROPOSED | P2 |
| [RFC-0010](rfc-0010-cognitive-layer.md) | Cognitive Layer | PROPOSED | P2 |
| [RFC-0011](rfc-0011-scheduler.md) | Scheduler | PROPOSED | P2 |
| [RFC-0012](rfc-0012-security-capability-model.md) | Security & Capability Model | PROPOSED | P2 |
| [RFC-0013](rfc-0013-workspace-engine.md) | Workspace Engine | PROPOSED | P2 |
| [RFC-0014](rfc-0014-quality-engine.md) | Quality Engine & Verification Pipelines | PROPOSED | P2 |
| [RFC-0015](rfc-0015-dod-engine.md) | Definition of Done (DoD) Engine | PROPOSED | P2 |
| [RFC-0016](rfc-0016-observation-runtime.md) | Observation Runtime & Telemetry Collector | PROPOSED | P2 |
| [RFC-0017](rfc-0017-workspace-snapshots-deterministic-replay.md) | Workspace Snapshots & Deterministic Replay | PROPOSED | P2 |
| [RFC-0018](rfc-0018-cost-engine-planning-cache.md) | Cost Engine & Planning Cache | PROPOSED | P2 |
| [RFC-0019](rfc-0019-adr-policy-versioning.md) | ADR (Architecture Decision Record) & Policy Versioning | PROPOSED | P2 |
| [RFC-0020](rfc-0020-agent-capabilities.md) | Agent Capabilities & Skill Tree Metrics | PROPOSED | P2 |
| [RFC-0021](rfc-0021-vector-search.md) | Vector Search & Local Graph Embeddings | PROPOSED | P2 |
| [RFC-0022](rfc-0022-multi-agent-collaboration.md) | Multi-Agent Collaboration & Event Routing | PROPOSED | P2 |
| [RFC-0023](rfc-0023-process-sandboxing.md) | Process Sandboxing & Container Isolation | PROPOSED | P2 |
| [RFC-0024](rfc-0024-api-gateways.md) | API Gateways (gRPC, REST, WebSockets) | PROPOSED | P2 |
| [RFC-0025](rfc-0025-dependency-tree.md) | Dependency Tree & AST Parser | PROPOSED | P2 |
| [RFC-0026](rfc-0026-benchmark-profiling.md) | Benchmark & Runtime Profiling | PROPOSED | P2 |
| [RFC-0027](rfc-0027-resilience-circuit-breakers.md) | Resilience, Circuit Breakers & Backoffs | PROPOSED | P2 |
| [RFC-0028](rfc-0028-audit-trail.md) | Audit Trail & Cryptographic Event Logs | PROPOSED | P2 |
| [RFC-0029](rfc-0029-web-ui.md) | Web UI & Mission Control Protocol | PROPOSED | P2 |
| [RFC-0030](rfc-0030-goal-engine.md) | Goal Engine (Goal -> Objectives -> Constraints -> Milestones) | PROPOSED | P2 |
| [RFC-0031](rfc-0031-world-model.md) | World Model (Workspace Ontology & Object Schema Mapping) | PROPOSED | P2 |
| [RFC-0032](rfc-0032-skill-graph.md) | Skill Graph (Dynamic Tech Dependencies and Multi-route Paths) | PROPOSED | P2 |
| [RFC-0033](rfc-0033-confidence-engine.md) | Confidence Engine (Dynamic Agent Self-Awareness and Routing Escalation) | PROPOSED | P2 |
| [RFC-0034](rfc-0034-advanced-quality-engine.md) | Advanced Quality Engine (Multi-dimensional Quality Scorecard checks) | PROPOSED | P2 |
| [RFC-0035](rfc-0035-capability-graph-agent-competency-model.md) | Capability Graph & Agent Competency Model | PROPOSED | P2 |
| [RFC-0036](rfc-0036-mission-simulation-dry-run-planner.md) | Mission Simulation & Dry-run Planner | PROPOSED | P2 |
| [RFC-0037](rfc-0037-adaptive-recovery-replanning-engine.md) | Adaptive Recovery & Replanning Engine | PROPOSED | P2 |
| [RFC-0038](rfc-0038-resource-planning-budget-estimation.md) | Resource Planning & Budget Estimation | PROPOSED | P2 |
| [RFC-0039](rfc-0039-evolution-engine.md) | Evolution Engine (Planner Learning) | PROPOSED | P2 |
| [RFC-0040](rfc-0040-intent-engine.md) | Intent Engine (Business, Priority, Target Constraints) | PROPOSED | P2 |
| [RFC-0041](rfc-0041-product-memory.md) | Product Memory (Business Ontologies and Domain Patterns) | PROPOSED | P2 |
| [RFC-0042](rfc-0042-product-manager-runtime.md) | Product Manager Runtime (Requirement Translation & DoD validation) | PROPOSED | P2 |
| [RFC-0043](rfc-0043-release-intelligence.md) | Release Intelligence (Canary loops, Rollbacks, and Production Telemetry) | PROPOSED | P2 |
| [RFC-0044](rfc-0044-economic-engine.md) | Economic Engine (Business Value, Latency, and expected ROI mapping) | PROPOSED | P2 |
| [RFC-0045](rfc-0045-digital-workforce.md) | Digital Workforce (Virtual Employee Competency and Budget models) | PROPOSED | P2 |
| [RFC-0046](rfc-0046-execution-graph-manager.md) | Execution Graph Manager (Dynamic DAG Merges, pause, and versioning) | PROPOSED | P2 |
| [RFC-0047](rfc-0047-workspace-transaction-engine.md) | Workspace Transaction Engine (Git-backed stashes and rollback stages) | PROPOSED | P2 |
| [RFC-0048](rfc-0048-prompt-registry.md) | Prompt Registry (Version control, hashes, and prompt benchmarking) | PROPOSED | P2 |
| [RFC-0049](rfc-0049-benchmark-framework.md) | Benchmark Framework (Planner comparative cost, latency, and compile metrics) | PROPOSED | P2 |
| [RFC-0050](rfc-0050-policy-simulator.md) | Policy Simulator (Dry-running old plan templates against new policy versions) | PROPOSED | P2 |
| [RFC-0051](rfc-0051-knowledge-decay.md) | Knowledge Decay & TTL (Decay score factor for obsolete graph nodes) | PROPOSED | P2 |
| [RFC-0052](rfc-0052-distributed-mission.md) | Distributed Mission (gRPC/WebSocket Event Store streaming to remote nodes) | PROPOSED | P2 |
| [RFC-0053](rfc-0053-artifact-lineage.md) | Artifact Lineage (Cryptographic metadata provenance tracing back to MissionID) | PROPOSED | P2 |
| [RFC-0054](rfc-0054-time-travel-debugging.md) | Time Travel Debugging (Read-only step playback and memory state inspection) | PROPOSED | P2 |
| [RFC-0055](rfc-0055-multi-workspace.md) | Multi-Workspace (Hierarchical submodules and project dependency maps) | PROPOSED | P2 |
| [RFC-0056](rfc-0056-trust-engine.md) | Trust Engine (Dynamic provider trust ratings audited by Truth Pipeline passes) | PROPOSED | P2 |
