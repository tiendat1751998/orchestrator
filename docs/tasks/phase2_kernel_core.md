# Phase 2: Kernel Core — Chi Tiết Từng Sub-Task

> [!IMPORTANT]
> **Nguyên tắc**: Kernel code KHÔNG import từ `plugins/`. Kernel chỉ import từ `contracts/`. Plugins import từ `sdk/` và `contracts/`. Vi phạm điều này = circular dependency = hệ thống không build được.

---

## Task 2.1: Kernel — Config & Logger

### Sub-task 2.1.1: Config Loader — Đọc YAML config
- **File**: `kernel/config/loader.go`
- **Chi tiết**:
  - Đọc config từ `.orchestrator/settings.yaml`
  - Hỗ trợ override bằng environment variables
  - Hỗ trợ syntax `${ENV_VAR}` trong YAML values
  - Hỗ trợ default values
- **Config structure**:
  ```yaml
  # .orchestrator/settings.yaml
  orchestrator:
    name: "my-orchestrator"
    log_level: "info"        # debug, info, warn, error
    log_format: "json"       # json, text
    data_dir: ".orchestrator/data"

  providers:
    default: "antigravity"
    antigravity:
      type: "cli"
      binary: "antigravity"
      model: "gemini-2.5-pro"
      timeout: "120s"

  agents:
    backend:
      provider: "antigravity"
      model: "gemini-2.5-pro"
      prompt_file: "prompts/backend/system.md"

  security:
    sandbox: true
    allowed_tools: ["git", "filesystem", "terminal"]
    blocked_commands: ["rm -rf /", "sudo", "chmod 777"]
  ```
- **⚠️ Pitfall #1**: Config file path resolution. Phải tìm config file theo thứ tự:
  1. `--config` flag (CLI argument)
  2. `.orchestrator/settings.yaml` (project root)
  3. `~/.orchestrator/settings.yaml` (user home)
  4. Default values (hardcoded)
- **⚠️ Pitfall #2**: Environment variable substitution phải xảy ra SAU khi parse YAML, TRƯỚC khi validate. Ví dụ: `api_key: "${GEMINI_API_KEY}"` → phải resolve thành giá trị thực.
- **⚠️ Pitfall #3**: `time.Duration` trong YAML. Go không tự parse `"120s"` từ YAML → cần custom unmarshal hoặc dùng library hỗ trợ.

### Sub-task 2.1.2: Config Validator
- **File**: `kernel/config/validator.go`
- **Chi tiết**:
  - Validate required fields (provider name, model)
  - Validate enum values (log_level phải là debug/info/warn/error)
  - Validate file paths tồn tại (binary, prompt_file)
  - Trả về tất cả lỗi cùng lúc (KHÔNG dừng ở lỗi đầu tiên)
- **⚠️ Pitfall**: Dùng multi-error pattern:
  ```go
  type ValidationErrors []error
  func (ve ValidationErrors) Error() string { ... }
  ```
  Trả về TẤT CẢ lỗi → user fix 1 lần, không phải chạy đi chạy lại.

### Sub-task 2.1.3: Config struct chính
- **File**: `kernel/config/config.go`
- **Chi tiết**: Go struct mapping 1:1 với YAML structure ở trên
- **⚠️ Pitfall**: KHÔNG dùng global variable `var GlobalConfig`. Truyền config qua constructor/dependency injection. Global state = nightmare khi test.

### Sub-task 2.1.4: Logger — Structured logging
- **File**: `kernel/logger/logger.go`
- **Chi tiết**:
  - Dựa trên `log/slog` (Go 1.21+ standard library)
  - Hỗ trợ JSON và Text output
  - Hỗ trợ log levels: Debug, Info, Warn, Error
  - Context-aware: tự động thêm request_id, task_id, agent_name
