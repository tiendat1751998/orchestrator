# Phase 3: Plugin System (SDK) — Chi Tiết Từng Sub-Task

> [!NOTE]
> SDK là "bộ công cụ" giúp viết plugins nhanh chóng. Mục tiêu: viết 1 agent mới chỉ cần <50 dòng code.

---

## Task 3.1: SDK — Agent SDK

### Sub-task 3.1.1: BaseAgent struct
- **File**: `sdk/agent/agent.go`
- **Chi tiết**:
  ```go
  // BaseAgent cung cấp implementation mặc định cho Agent interface
  // Developer chỉ cần embed và override methods cần thiết
  type BaseAgent struct {
      manifest *Manifest
      provider provider.Provider
      tools    map[string]tool.Tool
      logger   *slog.Logger
  }

  func NewBaseAgent(manifest *Manifest, p provider.Provider, logger *slog.Logger) *BaseAgent { ... }

  // Default implementations
  func (a *BaseAgent) Name() string { return a.manifest.Name }
  func (a *BaseAgent) Role() string { return a.manifest.Role }
  func (a *BaseAgent) Capabilities() []Capability { return a.manifest.Capabilities }
  func (a *BaseAgent) CanHandle(task *Task) bool {
      // Default: check nếu task.Type nằm trong capabilities
  }

  // Execute — implementation mặc định:
  // 1. Build prompt từ system prompt + task description
  // 2. Gọi provider.Send()
  // 3. Nếu response có tool calls → execute tools → gửi results lại
  // 4. Lặp lại cho đến khi không còn tool calls
  func (a *BaseAgent) Execute(ctx context.Context, task *Task) (*Result, error) { ... }
  ```
- **⚠️ Pitfall #1**: Tool call loop phải có MAX ITERATIONS limit (ví dụ: 20). Nếu AI liên tục gọi tool → infinite loop → goroutine leak. 
- **⚠️ Pitfall #2**: Mỗi iteration phải check `ctx.Done()`. Nếu context bị cancel giữa chừng → dừng ngay.
- **⚠️ Pitfall #3**: Provider có thể trả về tool calls KHÔNG hợp lệ (tool name sai, args sai format). BaseAgent phải handle gracefully — trả error message cho AI biết tool call sai, cho AI cơ hội sửa.

### Sub-task 3.1.2: Manifest Loader
- **File**: `sdk/agent/manifest.go`
- **Chi tiết**:
  ```go
  func LoadManifest(path string) (*Manifest, error) {
      // 1. Read YAML file
      // 2. Parse into Manifest struct
      // 3. Validate required fields
      // 4. Load system prompt from PromptFile if specified
      // 5. Return manifest
  }
  ```
- **⚠️ Pitfall**: `PromptFile` path resolution. Path nên relative to manifest file location, không phải working directory.

### Sub-task 3.1.3: Prompt Builder cho Agent
- **File**: `sdk/agent/prompt.go`
- **Chi tiết**:
  ```go
  func BuildPrompt(manifest *Manifest, task *Task) []provider.Message {
      messages := []provider.Message{
          {Role: provider.RoleSystem, Content: manifest.SystemPrompt},
          {Role: provider.RoleUser, Content: formatTaskAsPrompt(task)},
      }
      // Thêm context items từ task
      for _, ctx := range task.Context {
          messages = append(messages, ...)
      }
      return messages
  }
  ```
- **⚠️ Pitfall**: Token limit. Nếu system prompt + task context quá dài → vượt model context window. Cần truncation strategy.

### Sub-task 3.1.4: Agent lifecycle hooks
- **File**: `sdk/agent/lifecycle.go`
- **Chi tiết**: Default Init(), Start(), Stop() implementations cho Plugin interface

### Sub-task 3.1.5: Unit tests
- **Tests**:
  - BaseAgent.Execute() với mock provider (trả response không có tool calls)
  - BaseAgent.Execute() với tool calls (1 iteration)
  - BaseAgent.Execute() với max iterations limit
  - Manifest loading từ YAML
  - CanHandle matching

### Tiêu chí hoàn thành Task 3.1:
- [ ] BaseAgent embed pattern hoạt động
- [ ] Tool call loop có max iterations
- [ ] Context cancellation hoạt động
- [ ] Manifest loading từ YAML
- [ ] Unit tests ≥ 85% coverage

---

## Task 3.2: SDK — Provider SDK

### Sub-task 3.2.1: BaseProvider struct
- **File**: `sdk/provider/provider.go`
- **Chi tiết**:
  ```go
  type BaseProvider struct {
      config *provider.Config
      logger *slog.Logger
  }

  func (p *BaseProvider) Name() string { return p.config.Name }
  func (p *BaseProvider) IsAvailable(ctx context.Context) bool { ... }
  ```
- **Mục đích**: Giảm boilerplate khi viết provider mới. Developer chỉ cần implement `Send()` và `Stream()`.

