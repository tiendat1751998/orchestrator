# Phase 6: CLI, API & Polish — Chi Tiết Từng Sub-Task

> [!NOTE]
> Phase này biến hệ thống từ "library" thành "product". Sau phase này, người dùng có thể tương tác qua CLI hoặc API.

---

## Task 6.1: CLI — orchestrator-cli

### Sub-task 6.1.1: CLI Framework Setup
- **File**: `cmd/orchestrator/main.go`
- **Chi tiết**:
  - Dùng `cobra` library cho CLI framework (hoặc `urfave/cli`)
  - Root command: `orchestrator`
  - Version flag: `--version`
  - Config flag: `--config <path>`
  - Verbose flag: `--verbose` / `-v`
- **⚠️ Pitfall #1**: Dùng `cobra` (phổ biến nhất trong Go ecosystem, kubectl, docker dùng). KHÔNG tự implement CLI parser.
- **⚠️ Pitfall #2**: Global flags (--config, --verbose) PHẢI persistent (áp dụng cho tất cả sub-commands).

### Sub-task 6.1.2: Command — `orchestrator mission`
- **File**: `cmd/orchestrator/cmd/mission.go`
- **Usage**:
  ```bash
  # Interactive
  orchestrator mission "Build a REST API for user management with Go and Gin"

  # From file
  orchestrator mission --file mission.yaml

  # With constraints
  orchestrator mission "Build REST API" --lang go --framework gin --no-external-deps
  ```
- **Chi tiết**:
  1. Parse mission từ args hoặc file
  2. Khởi tạo Kernel
  3. Gọi `Orchestrator.ExecuteMission()`
  4. Stream progress to terminal (live updates)
  5. Print final result
- **⚠️ Pitfall #1**: Long-running command. Mission có thể chạy 30 phút. Cần:
  - Progress indicator (spinner, progress bar)
  - Live streaming task status
  - Ctrl+C graceful shutdown
- **⚠️ Pitfall #2**: Output formatting. Kết quả phải đẹp và dễ đọc trên terminal. Dùng color, indentation, boxes.

### Sub-task 6.1.3: Command — `orchestrator status`
- **File**: `cmd/orchestrator/cmd/status.go`
- **Usage**:
  ```bash
  orchestrator status              # Status of current/last mission
  orchestrator status <mission-id> # Status of specific mission
  ```

### Sub-task 6.1.4: Command — `orchestrator agents`
- **File**: `cmd/orchestrator/cmd/agents.go`
- **Usage**:
  ```bash
  orchestrator agents list         # List registered agents
  orchestrator agents info backend # Details of specific agent
  ```

### Sub-task 6.1.5: Command — `orchestrator providers`
- **File**: `cmd/orchestrator/cmd/providers.go`
- **Usage**:
  ```bash
  orchestrator providers list        # List registered providers
  orchestrator providers test gemini # Test connectivity
  ```

### Sub-task 6.1.6: Command — `orchestrator config`
- **File**: `cmd/orchestrator/cmd/config.go`
- **Usage**:
  ```bash
  orchestrator config show           # Show current config
  orchestrator config init           # Create default config file
  orchestrator config set key value  # Update config
  ```

### Sub-task 6.1.7: Terminal UI — Progress & Live Updates
- **File**: `cmd/orchestrator/ui/progress.go`
- **Chi tiết**: Live update trạng thái mission trên terminal
  ```
  🎯 Mission: Build REST API for user management
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  📋 Plan: 5 tasks

  ✅ [1/5] Design API schema        (Architect)    12s
  🔄 [2/5] Implement handlers       (Backend)      ...
  ⏳ [3/5] Write unit tests          (Backend)      waiting
  ⏳ [4/5] Review code              (Reviewer)     waiting
  ⏳ [5/5] Create Dockerfile        (DevOps)       waiting

  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ⏱️  Elapsed: 47s | 💰 Tokens: 12,456
  ```
- **⚠️ Pitfall**: Terminal clearing. Dùng ANSI escape codes. Trên Windows, cần `golang.org/x/sys/windows` để enable virtual terminal processing.

### Sub-task 6.1.8: Unit tests cho CLI
- **Tests**: Command parsing, flag handling, config loading

### Tiêu chí hoàn thành Task 6.1:
- [ ] `orchestrator mission "..."` works
- [ ] Live progress updates trên terminal
- [ ] Graceful shutdown (Ctrl+C)
- [ ] `orchestrator agents list` works
- [ ] `orchestrator providers list` works
- [ ] `orchestrator config init` creates default config

---

## Task 6.2: API — REST/gRPC Gateway