- **⚠️ Pitfall #1**: Dùng `slog` (standard library), KHÔNG dùng `logrus` hay `zap` lúc đầu. Lý do: ít dependencies, performance tốt, Go team maintain.
- **⚠️ Pitfall #2**: KHÔNG log sensitive data (API keys, user content). Tạo helper `redact()` function.
- **⚠️ Pitfall #3**: Log context fields PHẢI consistent. Ví dụ: LUÔN dùng `task_id`, không khi thì `taskId`, khi thì `task_id`, khi thì `TaskID`.

### Sub-task 2.1.5: Logger Formatter
- **File**: `kernel/logger/formatter.go`
- **Chi tiết**: Custom handler cho slog, format output đẹp cho terminal (colors, indentation)

### Sub-task 2.1.6: Unit tests
- **Files**: `kernel/config/config_test.go`, `kernel/logger/logger_test.go`
- **Tests cần viết**:
  - Load config từ YAML file
  - Environment variable override
  - Missing required field → error
  - Invalid enum value → error
  - Logger output format (JSON, Text)

### Tiêu chí hoàn thành Task 2.1:
- [ ] Load config từ YAML
- [ ] Env var override hoạt động
- [ ] Validation trả về tất cả lỗi
- [ ] Logger dùng slog
- [ ] Không có global state
- [ ] Unit tests ≥ 85% coverage

---

## Task 2.2: Kernel — EventBus

### Sub-task 2.2.1: In-memory Event Bus
- **File**: `kernel/eventbus/bus.go`
- **Chi tiết**:
  - Thread-safe (dùng `sync.RWMutex`)
  - Async publishing (dùng goroutine)
  - Wildcard subscription (`"task.*"` matches `"task.started"`, `"task.completed"`)
  - Buffered channel để tránh blocking publisher
- **⚠️ Pitfall #1**: `sync.RWMutex` cho subscriber map. Publisher dùng `RLock`, register/unregister dùng `Lock`. Nếu dùng `Mutex` → publisher bị block khi có subscriber đang register.
- **⚠️ Pitfall #2**: Goroutine leak. Mỗi subscriber handler chạy trong goroutine riêng. Nếu handler panic → goroutine chết nhưng không ai biết. Cần `recover()` trong handler wrapper.
- **⚠️ Pitfall #3**: Event ordering. In-memory bus nên đảm bảo ordering WITHIN A TOPIC. Cross-topic ordering không bắt buộc.
- **⚠️ Pitfall #4**: Slow subscriber. Nếu 1 subscriber xử lý chậm → KHÔNG được block publisher. Dùng buffered channel hoặc drop events khi buffer đầy.

### Sub-task 2.2.2: Subscriber management
- **File**: `kernel/eventbus/subscriber.go`
- **Chi tiết**:
  - `Subscribe()` trả về `unsubscribe func()`
  - Subscriber ID auto-generated
  - Wildcard matching dùng `path.Match()` hoặc simple string prefix
- **⚠️ Pitfall**: Unsubscribe PHẢI idempotent. Gọi 2 lần không được panic.

### Sub-task 2.2.3: Publisher
- **File**: `kernel/eventbus/publisher.go`
- **Chi tiết**: Helper functions để publish common events
  ```go
  func PublishTaskStarted(bus Bus, taskID, agentName string) { ... }
  func PublishTaskCompleted(bus Bus, taskID string, result *Result) { ... }
  func PublishTaskFailed(bus Bus, taskID string, err error) { ... }
  ```

### Sub-task 2.2.4: Unit tests
- **File**: `kernel/eventbus/bus_test.go`
- **Tests cần viết**:
  - Publish → Subscribe nhận được event
  - Wildcard subscription
  - Unsubscribe → không nhận event nữa
  - Concurrent publish/subscribe (race detector)
  - Slow subscriber không block publisher
  - Handler panic không crash hệ thống

### Tiêu chí hoàn thành Task 2.2:
- [ ] Thread-safe
- [ ] Wildcard subscription
- [ ] Unsubscribe không leak
- [ ] Handler panic recovery
- [ ] `go test -race` pass
- [ ] Unit tests ≥ 90% coverage

---

## Task 2.3: Kernel — Registry

