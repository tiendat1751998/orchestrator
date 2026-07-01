# Phase 5: Orchestration Engine — Chi Tiết Từng Sub-Task

> [!CAUTION]
> **Đây là Phase quan trọng nhất.** Phase này biến "bộ sưu tập agents riêng lẻ" thành "hệ thống điều phối tự động". Sau phase này, bạn có thể giao mission và nhận kết quả cuối cùng.

---

## Task 5.1: Kernel — Planner

### Sub-task 5.1.1: Mission struct
- **File**: `kernel/planner/mission.go`
- **Định nghĩa**:
  ```go
  type Mission struct {
      ID          string            `json:"id"`
      Title       string            `json:"title"`
      Description string            `json:"description"`
      Context     []ContextItem     `json:"context,omitempty"`
      Constraints []string          `json:"constraints,omitempty"` // "use Go", "no external deps"
      CreatedAt   time.Time         `json:"created_at"`
      Metadata    map[string]string `json:"metadata,omitempty"`
  }
  ```

### Sub-task 5.1.2: Task Decomposer
- **File**: `kernel/planner/decomposer.go`
- **Chi tiết**:
  ```go
  type Decomposer struct {
      provider provider.Provider
      logger   *slog.Logger
  }

  // Decompose sử dụng AI để phân rã mission thành danh sách tasks
  func (d *Decomposer) Decompose(ctx context.Context, mission *Mission) ([]*agent.Task, error) {
      // 1. Build meta-prompt: "Given this mission, break it down into sub-tasks..."
      // 2. Send to provider
      // 3. Parse response → list of tasks with dependencies
      // 4. Validate tasks (no circular deps, all agents exist)
      // 5. Return task list
  }
  ```
- **Meta-prompt guidelines**:
  ```
  You are a project planner. Given the following mission, break it down into
  specific, actionable sub-tasks. For each task, specify:
  - name: short descriptive name
  - type: one of [code_generation, code_review, testing, deployment, documentation]
  - description: what exactly needs to be done
  - dependencies: list of task names this depends on (empty if none)
  - priority: 1 (highest) to 5 (lowest)

  Output as JSON array.
  ```
- **⚠️ Pitfall #1**: AI output parsing. AI có thể trả về JSON không hợp lệ, hoặc JSON wrapped trong markdown code blocks. Parser phải handle:
  - Raw JSON
  - JSON trong ` ```json ... ``` ` block
  - JSON với trailing commas
  - Partial JSON (truncated response)
- **⚠️ Pitfall #2**: Task count. AI có thể trả về quá ít (1 task cho mission phức tạp) hoặc quá nhiều (20 tasks cho mission đơn giản). Cần heuristic: min 2, max 15 tasks.
- **⚠️ Pitfall #3**: Dependency correctness. AI có thể tạo dependency sai (task A depends on task C, nhưng task C chưa được định nghĩa). Validate sau khi parse.

### Sub-task 5.1.3: DAG Builder
- **File**: `kernel/planner/dag.go`
- **Chi tiết**:
  ```go
  type DAG struct {
      nodes map[string]*DAGNode
      edges map[string][]string // node → dependencies
  }

  type DAGNode struct {
      Task   *agent.Task
      Status Status
      Result *agent.Result
  }

  func NewDAG(tasks []*agent.Task) (*DAG, error) {
      // 1. Build adjacency list
      // 2. Detect circular dependencies (topological sort)
      // 3. Return DAG or error
  }

  func (d *DAG) ReadyTasks() []*agent.Task {
      // Return tasks whose ALL dependencies are completed
  }

  func (d *DAG) MarkCompleted(taskID string, result *agent.Result) {
      // Update node status, potentially unblock dependent tasks
  }
  ```
- **⚠️ Pitfall #1**: Circular dependency detection. Dùng topological sort (Kahn's algorithm). Nếu sort không hoàn thành (remaining nodes > 0) → có cycle.
- **⚠️ Pitfall #2**: `ReadyTasks()` phải thread-safe. Orchestrator gọi nó từ scheduler goroutine.
- **⚠️ Pitfall #3**: Diamond dependency. Task D depends on B and C, B and C depend on A. Khi A hoàn thành → B và C ready → chạy song song → khi cả 2 xong → D ready. Logic phải xử lý đúng trường hợp này.

