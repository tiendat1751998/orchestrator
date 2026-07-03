# RFC-0029: Web UI & Mission Control Protocol

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0024 (API Gateways), RFC-0016 (Observation Runtime)

## Summary

This RFC specifies the design of the **Web UI & Mission Control Protocol** in AEOS. It details the schema and design system for the real-time execution dashboard (`web/`), visualizing plan DAG states, active tasks, memory state maps, and logs streamed from the Kernel's WebSocket server.

## Motivation

Command-line interfaces (CLIs) are fast, but complex multi-agent schedules are easier to inspect and coordinate visually.
- Developers need a visual dashboard to inspect running tasks, track execution bottlenecks, view time-travel frames, and handle Human-in-the-Loop review approvals.

## Design

### 1. Architectural Placement

The Web UI connects to the running Kernel API Gateway via WebSocket channels.

```
  Go Kernel WebSocket API ──(Streams Events)──► Web UI (Vite + Vanilla CSS dashboard)
```

---

### 2. UI Features & Layout

- **Mission DAG View**: An interactive graph showing task nodes colored by state (Grey: pending, Orange: running, Green: completed, Red: failed).
- **Time-Travel Slider**: A frame-by-frame navigation bar linked to the `TimeTravelDebugger` (RFC-0054).
- **Security Intercept Modal**: Prompts the user with Approve/Reject buttons when an execution task requests unsandboxed permissions.

## Impact

- **Visual Control Panel**: Streamlines developer monitoring, debugging, and coordination loops.
- **Easy Approvals**: Integrates approval flows directly into the dashboard interface.
