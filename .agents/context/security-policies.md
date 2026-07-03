# Security Policies & Sandboxing

## 1. Agent Tool Permission System

- Tool execution is strictly gated by `kernel/security/permission.go`.
- An agent plugin can only invoke tools specified in its `tools` list in `plugin.yaml`.
- If an agent attempts to execute a tool not in its manifest, the execution fails with `contracts.ErrPermissionDenied` and is logged in the audit trail.

---

## 2. Command Sandboxing (`kernel/security/sandbox.go`)

### Dangerous Command Blacklist
All terminal executions are intercepted. Commands containing any of the following substrings are immediately blocked:
- `rm -rf /`
- `rm -rf /*`
- `sudo`
- `chmod 777`
- `reboot`
- `shutdown`
- `mkfs`
- `dd if=`

### File System Scope
- Write operations are jailed to the active workspace directory.
- Symbolic links resolving outside the workspace scope are blocked.

---

## 3. Logger Redaction & Credential Safety

- The slog handler implemented in `kernel/logger` intercepts all log records.
- Any key matching sensitive patterns (`api_key`, `secret`, `password`, `key`) has its value redacted as `[REDACTED]` before outputting.
- Secrets are never embedded in the agent prompts or task results.
- `secrets.yaml` containing provider keys must be kept locally at `~/.orchestrator/secrets.yaml` and added to `.gitignore`.
