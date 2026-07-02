# Micro-Task 1.02: Tạo .gitignore

## Thông tin
- **File tạo**: `.gitignore`
- **Dependencies trước**: 1.01
- **Thời gian**: 5 phút
- **Verify**: `git status` không hiển thị files bị ignore

## Nội dung CHÍNH XÁC cần tạo

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

## Quy tắc
1. PHẢI ignore `.env` — tránh lộ API keys
2. PHẢI ignore `/bin/` — binary output
3. PHẢI ignore IDE files (`.idea/`, `.vscode/`)
4. KHÔNG ignore `.orchestrator/settings.yaml` — đây là config, cần track

## Checklist
- [ ] File `.gitignore` tồn tại ở root
- [ ] `.env` bị ignore
- [ ] `/bin/` bị ignore
- [ ] `.idea/` và `.vscode/` bị ignore
- [ ] `.orchestrator/settings.yaml` KHÔNG bị ignore
