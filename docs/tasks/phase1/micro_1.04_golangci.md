# Micro-Task 1.04: Tạo .golangci.yml

## Thông tin
- **File tạo**: `.golangci.yml`
- **Dependencies trước**: 1.01
- **Thời gian**: 5 phút
- **Verify**: `golangci-lint run ./...` không config error

## Nội dung CHÍNH XÁC cần tạo

```yaml
# golangci-lint configuration
# Docs: https://golangci-lint.run/usage/configuration/

run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck      # Check error return values
    - govet         # Reports suspicious constructs
    - staticcheck   # Advanced static analysis
    - unused        # Detect unused code
    - gosimple      # Suggest code simplifications
    - gocritic      # Best practice checker
    - ineffassign   # Detect ineffectual assignments
    - typecheck     # Type checking
    - misspell      # Detect misspelled words
    - goimports     # Check import formatting

linters-settings:
  errcheck:
    check-type-assertions: true  # Check type assertion errors
    check-blank: false           # Don't check _ = err (intentional ignore)
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
  misspell:
    locale: US

issues:
  max-issues-per-linter: 50
  max-same-issues: 5
  exclude-rules:
    # Allow unused parameters in interface implementations
    - linters:
        - unused
      source: "func \\(.*\\) .* {"
    # Allow dot imports in test files
    - linters:
        - gocritic
      path: _test\.go
```

## Quy tắc
1. Chỉ bật linters CẦN THIẾT — quá nhiều sẽ gây noise
2. `timeout: 5m` — đủ cho project lớn
3. `check-type-assertions: true` — bắt lỗi type assertion không check ok
4. Exclude test files từ một số rules nghiêm ngặt

## ⚠️ Pitfalls cần tránh
1. **Quá nhiều linters**: Bật tất cả → hàng trăm warnings → developer bỏ qua tất cả. Chỉ bật 10-12 linters quan trọng nhất
2. **golangci-lint version**: Config format thay đổi giữa versions. Dùng v1.59+ 
3. **IDE conflict**: Một số IDE chạy linter riêng. Đảm bảo IDE dùng cùng `.golangci.yml`

## Checklist
- [ ] File `.golangci.yml` tồn tại ở root
- [ ] `errcheck` enabled
- [ ] `govet` enabled
- [ ] `staticcheck` enabled
- [ ] `timeout` ≥ 3m
- [ ] Test files có exclude rules
