# Micro-Task 6.12: Create kernel/gateway/websocket.go

## Info
- **File**: `kernel/gateway/websocket.go`
- **Package**: `gateway`
- **Depends on**: 6.09
- **Time**: 20 min
- **Verify**: `go build ./kernel/gateway/...`

## External dependencies
```bash
go get github.com/gorilla/websocket@latest
```

## Purpose
Implements bidirectional WebSocket handler for interactive mission sessions. Supports ping/pong heartbeat, connection pool cleanup, and structured JSON message protocol.

## EXACT code to create

```go
package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict origins in production
	},
}

// WSMessage is the structured WebSocket message format.
type WSMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

// WSConnection wraps a single WebSocket connection with context.
type WSConnection struct {
	conn   *websocket.Conn
	send   chan []byte
	logger *slog.Logger
}

// WSHub manages active WebSocket connections. Thread-safe.
type WSHub struct {
	mu          sync.RWMutex
	connections map[*WSConnection]bool
	logger      *slog.Logger
}

// NewWSHub constructs a new WebSocket connection hub.
func NewWSHub(logger *slog.Logger) *WSHub {
	return &WSHub{
		connections: make(map[*WSConnection]bool),
		logger:      logger,
	}
}

// HandleWebSocket upgrades HTTP to WebSocket and manages the connection lifecycle.
func (hub *WSHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	wsConn := &WSConnection{
		conn:   conn,
		send:   make(chan []byte, 256),
		logger: hub.logger,
	}

	hub.register(wsConn)
	defer hub.unregister(wsConn)

	// Configure connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Writer goroutine
	go wsConn.writePump(ctx)

	// Reader loop (blocks until disconnect)
	wsConn.readPump(hub)
}

func (ws *WSConnection) readPump(hub *WSHub) {
	defer ws.conn.Close()

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				ws.logger.Warn("websocket read error", "error", err)
			}
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			ws.logger.Warn("invalid websocket message", "error", err)
			continue
		}

		hub.handleMessage(ws, msg)
	}
}

func (ws *WSConnection) writePump(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-ws.send:
			if !ok {
				ws.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			ws.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (hub *WSHub) register(conn *WSConnection) {
	hub.mu.Lock()
	hub.connections[conn] = true
	hub.mu.Unlock()
}

func (hub *WSHub) unregister(conn *WSConnection) {
	hub.mu.Lock()
	if _, ok := hub.connections[conn]; ok {
		delete(hub.connections, conn)
		close(conn.send)
	}
	hub.mu.Unlock()
}

func (hub *WSHub) handleMessage(ws *WSConnection, msg WSMessage) {
	// Dispatch based on message type
	switch msg.Type {
	case "ping":
		data, _ := json.Marshal(WSMessage{Type: "pong"})
		ws.send <- data
	default:
		hub.logger.Debug("unhandled websocket message type", "type", msg.Type)
	}
}

// Broadcast sends a message to all connected clients.
func (hub *WSHub) Broadcast(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for conn := range hub.connections {
		select {
		case conn.send <- data:
		default:
			// Buffer full, skip this connection
		}
	}
}
```

## Rules
1. **Ping/Pong Heartbeat**: Send pings every 30s. If pong not received within 60s, connection is dead. Without this, half-open TCP connections leak.
2. **Buffered Send Channel**: Use buffered channel (256) for write queue. If buffer is full, drop the message rather than blocking the broadcaster.
3. **CheckOrigin**: Currently permissive. MUST be restricted to known origins in production deployment.

## Pitfalls

### Pitfall 1: Writing from multiple goroutines
```go
// WRONG:
go func() { conn.WriteMessage(websocket.TextMessage, data1) }()
go func() { conn.WriteMessage(websocket.TextMessage, data2) }() // RACE CONDITION!

// CORRECT:
conn.send <- data // Single writer goroutine reads from channel
```
WebSocket connections are NOT thread-safe for concurrent writes. Use a single writer goroutine consuming from a channel.

## Verify
```bash
go build ./kernel/gateway/...
```

## Checklist
- [ ] File `kernel/gateway/websocket.go` exists
- [ ] WebSocket upgrader with configurable origin check
- [ ] Ping/pong heartbeat (30s ping, 60s read deadline)
- [ ] Connection hub with register/unregister
- [ ] Broadcast to all connected clients
- [ ] Single writer goroutine pattern
- [ ] `go build ./kernel/gateway/...` passes
