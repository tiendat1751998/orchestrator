# Phase 1: Contracts & Foundation — Chi Tiết Từng Sub-Task

> [!IMPORTANT]
> **Nguyên tắc vàng của Phase này**: Contracts KHÔNG ĐƯỢC thay đổi sau khi Phase 2 bắt đầu. Mọi sai sót ở đây sẽ lan ra toàn bộ hệ thống. Dành thời gian suy nghĩ kỹ trước khi code.

---

## Task 1.1: Khởi tạo Go Module & Project Setup

### Sub-task 1.1.1: Tạo `go.mod`
- **File**: `go.mod`
- **Chi tiết**:
  - Module path: `github.com/tiendat1751998/orchestrator`
  - Go version: `go 1.23` (hoặc latest stable)
  - KHÔNG thêm dependencies nào vào lúc này — chỉ thêm khi cần
- **⚠️ Pitfall**: Đừng dùng `go 1.24` nếu chưa stable. Chọn version đã release ít nhất 3 tháng.

### Sub-task 1.1.2: Tạo `.gitignore`
- **File**: `.gitignore`
- **Nội dung cần ignore**:
  ```
  # Binaries
  /bin/
  *.exe
  *.dll
  *.so
  *.dylib

  # Test
  *.test
  *.out
  coverage.html
  coverage.txt

  # Vendor (nếu dùng)
  /vendor/

  # IDE
  .idea/
  .vscode/
  *.swp
  *.swo

  # OS
  .DS_Store
  Thumbs.db

  # Build
  /dist/
  /tmp/

  # Env
  .env
  .env.local
  ```
- **⚠️ Pitfall**: Nhiều người quên ignore `.env` → lộ API keys khi push lên GitHub.

### Sub-task 1.1.3: Tạo `Makefile`
- **File**: `Makefile`
- **Targets cần có**:
  ```makefile
  .PHONY: build test lint clean run fmt vet

  build:        # go build -o bin/orchestrator ./cmd/orchestrator/
  test:         # go test -v -race -coverprofile=coverage.out ./...
  lint:         # golangci-lint run ./...
  clean:        # rm -rf bin/ dist/ coverage.out
  run:          # go run ./cmd/orchestrator/
  fmt:          # gofmt -w .
  vet:          # go vet ./...
  ```
- **⚠️ Pitfall**: Makefile dùng TAB, không dùng SPACE. Trên Windows cần chú ý encoding.

### Sub-task 1.1.4: Tạo `.golangci.yml`
- **File**: `.golangci.yml`
- **Linters cần bật**:
  - `errcheck` — Bắt lỗi không check error return
  - `govet` — Phát hiện code suspicious
  - `staticcheck` — Phân tích tĩnh nâng cao
  - `unused` — Phát hiện code không dùng
  - `gosimple` — Gợi ý code đơn giản hơn
  - `gocritic` — Best practices
- **⚠️ Pitfall**: Đừng bật quá nhiều linters lúc đầu → sẽ bị overwhelmed bởi warnings.

### Sub-task 1.1.5: Tạo `cmd/orchestrator/main.go` (skeleton)
- **File**: `cmd/orchestrator/main.go`
- **Nội dung**: Chỉ có `package main` và `func main()` trống
- **Mục đích**: Đảm bảo `go build ./...` chạy được

### Tiêu chí hoàn thành Task 1.1:
- [ ] `go build ./...` không lỗi
- [ ] `go vet ./...` không warning
- [ ] `.gitignore` cover đủ các file cần thiết
- [ ] `make build` chạy thành công

---

## Task 1.2: Contracts — Provider Interface

> [!CAUTION]
> **Đây là interface QUAN TRỌNG NHẤT.** Mọi provider (Antigravity, Gemini, Claude, Ollama) đều phải implement interface này. Nếu thiết kế sai → phải sửa lại TẤT CẢ providers.

