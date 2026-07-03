# AEOS AI Coding Rules: Operational Guidelines

This document defines the strict coding rules and static constraints that every AI agent must follow when writing code in the AEOS workspace. Any violation of these guidelines will result in an immediate rejection and rollback of changes.

---

## 1. Core Principles of AI Discipline

### Principle 1: Think Before Coding
* **Explicit Assumptions**: Before implementing, state your assumptions explicitly. If uncertain, ask.
* **Surface Tradeoffs**: If multiple interpretations exist, present them to the User - do not pick silently.
* **Propose Simplifications**: If a simpler approach exists, say so. Push back on over-engineering when warranted.
* **Stop on Ambiguity**: If anything is unclear, stop immediately. Name what is confusing and ask.

### Principle 2: Simplicity First (KISS / YAGNI)
* **Minimum Viable Code**: Write the minimum code required to solve the problem. Nothing speculative.
* **No Feature Creep**: Do not write features or parameters beyond what was requested.
* **No Unnecessary Abstractions**: Do not create generic helper packages, single-use interfaces, or redundant wrappers.
* **LOC Minimization**: If you write 200 lines and it could be 50, rewrite it immediately.
* **Auditing Constraint**: Ask yourself: *"Would a senior engineer say this is overcomplicated?"* If yes, simplify.

### Principle 3: Surgical Changes
* **Local Scope Only**: Touch only what you must. Clean up only your own mess.
* **No Adjacent Improvements**: Do not "improve" or format adjacent code, comments, or styling that is unrelated to the task.
* **Match Existing Style**: Match the existing code style exactly, even if you would do it differently.
* **Orphan Cleanup**: Remove imports, variables, and functions that *your* changes made unused. Do not remove pre-existing dead code unless asked.

### Principle 4: Goal-Driven Execution
* **Verifiable Criteria**: Define concrete test cases for task validation. For example, instead of "Add validation", define: "Write TestInvalidInput in validation_test.go, make it fail, implement validation in parser.go, then verify `go test -run TestInvalidInput ./...` passes."
* **Sequential Verification Plans**: For multi-step tasks, outline execution under a strict, runnable plan. For example:
  1. Declare error types in contracts/errors.go → verify: `go build ./contracts/...`
  2. Implement error wrapping in parser.go → verify: `go vet ./...`
  3. Write unit tests in parser_test.go → verify: `go test -v ./...`


---

## 2. Complexity Budgets
Every code file, function, and package must stay within the following mathematical ceilings (budgets are maximum ceilings, not targets to reach):

| Metric | Budget Limit | Action on Overrun |
|---|---|---|
| **Package Size** | $\le 1,000$ LOC | Split into sub-packages / separate files |
| **File Size** | $\le 300$ LOC | Extract helper components into adjacent files |
| **Function Length** | $\le 80$ LOC | Extract inner blocks into helper functions |
| **Cyclomatic Complexity** | $\le 10$ | Simplify conditional loops, extract branches |
| **Interface Size** | $\le 8$ methods | Segregate interfaces into smaller traits |
| **Struct Size** | $\le 15$ fields | Decompose into nested value objects / structs |

---

## 3. Frozen Import Matrix & Boundaries
Code dependencies are strictly layered. Imports must never violate the following boundary matrix:

```
                  [cmd]
                    │
            ┌───────┴───────┐
            ▼               ▼
        [modules]       [kernel]
            │               │
            └───────┬───────┘
                    ▼
                  [sdk]
                    │
                    ▼
               [contracts]
```

### Dependency Rules:
1. **`contracts/`**: Stdlib only. **Zero** external library imports. **Zero** imports from other project layers.
2. **`sdk/`**: May only import from `contracts/` and stdlib.
3. **`plugins/`**: May only import from `contracts/`, `sdk/`, and stdlib. Plugins must never import other plugins or kernel core code.
4. **`kernel/`**: May only import from `contracts/`, `sdk/`, and stdlib. It must never import from `plugins/` or `modules/`.
5. **`modules/`**: May import from `contracts/`, `kernel/`, and stdlib.
6. **`cmd/`**: The wire layer. May import from any package.

---

## 4. Strict Coding Conventions & Forbidden Patterns

### Concurrency
* **No Unbounded Goroutines**: Never spawn raw `go func()` in execution loops. Use centralized runtimes (`kernel/runtime`) or bounded primitives (e.g., `errgroup.SetLimit`).

### Error Handling & Panic Controls
* **No Panic**: Never call `panic()` or `log.Fatal()` in library code. Always return errors cleanly up the stack.
* **Wrap Errors**: Always wrap lower-level errors with context using `fmt.Errorf("...: %w", err)`.
* **Zero Recovery Assumption**: Do not write `recover()` blocks in library functions. Panic recovery is handled exclusively at the EventBus dispatch or FSM runner level.

### Language Constraints
* **No Business Reflection**: The use of `reflect` is forbidden in domain business logic and plugin packages. Standard library reflection-based serialization (e.g. `encoding/json`, `yaml.v3`) is permitted exclusively for config loading and event store modules.
* **No Unsafe**: The use of the `unsafe` package is forbidden.
* **Named Struct Initialization**: Positional struct initialization is strictly forbidden. Always use named fields:
  ```go
  // CORRECT:
  node := &DAGNode{ID: "t1", Status: StatePending}
  
  // FORBIDDEN:
  node := &DAGNode{"t1", StatePending}
  ```

### Scope Modification Boundaries
* **No Package Renames**: Never rename existing Go packages.
* **No Unapproved Helpers**: Do not create generic helper packages (`util/`, `helper/`) unless requested. Place utility logic directly in the domain package.
* **No Global State**: Global variables are strictly forbidden unless they are read-only constants or package-level error variables.
