# Phase 1 — Micro-Tasks Index

> **Quy tắc**: Mỗi micro-task = 1 file duy nhất = 1 commit nhỏ.
> AI chỉ cần đọc 1 micro-task file và implement chính xác những gì được mô tả.
> Go version: **1.26**
>
> **Triết lý kiến trúc**: Hệ thống vừa là Orchestrator vừa là Bộ Não Thứ 2.
> AI là **công cụ tạo nội dung**, không phải bộ não. Brain = Go logic thuần.

## Thứ tự thực hiện (PHẢI theo đúng thứ tự)

```mermaid
graph TD
    A["1.01 go.mod"] --> B["1.02 .gitignore"]
    B --> C["1.03 Makefile"]
    C --> D["1.04 .golangci.yml"]
    D --> E["1.05 contracts/errors.go"]
    E --> F["1.06 contracts/types.go"]
    F --> G["1.07 contracts/status.go"]
    G --> H["1.08 provider/message.go"]
    H --> I["1.09 provider/request.go"]
    I --> J["1.10 provider/response.go"]
    J --> K["1.11 provider/config.go"]
    K --> L["1.12 provider/provider.go"]
    L --> M["1.13 provider/provider_test.go"]
    M --> N["1.14 tool/schema.go"]
    N --> O["1.15 tool/tool.go"]
    O --> P["1.16 tool/tool_test.go"]
    P --> Q["1.17 agent/capability.go"]
    Q --> R["1.18 agent/task.go"]
    R --> S["1.19 agent/result.go"]
    S --> T["1.20 agent/manifest.go"]
    T --> U["1.21 agent/agent.go"]
    U --> V["1.22 agent/agent_test.go"]
    V --> W["1.23 event/event.go"]
    W --> X["1.24 plugin/plugin.go"]
    X --> Y["1.25 memory/memory.go"]
    Y --> Z["1.26 search/search.go"]
    Z --> AA["1.27 workflow/workflow.go"]
    AA --> AB["1.28 context/context.go"]
    AB --> AC["1.29 planner/planner.go"]
    AC --> AD["1.30 orchestrator/orchestrator.go"]
    AD --> AE["1.31 resilience/resilience.go"]
    AE --> AF["1.32 security/security.go"]
    AF --> AG["1.33 gateway/gateway.go"]
    AG --> AH["1.34 feedback/feedback.go"]

    %% Brain & Knowledge — Second Brain contracts
    G --> AK["1.41 🧠 brain/brain.go"]
    AK --> AL["1.42 📚 knowledge/knowledge.go"]

    %% Entry point & verify depend on ALL contracts
    AH --> AM["1.43 cmd/main.go"]
    AL --> AM
    AM --> AN["1.44 verify: go build"]
```

## Danh sách Micro-Tasks (36 files riêng lẻ)

### Project Setup (4 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.01 | [micro_1.01_go_mod.md](micro_1.01_go_mod.md) | `go.mod` | 5 min |
| 1.02 | [micro_1.02_gitignore.md](micro_1.02_gitignore.md) | `.gitignore` | 5 min |
| 1.03 | [micro_1.03_makefile.md](micro_1.03_makefile.md) | `Makefile` | 10 min |
| 1.04 | [micro_1.04_golangci.md](micro_1.04_golangci.md) | `.golangci.yml` | 5 min |

### Shared Contracts (3 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.05 | [micro_1.05_errors.md](micro_1.05_errors.md) | `contracts/errors.go` | 10 min |
| 1.06 | [micro_1.06_types.md](micro_1.06_types.md) | `contracts/types.go` | 10 min |
| 1.07 | [micro_1.07_status.md](micro_1.07_status.md) | `contracts/status.go` | 10 min |

### Provider Contracts (6 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.08 | [micro_1.08_provider_message.md](micro_1.08_provider_message.md) | `contracts/provider/message.go` | 15 min |
| 1.09 | [micro_1.09_provider_request.md](micro_1.09_provider_request.md) | `contracts/provider/request.go` | 15 min |
| 1.10 | [micro_1.10_provider_response.md](micro_1.10_provider_response.md) | `contracts/provider/response.go` | 15 min |
| 1.11 | [micro_1.11_provider_config.md](micro_1.11_provider_config.md) | `contracts/provider/config.go` | 10 min |
| 1.12 | [micro_1.12_provider_interface.md](micro_1.12_provider_interface.md) | `contracts/provider/provider.go` | 10 min |
| 1.13 | [micro_1.13_provider_test.md](micro_1.13_provider_test.md) | `contracts/provider/provider_test.go` | 20 min |

### Tool Contracts (3 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.14 | [micro_1.14_tool_schema.md](micro_1.14_tool_schema.md) | `contracts/tool/schema.go` | 10 min |
| 1.15 | [micro_1.15_tool_interface.md](micro_1.15_tool_interface.md) | `contracts/tool/tool.go` | 10 min |
| 1.16 | [micro_1.16_tool_test.md](micro_1.16_tool_test.md) | `contracts/tool/tool_test.go` | 15 min |