### Sub-task 1.2.1: Định nghĩa `Message` types
- **File**: `contracts/provider/message.go`
- **Structs cần định nghĩa**:
  ```go
  // Role của message
  type Role string
  const (
      RoleSystem    Role = "system"
      RoleUser      Role = "user"
      RoleAssistant Role = "assistant"
      RoleTool      Role = "tool"
  )

  // Message trong conversation
  type Message struct {
      Role       Role        `json:"role"`
      Content    string      `json:"content"`
      ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
      ToolCallID string      `json:"tool_call_id,omitempty"`
  }

  // ToolCall — khi AI muốn gọi một tool
  type ToolCall struct {
      ID       string          `json:"id"`
      Name     string          `json:"name"`
      Args     json.RawMessage `json:"arguments"`
  }
  ```
- **⚠️ Pitfall #1**: Dùng `json.RawMessage` cho `Args` thay vì `map[string]interface{}`. Lý do: `RawMessage` giữ nguyên JSON gốc, không mất precision khi parse numbers.
- **⚠️ Pitfall #2**: `ToolCallID` dùng cho response message khi tool trả kết quả về. Nhiều người quên field này → agent không biết kết quả thuộc tool call nào.
- **⚠️ Pitfall #3**: KHÔNG dùng `int` cho enum Role → dùng `string` vì dễ debug, dễ serialize, và tương thích với mọi AI API.

### Sub-task 1.2.2: Định nghĩa `Request` struct
- **File**: `contracts/provider/request.go`
- **Struct cần định nghĩa**:
  ```go
  type Request struct {
      Model       string    `json:"model"`
      Messages    []Message `json:"messages"`
      Tools       []Tool    `json:"tools,omitempty"`
      Temperature *float64  `json:"temperature,omitempty"`
      MaxTokens   *int      `json:"max_tokens,omitempty"`
      TopP        *float64  `json:"top_p,omitempty"`
      StopWords   []string  `json:"stop,omitempty"`
      Stream      bool      `json:"stream,omitempty"`
  }

  // Tool definition cho function calling
  type Tool struct {
      Name        string          `json:"name"`
      Description string          `json:"description"`
      Parameters  json.RawMessage `json:"parameters"` // JSON Schema
  }
  ```
- **⚠️ Pitfall #1**: `Temperature` và `MaxTokens` dùng POINTER (`*float64`, `*int`) thay vì value. Lý do: phân biệt giữa "user set = 0" và "user không set" (nil). Nếu dùng value type, `Temperature: 0` và "không set" đều là `0` → bug.
- **⚠️ Pitfall #2**: `Tools` dùng JSON Schema format cho `Parameters`. Đây là chuẩn chung của OpenAI, Gemini, Claude. KHÔNG tự chế format riêng.
- **⚠️ Pitfall #3**: Field `Stream` nằm trong Request để provider biết user muốn streaming hay không. Đừng tạo 2 method `Send()` và `SendStream()` riêng — dùng 1 field boolean gọn hơn, nhưng VẪN cần 2 method ở interface vì return type khác nhau.

### Sub-task 1.2.3: Định nghĩa `Response` struct
- **File**: `contracts/provider/response.go`
- **Struct cần định nghĩa**:
  ```go
  type Response struct {
      ID           string     `json:"id"`
      Content      string     `json:"content"`
      ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
      FinishReason string     `json:"finish_reason"`
      Usage        Usage      `json:"usage"`
      Model        string     `json:"model"`
  }

  type Usage struct {
      PromptTokens     int `json:"prompt_tokens"`
      CompletionTokens int `json:"completion_tokens"`
      TotalTokens      int `json:"total_tokens"`
  }

  // StreamChunk — mỗi chunk khi streaming
  type StreamChunk struct {
      Delta      string     `json:"delta"`
      ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
      Done       bool       `json:"done"`
      FinishReason string   `json:"finish_reason,omitempty"`
  }
  ```
- **⚠️ Pitfall #1**: `FinishReason` dùng `string` thay vì enum. Lý do: mỗi provider trả giá trị khác nhau (`"stop"`, `"end_turn"`, `"max_tokens"`, `"tool_calls"`). Nếu dùng enum cứng sẽ break khi thêm provider mới.
- **⚠️ Pitfall #2**: `Usage` PHẢI có. Dù ban đầu không cần, nhưng sau này cần theo dõi chi phí API. Thêm sau sẽ phải sửa tất cả providers.
- **⚠️ Pitfall #3**: `StreamChunk.Done` flag rất quan trọng. Consumer cần biết khi nào stream kết thúc. Nếu thiếu → goroutine leak vì consumer đợi mãi.

