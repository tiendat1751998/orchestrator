# Workflow Helper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the base workflow structure `BaseWorkflow` and constructors in `sdk/workflow/workflow.go` and a topological sorting algorithm `SortSteps`.

**Architecture:** Embeds `sdkplugin.BasePlugin` and implements `contracts/workflow/Workflow` step verification, copy protection, and Kahn's algorithm for topological sorting and cycle detection.

**Tech Stack:** Go 1.26

## Global Constraints

- Enforce checks against duplicate step names inside constructors.
- Kahn's algorithm to resolve dependencies, checking `len(sorted) == len(steps)` to verify no circular dependencies.
- Return copies of the internal step list in `Steps()` to prevent modifications.
- Complete verification using `go test -v ./sdk/workflow/...` and formatting/vetting tools.

---

### Task 1: Create sdk/workflow/workflow.go

**Files:**
- Create: `sdk/workflow/workflow.go`

**Interfaces:**
- Consumes: `contracts/workflow/workflow.go`, `sdk/plugin/plugin.go`
- Produces: `BaseWorkflow`, `NewBaseWorkflow`, `SortSteps`

- [ ] **Step 1: Write the implementation**
  Create the file with the code provided in the specification.
- [ ] **Step 2: Run build to check syntactical correctness**
  Run: `go build ./sdk/workflow/...`
  Expected: Success.

### Task 2: Create sdk/workflow/workflow_test.go

**Files:**
- Create: `sdk/workflow/workflow_test.go`

**Interfaces:**
- Consumes: `sdk/workflow/workflow.go`
- Produces: Test suite for `BaseWorkflow` and `SortSteps`

- [ ] **Step 1: Write tests for duplicate steps, copy protection, topological sorting, and cycle detection**
  Create `sdk/workflow/workflow_test.go` with cases for validations and topological sorting.
- [ ] **Step 2: Run the test suite**
  Run: `go test -v ./sdk/workflow/...`
  Expected: All tests pass.
- [ ] **Step 3: Run formatting, linting, and quality checks**
  Run:
  - `go fmt ./...`
  - `go vet ./...`
  - `go build ./...`
  Expected: Clean exit for all checks.