### Sub-task 2.3.1: Plugin Registry
- **File**: `kernel/registry/registry.go`
- **Chi tiết**:
  ```go
  type Registry struct {
      mu       sync.RWMutex
      plugins  map[string]contracts.Plugin
      agents   map[string]agent.Agent
      providers map[string]provider.Provider
      tools    map[string]tool.Tool
  }

  func (r *Registry) Register(p contracts.Plugin) error { ... }
  func (r *Registry) Unregister(name string) error { ... }
  func (r *Registry) GetAgent(name string) (agent.Agent, error) { ... }
  func (r *Registry) GetProvider(name string) (provider.Provider, error) { ... }
  func (r *Registry) GetTool(name string) (tool.Tool, error) { ... }
  func (r *Registry) FindAgentForTask(task *agent.Task) (agent.Agent, error) { ... }
  ```
- **⚠️ Pitfall #1**: `FindAgentForTask()` iterate qua tất cả agents, gọi `CanHandle(task)`, trả về agent đầu tiên match. Nếu nhiều agents match → cần scoring/priority.
- **⚠️ Pitfall #2**: `Register()` phải check duplicate names. 2 plugins cùng tên = undefined behavior.
- **⚠️ Pitfall #3**: Thread-safety. Registry được access từ nhiều goroutines (scheduler, orchestrator, health checker).

### Sub-task 2.3.2: Plugin Lifecycle Manager
- **File**: `kernel/registry/plugin.go`
- **Chi tiết**:
  - `InitAll()` → Init tất cả plugins theo dependency order
  - `StartAll()` → Start tất cả plugins
  - `StopAll()` → Stop theo reverse order (LIFO)
- **⚠️ Pitfall**: Stop order PHẢI ngược với Start order. Ví dụ: start Provider → start Agent. Stop Agent → stop Provider. Nếu stop Provider trước → Agent gọi Provider sẽ panic.

### Sub-task 2.3.3: Capability Matching
- **File**: `kernel/registry/capability.go`
- **Chi tiết**: Tìm agent phù hợp nhất cho task dựa trên capabilities
- **⚠️ Pitfall**: Khi không tìm được agent → trả error rõ ràng, KHÔNG return nil silently.

### Sub-task 2.3.4: Unit tests
- **Tests**: Register/Unregister, FindAgentForTask, duplicate check, concurrent access

### Tiêu chí hoàn thành Task 2.3:
- [ ] Register/Unregister thread-safe
- [ ] FindAgentForTask hoạt động
- [ ] Plugin lifecycle: Init → Start → Stop (reverse order)
- [ ] Duplicate name detection
- [ ] Unit tests ≥ 85% coverage

---

## Task 2.4: Kernel — Runtime & Executor

### Sub-task 2.4.1: Task Executor
- **File**: `kernel/runtime/executor.go`
- **Chi tiết**:
  ```go
  type Executor struct {
      registry *registry.Registry
      eventbus event.Bus
      logger   *slog.Logger
  }

  // ExecuteTask chạy 1 task trên agent phù hợp
  func (e *Executor) ExecuteTask(ctx context.Context, task *agent.Task) (*agent.Result, error) {
      // 1. Find agent from registry
      // 2. Emit "task.started" event
      // 3. Call agent.Execute(ctx, task)
      // 4. Emit "task.completed" or "task.failed" event
      // 5. Return result
  }
  ```
- **⚠️ Pitfall #1**: Context propagation. LUÔN truyền ctx xuống agent.Execute(). Nếu user cancel → ctx.Done() → agent phải dừng.
- **⚠️ Pitfall #2**: Panic recovery. Agent code có thể panic. Executor phải `recover()` và trả về error thay vì crash toàn bộ hệ thống.
- **⚠️ Pitfall #3**: Timeout. Tạo child context với timeout từ task.Timeout:
  ```go
  ctx, cancel := context.WithTimeout(ctx, task.Timeout)
  defer cancel()
  ```

### Sub-task 2.4.2: Goroutine Pool / Worker Manager
- **File**: `kernel/runtime/manager.go`
- **Chi tiết**:
  - Giới hạn số goroutines chạy đồng thời (semaphore pattern)
  - Dùng `golang.org/x/sync/errgroup` hoặc custom worker pool
  - Configurable max workers (default: 5)