### Sub-task 1.2.4: Định nghĩa `Provider` interface
- **File**: `contracts/provider/provider.go`
- **Interface cần định nghĩa**:
  ```go
  type Provider interface {
      // Name trả về tên unique của provider (vd: "antigravity", "gemini")
      Name() string

      // Send gửi request và đợi response đầy đủ
      // Context dùng để cancel/timeout
      Send(ctx context.Context, req *Request) (*Response, error)

      // Stream gửi request và trả về channel streaming chunks
      // Channel sẽ được close khi stream kết thúc hoặc lỗi
      Stream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

      // IsAvailable kiểm tra provider có sẵn sàng không
      // (vd: API key hợp lệ, CLI binary tồn tại, service đang chạy)
      IsAvailable(ctx context.Context) bool

      // Models trả về danh sách models mà provider hỗ trợ
      Models(ctx context.Context) ([]string, error)
  }
  ```
- **⚠️ Pitfall #1**: `Send()` và `Stream()` PHẢI nhận `context.Context` là param đầu tiên. Lý do:
  - Cho phép timeout: `ctx, cancel := context.WithTimeout(ctx, 30*time.Second)`
  - Cho phép cancel: user nhấn Ctrl+C → cancel tất cả provider calls
  - Đây là Go convention bắt buộc
- **⚠️ Pitfall #2**: `Stream()` trả về `<-chan StreamChunk` (read-only channel), KHÔNG phải `chan StreamChunk`. Lý do: consumer không được phép ghi vào channel.
- **⚠️ Pitfall #3**: `Stream()` trả về `error` ngay lập tức nếu không kết nối được. Errors xảy ra giữa stream sẽ được gửi qua channel (StreamChunk có thể chứa error info). Đây là pattern chuẩn trong Go.
- **⚠️ Pitfall #4**: KHÔNG đặt `Close()` hoặc `Shutdown()` trong Provider interface. Provider lifecycle được quản lý bởi Plugin interface riêng (Task 1.5). Mixing concerns = disaster.

### Sub-task 1.2.5: Định nghĩa `ProviderConfig` struct
- **File**: `contracts/provider/config.go`
- **Struct cần định nghĩa**:
  ```go
  type Config struct {
      Name     string            `yaml:"name"`
      Type     string            `yaml:"type"` // "cli", "api", "local"
      Model    string            `yaml:"model"`
      BaseURL  string            `yaml:"base_url,omitempty"`
      APIKey   string            `yaml:"api_key,omitempty"`
      Binary   string            `yaml:"binary,omitempty"` // Path tới CLI binary
      Timeout  time.Duration     `yaml:"timeout"`
      MaxRetry int               `yaml:"max_retry"`
      Extra    map[string]string `yaml:"extra,omitempty"`
  }
  ```
- **⚠️ Pitfall #1**: `APIKey` KHÔNG nên đọc trực tiếp từ YAML. Nên hỗ trợ `${ENV_VAR}` syntax để đọc từ environment variable. Tránh commit API key vào Git.
- **⚠️ Pitfall #2**: `Binary` field cho CLI-based providers (Antigravity). API-based providers sẽ dùng `BaseURL` + `APIKey`.
- **⚠️ Pitfall #3**: `Extra` map cho các config đặc thù của từng provider mà chúng ta không biết trước. Đây là escape hatch để tránh phải sửa Config struct mỗi khi thêm provider mới.

### Sub-task 1.2.6: Viết unit tests cho types
- **File**: `contracts/provider/provider_test.go`
- **Tests cần viết**:
  - Test JSON serialization/deserialization của Message, Request, Response
  - Test pointer fields (Temperature = nil vs Temperature = 0)
  - Test ToolCall Args parse correctly
- **⚠️ Pitfall**: LUÔN test JSON round-trip. Nhiều bug xảy ra do `omitempty` tag bị thiếu hoặc sai.

