# Phase 4: Provider & Agent Plugins — Specifications

This phase implements the concrete plugins under `plugins/` that hook the Orchestrator up to actual coding platforms (Claude Code, Antigravity, Codex) and tools. All plugins must import only from `contracts/` and `sdk/`, never importing other plugins or kernel packages directly.

---

## Task 4.1: Claude Code Provider Plugin (`plugins/providers/claude/`)

Implement the provider port by wrapping the **Claude Code CLI** process.
- **Process Lifecycle & Execution**:
  - Spawn the `claude` CLI process inside the container sandbox.
  - Feed task prompts into the process stdin.
  - Monitor stdout/stderr in concurrent goroutines.
- **Output Diff Parser**:
  - Parse Claude Code CLI stdout lines to extract Diff blocks and files modified.
  - Compile the outputs into a structured `provider.Response` containing lists of edited files, success status, and compiler logs.

---

## Task 4.2: Antigravity CLI Provider Plugin (`plugins/providers/antigravity/`)

Implement the provider port by wrapping the **Antigravity CLI** process.
- **Process Execution & Windows Normalizer**:
  - Spawn `antigravity` process, multiplex stdin/stdout, and normalize Windows `\r\n` carriage returns.
- **CLI Stream Parser**:
  - Scan stdout for real-time status updates and extract task completion parameters.

---

## Task 4.3: Secure Tool Plugins (`plugins/tools/`)

Implement tools that interact with the filesystem under strict security controls.
- **Workspace Transaction Integration (RFC-0047)**:
  - Filesystem tools (`read_file`, `write_file`, `git_commit`) must register with the Workspace Transaction Engine.
  - The engine creates a transaction branch and checkpoint before modifications. If verification fails, it rolls back changes instantly.
- **WASM Compiler & Tool Sandboxing**:
  - Package helper compiler tools (e.g. Go syntax checks) to compile and execute inside the WASM sandbox environment.
