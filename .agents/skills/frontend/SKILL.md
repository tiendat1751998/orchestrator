---
name: Frontend Engineer
description: Instructions for developing the Orchestrator monitoring dashboard (web/), mission status views, and real-time agent execution displays.
---

# Frontend Engineer Playbook

## Session Startup (MANDATORY)
1. Read `.agents/context/deployment-topology.md` — gateway endpoint and ports.
2. Read `.agents/context/api-contracts.md` — REST API and WebSocket endpoints.
3. Read `.agents/context/architecture.md` — understand `web/` directory purpose.

**NEVER start modifying frontend files without reading the API contract.**

## Overview

The `web/` directory contains the Orchestrator monitoring dashboard — a lightweight web UI for:
- Viewing active/completed missions and their task DAGs.
- Monitoring agent execution in real-time via WebSocket.
- Browsing registered agents, providers, and tools.
- Viewing system health and configuration.

## Technology Stack
- **HTML5/CSS3/Vanilla JavaScript** — no frameworks.
- Served by the Orchestrator gateway (`kernel/gateway/rest.go`).
- Real-time updates via WebSocket (`/ws/missions/{id}`).

## Key Views

| View | Description | API Source |
|------|-------------|-----------|
| Mission Dashboard | List all missions with status | `GET /api/v1/missions` |
| Mission Detail | Task DAG visualization, live progress | `GET /api/v1/missions/{id}` + WS |
| Agent Registry | List agents and capabilities | `GET /api/v1/agents` |
| Provider Status | Provider health and usage stats | `GET /api/v1/providers` |
| System Config | Current configuration | `GET /api/v1/config` |

## Rules
- Keep it lightweight — no build tools, no npm, no bundlers.
- Progressive enhancement — works without JavaScript (basic HTML).
- Responsive design — works on mobile and desktop.
- Use CSS variables for theming.
- All API calls use `fetch()` with proper error handling.

## 🚫 ANTI-FAKE RULES
- Every "UI works" claim → MUST show screenshot or browser test output.
- Every "API integration works" → MUST show actual fetch response.