### Tiêu chí hoàn thành Task 1.2:
- [ ] Tất cả struct có JSON tags
- [ ] Tất cả struct có YAML tags (cho config)
- [ ] Pointer types cho optional fields
- [ ] `context.Context` ở mọi method có I/O
- [ ] Unit tests cho JSON serialization
- [ ] `go vet` không warning
- [ ] Godoc comment cho mọi exported type/method

---

## Task 1.3: Contracts — Agent Interface

### Sub-task 1.3.1: Định nghĩa `AgentCapability` type
- **File**: `contracts/agent/capability.go`
- **Structs cần định nghĩa**:
  ```go
  // Capability mô tả một khả năng cụ thể của agent
  type Capability string

  const (
      CapCodeGeneration   Capability = "code_generation"
      CapCodeReview       Capability = "code_review"
      CapArchitecture     Capability = "architecture"
      CapTesting          Capability = "testing"
      CapDocumentation    Capability = "documentation"
      CapDeployment       Capability = "deployment"
      CapDebugging        Capability = "debugging"
      CapRefactoring      Capability = "refactoring"
  )
  ```
- **⚠️ Pitfall**: Dùng `string` constants, KHÔNG dùng `iota`. Lý do: capabilities cần serialize vào YAML/JSON, `iota` sẽ trả về số → khó đọc, dễ break khi thêm/xóa item.

### Sub-task 1.3.2: Định nghĩa `Task` struct
- **File**: `contracts/agent/task.go`
- **Structs cần định nghĩa**:
  ```go
  type Task struct {
      ID           string            `json:"id"`
      Name         string            `json:"name"`
      Description  string            `json:"description"`
      Type         string            `json:"type"` // "code_gen", "review", "test"...
      Input        map[string]any    `json:"input"`
      Context      []ContextItem     `json:"context,omitempty"`
      Dependencies []string          `json:"dependencies,omitempty"` // IDs of tasks this depends on
      Priority     int               `json:"priority"`
      Timeout      time.Duration     `json:"timeout"`
      Metadata     map[string]string `json:"metadata,omitempty"`
  }

  type ContextItem struct {
      Type    string `json:"type"` // "file", "snippet", "url", "memory"
      Content string `json:"content"`
      Source  string `json:"source,omitempty"`
  }
  ```
- **⚠️ Pitfall #1**: `Input` dùng `map[string]any` thay vì struct cụ thể. Lý do: mỗi loại task có input khác nhau. Agent "backend" cần `{language: "go", framework: "gin"}`, agent "reviewer" cần `{diff: "..."}`. Struct cứng sẽ không linh hoạt.
- **⚠️ Pitfall #2**: `Dependencies` là list of Task IDs. Planner sẽ dùng field này để build DAG. KHÔNG dùng pointer tới Task khác → circular dependency → memory leak.
- **⚠️ Pitfall #3**: `Timeout` PER TASK, không phải global timeout. Mỗi task có độ phức tạp khác nhau → timeout khác nhau.

### Sub-task 1.3.3: Định nghĩa `Result` struct
- **File**: `contracts/agent/result.go`
- **Structs cần định nghĩa**:
  ```go
  type Result struct {
      TaskID    string        `json:"task_id"`
      AgentName string        `json:"agent_name"`
      Status    Status        `json:"status"`
      Output    string        `json:"output"`
      Artifacts []Artifact    `json:"artifacts,omitempty"`
      Error     string        `json:"error,omitempty"`
      Duration  time.Duration `json:"duration"`
      Usage     *Usage        `json:"usage,omitempty"` // Token usage from provider
      Metadata  map[string]string `json:"metadata,omitempty"`
  }

  type Artifact struct {
      Name    string `json:"name"`
      Type    string `json:"type"` // "file", "diff", "log", "report"
      Path    string `json:"path,omitempty"`
      Content string `json:"content,omitempty"`
  }

  type Status string
  const (
      StatusPending   Status = "pending"
      StatusRunning   Status = "running"
      StatusSuccess   Status = "success"
      StatusFailed    Status = "failed"
      StatusCancelled Status = "cancelled"
      StatusRetrying  Status = "retrying"
  )
  ```
