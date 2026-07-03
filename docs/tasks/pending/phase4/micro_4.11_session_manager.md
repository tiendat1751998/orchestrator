# Micro-Task 4.11: Create plugins/providers/antigravity/session/manager.go

## Info
- **File**: `plugins/providers/antigravity/session/manager.go`
- **Package**: `session`
- **Depends on**: 4.03 (CLI process manager)
- **Time**: 25 min
- **Verify**: `go build ./plugins/providers/antigravity/session/...`

## Purpose
Implements the session connection pool (`SessionManager` and `Session` structures) to manage multiple concurrent Antigravity CLI processes and clean up inactive sessions.

## EXACT code to create

```go
// Package session manages pools of CLI adapter connections to model providers.
package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/adapter"
)

// Session tracks the state of a single active CLI process.
type Session struct {
	ID       string
	Adapter  *adapter.CLIAdapter
	Created  time.Time
	LastUsed time.Time
}

// SessionManager pools active sessions and enforces resource limits.
// Thread-safe.
type SessionManager struct {
	mu           sync.RWMutex
	sessions     map[string]*Session
	binary       string
	maxSessions  int
	idleTimeout  time.Duration
	cleanupChan  chan struct{}
	cleanupWg    sync.WaitGroup
}

// NewSessionManager constructs a new SessionManager.
func NewSessionManager(binary string, maxSessions int, idleTimeout time.Duration) *SessionManager {
	if maxSessions <= 0 {
		maxSessions = 5 // Reasonable concurrent CLI process limit
	}
	if idleTimeout <= 0 {
		idleTimeout = 5 * time.Minute
	}

	sm := &SessionManager{
		sessions:    make(map[string]*Session),
		binary:      binary,
		maxSessions: maxSessions,
		idleTimeout: idleTimeout,
		cleanupChan: make(chan struct{}),
	}

	// Start background cleanup daemon
	sm.cleanupWg.Add(1)
	go sm.cleanupLoop()

	return sm
}

// GetOrCreate retrieves an existing session or spawns a new CLI process.
func (sm *SessionManager) GetOrCreate(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session: session ID cannot be empty")
	}

	sm.mu.Lock()
	// 1. Retrieve existing session
	if s, exists := sm.sessions[sessionID]; exists {
		s.LastUsed = time.Now()
		sm.mu.Unlock()
		return s, nil
	}

	// 2. Enforce session pool capacity limit
	if len(sm.sessions) >= sm.maxSessions {
		sm.mu.Unlock()
		return nil, fmt.Errorf("session: maximum concurrent sessions reached (%d)", sm.maxSessions)
	}

	// 3. Construct new CLI session
	cliAdapter := adapter.NewCLIAdapter(sm.binary)
	if err := cliAdapter.Start(ctx); err != nil {
		sm.mu.Unlock()
		return nil, fmt.Errorf("session: failed to start CLI adapter: %w", err)
	}

	s := &Session{
		ID:       sessionID,
		Adapter:  cliAdapter,
		Created:  time.Now(),
		LastUsed: time.Now(),
	}
	sm.sessions[sessionID] = s
	sm.mu.Unlock()

	return s, nil
}

// Close closes a specific session and terminates its CLI process.
func (sm *SessionManager) Close(sessionID string) error {
	sm.mu.Lock()
	s, exists := sm.sessions[sessionID]
	if !exists {
		sm.mu.Unlock()
		return nil
	}
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	return s.Adapter.Stop()
}

// Stop halts all active CLI processes and shuts down the cleanup daemon.
func (sm *SessionManager) Stop() error {
	close(sm.cleanupChan)
	sm.cleanupWg.Wait()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	var lastErr error
	for id, s := range sm.sessions {
		if err := s.Adapter.Stop(); err != nil {
			lastErr = err
		}
		delete(sm.sessions, id)
	}

	return lastErr
}

// cleanupLoop checks periodically and purges sessions that exceed idle limits.
func (sm *SessionManager) cleanupLoop() {
	defer sm.cleanupWg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.cleanupChan:
			return
		case <-ticker.C:
			sm.mu.Lock()
			now := time.Now()
			var expired []*Session

			for id, s := range sm.sessions {
				if now.Sub(s.LastUsed) > sm.idleTimeout {
					expired = append(expired, s)
					delete(sm.sessions, id)
				}
			}
			sm.mu.Unlock()

			// Terminate processes outside lock to prevent blocking updates
			for _, s := range expired {
				_ = s.Adapter.Stop()
			}
		}
	}
}
```

## Pitfalls

### Pitfall 1: Resource leaks from zombie CLI processes
Failing to implement background cleanup allows unused sessions to persist indefinitely. Since each session spawns a CLI process, this can quickly exhaust system memory and CPU resources. Use background scanners.

### Pitfall 2: Blocking requests during session termination
Calling `s.Adapter.Stop()` while holding the manager lock blocks all incoming session creation requests. Always unlock the manager before terminating processes.

## Verify
```bash
go build ./plugins/providers/antigravity/session/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/session/manager.go`
- [ ] Package name is `session`
- [ ] All exported types have Godoc
- [ ] `SessionManager` limits maximum concurrent sessions
- [ ] Spawns a background cleanup goroutine
- [ ] Build command passes