### Sub-task 3.2.2: Request Builder
- **File**: `sdk/provider/request.go`
- **Chi tiết**: Helper functions:
  ```go
  func NewRequest(model string) *provider.Request { ... }
  func (r *RequestBuilder) WithMessages(msgs ...provider.Message) *RequestBuilder { ... }
  func (r *RequestBuilder) WithTools(tools ...provider.Tool) *RequestBuilder { ... }
  func (r *RequestBuilder) WithTemperature(t float64) *RequestBuilder { ... }
  func (r *RequestBuilder) Build() *provider.Request { ... }
  ```
- **⚠️ Pitfall**: Builder pattern phải immutable. Mỗi method trả về new builder, không mutate original. Đảm bảo concurrent safety.

### Sub-task 3.2.3: Response Parser Helpers
- **File**: `sdk/provider/response.go`
- **Chi tiết**: Helper để extract thông tin từ response:
  ```go
  func HasToolCalls(resp *provider.Response) bool { ... }
  func ExtractToolCalls(resp *provider.Response) []provider.ToolCall { ... }
  func ExtractContent(resp *provider.Response) string { ... }
  ```

### Sub-task 3.2.4: Stream Processing Utilities
- **File**: `sdk/provider/stream.go`
- **Chi tiết**:
  ```go
  // CollectStream đọc toàn bộ stream và ghép thành Response đầy đủ
  func CollectStream(ctx context.Context, ch <-chan provider.StreamChunk) (*provider.Response, error) { ... }

  // ForwardStream forward chunks tới callback
  func ForwardStream(ctx context.Context, ch <-chan provider.StreamChunk, callback func(chunk provider.StreamChunk)) error { ... }
  ```
- **⚠️ Pitfall #1**: `CollectStream` PHẢI handle context cancellation. Nếu ctx cancelled → drain channel → return error.
- **⚠️ Pitfall #2**: Channel PHẢI được drained. Nếu consumer không đọc hết channel → producer goroutine bị block → leak.

### Sub-task 3.2.5: Unit tests
- **Tests**: Request builder, Response parser, Stream collector, stream cancellation

### Tiêu chí hoàn thành Task 3.2:
- [ ] BaseProvider giảm boilerplate
- [ ] Request builder pattern
- [ ] Stream utilities handle cancellation
- [ ] Unit tests ≥ 85% coverage

---

## Task 3.3: SDK — Tool, Workflow, và các SDK khác

### Sub-task 3.3.1: BaseTool struct
- **File**: `sdk/tool/tool.go`
- **Chi tiết**:
  ```go
  type BaseTool struct {
      name        string
      description string
      schema      *tool.Schema
  }

  func NewBaseTool(name, desc string, schema *tool.Schema) *BaseTool { ... }
  func (t *BaseTool) Name() string { ... }
  func (t *BaseTool) Description() string { ... }
  func (t *BaseTool) Schema() *tool.Schema { ... }
  ```
- **⚠️ Pitfall**: Schema validation. Trước khi execute, validate args against schema. Trả error rõ ràng nếu args không khớp schema.

### Sub-task 3.3.2: Tool Result Builder
- **File**: `sdk/tool/result.go`
- **Chi tiết**:
  ```go
  func Success(output string) *tool.Result { return &tool.Result{Output: output, ExitCode: 0} }
  func Failure(err string) *tool.Result { return &tool.Result{Error: err, ExitCode: 1} }
  func WithExitCode(output string, code int) *tool.Result { ... }
  ```

### Sub-task 3.3.3: Các SDK còn lại (skeleton)
- **Files**:
  - `sdk/workflow/workflow.go` — BaseWorkflow
  - `sdk/context/builder.go` — Context builder utilities
  - `sdk/memory/memory.go` — Memory store helpers
  - `sdk/search/search.go` — Search engine helpers
  - `sdk/plugin/plugin.go` — Generic BasePlugin
  - `sdk/event/event.go` — Event helpers
  - `sdk/task/task.go` — Task builder helpers
- **Lưu ý**: Phase 3 chỉ cần skeleton. Flesh out khi cần ở Phase 4+.

### Sub-task 3.3.4: Unit tests
- **Tests**: BaseTool, Result builders, Schema validation

### Tiêu chí hoàn thành Task 3.3:
- [ ] BaseTool hoạt động
- [ ] Result builders cho success/failure
- [ ] SDK skeleton cho tất cả plugin types
- [ ] Unit tests ≥ 80% coverage

---

## 📋 Checklist tổng Phase 3

- [ ] Task 3.1: Agent SDK (5 sub-tasks)
- [ ] Task 3.2: Provider SDK (5 sub-tasks)
- [ ] Task 3.3: Tool + other SDKs (4 sub-tasks)
- [ ] Viết 1 mock agent chỉ cần <50 dòng code → validate SDK usability
- [ ] Viết 1 mock provider chỉ cần implement Send() → validate SDK usability
- [ ] `go test ./sdk/...` pass
- [ ] Git commit: "Phase 3: SDK implementation"