- **⚠️ Pitfall #1**: `Error` là `string`, KHÔNG phải `error` interface. Lý do: `error` interface không serialize được sang JSON. Lưu error message dạng string.
- **⚠️ Pitfall #2**: `Artifacts` cho phép agent trả về nhiều output (files, diffs, reports). KHÔNG chỉ trả về 1 string output → sẽ hạn chế agent trong tương lai.
- **⚠️ Pitfall #3**: `Usage` là pointer vì có thể nil (khi agent không sử dụng provider, ví dụ agent chỉ chạy tool).

### Sub-task 1.3.4: Định nghĩa `Agent` interface
- **File**: `contracts/agent/agent.go`
- **Interface cần định nghĩa**:
  ```go
  type Agent interface {
      // Name trả về tên unique (vd: "backend", "reviewer")
      Name() string

      // Role mô tả vai trò (vd: "Backend Developer", "Code Reviewer")
      Role() string

      // Capabilities trả về danh sách khả năng
      Capabilities() []Capability

      // Execute thực hiện một task và trả về kết quả
      Execute(ctx context.Context, task *Task) (*Result, error)

      // CanHandle kiểm tra agent có xử lý được task này không
      CanHandle(task *Task) bool
  }
  ```
- **⚠️ Pitfall #1**: `CanHandle()` rất quan trọng. Orchestrator sẽ gọi method này để tìm agent phù hợp cho task. Nếu thiếu → phải hard-code mapping agent ↔ task type → không extensible.
- **⚠️ Pitfall #2**: `Execute()` trả về `(*Result, error)`. `error` ở đây là system error (network, timeout). Business error (code gen fail) nằm trong `Result.Status = StatusFailed`. PHẢI phân biệt 2 loại error này.
- **⚠️ Pitfall #3**: KHÔNG đặt `SetProvider()` trong Agent interface. Provider được inject qua constructor khi khởi tạo agent, không phải runtime method. Đây là Dependency Injection pattern.

### Sub-task 1.3.5: Định nghĩa `AgentManifest` struct
- **File**: `contracts/agent/manifest.go`
- **Struct cần định nghĩa**:
  ```go
  type Manifest struct {
      Name         string       `yaml:"name"`
      Version      string       `yaml:"version"`
      Role         string       `yaml:"role"`
      Description  string       `yaml:"description"`
      Capabilities []Capability `yaml:"capabilities"`
      Provider     string       `yaml:"provider"` // Provider name to use
      Model        string       `yaml:"model,omitempty"`
      Tools        []string     `yaml:"tools,omitempty"` // Tool names this agent can use
      SystemPrompt string       `yaml:"system_prompt,omitempty"`
      PromptFile   string       `yaml:"prompt_file,omitempty"` // Path to system prompt file
      MaxTokens    int          `yaml:"max_tokens,omitempty"`
      Temperature  float64      `yaml:"temperature,omitempty"`
  }
  ```
- **⚠️ Pitfall #1**: `SystemPrompt` và `PromptFile` — hỗ trợ cả inline prompt VÀ file path. Prompt dài nên để trong file riêng, prompt ngắn có thể inline.
- **⚠️ Pitfall #2**: `Provider` là tên (string), KHÔNG phải instance. Registry sẽ resolve tên → instance lúc runtime.

### Sub-task 1.3.6: Viết unit tests
- **File**: `contracts/agent/agent_test.go`
- **Tests cần viết**:
  - Test Status transitions hợp lệ
  - Test Task serialization với Dependencies
  - Test Result serialization với Artifacts
  - Test Manifest YAML parsing

### Tiêu chí hoàn thành Task 1.3:
- [ ] Task struct hỗ trợ dependencies
- [ ] Result struct hỗ trợ multiple artifacts
- [ ] Agent interface có CanHandle() method
- [ ] Phân biệt rõ system error vs business error
- [ ] Unit tests cho serialization
- [ ] Godoc comment đầy đủ

---

## Task 1.4: Contracts — Tool Interface