### Sub-task 6.2.1: REST API Server
- **File**: `kernel/gateway/rest.go`
- **Endpoints**:
  ```
  POST   /api/v1/missions              — Create new mission
  GET    /api/v1/missions               — List missions
  GET    /api/v1/missions/:id           — Get mission status
  DELETE /api/v1/missions/:id           — Cancel mission
  GET    /api/v1/missions/:id/stream    — SSE stream of mission progress
  GET    /api/v1/agents                 — List agents
  GET    /api/v1/providers              — List providers
  GET    /api/v1/health                 — Health check
  ```
- **⚠️ Pitfall #1**: Dùng `net/http` standard library + `chi` router (lightweight). KHÔNG dùng `gin` hay `echo` cho hệ thống internal.
- **⚠️ Pitfall #2**: SSE (Server-Sent Events) cho real-time streaming. Đơn giản hơn WebSocket cho unidirectional stream.
- **⚠️ Pitfall #3**: CORS headers. Nếu Web UI chạy trên domain khác → cần CORS.

### Sub-task 6.2.2: WebSocket Server
- **File**: `kernel/gateway/websocket.go`
- **Chi tiết**: Bidirectional communication cho interactive missions
- **⚠️ Pitfall #1**: Ping/pong heartbeat. WebSocket connections có thể die silently. Cần periodic ping.
- **⚠️ Pitfall #2**: Connection cleanup. Dùng context cancellation + defer close.

### Sub-task 6.2.3: gRPC Server (placeholder)
- **File**: `kernel/gateway/grpc.go`
- **Chi tiết**: Placeholder cho inter-service communication. Implement sau nếu cần.

### Sub-task 6.2.4: Message Queue (placeholder)
- **File**: `kernel/gateway/queue.go`
- **Chi tiết**: Placeholder cho NATS/Kafka integration. Implement sau nếu cần.

### Sub-task 6.2.5: Gateway Bootstrap
- **File**: `kernel/gateway/gateway.go`
- **Chi tiết**: Start/Stop gateway server, wire routes

### Sub-task 6.2.6: API tests
- **Tests**: HTTP endpoint tests, SSE streaming test

### Tiêu chí hoàn thành Task 6.2:
- [ ] REST API hoạt động
- [ ] SSE streaming cho mission progress
- [ ] Health check endpoint
- [ ] API tests pass

---

## Task 6.3: Modules — Mission, Workspace, Session

### Sub-task 6.3.1: Mission Module
- **File**: `modules/mission/manager.go`
- **Chi tiết**:
  ```go
  type Manager struct {
      store MissionStore
  }

  func (m *Manager) Create(mission *planner.Mission) error { ... }
  func (m *Manager) Get(id string) (*planner.Mission, error) { ... }
  func (m *Manager) List() ([]*planner.Mission, error) { ... }
  func (m *Manager) UpdateStatus(id string, status Status) error { ... }
  ```
- **Storage**: SQLite cho local, PostgreSQL cho production (qua interface)
- **⚠️ Pitfall**: Mission history phải persist. Khi restart → vẫn xem được missions cũ.

### Sub-task 6.3.2: Mission Store (SQLite)
- **File**: `modules/mission/store_sqlite.go`
- **Chi tiết**: SQLite implementation cho MissionStore interface
- **⚠️ Pitfall #1**: SQLite concurrent writes. Dùng `_journal_mode=WAL` và `_busy_timeout=5000`.
- **⚠️ Pitfall #2**: Migration. Schema changes cần migration mechanism. Dùng embedded SQL files hoặc Go migration library.

### Sub-task 6.3.3: Workspace Module
- **File**: `modules/workspace/workspace.go`
- **Chi tiết**: Quản lý working directory cho missions
- **⚠️ Pitfall**: Workspace isolation. 2 missions chạy cùng lúc trên cùng workspace → conflict. Cần locking mechanism.

### Sub-task 6.3.4: Session Module
- **File**: `modules/session/session.go`
- **Chi tiết**: Session state persistence (cho resume after crash)
- **⚠️ Pitfall**: Serialization. Session state phải serializable. Không lưu function pointers hay channels.

### Sub-task 6.3.5: Artifact Module
- **File**: `modules/artifact/artifact.go`
- **Chi tiết**: Quản lý output files, diffs, reports
- **⚠️ Pitfall**: Artifact storage location. Mặc định: `.orchestrator/artifacts/<mission-id>/`. Phải configurable.

### Sub-task 6.3.6: Execution Module
- **File**: `modules/execution/execution.go`
- **Chi tiết**: Execution history & logs per task
- **⚠️ Pitfall**: Log rotation. Execution logs có thể rất lớn. Auto-rotate sau 30 ngày.

