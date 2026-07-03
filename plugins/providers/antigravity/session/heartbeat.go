package session

import (
	"context"
	"log/slog"
	"time"
)

// StartHeartbeat initiates a background goroutine that pings the session's CLI adapter
// at regular intervals to verify process health.
func (sm *SessionManager) StartHeartbeat(ctx context.Context, sessionID string, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		interval = 60 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sm.mu.RLock()
				s, exists := sm.sessions[sessionID]
				sm.mu.RUnlock()

				if !exists {
					return // Session was closed externally: stop heartbeat
				}

				// Check process health
				_, stdout, _, err := s.Adapter.Pipes()
				if err != nil {
					// Process is not running or pipes are closed
					handleCrashedSession(ctx, sm, s, logger)
					continue
				}

				// Check if process has exited
				if stdout == nil {
					handleCrashedSession(ctx, sm, s, logger)
				}
			}
		}
	}()
}

func handleCrashedSession(ctx context.Context, sm *SessionManager, s *Session, logger *slog.Logger) {
	if logger != nil {
		logger.Warn("session process crashed or disconnected; attempting restart", "session_id", s.ID)
	}

	// Attempt process restart
	_ = s.Adapter.Stop()

	if err := s.Adapter.Start(ctx); err != nil {
		if logger != nil {
			logger.Error("failed to restart session process", "session_id", s.ID, "error", err.Error())
		}
		// Remove failed session from pool
		sm.mu.Lock()
		delete(sm.sessions, s.ID)
		sm.mu.Unlock()
	} else {
		if logger != nil {
			logger.Info("session process restarted successfully", "session_id", s.ID)
		}
		s.LastUsed = time.Now()
	}
}
