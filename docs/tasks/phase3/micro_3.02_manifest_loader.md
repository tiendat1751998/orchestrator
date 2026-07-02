# Micro-Task 3.02: Create sdk/agent/manifest.go

## Info
- **File**: `sdk/agent/manifest.go`
- **Package**: `agent`
- **Depends on**: 1.20 (manifest.go contract), 1.39 (input validation contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/agent/...`

## Purpose
Implements the manifest loader (`LoadManifest`) for agents. It parses the YAML agent configuration (`agent.yaml`), validates required properties, and correctly resolves the system prompt file path relative to the manifest file location rather than the working directory.

## EXACT code to create

```go
// Package agent provides SDK helper implementations for AI agent plugins.
package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"gopkg.in/yaml.v3"
)

// LoadManifest reads, parses, and validates an agent's YAML manifest file.
//
// Prompt Resolution:
//   - If PromptFile is defined, it reads the prompt from that file path.
//   - The path in PromptFile is resolved relative to the directory of the manifest file.
//   - The loaded prompt overwrites SystemPrompt.
func LoadManifest(manifestPath string) (*agent.Manifest, error) {
	if manifestPath == "" {
		return nil, fmt.Errorf("sdk/agent: manifest path cannot be empty")
	}

	// 1. Read the manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("sdk/agent: failed to read manifest file %q: %w", manifestPath, err)
	}

	// 2. Parse YAML content
	var manifest agent.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("sdk/agent: failed to parse manifest YAML: %w", err)
	}

	// 3. Resolve PromptFile if set
	if manifest.PromptFile != "" {
		manifestDir := filepath.Dir(manifestPath)
		resolvedPromptPath := filepath.Join(manifestDir, manifest.PromptFile)

		promptData, err := os.ReadFile(resolvedPromptPath)
		if err != nil {
			return nil, fmt.Errorf("sdk/agent: failed to read prompt file %q: %w", resolvedPromptPath, err)
		}
		manifest.SystemPrompt = string(promptData)
	}

	// 4. Validate manifest fields
	if err := validateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("sdk/agent: manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// validateManifest validates the loaded manifest fields.
func validateManifest(m *agent.Manifest) error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.Role == "" {
		return fmt.Errorf("role is required")
	}
	if len(m.Capabilities) == 0 {
		return fmt.Errorf("at least one capability is required")
	}
	for i, cap := range m.Capabilities {
		if !cap.IsValid() {
			return fmt.Errorf("invalid capability at index %d: %q", i, string(cap))
		}
	}
	if m.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if m.Temperature < 0 || m.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0 (got %f)", m.Temperature)
	}
	if m.MaxTokens < 0 {
		return fmt.Errorf("max_tokens cannot be negative (got %d)", m.MaxTokens)
	}
	return nil
}
```

## ⚠️ Pitfalls

### Pitfall 1: Incorrect PromptFile Path Resolution
```go
// ❌ WRONG:
promptData, err := os.ReadFile(manifest.PromptFile) // Tải từ working directory của process
// Sẽ lỗi nếu app chạy từ repo root nhưng plugins đặt sâu bên trong directories.

// ✅ CORRECT:
manifestDir := filepath.Dir(manifestPath)
resolvedPromptPath := filepath.Join(manifestDir, manifest.PromptFile) // Resolve relative to manifest
```
Các plugin agent thường đi kèm với tệp cấu hình và prompt đặt cùng thư mục (ví dụ: `plugins/agents/backend/agent.yaml` và `plugins/agents/backend/prompts/system.md`). File path phải tính tương đối theo tệp tin manifest.

### Pitfall 2: Silent Fail on Invalid Capability Types
```go
// ❌ WRONG:
// Không check các Capabilities trong manifest. Chấp nhận các chuỗi rác như "super_agent".
// Sẽ lỗi khi planner khớp task capabilities runtime.

// ✅ CORRECT:
for i, cap := range m.Capabilities {
    if !cap.IsValid() {
        return fmt.Errorf("invalid capability at index %d: %q", i, string(cap))
    }
}
```
Lọc lỗi sớm (Fail-Fast) khi nạp manifest giúp tránh lỗi cấu hình khó gỡ trong runtime.

## Verify
```bash
go build ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/manifest.go` tồn tại
- [ ] Package: `agent`
- [ ] Hàm `LoadManifest` đọc và unmarshal YAML cấu hình chính xác
- [ ] Tải `PromptFile` tương đối theo thư mục tệp tin `agent.yaml`
- [ ] Dữ liệu từ `PromptFile` ghi đè chính xác lên trường `SystemPrompt`
- [ ] Xác thực đầy đủ các thuộc tính của Manifest (`Name`, `Role`, `Capabilities`, `Provider`, `Temperature`, `MaxTokens`)
- [ ] `go build ./sdk/agent/...` chạy thành công
