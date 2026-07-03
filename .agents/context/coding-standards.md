# Coding Standards & Conventions

## 1. Language & Linter Setup

- **Language Version**: Go 1.26
- **Linters**: `golangci-lint` with the following active linters:
  - `errcheck`: Enforces checking of all returned errors.
  - `staticcheck`: Static analysis for bugs and performance optimizations.
  - `unused`: Detects dead and unused code (struct fields, functions, variables).
  - `gocritic`: Validates code style and idiomatic Go practices.

---

## 2. Strict Error Handling Guidelines

### Sentinel Errors Check
- Always use `errors.Is` or `errors.As` to check for sentinel errors defined in `contracts/errors.go`.
- **Never** perform string comparisons on errors.
```go
// ❌ WRONG
if err.Error() == "orchestrator: provider timeout" { ... }

// ✅ CORRECT
if errors.Is(err, contracts.ErrProviderTimeout) { ... }
```

### Error Wrapping
- Wrap errors using `fmt.Errorf("context: %w", err)` with the `%w` verb.
- **Never** use `%v` for wrapping since it destroys error chain metadata.
```go
// ❌ WRONG
return fmt.Errorf("failed execution: %v", err)

// ✅ CORRECT
return fmt.Errorf("failed execution: %w", err)
```

### Panic and Recover
- **Never** call `panic()` in library packages or kernel components. Always bubble errors back.
- Goroutine spawns must go through `kernel/runtime` or have a structured recovery wrapper.

---

## 3. Constructor & Initialization Patterns

- Constructors must be named `New<Type>` (e.g. `NewRegistry`).
- Inject all dependency resources through constructors — no global mutable state.
- Struct initialization must use named fields:
```go
// ❌ WRONG
node := TaskNode{"tsk-1", "Desc", "agt-1"}

// ✅ CORRECT
node := TaskNode{
	ID:          "tsk-1",
	Description: "Desc",
	AgentID:     "agt-1",
}
```

---

## 4. Import Conventions & Layer Boundaries

- Go standard imports must be sorted: stdlib → external → internal (`github.com/tiendat1751998/orchestrator/...`).
- Interfaces must be checked at compile time:
```go
var _ contracts.Plugin = (*MyPlugin)(nil)
```
- Multi-package contracts must be imported using their specific sub-packages:
```go
import (
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)
```
