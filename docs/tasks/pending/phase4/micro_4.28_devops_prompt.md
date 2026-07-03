# Micro-Task 4.28: Create plugins/agents/devops/prompts/system.md

## Info
- **File**: `plugins/agents/devops/prompts/system.md`
- **Package**: `none`
- **Depends on**: 4.27
- **Time**: 10 min
- **Verify**: `cat plugins/agents/devops/prompts/system.md`

## Purpose
Declares system prompts containing DevOps specific instructions and rules of behavior.

## EXACT code to create

```markdown
# Role
You are a DevOps Engineer Agent in an automated multi-agent team.
Your objective is to build CI/CD configurations, deploy scripts, coordinate releases, check environment configurations, and write system documentation.

# Capabilities
You have access to tools to read and write files, commit git changes, and run terminal build commands.

# Guidelines
1. **No Destructive Operations**: Never run commands that wipe disks or shut down servers.
2. **Detailed Docs**: Write comprehensive markdown guides detailing commands, ports, and structures.
3. **Environment Validations**: Verify files and environment flags before initiating builds.
```

## Pitfalls

### Pitfall 1: Bypassing destructive command guards
Failing to specify rules preventing dangerous operations can cause the model to execute dangerous commands like `rm -rf /` in shell tasks.

### Pitfall 2: Silent command blocks
DevOps instructions must detail documentation formatting guidelines to prevent generating plain unformatted texts.

## Verify
```bash
cat plugins/agents/devops/prompts/system.md
# Expected: markdown content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/devops/prompts/system.md`
- [ ] Role instructions are structured using clear markdown sections
- [ ] Guidelines block destructive shell commands
- [ ] Build command passes
