# Micro-Task 1.04: Create .golangci.yml

## Info
- **File**: `.golangci.yml`
- **Depends on**: 1.01
- **Time**: 5 min
- **Verify**: `golangci-lint run ./...`

## Purpose
Establishes the linting engine guidelines to enforce code style consistency, type-safety assertions, and check for unused methods or memory allocations.

## EXACT code to create

```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck      # Check error return values
    - govet         # Reports suspicious constructs
    - staticcheck   # Advanced static analysis
    - unused        # Detect unused code
    - gosimple      # Suggest code simplifications
    - gocritic      # Best practice checker
    - ineffassign   # Detect ineffectual assignments
    - typecheck     # Type checking
    - misspell      # Detect misspelled words
    - goimports     # Check import formatting

linters-settings:
  errcheck:
    check-type-assertions: true  # Check type assertion errors
    check-blank: false           # Don't check _ = err (intentional ignore)
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
  misspell:
    locale: US

issues:
  max-issues-per-linter: 50
  max-same-issues: 5
  exclude-rules:
    # Allow unused parameters in interface implementations
    - linters:
        - unused
      source: "func \\(.*\\) .* {"
    # Allow dot imports in test files
    - linters:
        - gocritic
      path: _test\.go
```

## ⚠️ Pitfalls

### Pitfall 1: Turning on too many conflicting linters
```yaml
# Chỉ bật khoảng 10-12 linters cốt lõi tập trung vào chất lượng code và hiệu năng.
```
Focus linting configurations on safety checks (like type check assertions and unused code detection) rather than cosmetic arguments.

### Pitfall 2: Too short run timeouts in CI environments
A default timeout (e.g. 1m) can cause linting stages to fail during parallel builds under constrained CI runner resources. Set `timeout: 5m` for safety.

## Verify
```bash
golangci-lint run ./...
```

## Checklist
- [ ] File `.golangci.yml` exists at workspace root
- [ ] Linter timeout is set to `5m` or higher
- [ ] Essential linters (`govet`, `errcheck`, `staticcheck`, `unused`, `typecheck`) are enabled
- [ ] Type assertions check (`check-type-assertions: true`) is active
- [ ] `golangci-lint run ./...` completes without configuration errors
