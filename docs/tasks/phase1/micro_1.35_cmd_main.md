# Micro-Task 1.35: Tạo cmd/orchestrator/main.go

## Thông tin
- **File tạo**: `cmd/orchestrator/main.go`
- **Package**: `main`
- **Dependencies trước**: 1.01 (go.mod)
- **Thời gian**: 5 phút
- **Verify**: `go build -o bin/orchestrator ./cmd/orchestrator/`

## Nội dung CHÍNH XÁC cần tạo

```go
// Package main is the entry point for the orchestrator CLI.
// Full implementation will be added in Phase 6.
package main

import "fmt"

func main() {
	fmt.Println("orchestrator v0.1.0-dev")
	fmt.Println("Use 'orchestrator --help' for usage information.")
}
```

## Mục đích
- Placeholder để đảm bảo `go build ./...` hoạt động với toàn bộ project
- Sẽ được thay thế bởi cobra CLI framework trong Phase 6
- Cho phép CI/CD verify build sớm

## ⚠️ Pitfalls cần tránh
1. **Path phải đúng**: `cmd/orchestrator/main.go` — convention Go cho multi-binary projects
2. **Package phải là `main`**: Nếu không → `go build` sẽ fail

## Lệnh verify
```bash
go build -o bin/orchestrator ./cmd/orchestrator/
./bin/orchestrator
# Expected output:
# orchestrator v0.1.0-dev
# Use 'orchestrator --help' for usage information.
```

## Checklist
- [ ] File `cmd/orchestrator/main.go` tồn tại
- [ ] Package: `package main`
- [ ] `main()` function exists
- [ ] `go build -o bin/orchestrator ./cmd/orchestrator/` thành công
- [ ] Binary chạy và in version
