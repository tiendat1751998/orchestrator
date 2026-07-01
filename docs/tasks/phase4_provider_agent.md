# Phase 4: Provider & Agent Plugins — Chi Tiết Từng Sub-Task

> [!CAUTION]
> **Phase này là nơi hệ thống bắt đầu "sống".** Provider plugin kết nối với AI thực (Antigravity CLI), Agent plugins sử dụng provider để thực hiện tasks, Tool plugins cho phép agents thao tác filesystem/git. Mọi lỗi ở đây sẽ ảnh hưởng trực tiếp đến trải nghiệm người dùng.

---

## Task 4.1: Plugin — Antigravity Provider

> [!IMPORTANT]
> **Đây là provider DUY NHẤT bạn cần implement lúc này.** Antigravity CLI là CLI-based, không phải API-based → cần adapter đặc biệt để giao tiếp qua stdin/stdout.

### Sub-task 4.1.1: Plugin manifest & registration
- **File**: `plugins/providers/antigravity/plugin.yaml`
- **Nội dung**:
  ```yaml
  name: "antigravity"
  type: "provider"
  version: "0.1.0"
  description: "Antigravity CLI provider (Gemini backend)"
  config:
    binary: "antigravity"  # hoặc đường dẫn đầy đủ
    model: "gemini-2.5-pro"
    timeout: "120s"
  ```
- **File**: `plugins/providers/antigravity/plugin.go`
- **Chi tiết**: Implement `contracts/plugin.Plugin` interface (Init, Start, Stop, Health)

### Sub-task 4.1.2: CLI Process Manager
- **File**: `plugins/providers/antigravity/adapter/cli.go`
- **Chi tiết**:
  ```go
  type CLIAdapter struct {
      binary  string
      cmd     *exec.Cmd
      stdin   io.WriteCloser
      stdout  io.ReadCloser
      stderr  io.ReadCloser
      mu      sync.Mutex
  }

  func NewCLIAdapter(binary string) *CLIAdapter { ... }
  func (a *CLIAdapter) Start(ctx context.Context) error { ... }
  func (a *CLIAdapter) Stop() error { ... }
  func (a *CLIAdapter) Send(input string) (string, error) { ... }
  ```
- **⚠️ Pitfall #1**: Process lifecycle. CLI process có thể crash bất cứ lúc nào. Cần detect process death và restart.
- **⚠️ Pitfall #2**: `exec.Cmd` chỉ Start() được 1 lần. Nếu process chết → phải tạo `exec.Cmd` mới.
- **⚠️ Pitfall #3**: stdin/stdout pipes. Trên Windows, line ending là `\r\n`, trên Linux/Mac là `\n`. Cần normalize.
- **⚠️ Pitfall #4**: Process kill trên Windows. `cmd.Process.Kill()` chỉ kill process chính, không kill child processes. Cần `taskkill /F /T /PID` hoặc `cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}`.

### Sub-task 4.1.3: Stdin Writer
- **File**: `plugins/providers/antigravity/adapter/stdin.go`
- **Chi tiết**:
  ```go
  func (a *CLIAdapter) WritePrompt(prompt string) error {
      // 1. Encode prompt thành format mà Antigravity CLI hiểu
      // 2. Write to stdin pipe
      // 3. Flush
  }
  ```
- **⚠️ Pitfall**: Concurrent writes to stdin. PHẢI dùng mutex. 2 goroutines write cùng lúc = garbled output.

### Sub-task 4.1.4: Stdout Reader
- **File**: `plugins/providers/antigravity/adapter/stdout.go`
- **Chi tiết**:
  ```go
  func (a *CLIAdapter) ReadResponse() (string, error) {
      // 1. Read from stdout pipe
      // 2. Detect end of response (delimiter hoặc EOF marker)
      // 3. Return complete response
  }
  ```
- **⚠️ Pitfall #1**: Response boundary detection. CLI có thể output nhiều dòng. Cần biết khi nào response kết thúc. Options:
  - Dùng sentinel marker (vd: `---END---`)
  - Dùng timeout (không nhận thêm output sau N giây)
  - Dùng JSON framing (mỗi response là 1 JSON object)