### Sub-task 1.4.1: Định nghĩa `ToolSchema` (JSON Schema)
- **File**: `contracts/tool/schema.go`
- **Struct cần định nghĩa**:
  ```go
  // Schema mô tả input parameters của tool, dùng JSON Schema format
  type Schema struct {
      Type        string                `json:"type"` // always "object"
      Properties  map[string]Property   `json:"properties"`
      Required    []string              `json:"required,omitempty"`
  }

  type Property struct {
      Type        string   `json:"type"` // "string", "integer", "boolean", "array"
      Description string   `json:"description"`
      Enum        []string `json:"enum,omitempty"`
      Default     any      `json:"default,omitempty"`
  }
  ```
- **⚠️ Pitfall**: Dùng JSON Schema format chuẩn. Đây là format mà OpenAI, Gemini, Claude đều sử dụng cho function calling. KHÔNG tự chế format riêng.

### Sub-task 1.4.2: Định nghĩa `Tool` interface
- **File**: `contracts/tool/tool.go`
- **Interface cần định nghĩa**:
  ```go
  type Tool interface {
      // Name trả về tên unique (vd: "git_commit", "read_file")
      Name() string

      // Description mô tả tool làm gì (sẽ gửi cho AI)
      Description() string

      // Schema trả về JSON Schema cho input parameters
      Schema() *Schema

      // Execute chạy tool với input arguments
      Execute(ctx context.Context, args json.RawMessage) (*Result, error)
  }

  type Result struct {
      Output   string `json:"output"`
      Error    string `json:"error,omitempty"`
      ExitCode int    `json:"exit_code"`
  }
  ```
- **⚠️ Pitfall #1**: `Execute()` nhận `json.RawMessage` thay vì `map[string]any`. Lý do: mỗi tool tự parse args theo struct riêng, type-safe hơn.
- **⚠️ Pitfall #2**: `ExitCode` quan trọng cho shell-based tools (git, docker). AI cần biết command thành công (0) hay thất bại (non-zero).
- **⚠️ Pitfall #3**: `Description()` sẽ được gửi cho AI trong system prompt. Viết rõ ràng, súc tích. AI dựa vào description này để quyết định dùng tool nào.

### Sub-task 1.4.3: Viết unit tests
- **File**: `contracts/tool/tool_test.go`
- **Tests**: Schema serialization, Result serialization

### Tiêu chí hoàn thành Task 1.4:
- [ ] Schema tương thích JSON Schema chuẩn
- [ ] Execute nhận json.RawMessage
- [ ] ExitCode cho shell tools
- [ ] Unit tests pass

---

## Task 1.5: Contracts — Event, Memory, Search, Workflow, Plugin, Context

### Sub-task 1.5.1: Event contracts
- **File**: `contracts/event/event.go`
- **Định nghĩa**:
  ```go
  type Event struct {
      ID        string    `json:"id"`
      Type      string    `json:"type"` // "task.started", "task.completed", "agent.error"
      Source    string    `json:"source"`
      Payload   any       `json:"payload"`
      Timestamp time.Time `json:"timestamp"`
  }

  type Bus interface {
      Publish(ctx context.Context, event Event) error
      Subscribe(eventType string, handler func(Event)) (unsubscribe func(), err error)
  }
  ```
- **⚠️ Pitfall #1**: `Subscribe` trả về `unsubscribe func()`. Đây là pattern để tránh memory leak — consumer PHẢI gọi unsubscribe khi không cần nữa.
- **⚠️ Pitfall #2**: Event `Type` dùng dot notation (`"task.started"`) để hỗ trợ wildcard subscribe (`"task.*"`).
- **⚠️ Pitfall #3**: `Payload` là `any` vì mỗi event type có payload khác nhau. Consumer cần type assertion.