### Tiêu chí hoàn thành Task 6.3:
- [ ] Mission CRUD with SQLite
- [ ] Mission history persist across restarts
- [ ] Workspace management
- [ ] Artifact storage
- [ ] Unit tests ≥ 80% coverage

---

## Task 6.4: Feedback Loop & Metrics

### Sub-task 6.4.1: Evaluator
- **File**: `kernel/feedback/evaluator.go`
- **Chi tiết**: Đánh giá chất lượng output
  - Code output: `go build`, `go test` pass?
  - Review output: có actionable feedback?
- **⚠️ Pitfall**: Automated evaluation khó. Phase đầu dùng simple heuristics (build pass = success). Phase sau dùng AI-as-judge.

### Sub-task 6.4.2: Scorer
- **File**: `kernel/feedback/scorer.go`
- **Chi tiết**: Score agents theo hiệu suất
  - Success rate
  - Average duration
  - Token efficiency (result quality / tokens used)

### Sub-task 6.4.3: Learner
- **File**: `kernel/feedback/learner.go`
- **Chi tiết**: Học từ missions trước để improve planning
- **⚠️ Pitfall**: Phase đầu: chỉ lưu statistics. Phase sau: dùng statistics để optimize agent selection.

### Sub-task 6.4.4: Ranking
- **File**: `kernel/feedback/ranking.go`
- **Chi tiết**: Xếp hạng agent phù hợp nhất cho từng loại task

### Sub-task 6.4.5: Metrics
- **File**: `kernel/metrics/metrics.go`
- **Chi tiết**: Prometheus-compatible metrics (counters, histograms, gauges)
- **⚠️ Pitfall**: Metric cardinality. KHÔNG tạo metric per task ID (high cardinality). Dùng labels: `agent_name`, `task_type`, `status`.

### Tiêu chí hoàn thành Task 6.4:
- [ ] Agent scoring hoạt động
- [ ] Metrics exportable
- [ ] Statistics persisted

---

## Task 6.5: Documentation & Examples

### Sub-task 6.5.1: README.md
- **File**: `README.md`
- **Sections**: Introduction, Features, Quick Start, Installation, Configuration, Architecture

### Sub-task 6.5.2: Architecture docs
- **File**: `docs/architecture.md`
- **Nội dung**: Diagrams, component descriptions, data flow

### Sub-task 6.5.3: Plugin Development Guide
- **File**: `docs/plugin-development.md`
- **Nội dung**: How to create custom agents, providers, tools

### Sub-task 6.5.4: API docs
- **File**: `docs/api.md`
- **Nội dung**: REST API reference, request/response examples

### Sub-task 6.5.5: Examples
- **Files**:
  - `examples/hello-world/` — Simplest mission
  - `examples/rest-api/` — Build REST API mission
  - `examples/custom-agent/` — Create custom agent
  - `examples/custom-provider/` — Create custom provider

### Tiêu chí hoàn thành Task 6.5:
- [ ] README có Quick Start
- [ ] Architecture diagram
- [ ] Plugin development guide
- [ ] Ít nhất 2 working examples

---

## 📋 Checklist tổng Phase 6

- [ ] Task 6.1: CLI (8 sub-tasks)
- [ ] Task 6.2: API Gateway (6 sub-tasks)
- [ ] Task 6.3: Modules (6 sub-tasks)
- [ ] Task 6.4: Feedback & Metrics (5 sub-tasks)
- [ ] Task 6.5: Documentation (5 sub-tasks)
- [ ] **Milestone M5: Production Ready**
- [ ] End-to-end demo: CLI → mission → result
- [ ] Git commit: "Phase 6: CLI, API, and polish"
- [ ] Tag release: `v0.1.0`

---

## 📊 Tổng kết toàn bộ Project

| Phase | Tasks | Sub-tasks | Ước lượng |
|---|---|---|---|
| Phase 1: Contracts | 6 | 25 | 2 tuần |
| Phase 2: Kernel | 6 | 27 | 3 tuần |
| Phase 3: SDK | 3 | 14 | 2 tuần |
| Phase 4: Plugins | 3 | 26 | 3 tuần |
| Phase 5: Orchestration | 4 | 26 | 3 tuần |
| Phase 6: Polish | 5 | 30 | 4 tuần |
| **Tổng** | **27** | **148** | **~17 tuần** |

> [!TIP]
> **Gợi ý workflow**: Mỗi lần làm việc, chọn 1 sub-task, nói "Implement sub-task X.Y.Z" và tôi sẽ code chi tiết cho sub-task đó. Mỗi sub-task = 1 commit nhỏ, dễ review, dễ rollback.