- **⚠️ Pitfall #2**: Large responses. Nếu AI trả về response rất dài → stdout buffer đầy → process block → deadlock. Cần đọc stdout LIÊN TỤC trong background goroutine.

### Sub-task 4.1.5: Stderr Handler
- **File**: `plugins/providers/antigravity/adapter/stderr.go`
- **Chi tiết**: Đọc stderr trong background goroutine, log errors
- **⚠️ Pitfall**: stderr và stdout PHẢI đọc đồng thời. Nếu chỉ đọc stdout mà stderr buffer đầy → process block → deadlock. Đây là bug CỰC KỲ phổ biến với `exec.Cmd`.

### Sub-task 4.1.6: Response Parser — Markdown
- **File**: `plugins/providers/antigravity/parser/markdown.go`
- **Chi tiết**: Parse markdown response từ Antigravity CLI thành structured data
- **⚠️ Pitfall**: AI response format không cố định. Cần robust parsing, graceful degradation khi format lạ.

### Sub-task 4.1.7: Response Parser — Tool Calls
- **File**: `plugins/providers/antigravity/parser/toolcall.go`
- **Chi tiết**: Detect và parse tool call blocks trong response
  ```
  Pattern ví dụ:
  ```json
  {"tool": "read_file", "args": {"path": "/foo/bar.go"}}
  ```
  ```
- **⚠️ Pitfall**: Tool call format có thể thay đổi giữa các versions của Antigravity CLI. Parser phải flexible.

### Sub-task 4.1.8: Response Parser — JSON
- **File**: `plugins/providers/antigravity/parser/json.go`
- **Chi tiết**: Parse JSON structured output (nếu CLI trả về JSON mode)

### Sub-task 4.1.9: Response Parser — Error
- **File**: `plugins/providers/antigravity/parser/error.go`
- **Chi tiết**: Detect error patterns (rate limit, invalid API key, model not found)

### Sub-task 4.1.10: Session Manager
- **File**: `plugins/providers/antigravity/session/manager.go`
- **Chi tiết**:
  ```go
  type SessionManager struct {
      sessions map[string]*Session
      mu       sync.RWMutex
  }

  type Session struct {
      ID       string
      Adapter  *CLIAdapter
      Messages []provider.Message
      Created  time.Time
      LastUsed time.Time
  }
  ```
- **⚠️ Pitfall #1**: Session cleanup. Sessions không dùng lâu phải được tự động đóng (close CLI process). Dùng background goroutine kiểm tra `LastUsed`.
- **⚠️ Pitfall #2**: Session pool size. Giới hạn số sessions đồng thời. Mỗi session = 1 CLI process = RAM + CPU.

### Sub-task 4.1.11: Heartbeat
- **File**: `plugins/providers/antigravity/session/heartbeat.go`
- **Chi tiết**: Ping CLI process định kỳ để đảm bảo nó còn sống
- **⚠️ Pitfall**: Heartbeat interval không nên quá ngắn (tốn resource) hoặc quá dài (phát hiện chậm).

### Sub-task 4.1.12: Prompt Builder
- **File**: `plugins/providers/antigravity/prompt/builder.go`
- **Chi tiết**: Convert `provider.Request` → format cụ thể cho Antigravity CLI

### Sub-task 4.1.13: Main Provider implementation
- **File**: `plugins/providers/antigravity/provider.go`
- **Chi tiết**: Implement `contracts/provider.Provider` interface bằng cách sử dụng tất cả components trên
- **⚠️ Pitfall**: Provider.Send() phải thread-safe. Nhiều agents có thể gọi cùng lúc. Dùng session pool.

### Sub-task 4.1.14: Chiến lược thay thế — Gemini API trực tiếp
- **Lưu ý**: Nếu việc giao tiếp qua CLI quá phức tạp/không ổn định, có thể chuyển sang dùng Gemini API trực tiếp:
  ```go
  // Alternative: Dùng Gemini API SDK
  import "google.golang.org/genai"
  ```
  Ưu điểm: API ổn định, có SDK chính thức, streaming native
  Nhược điểm: Cần API key riêng, không tận dụng được Antigravity features

