# Micro-Task 4.27: Create plugins/agents/devops/agent.yaml

## Info
- **File**: `plugins/agents/devops/agent.yaml`
- **Package**: `none`
- **Depends on**: 4.26
- **Time**: 10 min
- **Verify**: `cat plugins/agents/devops/agent.yaml`

## Purpose
Declares the capabilities and tools configuration settings for the DevOps Engineer Agent.

## EXACT code to create

```yaml
name: "devops"
version: "0.1.0"
role: "DevOps Engineer"
description: "Manages deployment processes, CI/CD configurations, environments, and documentation"
capabilities:
  - "deployment"
  - "documentation"
provider: "antigravity"
model: "gemini-3.5-pro"
tools:
  - "read_file"
  - "write_file"
  - "list_dir"
  - "git_status"
  - "git_diff"
  - "git_add"
  - "git_commit"
  - "run_command"
prompt_file: "prompts/system.md"
temperature: 0.2
max_tokens: 8192
```

## Pitfalls

### Pitfall 1: Mismatching capability tags
Ensure all capability tags match the core system constants exactly to prevent registration errors.

### Pitfall 2: Bypassing security tools settings
Failing to assign terminal run tools prevents the agent from validating builds. Always bind required tools.

## Verify
```bash
cat plugins/agents/devops/agent.yaml
# Expected: YAML content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/devops/agent.yaml`
- [ ] Capabilities map deployment and documentation tags
- [ ] Model target points to `"gemini-3.5-pro"`
- [ ] Build command passes
