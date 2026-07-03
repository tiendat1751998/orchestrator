# Deployment Topology & CLI Reference

## 1. Runtime Deployment Model

The system is deployed as a single, compiled Go CLI binary (`orchestrator`) with zero external dependencies (databases, message buses, caching servers). Everything executes locally in-process.

```
                  ┌──────────────────────────────┐
                  │      Terminal CLI Client     │
                  └──────────────┬───────────────┘
                                 │ execution & IPC
                                 ▼
                  ┌──────────────────────────────┐
                  │      Orchestrator Binary     │
                  │                              │
                  │  ┌────────────────────────┐  │
                  │  │     Cobra CLI Engine   │  │
                  │  └───────────┬────────────┘  │
                  │              │               │
                  │  ┌───────────▼────────────┐  │
                  │  │    Kernel Core Engine  │  │
                  │  └───────────┬────────────┘  │
                  │              │               │
                  │  ┌───────────▼────────────┐  │
                  │  │       Plugin Registry  │  │
                  │  └───────────┬────────────┘  │
                  │              │               │
                  │  ┌───────────▼────────────┐  │
                  │  │       REST/WS Gateway  │  │
                  │  └────────────────────────┘  │
                  └──────────────────────────────┘
```

---

## 2. CLI Command Specifications

All commands use a cobra-based parser. The binary supports the following options:

### Version Info
```bash
orchestrator --version
# Output format:
# orchestrator version dev (commit: unknown, built: unknown)
```

### Config Management
- **Initialize config**:
  ```bash
  orchestrator config init --output /path/to/settings.yaml
  ```
- **Show configuration** (redacts sensitive API keys):
  ```bash
  orchestrator config show --config /path/to/settings.yaml
  ```

### Plugin Inspection
- **List registered agents**:
  ```bash
  orchestrator agents list --config /path/to/settings.yaml
  ```
- **List registered providers**:
  ```bash
  orchestrator providers list --config /path/to/settings.yaml
  ```

### Daemon Server (API Gateway Mode)
- **Start REST/WS gateway**:
  ```bash
  orchestrator serve --config /path/to/settings.yaml
  ```
  Starts a web gateway serving:
  - `GET /api/v1/health` (returns `{"data":{"status":"healthy"}}`)
  - `GET /api/v1/missions` (returns active and completed mission list)
  - `WebSocket /ws/missions/{id}` (real-time stream of task status updates)