### Sub-task 1.5.2: Memory contracts
- **File**: `contracts/memory/memory.go`
- **Định nghĩa**:
  ```go
  type Store interface {
      Save(ctx context.Context, key string, value any, opts ...SaveOption) error
      Load(ctx context.Context, key string, dest any) error
      Delete(ctx context.Context, key string) error
      Search(ctx context.Context, query string, limit int) ([]Entry, error)
      List(ctx context.Context, prefix string) ([]string, error)
  }

  type Entry struct {
      Key       string    `json:"key"`
      Value     any       `json:"value"`
      Score     float64   `json:"score,omitempty"` // Relevance score cho search
      CreatedAt time.Time `json:"created_at"`
  }

  type SaveOption func(*saveOptions)
  ```
- **⚠️ Pitfall**: `SaveOption` dùng functional options pattern. Cho phép thêm options (TTL, tags, metadata) mà không break interface.

### Sub-task 1.5.3: Search contracts
- **File**: `contracts/search/search.go`
- **Định nghĩa**:
  ```go
  type Engine interface {
      Index(ctx context.Context, items []Indexable) error
      Search(ctx context.Context, query string, opts ...SearchOption) ([]SearchResult, error)
  }

  type Indexable interface {
      ID() string
      Content() string
      Metadata() map[string]string
  }

  type SearchResult struct {
      ID       string            `json:"id"`
      Content  string            `json:"content"`
      Score    float64           `json:"score"`
      Metadata map[string]string `json:"metadata"`
  }

  type SearchOption func(*searchOptions)
  ```

### Sub-task 1.5.4: Workflow contracts
- **File**: `contracts/workflow/workflow.go`
- **Định nghĩa**:
  ```go
  type Workflow interface {
      Name() string
      Steps() []Step
      Execute(ctx context.Context, input map[string]any) (*WorkflowResult, error)
  }

  type Step struct {
      Name         string   `yaml:"name"`
      Agent        string   `yaml:"agent"`
      Task         string   `yaml:"task"`
      DependsOn    []string `yaml:"depends_on,omitempty"`
      Condition    string   `yaml:"condition,omitempty"` // Expression: "previous.status == 'success'"
      OnFailure    string   `yaml:"on_failure,omitempty"` // "retry", "skip", "abort"
      MaxRetries   int      `yaml:"max_retries,omitempty"`
  }

  type WorkflowResult struct {
      Status     Status                   `json:"status"`
      Steps      map[string]*StepResult   `json:"steps"`
      Duration   time.Duration            `json:"duration"`
  }

  type StepResult struct {
      Status   Status        `json:"status"`
      Output   any           `json:"output"`
      Error    string        `json:"error,omitempty"`
      Duration time.Duration `json:"duration"`
  }
  ```
- **⚠️ Pitfall**: `Condition` dùng expression string thay vì code. Cho phép config workflow trong YAML mà không cần viết Go code.

### Sub-task 1.5.5: Plugin contracts
- **File**: `contracts/plugin/plugin.go`
- **Định nghĩa**:
  ```go
  type Type string
  const (
      TypeAgent    Type = "agent"
      TypeProvider Type = "provider"
      TypeTool     Type = "tool"
      TypeSearch   Type = "search"
      TypeMemory   Type = "memory"
      TypeWorkflow Type = "workflow"
      TypeContext  Type = "context"
  )

  type Plugin interface {
      Name() string
      Type() Type
      Version() string
      Init(ctx context.Context, config map[string]any) error
      Start(ctx context.Context) error
      Stop(ctx context.Context) error
      Health(ctx context.Context) error
  }
  ```
- **⚠️ Pitfall #1**: `Init()` và `Start()` tách riêng. `Init` = load config, validate. `Start` = bắt đầu chạy (mở connections, start goroutines). Tách ra để có thể init tất cả trước, rồi start theo thứ tự dependency.
- **⚠️ Pitfall #2**: `Health()` trả về error nếu plugin không healthy. Dùng cho health check endpoint và circuit breaker.

### Sub-task 1.5.6: Planner, Orchestrator, Resilience, Security, Gateway, Feedback contracts
- **Files**:
  - `contracts/planner/planner.go`
  - `contracts/orchestrator/orchestrator.go`
  - `contracts/resilience/resilience.go`
  - `contracts/security/security.go`
  - `contracts/gateway/gateway.go`
  - `contracts/feedback/feedback.go`
  - `contracts/context/context.go`