- **⚠️ Pitfall #1**: Không giới hạn goroutines = OOM khi có 1000 tasks. PHẢI có semaphore.
- **⚠️ Pitfall #2**: `errgroup` tự cancel context khi 1 goroutine lỗi. Cân nhắc: có muốn cancel TẤT CẢ tasks khi 1 task fail không? Thường là KHÔNG → dùng custom pool thay vì errgroup.

### Sub-task 2.4.3: Dispatcher
- **File**: `kernel/runtime/dispatcher.go`
- **Chi tiết**: Nhận task từ scheduler → chọn worker → gửi cho executor
- **⚠️ Pitfall**: Dispatcher KHÔNG nên block. Nếu tất cả workers bận → đưa task vào waiting queue, không block caller.

### Sub-task 2.4.4: Runtime Engine
- **File**: `kernel/runtime/runtime.go`
- **Chi tiết**: Kết nối Executor + Manager + Dispatcher. Start/Stop runtime.
- **⚠️ Pitfall**: Graceful shutdown. Khi Stop():
  1. Ngừng nhận task mới
  2. Đợi các tasks đang chạy hoàn thành (với timeout)
  3. Cancel các tasks quá timeout
  4. Cleanup resources

### Sub-task 2.4.5: Unit tests
- **Tests**:
  - Execute 1 task thành công
  - Execute task với timeout → context deadline exceeded
  - Execute task, agent panic → error returned (không crash)
  - Concurrent execution (race detector)
  - Graceful shutdown (tasks finish before shutdown)
  - Worker pool limit (max 5 concurrent)

### Tiêu chí hoàn thành Task 2.4:
- [ ] Task execution với timeout
- [ ] Panic recovery
- [ ] Worker pool có giới hạn
- [ ] Graceful shutdown
- [ ] Event emission (started, completed, failed)
- [ ] `go test -race` pass

---

## Task 2.5: Kernel — Scheduler

### Sub-task 2.5.1: Priority Queue
- **File**: `kernel/scheduler/queue.go`
- **Chi tiết**:
  - Implement priority queue dùng `container/heap`
  - Priority cao hơn = chạy trước
  - Hỗ trợ dependency resolution: task B chờ task A hoàn thành
- **⚠️ Pitfall #1**: `container/heap` KHÔNG thread-safe. Phải wrap bằng mutex.
- **⚠️ Pitfall #2**: Circular dependency detection. Nếu A depends on B, B depends on A → deadlock. PHẢI detect và trả error khi enqueue.

### Sub-task 2.5.2: Scheduler Loop
- **File**: `kernel/scheduler/scheduler.go`
- **Chi tiết**:
  ```go
  func (s *Scheduler) Run(ctx context.Context) {
      for {
          select {
          case <-ctx.Done():
              return
          default:
              task := s.queue.Dequeue()
              if task == nil {
                  // No ready tasks, wait
                  time.Sleep(100 * time.Millisecond) // hoặc dùng condition variable
                  continue
              }
              s.dispatcher.Dispatch(task)
          }
      }
  }
  ```
- **⚠️ Pitfall #1**: Busy loop. Nếu queue trống → `time.Sleep` hoặc `sync.Cond` để tránh CPU 100%.
- **⚠️ Pitfall #2**: Task readiness. Task chỉ "ready" khi TẤT CẢ dependencies đã hoàn thành. Scheduler phải check dependency status trước khi dequeue.

### Sub-task 2.5.3: Priority Calculation
- **File**: `kernel/scheduler/priority.go`
- **Chi tiết**: Tính priority dựa trên: user-defined priority, dependency depth, wait time

### Sub-task 2.5.4: Unit tests
- **Tests**: Enqueue/Dequeue order, dependency resolution, circular dependency detection, concurrent access

### Tiêu chí hoàn thành Task 2.5:
- [ ] Priority queue thread-safe
- [ ] Dependency resolution
- [ ] Circular dependency detection
- [ ] No busy loop
- [ ] Unit tests ≥ 85% coverage