### Sub-task 4.1.15: Unit & Integration tests
- **Tests**:
  - Mock CLI adapter (không cần CLI thật)
  - Parser tests (markdown, tool calls, JSON, errors)
  - Session manager lifecycle
  - Provider.Send() với mock adapter
  - Provider.Stream() với mock adapter
  - **Integration test**: Provider.Send() với Antigravity CLI thật (gated, chỉ chạy khi có CLI)

### Tiêu chí hoàn thành Task 4.1:
- [ ] Provider giao tiếp được với Antigravity CLI (hoặc Gemini API)
- [ ] Send() trả về Response đầy đủ
- [ ] Stream() trả về channel chunks
- [ ] Session management (pool, cleanup)
- [ ] Error handling (rate limit, timeout, crash)
- [ ] Parser xử lý được markdown, tool calls, JSON
- [ ] Unit tests ≥ 80% coverage

---

## Task 4.2: Plugin — Core Tools

### Sub-task 4.2.1: Filesystem Tool — read_file
- **File**: `plugins/tools/filesystem/read_file.go`
- **Schema**:
  ```json
  {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "Absolute path to file"},
      "start_line": {"type": "integer", "description": "Start line (1-indexed)"},
      "end_line": {"type": "integer", "description": "End line (1-indexed, inclusive)"}
    },
    "required": ["path"]
  }
  ```
- **⚠️ Pitfall #1**: Path traversal attack. Agent có thể yêu cầu đọc `/etc/passwd` hoặc `C:\Windows\System32\config`. PHẢI validate path nằm trong workspace.
- **⚠️ Pitfall #2**: Large files. Nếu file > 1MB → KHÔNG đọc toàn bộ, chỉ đọc range (start_line → end_line). Tránh OOM.
- **⚠️ Pitfall #3**: Binary files. Kiểm tra file có phải text không. Đọc binary file vào string = garbage.

### Sub-task 4.2.2: Filesystem Tool — write_file
- **File**: `plugins/tools/filesystem/write_file.go`
- **⚠️ Pitfall #1**: Atomic write. Write vào temp file trước → rename. Tránh corrupt file nếu crash giữa chừng.
- **⚠️ Pitfall #2**: Path validation (giống read_file).
- **⚠️ Pitfall #3**: Tạo parent directories nếu chưa tồn tại (`os.MkdirAll`).

### Sub-task 4.2.3: Filesystem Tool — list_dir
- **File**: `plugins/tools/filesystem/list_dir.go`
- **⚠️ Pitfall**: Giới hạn depth và số lượng entries. Listing `/` recursive = rất chậm.

### Sub-task 4.2.4: Filesystem Tool — search (grep)
- **File**: `plugins/tools/filesystem/search.go`
- **⚠️ Pitfall**: Giới hạn kết quả (max 50 matches). Regex phải có timeout.

### Sub-task 4.2.5: Git Tool — git operations
- **File**: `plugins/tools/git/git.go`
- **Operations**:
  - `git_status` — Status hiện tại
  - `git_diff` — Xem changes
  - `git_add` — Stage files
  - `git_commit` — Commit
  - `git_log` — View history
  - `git_clone` — Clone repo
- **⚠️ Pitfall #1**: Git credentials. Clone private repo cần auth. KHÔNG hardcode credentials.
- **⚠️ Pitfall #2**: Large diffs. `git diff` trên file lớn → output rất dài. Cần truncate.
- **⚠️ Pitfall #3**: Working directory. Git commands phải chạy trong đúng repo directory. Validate trước khi execute.

### Sub-task 4.2.6: Terminal Tool — run_command
- **File**: `plugins/tools/terminal/terminal.go`
- **⚠️ Pitfall #1**: Command injection. Agent có thể gửi `rm -rf /`. Security policy PHẢI block.
- **⚠️ Pitfall #2**: Long-running commands. Timeout PHẢI có. Default 30s.
- **⚠️ Pitfall #3**: Trên Windows, dùng `cmd /c` hoặc `powershell -Command`. Trên Linux, dùng `sh -c`.
- **⚠️ Pitfall #4**: Output size limit. Command có thể output hàng GB. Truncate ở 100KB.

