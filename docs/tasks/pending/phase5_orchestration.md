# Phase 5: Orchestration Engine — Specifications

This phase implements the master coordination logic (`kernel/orchestrator/` and `kernel/planner/`) that decomposes goals, scores candidates, explains selections, and manages learning loops.

---

## Task 5.1: Master Planner & Re-planner (`kernel/planner/`)

- **Decomposer, CSP Filter, & Beam Search (RFC-0030, RFC-0031, RFC-0032)**:
  - Decompose a high-level user `goal.Goal` into a structured set of objectives.
  - Implement a Go-native **Constraint Satisfaction Programming (CSP)** module that queries constraints (e.g. Budget, Language) and prunes invalid nodes from the Knowledge Graph before invoking any LLM search.
  - Implement **Beam Search** with a configured width $K$ to generate parallel candidate plan DAGs (`[]fsm.DAG`).
- **Pareto Frontier Scoring & Explainable Rationale**:
  - Implement the mathematical scoring function:
    $$Score(P_i) = w_{quality} \cdot Q(P_i) + w_{cost} \cdot C(P_i) + w_{time} \cdot T(P_i) + w_{confidence} \cdot Conf(P_i) - w_{risk} \cdot R(P_i)$$
  - Implement **Explainable Planning** to output a contrastive mathematical report detailing why the plan was chosen.
  - Implement **UCB-1 Exploration Bonus** to prevent new template starvation (Axiom 17):
    $$UCB(T_i) = SuccessRate(T_i) + c \cdot \sqrt{\frac{\ln(TotalMissions)}{UsageCount(T_i)}}$$
- **Dynamic Replanning & Recovery (RFC-0037)**:
  - If a task fails, the Replanner generates a corrective sub-graph and merges it into the active execution DAG, managing versioning and state via the `ExecutionGraphManager` (RFC-0046).

---

## Task 5.2: Orchestrator Coordinator & FSM Runner

- **FSM Runner**:
  - Implement the core mission state machine coordinating transitions (`Created -> Planning -> Running -> Validating -> Completed/Failed`).
- **Workspace Transaction checkpoints (RFC-0047)**:
  - Integrate transaction checkpoints at FSM boundaries: Git stashing dirty files before execution and running hard rollbacks if DoD validation fails.
- **DoD validation Gate (RFC-0015, RFC-0034)**:
  - Implement the DoD Engine to compile a multi-dimensional Quality Scorecard (compile pass, lint warnings count, test coverage metrics) to verify code correctness before commit.
