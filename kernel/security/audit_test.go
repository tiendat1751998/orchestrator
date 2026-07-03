package security

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAuditLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	// 1. Construct logger
	logger, err := NewAuditLogger(logPath)
	if err != nil {
		t.Fatalf("failed to construct AuditLogger: %v", err)
	}

	// 2. Perform concurrent logs
	const numGoroutines = 10
	const logsPerGoroutine = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				err := logger.Log("test-agent", "run", "cmd", true, "success detail")
				if err != nil {
					t.Errorf("failed to log: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 3. Close logger
	err = logger.Close()
	if err != nil {
		t.Fatalf("failed to close AuditLogger: %v", err)
	}

	// 4. Verify log file content and line count
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open generated log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var entry AuditEntry
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON line %d: %v", lineCount, err)
		}

		if entry.Agent != "test-agent" {
			t.Errorf("unexpected agent: got %q, want %q", entry.Agent, "test-agent")
		}
		if entry.Action != "run" {
			t.Errorf("unexpected action: got %q, want %q", entry.Action, "run")
		}
		if entry.Target != "cmd" {
			t.Errorf("unexpected target: got %q, want %q", entry.Target, "cmd")
		}
		if !entry.Allowed {
			t.Errorf("unexpected allowed flag: got %v, want true", entry.Allowed)
		}
		if entry.Details != "success detail" {
			t.Errorf("unexpected details: got %q, want %q", entry.Details, "success detail")
		}
		if entry.Timestamp.IsZero() {
			t.Errorf("timestamp was not set")
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("error during scan: %v", err)
	}

	expectedLogs := numGoroutines * logsPerGoroutine
	if lineCount != expectedLogs {
		t.Errorf("got %d lines in log file, expected %d", lineCount, expectedLogs)
	}
}

func TestAuditLogger_InvalidPath(t *testing.T) {
	// Attempt to create logger in a non-existent directory without permissions
	_, err := NewAuditLogger("/nonexistent-dir/audit.log")
	if err == nil {
		t.Error("expected error when creating AuditLogger with invalid path")
	}
}