### Agent Contracts (6 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.17 | [micro_1.17_agent_capability.md](micro_1.17_agent_capability.md) | `contracts/agent/capability.go` | 10 min |
| 1.18 | [micro_1.18_agent_task.md](micro_1.18_agent_task.md) | `contracts/agent/task.go` | 15 min |
| 1.19 | [micro_1.19_agent_result.md](micro_1.19_agent_result.md) | `contracts/agent/result.go` | 15 min |
| 1.20 | [micro_1.20_agent_manifest.md](micro_1.20_agent_manifest.md) | `contracts/agent/manifest.go` | 10 min |
| 1.21 | [micro_1.21_agent_interface.md](micro_1.21_agent_interface.md) | `contracts/agent/agent.go` | 10 min |
| 1.22 | [micro_1.22_agent_test.md](micro_1.22_agent_test.md) | `contracts/agent/agent_test.go` | 20 min |

### Other Contracts (12 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.23 | [micro_1.23_event.md](micro_1.23_event.md) | `contracts/event/event.go` | 15 min |
| 1.24 | [micro_1.24_plugin.md](micro_1.24_plugin.md) | `contracts/plugin/plugin.go` | 10 min |
| 1.25 | [micro_1.25_memory.md](micro_1.25_memory.md) | `contracts/memory/memory.go` | 10 min |
| 1.26 | [micro_1.26_search.md](micro_1.26_search.md) | `contracts/search/search.go` | 10 min |
| 1.27 | [micro_1.27_workflow.md](micro_1.27_workflow.md) | `contracts/workflow/workflow.go` | 10 min |
| 1.28 | [micro_1.28_context.md](micro_1.28_context.md) | `contracts/context/context.go` | 10 min |
| 1.29 | [micro_1.29_planner.md](micro_1.29_planner.md) | `contracts/planner/planner.go` | 10 min |
| 1.30 | [micro_1.30_orchestrator.md](micro_1.30_orchestrator.md) | `contracts/orchestrator/orchestrator.go` | 10 min |
| 1.31 | [micro_1.31_resilience.md](micro_1.31_resilience.md) | `contracts/resilience/resilience.go` | 10 min |
| 1.32 | [micro_1.32_security.md](micro_1.32_security.md) | `contracts/security/security.go` | 10 min |
| 1.33 | [micro_1.33_gateway.md](micro_1.33_gateway.md) | `contracts/gateway/gateway.go` | 5 min |
| 1.34 | [micro_1.34_feedback.md](micro_1.34_feedback.md) | `contracts/feedback/feedback.go` | 10 min |

### 🧠 Brain & Knowledge Contracts — Second Brain (2 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.41 | [micro_1.41_brain_contracts.md](micro_1.41_brain_contracts.md) | `contracts/brain/brain.go` | 15 min |
| 1.42 | [micro_1.42_knowledge_contracts.md](micro_1.42_knowledge_contracts.md) | `contracts/knowledge/knowledge.go` | 15 min |

### Production Hardening (4 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.37 | [micro_1.37_structured_errors.md](micro_1.37_structured_errors.md) | `contracts/errors.go` (Enhanced) | 20 min |
| 1.38 | [micro_1.38_context_metadata.md](micro_1.38_context_metadata.md) | `contracts/context/metadata.go` | 15 min |
| 1.39 | [micro_1.39_input_validation.md](micro_1.39_input_validation.md) | `contracts/agent/task.go`, `contracts/provider/request.go` | 20 min |
| 1.40 | [micro_1.40_health_report.md](micro_1.40_health_report.md) | `contracts/plugin/health.go`, `contracts/plugin/plugin.go` | 20 min |

### Entry Point & Verification (2 tasks)
| # | Micro-task file | Target file | Thời gian |
|---|---|---|---|
| 1.43 | [micro_1.35_cmd_main.md](micro_1.35_cmd_main.md) | `cmd/orchestrator/main.go` | 5 min |
| 1.44 | [micro_1.36_verify.md](micro_1.36_verify.md) | — (verification only) | 15 min |

---

## Tổng kết

| Nhóm | Số tasks | Ước lượng |
|---|---|---|
| Project Setup | 4 | 25 min |
| Shared Contracts | 3 | 30 min |
| Provider Contracts | 6 | 85 min |
| Tool Contracts | 3 | 35 min |
| Agent Contracts | 6 | 80 min |
| Other Contracts | 12 | 120 min |
| 🧠 Brain & Knowledge | 2 | 30 min |
| Production Hardening | 4 | 75 min |
| Entry Point + Verify | 2 | 20 min |
| **Tổng** | **42** | **~8.3 giờ** |

## Cách sử dụng

Đưa cho AI bất kỳ nội dung sau:

```
Hãy đọc file docs/tasks/phase1/micro_1.XX_name.md và implement CHÍNH XÁC
những gì được mô tả trong đó. Không thêm, không bớt.
```
