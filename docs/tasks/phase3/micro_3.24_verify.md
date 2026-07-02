# Micro-Task 3.24: Verification — Build & Test toàn bộ Phase 3

## Thông tin
- **File tạo**: Không tạo file nào (chỉ verify)
- **Dependencies trước**: TẤT CẢ micro-tasks 3.01 → 3.23
- **Thời gian**: 15 phút
- **Mục đích**: Đảm bảo TẤT CẢ các thư viện SDK (Base classes, Middlewares, và Helpers) biên dịch thành công, không bị lỗi import vòng tròn (import cycles), và vượt qua tất cả các tests với bộ phát hiện tranh chấp bộ nhớ (race detector).

## Lệnh verify (PHẢI chạy theo đúng thứ tự)

### Step 1: Kiểm tra các tệp tin SDK tồn tại
```bash
# Plugin SDK
ls sdk/plugin/plugin.go

# Agent SDK
ls sdk/agent/manifest.go
ls sdk/agent/prompt.go
ls sdk/agent/agent.go
ls sdk/agent/agent_test.go

# Provider SDK
ls sdk/provider/provider.go
ls sdk/provider/request.go
ls sdk/provider/stream.go
ls sdk/provider/provider_test.go

# Tool SDK
ls sdk/tool/tool.go
ls sdk/tool/result.go
ls sdk/tool/tool_test.go

# Support Skeletons
ls sdk/workflow/workflow.go
ls sdk/context/builder.go
ls sdk/memory/memory.go
ls sdk/search/search.go
ls sdk/task/task.go

# Middlewares & Helpers
ls sdk/middleware/agent.go
ls sdk/middleware/provider.go
ls sdk/middleware/middleware_test.go
ls sdk/helpers/ratelimit.go

# Testing SDK Mocks
ls sdk/testing/mocks.go
ls sdk/testing/mocks_test.go
```

### Step 2: Go Build (kiểm tra compile)
```bash
go build ./sdk/...
# Expected: không output, không error
```

### Step 3: Go Vet (kiểm tra tĩnh)
```bash
go vet ./sdk/...
# Expected: không output, không warning
```

### Step 4: Go Test (chạy unit tests)
```bash
go test -v ./sdk/...
# Expected: ALL PASS
```

### Step 5: Go Test với Race Detector
```bash
go test -race ./sdk/...
# Expected: ALL PASS, no race conditions detected
```

### Step 6: Import Cycle Check
```bash
go build ./...
# Đảm bảo không xảy ra import cycles giữa contracts, kernel, và sdk.
# Đồ thị import hợp lệ:
#   contracts/ ← kernel/ ← sdk/
#   sdk/ chỉ được import từ contracts/ hoặc kernel/
#   sdk/ tuyệt đối KHÔNG được import từ plugins/, modules/ hay api/
```

### Step 7: Git Commit
```bash
git add -A
git commit -m "Phase 3: SDK Developer Helpers implementation (24 micro-tasks)"
git push origin main
```

## Checklist kiểm tra chất lượng Phase 3

### SDK Core Packages
- [ ] `sdk/plugin/plugin.go` — Định nghĩa `BasePlugin` hỗ trợ đầy đủ trạng thái `initialized`, `started` và report health 1.40.
- [ ] `sdk/agent/manifest.go` — Hàm `LoadManifest` phân giải PromptFile tương đối theo thư mục tệp YAML.
- [ ] `sdk/agent/prompt.go` — `BuildPrompt` tách biệt rõ ràng Instruction và Context Items thành các messages riêng.
- [ ] `sdk/agent/agent.go` — `BaseAgent` thực hiện ReAct loop song song hóa tool calls (`sync.WaitGroup`) và bảo vệ chống lặp vô hạn.
- [ ] `sdk/provider/provider.go` — `BaseProvider` sao chép slice `models` trước khi trả về để tránh đột biến bộ nhớ ngoài luồng.
- [ ] `sdk/provider/request.go` — `RequestBuilder` triển khai dạng bất biến (immutable) an toàn đa luồng.
- [ ] `sdk/provider/stream.go` — `CollectStream` tự động drain channel trong background khi hủy context để tránh rò rỉ producer.
- [ ] `sdk/tool/tool.go` — `BaseTool` tự động xác thực kiểu dữ liệu JSON thô (bao gồm ép kiểu integer của float64) đối chiếu với schema.
- [ ] `sdk/tool/result.go` — Trình sinh kết quả `JSON()` trả về system error nếu marshal thất bại.
- [ ] `sdk/middleware/agent.go` — Đầy đủ Agent Logging, Metrics, và Recovery middlewares.
- [ ] `sdk/middleware/provider.go` — Đầy đủ Provider Logging, Retry, CircuitBreaker, và Metrics middlewares.
- [ ] `sdk/helpers/ratelimit.go` — Triển khai Rate Limiter dạng Token Bucket an toàn đa luồng, giải phóng lock trước khi chờ.
- [ ] `sdk/testing/mocks.go` — `MockEventBus` bảo vệ ghi đồng thời bằng `sync.RWMutex`.

### Quality Gates
- [ ] `go build ./...` ✅ (không lỗi compile toàn dự án)
- [ ] `go test ./sdk/...` ALL PASS
- [ ] `go test -race ./sdk/...` NO RACES
- [ ] Không chứa import cycles
- [ ] Commit và push lên origin main thành công