### Sub-task 5.1.4: Planning Strategies
- **File**: `kernel/planner/strategy.go`
- **Strategies**:
  ```go
  type Strategy string
  const (
      StrategySequential Strategy = "sequential" // Chạy từng task một
      StrategyParallel   Strategy = "parallel"   // Chạy song song tất cả tasks không phụ thuộc
      StrategyHybrid     Strategy = "hybrid"     // Tự động chọn (default)
  )
  ```
- **⚠️ Pitfall**: Sequential strategy vẫn phải tôn trọng DAG order. Không được chạy task B trước task A nếu B depends on A.

### Sub-task 5.1.5: Replanner
- **File**: `kernel/planner/replanner.go`
- **Chi tiết**:
  ```go
  func (r *Replanner) Replan(ctx context.Context, mission *Mission, failedTask *agent.Task, err error) ([]*agent.Task, error) {
      // 1. Build prompt: "Task X failed with error Y. How should we recover?"
      // 2. Options:
      //    a. Retry same task with different approach
      //    b. Replace task with alternative tasks
      //    c. Skip task if not critical
      //    d. Abort mission
      // 3. Return new task list
  }
  ```
- **⚠️ Pitfall #1**: Re-plan loop. Replanner gọi AI → AI tạo new plan → new plan cũng fail → replan lại → infinite loop. PHẢI có max replan attempts (default: 3).
- **⚠️ Pitfall #2**: Partial completion. Khi replan, phải biết tasks nào đã hoàn thành → không lặp lại.

### Sub-task 5.1.6: Plan Optimizer
- **File**: `kernel/planner/optimizer.go`
- **Chi tiết**: Tối ưu plan:
  - Merge tasks tương tự (2 tasks "write test for X" → 1 task "write tests for X and Y")
  - Parallelize independent tasks
  - Remove redundant tasks
- **⚠️ Pitfall**: Over-optimization có thể merge tasks mà thực tế nên tách riêng. Cần conservative approach.

### Sub-task 5.1.7: Unit tests
- **Tests**:
  - Decompose mission → task list (mock provider)
  - DAG build → no cycles
  - DAG build → detect cycle → error
  - ReadyTasks → returns correct tasks
  - Diamond dependency resolution
  - Replanner với max attempts
  - Strategy: sequential vs parallel vs hybrid

### Tiêu chí hoàn thành Task 5.1:
- [ ] Mission → decompose → DAG of tasks
- [ ] Circular dependency detection
- [ ] ReadyTasks() thread-safe
- [ ] Replan có max attempts
- [ ] 3 strategies (sequential, parallel, hybrid)
- [ ] Unit tests ≥ 85% coverage

---

## Task 5.2: Kernel — Orchestrator

### Sub-task 5.2.1: Main Orchestrator
- **File**: `kernel/orchestrator/orchestrator.go`
- **Chi tiết**:
  ```go
  type Orchestrator struct {
      planner   *planner.Planner
      scheduler *scheduler.Scheduler
      runtime   *runtime.Runtime
      registry  *registry.Registry
      eventbus  event.Bus
      logger    *slog.Logger
  }

  // ExecuteMission — Main entry point
  func (o *Orchestrator) ExecuteMission(ctx context.Context, mission *planner.Mission) (*MissionResult, error) {
      // 1. Plan: Decompose mission → DAG
      // 2. Schedule: Push tasks to scheduler
      // 3. Execute: Scheduler dispatches to runtime → agents execute
      // 4. Monitor: Watch for completions/failures
      // 5. Replan: If task fails → replan
      // 6. Aggregate: Collect all results
      // 7. Return: Final result
  }
  ```
- **⚠️ Pitfall #1**: Error handling strategy. Khi 1 task fail:
  - Option A: Abort toàn bộ mission → simple nhưng wasteful
  - Option B: Replan → smarter nhưng complex
  - Option C: Skip failed task, continue others → risky nếu có dependencies
  - **Recommendation**: Default = replan (max 3 times), sau đó abort. Configurable.
- **⚠️ Pitfall #2**: Context propagation. Mission context → Plan context → Task context. Cancel mission → cancel all tasks. Chain contexts:
  ```go
  missionCtx, missionCancel := context.WithCancel(ctx)
  defer missionCancel()
  ```
- **⚠️ Pitfall #3**: Deadlock. Nếu tất cả workers bận VÀ tất cả tasks đang chờ dependency → deadlock. Cần detect: nếu no progress trong timeout → abort.