---

## Task 2.6: Kernel — Lifecycle & Bootstrap

### Sub-task 2.6.1: Kernel Bootstrap
- **File**: `kernel/kernel.go`
- **Chi tiết**:
  ```go
  type Kernel struct {
      config    *config.Config
      logger    *slog.Logger
      eventbus  *eventbus.Bus
      registry  *registry.Registry
      runtime   *runtime.Runtime
      scheduler *scheduler.Scheduler
  }

  func New(cfg *config.Config) (*Kernel, error) {
      // 1. Create logger
      // 2. Create eventbus
      // 3. Create registry
      // 4. Create runtime
      // 5. Create scheduler
      // 6. Wire everything together
      return &Kernel{...}, nil
  }

  func (k *Kernel) Start(ctx context.Context) error {
      // 1. Init all plugins
      // 2. Start all plugins
      // 3. Start scheduler
      // 4. Start runtime
      // 5. Emit "kernel.started" event
  }

  func (k *Kernel) Stop(ctx context.Context) error {
      // 1. Stop scheduler (no new tasks)
      // 2. Stop runtime (wait for running tasks)
      // 3. Stop all plugins (reverse order)
      // 4. Emit "kernel.stopped" event
  }
  ```
- **⚠️ Pitfall #1**: Startup order matters. EventBus trước, rồi Registry, rồi Runtime, rồi Scheduler. Đảo thứ tự = nil pointer panic.
- **⚠️ Pitfall #2**: Shutdown order = NGƯỢC LẠI startup order.
- **⚠️ Pitfall #3**: Shutdown timeout. Nếu tasks chạy quá lâu → force cancel sau timeout.
  ```go
  ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
  defer cancel()
  ```

### Sub-task 2.6.2: Signal Handling
- **File**: `kernel/lifecycle/lifecycle.go`
- **Chi tiết**:
  ```go
  func WaitForShutdown(ctx context.Context, kernel *Kernel) {
      sigChan := make(chan os.Signal, 1)
      signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

      select {
      case <-sigChan:
          kernel.Stop(ctx)
      case <-ctx.Done():
          kernel.Stop(ctx)
      }
  }
  ```
- **⚠️ Pitfall**: Buffer size 1 cho signal channel. Nếu unbuffered → signal có thể bị miss.

### Sub-task 2.6.3: State Machine
- **File**: `kernel/state.go`
- **Chi tiết**: Kernel states: `Created` → `Initializing` → `Running` → `ShuttingDown` → `Stopped`
- **⚠️ Pitfall**: State transitions phải atomic. Dùng `atomic.Value` hoặc mutex.

### Sub-task 2.6.4: Integration test
- **File**: `kernel/kernel_test.go`
- **Tests**:
  - Kernel Start → Stop cycle
  - Signal handling (SIGINT)
  - Start with invalid config → error
  - Double Start → error
  - Stop before Start → no-op

### Tiêu chí hoàn thành Task 2.6:
- [ ] Kernel Start/Stop hoạt động
- [ ] Signal handling (Ctrl+C)
- [ ] State machine transitions
- [ ] Shutdown timeout
- [ ] Integration test pass

---

## 📋 Checklist tổng Phase 2

- [ ] Task 2.1: Config loader + Logger (6 sub-tasks)
- [ ] Task 2.2: EventBus (4 sub-tasks)
- [ ] Task 2.3: Registry (4 sub-tasks)
- [ ] Task 2.4: Runtime & Executor (5 sub-tasks)
- [ ] Task 2.5: Scheduler (4 sub-tasks)
- [ ] Task 2.6: Lifecycle & Bootstrap (4 sub-tasks)
- [ ] Kernel Start → chạy → Stop cycle hoạt động
- [ ] `go test -race ./kernel/...` pass
- [ ] `go test -coverprofile=coverage.out ./kernel/...` ≥ 80%
- [ ] Git commit: "Phase 2: Kernel core implementation"
