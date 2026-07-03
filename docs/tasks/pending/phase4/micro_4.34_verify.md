# Micro-Task 4.34: Verification — Build & Test All Phase 4

## Info
- **File**: None (verification task definition only)
- **Depends on**: ALL micro-tasks 4.01 → 4.33
- **Time**: 15 min
- **Verify**: `go test -race ./plugins/...`

## Purpose
Verifies that all provider plugins, tool plugins, and agent plugins compile cleanly without import cycle issues, and pass all verification tests with the race detector.

## Verification Commands (Execute in exact order)

### Step 1: Verify Plugin Files Exist
```bash
# Antigravity Provider
ls plugins/providers/antigravity/plugin.yaml
ls plugins/providers/antigravity/plugin.go
ls plugins/providers/antigravity/provider.go
ls plugins/providers/antigravity/provider_gemini.go
ls plugins/providers/antigravity/provider_test.go
ls plugins/providers/antigravity/adapter/cli.go
ls plugins/providers/antigravity/adapter/stdin.go
ls plugins/providers/antigravity/adapter/stdout.go
ls plugins/providers/antigravity/adapter/stderr.go
ls plugins/providers/antigravity/parser/markdown.go
ls plugins/providers/antigravity/parser/toolcall.go
ls plugins/providers/antigravity/parser/json.go
ls plugins/providers/antigravity/parser/error.go
ls plugins/providers/antigravity/session/manager.go
ls plugins/providers/antigravity/session/heartbeat.go
ls plugins/providers/antigravity/prompt/builder.go

# Filesystem, Git, and Terminal Tools
ls plugins/tools/filesystem/read_file.go
ls plugins/tools/filesystem/write_file.go
ls plugins/tools/filesystem/list_dir.go
ls plugins/tools/filesystem/search.go
ls plugins/tools/git/git.go
ls plugins/tools/terminal/terminal.go
ls plugins/tools/tools_test.go

# Core Developer Agents
ls plugins/agents/backend/agent.yaml
ls plugins/agents/backend/agent.go
ls plugins/agents/backend/prompts/system.md
ls plugins/agents/devops/agent.yaml
ls plugins/agents/devops/agent.go
ls plugins/agents/devops/prompts/system.md
ls plugins/agents/reviewer/agent.yaml
ls plugins/agents/reviewer/agent.go
ls plugins/agents/reviewer/prompts/system.md
ls plugins/agents/agent_test.go
```

### Step 2: Go Build (Compiler Check)
```bash
go build ./plugins/...
```

### Step 3: Go Vet (Linter Check)
```bash
go vet ./plugins/...
```

### Step 4: Go Test (Unit Tests)
```bash
go test -v ./plugins/...
```

### Step 5: Go Test with Race Detector
```bash
go test -race ./plugins/...
```

### Step 6: Full Workspace Check
```bash
go build ./...
```

### Step 7: Git Commit
```bash
git add -A
git commit -m "Phase 4: Provider, tools, and agents plugins implementation (34 micro-tasks)"
git push origin main
```

## Phase 4 Quality Checklist

### Provider Plugin
- [ ] `plugins/providers/antigravity/adapter/cli.go` — CLI Process Manager handles process restarts and Windows process groups.
- [ ] `plugins/providers/antigravity/adapter/stdin.go` — Safe concurrent prompt writes with line normalization.
- [ ] `plugins/providers/antigravity/adapter/stdout.go` — Reads stdout in background goroutine with timeout guards.
- [ ] `plugins/providers/antigravity/adapter/stderr.go` — Drains stderr in the background to prevent OS buffer blockages.
- [ ] `plugins/providers/antigravity/parser/markdown.go` — Cleans delimiter tokens and code block fences.
- [ ] `plugins/providers/antigravity/parser/toolcall.go` — Parses single or list tool call outputs, generating unique IDs.
- [ ] `plugins/providers/antigravity/parser/json.go` — Safely strips markdown blocks and decodes JSON responses.
- [ ] `plugins/providers/antigravity/parser/error.go` — Maps CLI output warning keywords to contracts error sentinels.
- [ ] `plugins/providers/antigravity/session/manager.go` — Pools process sessions and terminates idle sessions.
- [ ] `plugins/providers/antigravity/session/heartbeat.go` — Periodically monitors process pipe responsiveness.
- [ ] `plugins/providers/antigravity/provider_gemini.go` — Native HTTP REST alternative fallback is compilable.

### Tool Plugins
- [ ] `plugins/tools/filesystem/read_file.go` — Enforces path traversal filters and blocks binary content.
- [ ] `plugins/tools/filesystem/write_file.go` — Implements atomic temp file writes followed by directory renames.
- [ ] `plugins/tools/filesystem/list_dir.go` — Provides shallow directory scanners with empty path fallbacks.
- [ ] `plugins/tools/filesystem/search.go` — Grep walks files while skipping repository structures and matching limits.
- [ ] `plugins/tools/git/git.go` — Executes git wrappers inside workspace directories and truncates diffs.
- [ ] `plugins/tools/terminal/terminal.go` — Runs shells on native OS platforms with 30s timeouts and blocklists.

### Agent Plugins
- [ ] `plugins/agents/backend/agent.go` — Constructor loads YAML manifest config and returns SDK BaseAgent.
- [ ] `plugins/agents/devops/agent.go` — Integrates capabilities deployment and documentation.
- [ ] `plugins/agents/reviewer/agent.go` — Limits capabilities to read-only tool sets (no write file tools).

### Quality Gates
- [ ] `go build ./...` ✅ (clean compilation of entire workspace)
- [ ] `go test ./plugins/...` ALL PASS
- [ ] `go test -race ./plugins/...` NO RACES
- [ ] No circular package imports are found
- [ ] Git commit and push succeeds
