# RFC-0024: API Gateways (gRPC, REST, WebSockets)

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0008 (Event Model)

## Summary

This RFC specifies the design of **API Gateways (gRPC, REST, WebSockets)** in AEOS. It defines the communication protocols that allow external CLI tools (like Antigravity CLI) or Web UIs to interact with the central Go Kernel: streaming logs, starting missions, and subscribing to Event Store updates.

## Motivation

To operate as a unified platform, the AEOS Kernel must expose clean API ports.
- External monitoring tools, IDE plugins, and browser test engines need to connect to the running kernel.
- Exposing standard REST, gRPC, and WebSocket ports ensures language-independent connectivity.

## Design

### 1. Architectural Placement

The API Gateway is a service inside the CLI/Command layer, exposing external network ports.

```
  Antigravity CLI ──(gRPC/REST)──► [API Gateway] ──► [Kernel Core] ◄──(WebSockets)── Web UI
```

---

### 2. Contracts (`contracts/api/gateway.go`)

```go
package api

import (
	"context"
)

// GatewayConfig defines ports and protocol targets.
type GatewayConfig struct {
	GRPCPort int `json:"grpc_port"`
	RESTPort int `json:"rest_port"`
	WSPort   int `json:"ws_port"`
}

// APIGateway manages external API services.
type APIGateway interface {
	// Start opens network ports and begins listening.
	Start(ctx context.Context, config GatewayConfig) error
	
	// Stop shuts down all listener servers.
	Stop(ctx context.Context) error
}
```

## Impact

- **Language Interoperability**: External tools (like Python benchmark runners) can interface with the Go kernel.
- **Real-Time UI Updates**: WebSocket streaming pushes active FSM state changes to the Web UI instantly.
