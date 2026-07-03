# Micro-Task 6.30: Final Verification Gate

## Info
- **File**: N/A (verification task)
- **Depends on**: All Phase 6 tasks (6.01-6.29)
- **Time**: 30 min
- **Verify**: Full build + test + demo

## Purpose
End-to-end verification that the entire system builds, tests pass, and the CLI demo works correctly. This is the final gate before v0.1.0 release.

## Verification Steps

### Step 1: Full Build
```bash
go build ./...
```
Expected: Zero errors. All packages compile.

### Step 2: Static Analysis
```bash
go vet ./...
```
Expected: Zero warnings.

### Step 3: Full Test Suite
```bash
go test ./... -count=1 -race -timeout=5m
```
Expected: All tests pass. No data races.

### Step 4: CLI Demo
```bash
# 4.1: Version
orchestrator --version
# Expected: orchestrator version dev (commit: unknown, built: unknown)

# 4.2: Config init
orchestrator config init --output /tmp/test-config.yaml
# Expected: ✅ Configuration file created

# 4.3: Config show
orchestrator config show --config /tmp/test-config.yaml
# Expected: YAML output with API keys redacted

# 4.4: Agents list
orchestrator agents list --config /tmp/test-config.yaml
# Expected: table or "No agents registered"

# 4.5: Providers list
orchestrator providers list --config /tmp/test-config.yaml
# Expected: table or "No providers registered"

# 4.6: Help
orchestrator --help
# Expected: usage text with all subcommands
```

### Step 5: API Gateway Smoke Test
```bash
# Start gateway in background
orchestrator serve --config /tmp/test-config.yaml &

# Health check
curl http://localhost:8080/api/v1/health
# Expected: {"data":{"status":"healthy"}}

# List missions (empty)
curl http://localhost:8080/api/v1/missions
# Expected: {"data":[]}

# Stop gateway
kill %1
```

### Step 6: Git Tag
```bash
git add -A
git commit -m "Phase 6: CLI, API, and polish"
git tag v0.1.0
```

## Acceptance Criteria
- [ ] `go build ./...` — zero errors
- [ ] `go vet ./...` — zero warnings
- [ ] `go test ./...` — all pass, no races
- [ ] CLI commands functional
- [ ] REST API health check responds
- [ ] Tagged as v0.1.0

## Milestone M5: Production Ready ✅
