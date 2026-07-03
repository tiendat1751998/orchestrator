# Micro-Task 4.25: Create plugins/agents/backend/prompts/system.md

## Info
- **File**: `plugins/agents/backend/prompts/system.md`
- **Package**: `none`
- **Depends on**: 4.24
- **Time**: 10 min
- **Verify**: `cat plugins/agents/backend/prompts/system.md`

## Purpose
Specifies the system instruction guidelines for the Backend Developer Agent. This styles the role directives and rules of behavior passed to the provider request messages.

## EXACT code to create

```markdown
# Role
You are a master Backend Developer Agent in an automated multi-agent team.
Your objective is to design, code, debug, refactor, and write unit tests for high-quality backend application logic.

# Capabilities
You have access to a variety of tools to read and write files, view repository statuses, and run terminal commands in the workspace sandbox.

# Action Guidelines
1. **Analyze First**: Inspect existing schemas and files before writing new code.
2. **Atomic Steps**: Write complete compilable structures. Do not use placeholders or comments like `// TODO: implement later`.
3. **Run Checks**: After updating files, compile and run tests using shell command tools to verify stability.
```

## Pitfalls

### Pitfall 1: Formatting prompt markers as raw HTML
Using unsupported HTML formatting tags within system prompts can confuse CLI parsers or yield raw tag strings in response outputs. Use standard markdown.

### Pitfall 2: Permitting placeholder mockups in instructions
Failing to explicitly instruct the agent to avoid placeholders like `// TODO` causes models to output partial, incomplete answers.

## Verify
```bash
cat plugins/agents/backend/prompts/system.md
# Expected: markdown content printed cleanly
```

## Checklist
- [ ] File exists at `plugins/agents/backend/prompts/system.md`
- [ ] Defines role directives clearly
- [ ] Contains guideline constraints blocking placeholder mockups
- [ ] Markdown syntax is valid
