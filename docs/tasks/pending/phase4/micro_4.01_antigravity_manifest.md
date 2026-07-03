# Micro-Task 4.01: Create plugins/providers/antigravity/plugin.yaml

## Info
- **File**: `plugins/providers/antigravity/plugin.yaml`
- **Package**: `none`
- **Depends on**: 3.26
- **Time**: 10 min
- **Verify**: `cat plugins/providers/antigravity/plugin.yaml`

## Purpose
Declares the plugin registration manifest for the Antigravity provider plugin. This defines the plugin type, name, version, and initial configuration parameters used by the kernel registry.

## EXACT code to create

```yaml
name: "antigravity"
type: "provider"
version: "0.1.0"
description: "Antigravity CLI provider (Gemini backend)"
config:
  binary: "antigravity"
  model: "gemini-2.5-pro"
  timeout: "120s"
```

## Pitfalls

### Pitfall 1: Indentation structure violations in YAML
```yaml
# WRONG:
name: "antigravity"
  type: "provider" # Invalid indentation causes YAML parser to crash during plugin load phases!

# CORRECT:
name: "antigravity"
type: "provider"
```
Ensure all root attributes (`name`, `type`, `version`, `description`, `config`) share matching zero-level indentations.

### Pitfall 2: Typos in reserved metadata keys
Using camelCase keys (e.g. `binaryPath`) instead of snake_case or standard schema labels prevents loader engines from locating settings. Use exact matching tags.

## Verify
```bash
# Verify file exists and is readable YAML
cat plugins/providers/antigravity/plugin.yaml
# Expected: YAML contents printed cleanly
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/plugin.yaml`
- [ ] Plugin name property is set to `"antigravity"`
- [ ] Plugin type matches `"provider"`
- [ ] YAML structure is indentation-compliant
- [ ] Config fields contain binary, model, and timeout keys
