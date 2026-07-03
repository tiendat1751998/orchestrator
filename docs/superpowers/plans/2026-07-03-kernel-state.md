# Kernel State Machine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `kernel/state.go` to manage the kernel's lifecycle state machine, ensuring valid transitions.

**Architecture:** Implement a thread-safe `StateMachine` wrapping a `State` value with a `sync.RWMutex` for transition validation and state changes.

**Tech Stack:** Go 1.26, standard library (`fmt`, `sync`).

## Global Constraints
- Strictly adhere to `kernel` package boundaries (only import standard library).
- Named struct initialization only.
- Run Go verification commands (`go fmt`, `go vet`, `go test`, `golangci-lint`) to ensure zero errors.

---

### Task 1: Create State Machine Code

**Files:**
- Create: `kernel/state.go`

**Interfaces:**
- Produces: `State`, `StateMachine`, `NewStateMachine()`, `Transition()`, `Is()`, `IsRunning()`, `IsStopped()`

- [ ] **Step 1: Write the implementation of `kernel/state.go`**
  Write the exact Go code defined in the specification.

- [ ] **Step 2: Verify compilation**
  Run: `go build ./kernel/...`
  Expected: Command completes successfully with no output.

---

### Task 2: Create State Machine Unit Tests

**Files:**
- Create: `kernel/state_test.go`

- [ ] **Step 1: Write the unit test file**
  Create unit tests verifying valid and invalid transitions, and thread-safety of the state machine.

- [ ] **Step 2: Run tests**
  Run: `go test -v -race ./kernel/...`
  Expected: Tests pass successfully.

- [ ] **Step 3: Run full verification suite**
  Run: `go fmt ./kernel/...` and `go vet ./kernel/...` and `golangci-lint run ./kernel/...`
  Expected: No formatting or lint errors.
