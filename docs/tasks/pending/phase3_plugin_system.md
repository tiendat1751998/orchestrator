# Phase 3: Plugin System & SDK — Specifications

This phase implements the SDK (`sdk/`) that developers use to write providers, tools, and agents. All SDK designs must adhere strictly to the hexagonal layer boundaries, preventing dependencies from crossing from plugins to the kernel core.

---

## Task 3.1: SDK Process-Level Provider Base (`sdk/provider/`)

- **Base process Executor**:
  - Implement a standard SDK wrapper around Go `os/exec` for spawning external CLI platforms (Claude Code, Antigravity, Codex) inside isolated sandboxes.
  - Implement stdin writers, stdout readers, and stderr readers.
  - Implement concurrent reading for stdout/stderr in separate background goroutines to prevent OS pipe blockages.
- **Windows Pipe Normalizer**:
  - Automatically normalize Windows `\r\n` carriage returns to standard `\n` line endings.

---

## Task 3.2: SDK WASM Sandbox & Tool Base (`sdk/tool/` & `sdk/security/`)

- **Wazero WASM Sandbox Helper (RFC-0023)**:
  - Provide a default WASM sandboxing runtime using `Wazero`.
  - Jail filesystems to a designated workspace subfolder with strict, deny-by-default capability tokens (Axiom 12).
  - Enforce memory limit caps (e.g. 64MB) and CPU execution time limits.
- **Tool Input Schema Validator**:
  - Validate parameters against JSON schema declarations before calling `Execute()`.
  - Automatically acquire workspace locks (Workspace Transaction Engine - RFC-0047) prior to executing tool edits, preventing concurrent modification conflicts.
