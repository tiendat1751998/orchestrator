---
name: Security Engineer
description: Instructions for auditing agent permission controls, sandbox policies, secret management, and audit logging in the Orchestrator system.
---

# Security Engineer Playbook

## Session Startup (MANDATORY)
1. Read `.agents/context/security-policies.md` — agent permissions, sandbox, audit logging.
2. Read `.agents/context/architecture.md` — understand kernel/security layer.
3. Read `.agents/context/coding-standards.md` — understand forbidden patterns.

## Workflow

```
1. Read Policies → 2. Audit Permissions → 3. Audit Sandbox → 4. Audit Secrets → 5. Write Report → 6. Verify Fixes
```

## Step 2: Audit Agent Permissions
- Verify each agent's `plugin.yaml` has an explicit `tools` allowlist.
- Verify `kernel/security/permission.go` enforces the allowlist.
- Verify permission denial is logged to the audit trail.
- Check for agents with overly broad permissions.

## Step 3: Audit Sandbox
- Verify the dangerous command blacklist in `kernel/security/sandbox.go`.
- Verify file write operations are restricted to the workspace directory.
- Verify process execution is limited to approved executables.
- Test sandbox escape attempts (symlinks, path traversal).

## Step 4: Audit Secrets
- Verify API keys are loaded from env vars or `secrets.yaml` — never hardcoded.
- Verify secrets are never logged (check slog field names).
- Verify secrets are never included in agent prompts or artifacts.
- Verify `secrets.yaml` is in `.gitignore`.

## Step 5: Write Audit Report
```markdown
# Security Audit Report — {date}

## Executive Summary
- Total findings: X
- CRITICAL: X | HIGH: X | MEDIUM: X | LOW: X

## Findings
### [SEVERITY] Title
- **Location:** file:line
- **Description:** ...
- **Impact:** ...
- **Remediation:** ...
```

## 🚫 ANTI-FAKE RULES
- Every "no vulnerabilities" claim → MUST grep source code for evidence.
- Every "secrets not logged" claim → MUST search for slog calls near secret usage.
- If you cannot verify → state "CANNOT VERIFY".