### Sub-task 5.2.2: Coordinator
- **File**: `kernel/orchestrator/coordinator.go`
- **Chi tiết**: Quản lý inter-agent communication
  ```go
  // PassResult chuyển kết quả từ task A sang input của task B
  func (c *Coordinator) PassResult(fromTask, toTask string, result *agent.Result) error {
      // 1. Get task B from DAG
      // 2. Add result to task B's Context
      // 3. Mark dependency as satisfied
  }
  ```
- **⚠️ Pitfall**: Result size. Nếu task A trả về 100KB code → đưa toàn bộ vào context của task B → token limit exceeded. Cần summarize hoặc chỉ truyền relevant parts.

### Sub-task 5.2.3: Pipeline Manager
- **File**: `kernel/orchestrator/pipeline.go`
- **Chi tiết**: Quản lý luồng plan → schedule → execute → collect
- **States**: `Planning` → `Scheduling` → `Executing` → `Aggregating` → `Completed` / `Failed`
- **⚠️ Pitfall**: State transitions phải atomic. Race condition có thể khiến pipeline ở trạng thái inconsistent.

### Sub-task 5.2.4: Supervisor
- **File**: `kernel/orchestrator/supervisor.go`
- **Chi tiết**:
  ```go
  type Supervisor struct {
      running  map[string]*RunningTask
      mu       sync.RWMutex
  }

  func (s *Supervisor) Monitor(ctx context.Context) {
      ticker := time.NewTicker(5 * time.Second)
      for {
          select {
          case <-ctx.Done():
              return
          case <-ticker.C:
              s.checkHealth()
          }
      }
  }

  func (s *Supervisor) checkHealth() {
      // Check for:
      // 1. Tasks running longer than timeout
      // 2. Agents that stopped responding
      // 3. Provider health
  }
  ```
- **⚠️ Pitfall**: Supervisor check interval. Quá ngắn (100ms) → CPU waste. Quá dài (60s) → slow detection. Default: 5s.

### Sub-task 5.2.5: Aggregator
- **File**: `kernel/orchestrator/aggregator.go`
- **Chi tiết**:
  ```go
  type MissionResult struct {
      MissionID  string                    `json:"mission_id"`
      Status     Status                    `json:"status"`
      Tasks      map[string]*agent.Result  `json:"tasks"`
      Summary    string                    `json:"summary"`
      Artifacts  []agent.Artifact          `json:"artifacts"`
      Duration   time.Duration             `json:"duration"`
      TotalUsage *provider.Usage           `json:"total_usage"`
  }

  func (a *Aggregator) Aggregate(dag *planner.DAG) *MissionResult {
      // 1. Collect all task results from DAG
      // 2. Merge artifacts
      // 3. Sum up token usage
      // 4. Generate summary (optionally using AI)
      // 5. Return MissionResult
  }
  ```
- **⚠️ Pitfall**: Summary generation. Có thể dùng AI để summarize toàn bộ mission result. Nhưng phải consider: thêm 1 API call = thêm chi phí + latency. Cũng có thể dùng template-based summary.

### Sub-task 5.2.6: Feedback Collector
- **File**: `kernel/orchestrator/feedback.go`
- **Chi tiết**: Thu thập metrics cho feedback loop (Phase 6):
  - Thời gian hoàn thành mỗi task
  - Agent nào xử lý task nào
  - Success/failure rate per agent
  - Token usage per agent

### Sub-task 5.2.7: Unit & Integration tests
- **Tests**:
  - ExecuteMission → plan → execute → result (mock provider + agents)
  - Task failure → replan → retry → success
  - Task failure → max retries → abort
  - Parallel task execution
  - Diamond dependency resolution end-to-end
  - Deadlock detection
  - Context cancellation (cancel mission mid-flight)

### Tiêu chí hoàn thành Task 5.2:
- [ ] ExecuteMission end-to-end flow
- [ ] Task result passing between agents
- [ ] Failure handling (replan, retry, abort)
- [ ] Supervisor health monitoring
- [ ] Result aggregation
- [ ] Unit tests ≥ 80% coverage
- [ ] Integration test with mock components

---

## Task 5.3: Kernel — Resilience

