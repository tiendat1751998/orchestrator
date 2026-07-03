# Micro-Task 4.31: Create plugins/agents/reviewer/prompts/system.md

## Info
- **File**: `plugins/agents/reviewer/prompts/system.md`
- **Package**: `none`
- **Depends on**: 4.30
- **Time**: 10 min
- **Verify**: `cat plugins/agents/reviewer/prompts/system.md`

## Purpose
Declares the system instructions configuration file for the Code Reviewer Agent.

## EXACT code to create

```markdown
# Role
You are a Code Reviewer Agent in an automated multi-agent team.
Your objective is to inspect code changes, locate syntax errors, verify test coverage, and audit security vulnerabilities.

# Capabilities
You have access to read-only tools to inspect files and view Git logs. You cannot modify files or execute terminal commands directly.

# Guidelines
1. **Audit Security**: Verify that files contain no exposed API keys or secrets.
2. **Review Style**: Assert standard formatting guidelines and warn against code smells.
3. **Approve Rules**: Provide detailed score-based reviews, explaining issues clearly.
```

## Pitfalls

### Pitfall 1: Formatting guidelines with loose constraints
Failing to specify standard audit boundaries results in agents approving insecure code changes containing plain passwords or tokens. Add strict API key checks in guidelines.

### Pitfall 2: Confusing capabilities scopes
Reviewer prompts must detail that the agent is read-only to avoid the agent attempting to request write tools from prompt loops.

## Verify
```bash
cat plugins/agents/reviewer/prompts/system.md
# Expected: markdown content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/reviewer/prompts/system.md`
- [ ] Role specifies Code Reviewer properties
- [ ] Guidelines detail security checks for API keys and formatting rules
- [ ] Build command passes
