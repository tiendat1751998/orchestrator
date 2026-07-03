# Micro-Task 4.30: Create plugins/agents/reviewer/agent.yaml

## Info
- **File**: `plugins/agents/reviewer/agent.yaml`
- **Package**: `none`
- **Depends on**: 4.29
- **Time**: 10 min
- **Verify**: `cat plugins/agents/reviewer/agent.yaml`

## Purpose
Declares the capabilities and tools configuration settings for the Code Reviewer Agent.

## EXACT code to create

```yaml
name: "reviewer"
version: "0.1.0"
role: "Code Reviewer"
description: "Reviews source code modifications, verifies tests, audits security, and asserts standards"
capabilities:
  - "code_review"
  - "testing"
provider: "antigravity"
model: "gemini-3.5-pro"
tools:
  - "read_file"
  - "list_dir"
  - "search"
  - "git_status"
  - "git_diff"
  - "git_log"
prompt_file: "prompts/system.md"
temperature: 0.1
max_tokens: 8192
```

## Pitfalls

### Pitfall 1: Granting file write tools to auditing agents
Granting `write_file` or `run_command` tools to the Reviewer Agent violates read-only audit safety designs. Keep its tools restricted to read-only scopes.

### Pitfall 2: Silent configuration file paths errors
If prompt files are not found, constructors will fail. Verify configuration targets.

## Verify
```bash
cat plugins/agents/reviewer/agent.yaml
# Expected: YAML content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/reviewer/agent.yaml`
- [ ] Capabilities include code_review and testing
- [ ] Model target points to `"gemini-3.5-pro"`
- [ ] Tools omit write_file and run_command to enforce audit boundaries
- [ ] Build command passes
