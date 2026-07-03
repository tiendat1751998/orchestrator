# Data Model & Configuration Specs

## 1. System Settings Schema (`settings.yaml`)

The primary configuration file resides at `.orchestrator/settings.yaml` or `~/.orchestrator/settings.yaml`.

```yaml
orchestrator:
  name: "my-orchestrator"
  log_level: "info"        # debug, info, warn, error
  log_format: "json"       # json, text
  data_dir: ".orchestrator/data"

providers:
  default: "antigravity"
  antigravity:
    type: "cli"
    binary: "antigravity"
    model: "gemini-2.5-pro"
    timeout: "120s"
  gemini:
    type: "api"
    api_key: "${GEMINI_API_KEY}"
    model: "gemini-2.5-flash"
    timeout: "60s"

agents:
  backend:
    provider: "antigravity"
    model: "gemini-2.5-pro"
    prompt_file: "prompts/backend/system.md"
  reviewer:
    provider: "gemini"
    model: "gemini-2.5-flash"
    prompt_file: "prompts/reviewer/system.md"

security:
  sandbox: true
  allowed_tools: ["git", "filesystem", "terminal"]
  blocked_commands: ["rm -rf /", "sudo", "chmod 777"]
```

---

## 2. Plugin Manifest Schema (`plugin.yaml`)

Every agent, provider, or tool plugin must contain a `plugin.yaml` file in its root.

```yaml
name: "backend"
version: "0.1.0"
type: "agent"              # "agent", "provider", "tool"
role: "Backend Engineer"
description: "AI developer for writing Go code and API tests"
capabilities:
  - "code_generation"
  - "testing"
  - "refactoring"
tools:
  - "git"
  - "filesystem"
  - "terminal"
provider: "antigravity"
model: "gemini-2.5-pro"
prompt_file: "prompts/system.md"
max_tokens: 8192
temperature: 0.2
```

---

## 3. Mission & Task DAG Model

The runtime state compiles tasks into a Directed Acyclic Graph (DAG).

### Mission Structure
- **ID**: `msn-<hex>` (Unique ID generated using crypto/rand bytes)
- **Title**: Human readable title
- **Description**: Detailed description of what the user wants to build
- **Constraints**: List of rules (`use Go`, `no external deps`, etc.)

### Task Structure
- **ID**: `tsk-<hex>`
- **Name**: Short task name
- **Type**: Capability string (e.g. `code_generation`, `testing`)
- **Input**: Map of string to any (`map[string]any`)
- **Dependencies**: List of predecessor Task IDs
- **Priority**: Execution priority (1 to 5)
- **Timeout**: Timeout duration per task
