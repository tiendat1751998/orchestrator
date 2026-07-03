package resilience_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/tiendat1751998/orchestrator/kernel/resilience"
)

func TestCheckpointStore_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recovery-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := resilience.NewCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create checkpoint store: %v", err)
	}

	// 1. Load missing checkpoint -> returns nil, nil
	cp, err := store.Load("missing-mission")
	if err != nil {
		t.Fatalf("expected no error on loading missing, got %v", err)
	}
	if cp != nil {
		t.Fatalf("expected nil checkpoint, got %v", cp)
	}

	// 2. Save valid checkpoint
	expected := &resilience.Checkpoint{
		MissionID: "mission-1",
		State:     "running",
		Results: map[string]string{
			"step-1": "success",
		},
	}
	err = store.Save(expected)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// 3. Load checkpoint
	cp, err = store.Load("mission-1")
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}
	if cp == nil {
		t.Fatal("expected loaded checkpoint, got nil")
	}

	if cp.MissionID != expected.MissionID {
		t.Errorf("expected MissionID %q, got %q", expected.MissionID, cp.MissionID)
	}
	if cp.State != expected.State {
		t.Errorf("expected State %q, got %q", expected.State, cp.State)
	}
	if cp.Results["step-1"] != "success" {
		t.Errorf("expected results step-1 to be success, got %v", cp.Results["step-1"])
	}
}

func TestCheckpointStore_InvalidParams(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recovery-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := resilience.NewCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create checkpoint store: %v", err)
	}

	// Save nil checkpoint
	err = store.Save(nil)
	if err == nil {
		t.Error("expected error saving nil checkpoint")
	}

	// Save empty MissionID checkpoint
	err = store.Save(&resilience.Checkpoint{MissionID: ""})
	if err == nil {
		t.Error("expected error saving checkpoint with empty MissionID")
	}

	// Load empty mission ID
	_, err = store.Load("")
	if err == nil {
		t.Error("expected error loading with empty mission ID")
	}
}

func TestCheckpointStore_Concurrency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recovery-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := resilience.NewCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create checkpoint store: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10
	iterations := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cp := &resilience.Checkpoint{
					MissionID: "concurrent-mission",
					State:     "running",
					Results: map[string]string{
						"worker": string(rune(workerID)),
					},
				}
				_ = store.Save(cp)
				_, _ = store.Load("concurrent-mission")
			}
		}(i)
	}

	wg.Wait()
}

func TestCheckpointStore_NewCheckpointStore_InvalidDir(t *testing.T) {
	// Trying to create a store inside a path that is actually a file should fail directory creation.
	tmpFile, err := os.CreateTemp("", "recovery-file-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Use the file name as directory path (MkdirAll should fail)
	_, err = resilience.NewCheckpointStore(filepath.Join(tmpFile.Name(), "subdir"))
	if err == nil {
		t.Error("expected error when creating store in an invalid path")
	}
}
