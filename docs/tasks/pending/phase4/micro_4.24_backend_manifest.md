# Micro-Task 4.24: Create plugins/agents/backend/agent.yaml

## Info
- **File**: `plugins/agents/backend/agent.yaml`
- **Package**: `none`
- **Depends on**: 4.23
- **Time**: 10 min
- **Verify**: `cat plugins/agents/backend/agent.yaml`

## Purpose
Declares the capability and tool configurations for the Backend Developer Agent. This enables the runtime registry to discover and associate capabilities to this specific agent.

## EXACT code to create

```yaml
name: "backend"
version: "0.1.0"
role: "Backend Developer"
description: "Generates backend code, APIs, database schemas, and unit tests"
capabilities:
  - "code_generation"
  - "testing"
  - "debugging"
  - "refactoring"
provider: "antigravity"
model: "gemini-3.5-pro"
tools:
  - "read_file"
  - "write_file"
  - "list_dir"
  - "search"
  - "git_status"
  - "git_diff"
  - "git_add"
  - "git_commit"
  - "run_command"
prompt_file: "prompts/system.md"
temperature: 0.3
max_tokens: 8192
```

## Pitfalls

### Pitfall 1: Incorrect indentation in capability lists
Ensure array lists are properly indented with hyphens under the parent key to prevent parsing crashes.

### Pitfall 2: Typos in capabilities naming
Using custom or informal capability tags like `"build-api"` instead of standardized tags (`"code_generation"`) will prevent the scheduler and registry matching routines from selecting this agent.

## Verify
```bash
cat plugins/agents/backend/agent.yaml
# Expected: YAML content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/backend/agent.yaml`
- [ ] Capabilities include code_generation, testing, debugging, and refactoring
- [ ] Provider is configured to `"antigravity"`
- [ ] Model targets `"gemini-3.5-pro"`
- [ ] Prompt file is mapped to `"prompts/system.md"`
