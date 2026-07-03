# Input Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrades input validation models for both tasks and requests to return structured global `contracts.ValidationError` records instead of localized sentinel error strings.

**Architecture:** We will implement the `Validate()` method on the `Task` struct inside `contracts/agent/task.go` and completely replace request validation in `contracts/provider/request.go` to use `contracts.ValidationError`, then verify it using the package unit tests.

**Tech Stack:** Go 1.26, `testing` package.

## Global Constraints
- Do not create/modify files outside the task scope.
- Public interfaces in `contracts/` are strictly frozen.
- Zero tolerance for compiler/quality gate failures: roll back on error.
- Do not add ad-hoc optimizations, caches, or unapproved dependencies.
- Limit code modifications to a maximum of 8 files per task execution.

---

### Task 1: Add Validate Method to Task Struct

**Files:**
- Modify: [task.go](file:///d:/project/orchestrator/contracts/agent/task.go)
- Modify: [task_test.go](file:///d:/project/orchestrator/contracts/agent/task_test.go)

**Interfaces:**
- Consumes: `contracts.TaskID`
- Produces: `func (t *Task) Validate() error`

- [ ] **Step 1: Write the failing test**
  Add `TestTaskValidate` in [task_test.go](file:///d:/project/orchestrator/contracts/agent/task_test.go) verifying validation errors for missing ID, missing Name, missing Type, negative Timeout, duplicate dependencies, empty dependency IDs, and self-dependencies.
  
  ```go
  func TestTaskValidate(t *testing.T) {
  	// Test successful validation
  	task := NewTask("test-task", "Description", "execute")
  	if err := task.Validate(); err != nil {
  		t.Errorf("expected validation to pass, got: %v", err)
  	}

  	// Test empty ID
  	invalidTask := NewTask("test-task", "Description", "execute")
  	invalidTask.ID = ""
  	if err := invalidTask.Validate(); err == nil {
  		t.Error("expected error for empty task ID, got nil")
  	}
  }
  ```

- [ ] **Step 2: Run test to verify it fails**
  Run: `go test -v ./contracts/agent -run TestTaskValidate`
  Expected: FAIL with "task.Validate undefined"

- [ ] **Step 3: Write minimal implementation**
  Add the `Validate()` method to `Task` inside [task.go](file:///d:/project/orchestrator/contracts/agent/task.go).

- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./contracts/agent -run TestTaskValidate`
  Expected: PASS

- [ ] **Step 5: Commit**
  Run: `git add contracts/agent/task.go contracts/agent/task_test.go`

---

### Task 2: Update Request Validation to Use Global ValidationError

**Files:**
- Modify: [request.go](file:///d:/project/orchestrator/contracts/provider/request.go)
- Modify: [request_test.go](file:///d:/project/orchestrator/contracts/provider/request_test.go)

**Interfaces:**
- Consumes: `contracts.ValidationError`
- Produces: `func (r *Request) Validate() error`

- [ ] **Step 1: Update tests in request_test.go to assert ValidationError types**
  Modify [request_test.go](file:///d:/project/orchestrator/contracts/provider/request_test.go) to verify that `Validate()` returns a pointer to `contracts.ValidationError` and contains correct `Component`, `Field`, and `Reason` values, including role validations (system/user messages must contain content, tool messages must have a tool_call_id).

- [ ] **Step 2: Run test to verify it fails**
  Run: `go test -v ./contracts/provider`
  Expected: FAIL (or compilation errors due to type assertion failure or new validation rules)

- [ ] **Step 3: Write minimal implementation**
  Replace `Request.Validate()` in [request.go](file:///d:/project/orchestrator/contracts/provider/request.go) and ensure imports are correct (include `fmt`, `github.com/tiendat1751998/orchestrator/contracts` if needed, etc.). Remove any local ValidationError struct.

- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./contracts/provider`
  Expected: PASS

- [ ] **Step 5: Commit**
  Run: `git add contracts/provider/request.go contracts/provider/request_test.go`

---

### Task 3: Global Verification and Quality Gates

**Files:**
- None (Verification step)

- [ ] **Step 1: Build packages under contracts**
  Run: `go build ./contracts/...`
  Expected: PASS

- [ ] **Step 2: Run all tests in contracts**
  Run: `go test ./contracts/...`
  Expected: PASS

- [ ] **Step 3: Run golangci-lint**
  Run: `golangci-lint run ./contracts/...`
  Expected: PASS (or zero errors/warnings)
