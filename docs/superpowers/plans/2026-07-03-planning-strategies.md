# Pareto Frontier Scoring and UCB-1 Exploration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Pareto multi-objective scoring and UCB-1 exploration functions in `kernel/planner/pareto.go` and unit tests in `kernel/planner/pareto_test.go`.

**Architecture:** A scorer struct that evaluates quality, cost, time, confidence, and risk using configurable weights, with a UCB-1 exploration factor added for candidate selection.

**Tech Stack:** Go 1.26

## Global Constraints

- Handle totalRuns and usageCount bounds safely (e.g., divide-by-zero, log of zero, etc.).
- Ensure zero external dependencies.
- Positional struct initialization is strictly forbidden. Always use named fields.
- Complete verification using `go test -v ./kernel/planner/...` and formatting/vetting tools.

---

### Task 1: Create kernel/planner/pareto.go

**Files:**
- Create: `kernel/planner/pareto.go`

**Interfaces:**
- Consumes: `contracts/fsm`
- Produces: `Weights`, `Scorer`, `NewScorer`, `ScoreCandidate`

- [ ] **Step 1: Write the implementation**
  Create the file with the code exactly as described in the specification.
- [ ] **Step 2: Run build to check syntactical correctness**
  Run: `go build ./kernel/planner/...`
  Expected: Success.

### Task 2: Create kernel/planner/pareto_test.go

**Files:**
- Create: `kernel/planner/pareto_test.go`

**Interfaces:**
- Consumes: `kernel/planner/pareto.go`
- Produces: Test suite for `Scorer`

- [ ] **Step 1: Write tests for Pareto Frontier Scoring**
  Write tests in `kernel/planner/pareto_test.go` covering:
  - Scorer initialization with weights.
  - Multi-objective scoring math (combining Q, C, T, Conf, R correctly).
  - UCB-1 exploration logic (usageCount > 0 vs usageCount == 0).
- [ ] **Step 2: Run the test suite**
  Run: `go test -v ./kernel/planner/...`
  Expected: All tests pass.
- [ ] **Step 3: Run formatting, linting, and quality checks**
  Run:
  - `go fmt ./...`
  - `go vet ./...`
  - `go build ./...`
  Expected: Clean exit for all checks.
