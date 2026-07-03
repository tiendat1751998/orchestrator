# Search Helpers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `sdk/search/search.go` implementing `contracts/search.Engine` interface, along with `sdk/search/search_test.go` verifying its features.

**Architecture:** Implement a thread-safe `InMemorySearchEngine` utilizing a mutex-protected map. Scores search matches based on string rules, supports metadata filters, and sorts results by score (descending) and ID (lexicographically ascending) for ties.

**Tech Stack:** Go 1.26 (Standard Library).

## Global Constraints

- docs/adp.md (Architectural Decision Principles)
- .agents/rules/ai_rules.md (Go code conventions and complexity budgets)
- .agents/rules/ponytail.md (Simplest minimal solution)
- .agents/rules/superpowers.md (Process discipline)

---

### Task 1: Create `sdk/search/search.go`

**Files:**
- Create: `sdk/search/search.go`

**Interfaces:**
- Consumes: `contracts/plugin`, `contracts/search`, `sdk/plugin`
- Produces: `InMemorySearchEngine`, `NewInMemorySearchEngine`

- [ ] **Step 1: Write the minimal implementation code**

Write the implementation of `InMemorySearchEngine` to `sdk/search/search.go` exactly as specified in the micro-task description, including deep copying of metadata and thread safety.

- [ ] **Step 2: Compile the search package**

Run: `go build ./sdk/search/...`
Expected: Compile success with no errors.

---

### Task 2: Create `sdk/search/search_test.go`

**Files:**
- Create: `sdk/search/search_test.go`

**Interfaces:**
- Consumes: `InMemorySearchEngine`, `NewInMemorySearchEngine`
- Produces: Unit tests for Search Engine features.

- [ ] **Step 1: Write the test code**

Write the unit tests checking:
1. Indexing and searching items.
2. Relevance scoring weights (1.0, 0.8, 0.5, 0.3, 0.1).
3. Metadata filter matching and exclusion.
4. Sorting order by Score descending, and lexicographical ID sorting for ties.
5. Limit (MaxResults) and MinScore configurations.

- [ ] **Step 2: Run tests to verify all test cases pass**

Run: `go test -v ./sdk/search/...`
Expected: PASS

- [ ] **Step 3: Run golangci-lint, vet, and check race conditions**

Run: `go vet ./sdk/search/...`
Run: `go test -race ./sdk/search/...`
Expected: PASS with zero warnings/errors.