### Sub-task 4.2.7: Unit tests cho tất cả tools
- **Tests**:
  - read_file: file tồn tại, file không tồn tại, path traversal
  - write_file: tạo file mới, overwrite, atomic write
  - list_dir: empty dir, nested dir, depth limit
  - git: mock git binary
  - terminal: echo command, timeout, blocked command

### Tiêu chí hoàn thành Task 4.2:
- [ ] Filesystem tools hoạt động (read, write, list, search)
- [ ] Git tools hoạt động (status, diff, add, commit, log)
- [ ] Terminal tool có timeout và security
- [ ] Path validation chống traversal
- [ ] Output size limits
- [ ] Unit tests ≥ 85% coverage

---

## Task 4.3: Plugin — Core Agents

### Sub-task 4.3.1: Backend Agent
- **Files**:
  - `plugins/agents/backend/agent.go` — Implement Agent interface
  - `plugins/agents/backend/agent.yaml` — Manifest
  - `plugins/agents/backend/prompts/system.md` — System prompt
- **Manifest ví dụ**:
  ```yaml
  name: "backend"
  version: "0.1.0"
  role: "Backend Developer"
  description: "Generates backend code, APIs, database schemas"
  capabilities:
    - code_generation
    - testing
    - debugging
    - refactoring
  provider: "antigravity"
  model: "gemini-2.5-pro"
  tools:
    - read_file
    - write_file
    - list_dir
    - search
    - git_status
    - git_diff
    - git_add
    - git_commit
    - run_command
  prompt_file: "prompts/system.md"
  temperature: 0.3
  max_tokens: 8192
  ```
- **System prompt guidelines**:
  - Rõ ràng về role và responsibilities
  - Liệt kê tools available và khi nào dùng
  - Output format expectations
  - Constraints (language, framework, coding style)
- **⚠️ Pitfall #1**: System prompt quá dài → tốn tokens, giảm context window cho task. Giới hạn ≤ 2000 tokens.
- **⚠️ Pitfall #2**: System prompt quá mơ hồ → AI output không nhất quán.

### Sub-task 4.3.2: DevOps Agent
- **Files**:
  - `plugins/agents/devops/agent.go`
  - `plugins/agents/devops/agent.yaml`
  - `plugins/agents/devops/prompts/system.md`
- **Capabilities**: `deployment`, `documentation`
- **Tools**: git, terminal, read_file, write_file

### Sub-task 4.3.3: Reviewer Agent
- **Files**:
  - `plugins/agents/reviewer/agent.go`
  - `plugins/agents/reviewer/agent.yaml`
  - `plugins/agents/reviewer/prompts/system.md`
- **Capabilities**: `code_review`, `testing`
- **Đặc biệt**: Agent này KHÔNG tạo code mới, chỉ review code từ agents khác
- **⚠️ Pitfall**: Reviewer agent có thể approve code lỗi. Cần scoring mechanism (Phase 6).

### Sub-task 4.3.4: Integration test — End-to-end agent execution
- **Tests**:
  - Tạo Backend agent → giao task "write hello world in Go" → nhận Result
  - Agent sử dụng tools (read_file, write_file) trong quá trình execute
  - Agent handle error gracefully (provider timeout, tool failure)

### Tiêu chí hoàn thành Task 4.3:
- [ ] 3 agents (backend, devops, reviewer) hoạt động
- [ ] Mỗi agent có manifest YAML + system prompt
- [ ] Agents sử dụng tools qua tool call loop
- [ ] Integration test: agent → provider → tools → result
- [ ] Unit tests ≥ 80% coverage

---

## 📋 Checklist tổng Phase 4

- [ ] Task 4.1: Antigravity Provider (15 sub-tasks)
- [ ] Task 4.2: Core Tools (7 sub-tasks)
- [ ] Task 4.3: Core Agents (4 sub-tasks)
- [ ] Milestone M3: First Agent Call thành công
- [ ] Integration test: Agent → Provider → Tools → Result
- [ ] `go test ./plugins/...` pass
- [ ] Git commit: "Phase 4: Provider, tools, and agents"
