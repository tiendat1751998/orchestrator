# Micro-Task 1.16: Tạo contracts/tool/tool_test.go

## Thông tin
- **File tạo**: `contracts/tool/tool_test.go`
- **Package**: `tool_test`
- **Dependencies trước**: 1.14, 1.15
- **Thời gian**: 15 phút

## Tests CHÍNH XÁC cần viết

1. **TestSchema_JSONRoundTrip**: Tạo Schema → marshal → unmarshal → compare
2. **TestSchema_Builder**: `NewSchema().AddProperty("path", StringProperty("...")).AddRequired("path")` → verify properties
3. **TestResult_IsSuccess**: ExitCode=0 → true, ExitCode=1 → false
4. **TestResult_String**: Success → output, Failure → "error: ..."
5. **TestStringProperty**: Verify type="string" và description set đúng
6. **TestEnumProperty**: Verify enum values preserved
7. **TestProperty_WithItems**: Array property với Items

## Lệnh verify
```bash
go test -v ./contracts/tool/...
# Expected: ALL PASS
```

## Checklist
- [ ] ≥ 7 test functions
- [ ] Schema builder tests
- [ ] Result helper tests
- [ ] JSON round-trip tests
- [ ] ALL PASS
