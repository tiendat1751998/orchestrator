# RFC-0007: Provider and Runtime Separation

- **Status**: PROPOSED
- **Priority**: P1 — Core
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0006 (Plugin SDK & Registry)

## Summary

This RFC details the decoupling of the **Provider layer** (plugins) and the **Execution Runtime** (kernel). Providers are simplified to act purely as translation engines. They translate system request data structures into model-specific API structures, and parse raw results back. The **Execution Runtime** handles all network transportation, standard input/output streaming, OS process spawning, timeout management, and circuit breakers.

## Motivation

Issue 8 from the architecture review noted that Providers should not spawn processes. In previous designs, CLI providers were responsible for calling shell command tools internally, spawning subprocesses, and waiting for outputs. This creates coupling:
- Providers need filesystem access and process permissions.
- Security controls (sandboxing, CPU limits) are scattered across individual plugins.
- Testability is poor (testing a provider requires running mock shell environments).

Decoupling these domains ensures that the execution layer retains strict, centralized control over system resources.

## Design

### Responsibility Split

| Feature | Provider Plugin (Translator) | Execution Runtime (Kernel) |
|---|---|---|
| CLI Command Spawning | ❌ No | ✅ Yes (Process Manager) |
| Network Call Transportation | ❌ No (HTTP client injected/managed by kernel) | ✅ Yes (HTTP Client Pool) |
| Text Request Mapping | ✅ Yes (translates prompt tags) | ❌ No |
| Subprocess Stdin/Stdout | ❌ No | ✅ Yes (Spawns, monitors, and streams) |
| Circuit Breaking & Retries | ❌ No | ✅ Yes (Resilience Manager) |

---

### Decoupled Contracts (`contracts/provider/`)

#### 1. Common Types

```go
// contracts/provider/request.go
package provider

type MessageRole string
const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

type Request struct {
	Messages    []Message      `json:"messages"`
	Temperature float64        `json:"temperature"`
	MaxTokens   int            `json:"max_tokens"`
	StopSignals []string       `json:"stop_signals,omitempty"`
	Tools       []any          `json:"tools,omitempty"` // schemas agent can call
}

type Response struct {
	Content      string         `json:"content"`
	ToolCalls    []ToolCall     `json:"tool_calls,omitempty"`
	TokensUsed   int            `json:"tokens_used"`
	FinishReason string         `json:"finish_reason"`
}

type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}
```

#### 2. Provider Interfaces

```go
// contracts/provider/api.go
package provider

import (
	"context"
)

// APIRequest represents the prepared request to send over HTTP/gRPC.
type APIRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
}

type APIProvider interface {
	Provider
	// TranslateRequest converts AEOS Request to raw API call specs.
	TranslateRequest(req Request) (*APIRequest, error)
	// ParseResponse translates raw HTTP response bodies into AEOS Response.
	ParseResponse(statusCode int, body []byte) (*Response, error)
}

// StreamChunk represents a single token or delta chunk received during streaming.
type StreamChunk struct {
	Content      string `json:"content"`
	TokensUsed   int    `json:"tokens_used,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// StreamingAPIProvider represents an API provider that supports SSE/streaming.
type StreamingAPIProvider interface {
	APIProvider
	// ParseStreamChunk translates a raw network byte chunk (SSE frame or HTTP chunk)
	// into a unified system StreamChunk structure.
	ParseStreamChunk(chunk []byte) (*StreamChunk, error)
}
```

```go
// contracts/provider/cli.go
package provider

// CLICommand defines command details to be spawned by the Kernel.
type CLICommand struct {
	Program string            `json:"program"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Dir     string            `json:"dir,omitempty"`
	Stdin   []byte            `json:"stdin,omitempty"`
}

type CLIProvider interface {
	Provider
	// BuildCommand translates AEOS Request into CLI command configuration.
	BuildCommand(req Request) (*CLICommand, error)
	// ParseOutput interprets standard output/error results from the process execution.
	ParseOutput(exitCode int, stdout, stderr []byte) (*Response, error)
}

// CLIStreamProvider represents a CLI provider that supports stdout stream processing.
type CLIStreamProvider interface {
	CLIProvider
	// ParseStreamLine processes a single stdout/stderr line from the running process.
	ParseStreamLine(isStderr bool, line []byte) (*StreamChunk, error)
}
```

---

### Execution Runtime Engine Flow

```
   Agent Action
       │
       ▼
 ┌───────────────┐
 │ APIProvider   │ ── TranslateRequest ──► APIRequest details
 └───────────────┘
       │
       ▼
 ┌────────────────────────────────────────────────────────┐
 │ Execution Runtime (Kernel HTTP Transport)              │
 │  - Applies timeouts, tracing headers, retries          │
 │  - Dispatches HTTP Request to API endpoint             │
 │  - Collects HTTP Response                              │
 └────────────────────────────────────────────────────────┘
       │
       ▼
 ┌───────────────┐
 │ APIProvider   │ ── ParseResponse ──► Unified Response
 └───────────────┘
```

For CLIs:

```
   Agent Action
       │
       ▼
 ┌───────────────┐
 │ CLIProvider   │ ── BuildCommand ──► CLICommand description
 └───────────────┘
       │
       ▼
 ┌────────────────────────────────────────────────────────┐
 │ Execution Runtime (Process Manager)                    │
 │  - Verifies safety permissions (Sandbox/Docker check)  │
 │  - Spawns subprocess, streams stdin                    │
 │  - Collects stdout, stderr, exit code                  │
 └────────────────────────────────────────────────────────┘
       │
       ▼
 ┌───────────────┐
 │ CLIProvider   │ ── ParseOutput ──► Unified Response
 └───────────────┘
```

### Process Manager Implementation (`kernel/execution/process/`)

The process execution is handled exclusively inside the kernel:

```go
// kernel/execution/process/manager.go
package process

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

type ProcessManager interface {
	Execute(ctx context.Context, cmd *provider.CLICommand, timeout time.Duration) (*provider.ProcessResult, error)
}

type processManager struct{}

func NewProcessManager() ProcessManager {
	return &processManager{}
}

func (m *processManager) Execute(ctx context.Context, cmd *provider.CLICommand, timeout time.Duration) (*provider.ProcessResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	osCmd := exec.CommandContext(ctx, cmd.Program, cmd.Args...)
	osCmd.Dir = cmd.Dir
	if len(cmd.Env) > 0 {
		env := make([]string, 0, len(cmd.Env))
		for k, v := range cmd.Env {
			env = append(env, k+"="+v)
		}
		osCmd.Env = env
	}

	if len(cmd.Stdin) > 0 {
		osCmd.Stdin = bytes.NewReader(cmd.Stdin)
	}

	var stdout, stderr bytes.Buffer
	osCmd.Stdout = &stdout
	osCmd.Stderr = &stderr

	startTime := time.Now()
	err := osCmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}

	return &provider.ProcessResult{
		ExitCode: exitCode,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}, nil
}
```

## Impact

- **Security Isolation**: The subprocess or network activity is governed directly in the kernel. Sandboxing policies can be applied once at the kernel level without relying on plugin-specific implementations.
- **Provider Simplification**: Providers only manage data mapping. No file system permissions, no socket creation, no process forks.

## Open Questions

1. **How does Streaming work under this decoupled design?**
   - For CLI providers, stdout can be read line-by-line using a pipe channel injected by the Process Manager.
   - For HTTP streaming, the kernel uses a stream response body and feeds chunks back to the APIProvider to parse in real-time. This will be specified in RFC-0008 (Event Model).
