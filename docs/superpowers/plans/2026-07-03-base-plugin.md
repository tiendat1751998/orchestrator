# Base Plugin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a thread-safe reusable `BasePlugin` helper implementing the `contracts/plugin.Plugin` interface, and write unit tests to verify its correctness.

**Architecture:** `BasePlugin` will protect state transitions (Init -> Start -> Stop) and status getters using a `sync.RWMutex`.

**Tech Stack:** Go 1.26.3, Go testing package.

## Global Constraints

- Must import `contracts/plugin` package aliasing it as `cplugin` to avoid namespace collisions.
- Strictly adhere to `contracts/` and `sdk/` import layering (no imports outside contracts/ and stdlib).
- Thread-safe state checks and updates under mutex locks.
- Idempotency for `Stop` calls.
- Always use named field initialization for all structs.

---

### Task 1: Create Base Plugin

**Files:**
- Create: `sdk/plugin/plugin.go`

**Interfaces:**
- Consumes: `github.com/tiendat1751998/orchestrator/contracts/plugin`
- Produces: `BasePlugin` struct and `NewBasePlugin` constructor.

- [ ] **Step 1: Write the minimal implementation code**

Create `sdk/plugin/plugin.go` containing the BasePlugin implementation.

- [ ] **Step 2: Build the code to verify it compiles**

Run: `go build ./sdk/plugin/...`
Expected: Passes with no errors.

---

### Task 2: Create Base Plugin Tests

**Files:**
- Create: `sdk/plugin/plugin_test.go`

**Interfaces:**
- Consumes: `BasePlugin`
- Produces: Unit tests for state transitions, validation, and thread-safety.

- [ ] **Step 1: Write test cases in sdk/plugin/plugin_test.go**

Create `sdk/plugin/plugin_test.go` containing lifecycle, validation, and concurrency tests.

- [ ] **Step 2: Run the test suite and verify passing**

Run: `go test -v -race ./sdk/plugin/...`
Expected: PASS and no race conditions detected.

---

### Task 3: Final Verification and Code Formatting

**Files:**
- Modify: none

- [ ] **Step 1: Run format and vet tools**

Run: `go fmt ./...` and `go vet ./...`
Expected: No errors or changes needed.

- [ ] **Step 2: Run golangci-lint if installed**

Run: `golangci-lint run ./...`
Expected: No errors.