### Sub-task 5.3.1: Circuit Breaker
- **File**: `kernel/resilience/circuit_breaker.go`
- **Chi tiết**:
  ```go
  type CircuitBreaker struct {
      failureThreshold int           // Số lần fail liên tiếp trước khi open circuit
      resetTimeout     time.Duration  // Thời gian chờ trước khi thử lại (half-open)
      state            State          // Closed, Open, HalfOpen
      failures         int
      lastFailure      time.Time
      mu               sync.Mutex
  }

  type State int
  const (
      StateClosed   State = iota // Normal, cho phép requests
      StateOpen                  // Blocked, reject requests ngay lập tức
      StateHalfOpen              // Cho phép 1 request thử, nếu thành công → Closed
  )

  func (cb *CircuitBreaker) Execute(fn func() error) error {
      // 1. Check state
      // 2. If Open && timeout chưa hết → return ErrCircuitOpen
      // 3. If Open && timeout hết → move to HalfOpen, allow 1 request
      // 4. Execute fn()
      // 5. If success → reset failures, move to Closed
      // 6. If fail → increment failures, if >= threshold → Open
  }
  ```
- **⚠️ Pitfall #1**: Thread-safety. Multiple goroutines gọi Execute() cùng lúc. State changes PHẢI dùng mutex.
- **⚠️ Pitfall #2**: Half-open state. Chỉ cho phép 1 request trong half-open. Nếu cho nhiều → tất cả đều fail → không có ý nghĩa.

### Sub-task 5.3.2: Retry với Exponential Backoff
- **File**: `kernel/resilience/retry.go`
- **Chi tiết**:
  ```go
  type RetryConfig struct {
      MaxAttempts  int
      InitialDelay time.Duration
      MaxDelay     time.Duration
      Multiplier   float64
      Jitter       bool // Random jitter để tránh thundering herd
  }

  func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
      for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
          err := fn()
          if err == nil {
              return nil
          }
          if !isRetryable(err) {
              return err // Non-retryable error
          }
          delay := calculateDelay(attempt, cfg)
          select {
          case <-ctx.Done():
              return ctx.Err()
          case <-time.After(delay):
          }
      }
      return ErrMaxRetriesExceeded
  }
  ```
- **⚠️ Pitfall #1**: `isRetryable()` phải phân biệt retryable errors (timeout, rate limit, 503) và non-retryable errors (invalid API key, 401, 403). Retry lỗi 401 = vô nghĩa.
- **⚠️ Pitfall #2**: Jitter. Nếu 100 requests đều retry sau 1s → thundering herd → provider lại overload. Thêm random jitter: `delay = baseDelay + random(0, baseDelay * 0.5)`.
- **⚠️ Pitfall #3**: Context cancellation trong retry loop. Nếu context bị cancel → DỪNG retry ngay, không đợi delay.

### Sub-task 5.3.3: Fallback
- **File**: `kernel/resilience/fallback.go`
- **Chi tiết**: Khi provider chính lỗi → chuyển sang provider backup
  ```go
  func WithFallback(primary, fallback func() error) error {
      err := primary()
      if err != nil {
          return fallback()
      }
      return nil
  }
  ```
- **⚠️ Pitfall**: Hiện tại chỉ có 1 provider (Antigravity). Fallback sẽ hữu ích khi thêm providers sau.

### Sub-task 5.3.4: Timeout Manager
- **File**: `kernel/resilience/timeout.go`
- **Chi tiết**: Centralized timeout management
- **⚠️ Pitfall**: Timeout cascading. Mission timeout = 10m, nhưng mỗi task timeout = 5m, 3 tasks sequential = 15m > 10m. Mission timeout PHẢI override task timeout nếu remaining time < task timeout.

### Sub-task 5.3.5: Health Checker
- **File**: `kernel/resilience/health.go`
- **Chi tiết**: Kiểm tra health của providers, agents, tools

### Sub-task 5.3.6: Auto Recovery
- **File**: `kernel/resilience/recovery.go`
- **Chi tiết**: Tự động khôi phục sau crash
- **⚠️ Pitfall**: Recovery state. Lưu mission state vào disk. Khi restart → load state → resume.

### Sub-task 5.3.7: Unit tests
- **Tests**:
  - Circuit breaker: Closed → Open → HalfOpen → Closed
  - Retry: success on 3rd attempt, non-retryable error, max retries
  - Fallback: primary fails → fallback succeeds
  - Timeout cascading

