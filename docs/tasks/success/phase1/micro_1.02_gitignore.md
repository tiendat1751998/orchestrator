# Micro-Task 1.02: Create .gitignore

## Info
- **File**: `.gitignore`
- **Depends on**: 1.01
- **Time**: 5 min
- **Verify**: `git status`

## Purpose
Configures git exclusion rules to prevent committing compiled binaries, IDE scratch files, developer secrets, or test coverage outlays into source control.

## EXACT code to create

```gitignore
# ===== Binaries =====
/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# ===== Test =====
*.test
*.out
coverage.html
coverage.txt
coverage.out

# ===== Vendor =====
/vendor/

# ===== IDE =====
.idea/
.vscode/
*.swp
*.swo
*~
.gemini/

# ===== OS =====
.DS_Store
Thumbs.db

# ===== Build =====
/dist/
/tmp/

# ===== Environment =====
.env
.env.local
.env.*.local

# ===== Orchestrator data =====
.orchestrator/data/
.orchestrator/logs/
.orchestrator/artifacts/
```

## ⚠️ Pitfalls

### Pitfall 1: Committing local env files (Secret Leak)
```gitignore
.env
.env.local
```
Always verify `.env` is listed to block secret key exposure.

### Pitfall 2: Ignoring required project configuration templates
```gitignore
.orchestrator/data/
.orchestrator/logs/
.orchestrator/artifacts/ // Chỉ ignore các folder dữ liệu chạy, giữ lại settings.yaml
```
Never wildcard-ignore config folders. Only ignore runtime data directories within them.

## Verify
```bash
git status
# Verify that untracked temporary files do not appear
```

## Checklist
- [ ] File `.gitignore` exists at workspace root
- [ ] `.env` is ignored
- [ ] `/bin/` is ignored
- [ ] IDE directories (`.idea/`, `.vscode/`, `.gemini/`) are ignored
- [ ] `.orchestrator/data/` is ignored but config templates remain trackable
