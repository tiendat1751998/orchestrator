# Micro-Task 1.01: Tạo go.mod

## Thông tin
- **File tạo**: `go.mod`
- **Dependencies trước**: Không
- **Thời gian**: 5 phút
- **Verify**: `go mod tidy` không lỗi

## Nội dung CHÍNH XÁC cần tạo

```go
module github.com/tiendat1751998/orchestrator

go 1.23.0
```

## Quy tắc
1. KHÔNG thêm bất kỳ `require` nào lúc này
2. Module path PHẢI khớp với GitHub repo URL
3. Go version: dùng version stable mới nhất (≥ 1.23)

## Lệnh verify
```bash
go mod tidy
# Expected: không output, không error
```

## Checklist
- [ ] File `go.mod` tồn tại ở root
- [ ] Module path đúng: `github.com/tiendat1751998/orchestrator`
- [ ] Go version ≥ 1.23
- [ ] `go mod tidy` không lỗi