- **Lưu ý**: Mỗi file định nghĩa interface chính + supporting types. Chi tiết sẽ được refine trong Phase 2 và 5 khi implement.

### Sub-task 1.5.7: Viết unit tests cho tất cả contracts
- **Mục tiêu**: Mỗi contract package có ít nhất 1 test file
- **Focus**: Serialization, type safety, interface compliance

### Tiêu chí hoàn thành Task 1.5:
- [ ] Mỗi contract package có ít nhất 1 interface
- [ ] Tất cả dùng functional options cho extensibility
- [ ] Event bus hỗ trợ unsubscribe
- [ ] Plugin lifecycle: Init → Start → Stop
- [ ] Unit tests pass

---

## Task 1.6: Contracts — Shared Types & Errors

### Sub-task 1.6.1: Shared ID types
- **File**: `contracts/types.go`
- **Định nghĩa**:
  ```go
  package contracts

  type (
      MissionID  string
      TaskID     string
      AgentID    string
      ProviderID string
      SessionID  string
      PluginID   string
  )

  // NewID tạo ID mới dạng UUID v4
  func NewID() string { ... }
  ```
- **⚠️ Pitfall**: Dùng named types (`MissionID`) thay vì plain `string`. Compiler sẽ bắt lỗi khi pass nhầm `TaskID` vào chỗ cần `MissionID`. Type safety miễn phí.

### Sub-task 1.6.2: Shared errors
- **File**: `contracts/errors.go`
- **Định nghĩa**:
  ```go
  package contracts

  import "errors"

  var (
      ErrProviderUnavailable = errors.New("provider unavailable")
      ErrProviderTimeout     = errors.New("provider timeout")
      ErrAgentBusy           = errors.New("agent is busy")
      ErrTaskCancelled       = errors.New("task cancelled")
      ErrTaskTimeout         = errors.New("task timeout")
      ErrTaskFailed          = errors.New("task failed")
      ErrInvalidConfig       = errors.New("invalid configuration")
      ErrPluginNotFound      = errors.New("plugin not found")
      ErrPermissionDenied    = errors.New("permission denied")
      ErrRateLimited         = errors.New("rate limited")
  )
  ```
- **⚠️ Pitfall #1**: Dùng `errors.New()` cho sentinel errors. Consumer dùng `errors.Is(err, contracts.ErrProviderTimeout)` để kiểm tra.
- **⚠️ Pitfall #2**: KHÔNG dùng string matching (`err.Error() == "timeout"`) → fragile, dễ break.
- **⚠️ Pitfall #3**: Khi wrap error, LUÔN dùng `fmt.Errorf("...: %w", err)` với `%w` verb. Đây là Go 1.13+ convention để unwrap errors.

### Sub-task 1.6.3: Shared status types
- **File**: `contracts/status.go`
- **Đã định nghĩa trong Task 1.3 (agent/result.go)**. Import từ đó hoặc di chuyển Status sang `contracts/status.go` nếu nhiều package cùng dùng.

### Tiêu chí hoàn thành Task 1.6:
- [ ] Named ID types (không dùng plain string)
- [ ] Sentinel errors cho mọi error case đã biết
- [ ] `go build ./...` thành công
- [ ] `go vet ./...` không warning
- [ ] Toàn bộ Phase 1 có ≥ 80% test coverage cho contracts

---

## 📋 Checklist tổng Phase 1

- [ ] Task 1.1: Go module, .gitignore, Makefile, linter
- [ ] Task 1.2: Provider contracts (6 sub-tasks)
- [ ] Task 1.3: Agent contracts (6 sub-tasks)
- [ ] Task 1.4: Tool contracts (3 sub-tasks)
- [ ] Task 1.5: Event, Memory, Search, Workflow, Plugin, Context contracts (7 sub-tasks)
- [ ] Task 1.6: Shared types & errors (3 sub-tasks)
- [ ] `go build ./...` thành công
- [ ] `go test ./...` ≥ 80% coverage
- [ ] `golangci-lint run` không error
- [ ] Git commit: "Phase 1: Complete contracts foundation"