### Tiêu chí hoàn thành Task 5.3:
- [ ] Circuit breaker 3 states
- [ ] Retry với exponential backoff + jitter
- [ ] Fallback mechanism
- [ ] Timeout cascading
- [ ] Health checks
- [ ] Unit tests ≥ 85% coverage

---

## Task 5.4: Kernel — Security

### Sub-task 5.4.1: Permission Manager
- **File**: `kernel/security/permission.go`
- **Chi tiết**:
  ```go
  type PermissionManager struct {
      policies map[string]*Policy // agent name → policy
  }

  type Policy struct {
      AllowedTools    []string `yaml:"allowed_tools"`
      BlockedCommands []string `yaml:"blocked_commands"`
      AllowedPaths    []string `yaml:"allowed_paths"`
      BlockedPaths    []string `yaml:"blocked_paths"`
      MaxFileSize     int64    `yaml:"max_file_size"`    // bytes
      MaxOutputSize   int64    `yaml:"max_output_size"`  // bytes
  }

  func (pm *PermissionManager) CanUseTool(agentName, toolName string) bool { ... }
  func (pm *PermissionManager) CanAccessPath(agentName, path string) bool { ... }
  func (pm *PermissionManager) CanRunCommand(agentName, command string) bool { ... }
  ```
- **⚠️ Pitfall #1**: Default deny. Nếu agent không có policy → block tất cả. KHÔNG default allow.
- **⚠️ Pitfall #2**: Path matching. `BlockedPaths: ["/etc", "C:\\Windows"]` phải match subdirectories. Dùng `strings.HasPrefix()` hoặc `filepath.Match()`.
- **⚠️ Pitfall #3**: Command matching. `BlockedCommands: ["rm -rf"]` phải match `rm -rf /`, `rm  -rf /tmp` (extra spaces). Normalize trước khi match.

### Sub-task 5.4.2: Sandbox
- **File**: `kernel/security/sandbox.go`
- **Chi tiết**: Chạy commands trong môi trường hạn chế
- **⚠️ Pitfall**: Full sandbox (Docker/chroot) phức tạp. Phase đầu: dùng permission-based sandbox (check trước khi execute). Phase sau: Docker-based sandbox.

### Sub-task 5.4.3: Audit Logger
- **File**: `kernel/security/audit.go`
- **Chi tiết**: Log toàn bộ hành động của agents vào audit log riêng
  ```go
  type AuditEntry struct {
      Timestamp time.Time `json:"timestamp"`
      Agent     string    `json:"agent"`
      Action    string    `json:"action"` // "tool_call", "file_read", "file_write", "command"
      Target    string    `json:"target"` // tool name, file path, command
      Allowed   bool      `json:"allowed"`
      Details   string    `json:"details,omitempty"`
  }
  ```
- **⚠️ Pitfall**: Audit log KHÔNG dùng standard logger. File riêng, append-only, không rotate bừa bãi.

### Sub-task 5.4.4: Secrets Manager
- **File**: `kernel/security/secrets.go`
- **Chi tiết**: Quản lý API keys, credentials
- **⚠️ Pitfall #1**: KHÔNG lưu secrets trong config YAML. Đọc từ environment variables hoặc secret store.
- **⚠️ Pitfall #2**: KHÔNG log secrets. Redact trong log output.

### Sub-task 5.4.5: Unit tests
- **Tests**:
  - Permission: allowed tool, blocked tool
  - Permission: allowed path, blocked path, subdirectory
  - Permission: blocked command with variations
  - Audit log entries
  - Default deny policy

### Tiêu chí hoàn thành Task 5.4:
- [ ] Permission-based access control
- [ ] Default deny
- [ ] Audit logging
- [ ] Secrets from environment
- [ ] No secrets in logs
- [ ] Unit tests ≥ 85% coverage

---

## 📋 Checklist tổng Phase 5

- [ ] Task 5.1: Planner (7 sub-tasks)
- [ ] Task 5.2: Orchestrator (7 sub-tasks)
- [ ] Task 5.3: Resilience (7 sub-tasks)
- [ ] Task 5.4: Security (5 sub-tasks)
- [ ] **Milestone M4: First Mission** — Mission → Plan → Execute → Result end-to-end
- [ ] Integration test: full orchestration flow
- [ ] `go test -race ./kernel/...` pass
- [ ] Git commit: "Phase 5: Orchestration engine"
