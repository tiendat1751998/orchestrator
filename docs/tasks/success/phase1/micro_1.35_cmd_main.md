# Micro-Task 1.35: Create cmd/orchestrator/main.go

## Info
- **File**: `cmd/orchestrator/main.go`
- **Package**: `main`
- **Depends on**: 1.01 (go.mod)
- **Time**: 5 min
- **Verify**: `go build -o bin/orchestrator ./cmd/orchestrator/`

## Purpose
Initializes a skeleton placeholder main entry point binary for the orchestrator CLI. This is replaced with a complete Cobra command wrapper in Phase 6, but acts as a compilation check for CI/CD pipelines right now.

## EXACT code to create

```go
// Package main is the entry point for the orchestrator CLI.
// Full implementation will be added in Phase 6.
package main

import "fmt"

func main() {
	fmt.Println("orchestrator v0.1.0-dev")
	fmt.Println("Use 'orchestrator --help' for usage information.")
}
```

## Rules
1. **Directory layout convention**: Placed in `cmd/orchestrator/main.go` (Go's standard layout design for multi-binary repos).
2. **Package name**: Package namespace must be strictly declared as `main`.

## ⚠️ Pitfalls

### Pitfall 1: Declaring package name other than `main` in entry files
```go
package main
```
The compiler requires package name `main` containing a `main()` function to build standalone executable binaries.

### Pitfall 2: Using non-standard directory structures
Placing the main entry file in the repository root (`/main.go`) makes managing multiple utilities or binaries difficult. Always stick to the `/cmd/<app_name>/main.go` layout convention.

## Verify
```bash
go build -o bin/orchestrator ./cmd/orchestrator/
./bin/orchestrator
```

## Checklist
- [ ] File `cmd/orchestrator/main.go` exists
- [ ] Package: `main`
- [ ] `main()` function is declared
- [ ] Compilation via `go build` produces an executable in `bin/`
- [ ] Executable prints version string upon invocation
