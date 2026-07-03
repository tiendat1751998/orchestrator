# Micro-Task 6.16: Create modules/mission/store_sqlite.go

## Info
- **File**: `modules/mission/store_sqlite.go`
- **Package**: `mission`
- **Depends on**: 6.15
- **Time**: 30 min
- **Verify**: `go build ./modules/mission/...`

## External dependencies
```bash
go get modernc.org/sqlite@latest
```
Using `modernc.org/sqlite` (pure Go) instead of `mattn/go-sqlite3` (CGO) for cross-platform compilation.

## Purpose
Implements SQLite-backed `Store` interface with WAL mode, auto-migration, and parameterized queries.

## EXACT code to create

```go
package mission

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite. Thread-safe via WAL mode.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates a SQLite database at the given data directory.
func NewSQLiteStore(dataDir string) (*SQLiteStore, error) {
	dbPath := filepath.Join(dataDir, "missions.db")

	// WAL mode + busy timeout for concurrent access
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("mission/sqlite: failed to open database: %w", err)
	}

	// Connection pool settings for SQLite
	db.SetMaxOpenConns(1) // SQLite handles one writer at a time
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Keep connection alive

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("mission/sqlite: migration failed: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS missions (
		id          TEXT PRIMARY KEY,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT 'pending',
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		duration_ms INTEGER NOT NULL DEFAULT 0,
		total_tokens INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS mission_tasks (
		id         TEXT PRIMARY KEY,
		mission_id TEXT NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
		name       TEXT NOT NULL,
		agent      TEXT NOT NULL DEFAULT '',
		status     TEXT NOT NULL DEFAULT 'pending',
		duration_ms INTEGER NOT NULL DEFAULT 0,
		output     TEXT NOT NULL DEFAULT '',
		error_msg  TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_mission_tasks_mission_id ON mission_tasks(mission_id);
	CREATE INDEX IF NOT EXISTS idx_missions_created_at ON missions(created_at DESC);`

	_, err := s.db.Exec(schema)
	return err
}

// Create inserts a new mission record.
func (s *SQLiteStore) Create(record *MissionRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO missions (id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		record.ID, record.Title, record.Description, record.Status, record.CreatedAt, record.UpdatedAt,
	)
	return err
}

// Get retrieves a mission by ID with its task records.
func (s *SQLiteStore) Get(id string) (*MissionRecord, error) {
	row := s.db.QueryRow(
		`SELECT id, title, description, status, created_at, updated_at, duration_ms, total_tokens FROM missions WHERE id = ?`, id,
	)

	var r MissionRecord
	var durationMs int64
	if err := row.Scan(&r.ID, &r.Title, &r.Description, &r.Status, &r.CreatedAt, &r.UpdatedAt, &durationMs, &r.TotalTokens); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mission %q not found", id)
		}
		return nil, err
	}
	r.Duration = time.Duration(durationMs) * time.Millisecond

	// Load associated tasks
	tasks, err := s.loadTasks(id)
	if err != nil {
		return nil, err
	}
	r.Tasks = tasks

	return &r, nil
}

func (s *SQLiteStore) loadTasks(missionID string) ([]TaskRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, name, agent, status, duration_ms, output, error_msg FROM mission_tasks WHERE mission_id = ? ORDER BY created_at`,
		missionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskRecord
	for rows.Next() {
		var t TaskRecord
		var durationMs int64
		if err := rows.Scan(&t.ID, &t.Name, &t.Agent, &t.Status, &durationMs, &t.Output, &t.Error); err != nil {
			return nil, err
		}
		t.Duration = time.Duration(durationMs) * time.Millisecond
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// List returns all missions ordered by creation time (newest first).
func (s *SQLiteStore) List() ([]*MissionRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, title, description, status, created_at, updated_at, duration_ms, total_tokens FROM missions ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var missions []*MissionRecord
	for rows.Next() {
		var r MissionRecord
		var durationMs int64
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.Status, &r.CreatedAt, &r.UpdatedAt, &durationMs, &r.TotalTokens); err != nil {
			return nil, err
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
		missions = append(missions, &r)
	}
	return missions, rows.Err()
}

// UpdateStatus updates a mission's status.
func (s *SQLiteStore) UpdateStatus(id string, status string) error {
	_, err := s.db.Exec(
		`UPDATE missions SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	return err
}

// UpdateTokens updates a mission's total token count.
func (s *SQLiteStore) UpdateTokens(id string, tokens int) error {
	_, err := s.db.Exec(
		`UPDATE missions SET total_tokens = ?, updated_at = ? WHERE id = ?`,
		tokens, time.Now(), id,
	)
	return err
}

// AddTask inserts a task record associated with a mission.
func (s *SQLiteStore) AddTask(missionID string, task TaskRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO mission_tasks (id, mission_id, name, agent, status, duration_ms, output, error_msg) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, missionID, task.Name, task.Agent, task.Status, task.Duration.Milliseconds(), task.Output, task.Error,
	)
	return err
}

// Delete removes a mission and cascading tasks.
func (s *SQLiteStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM missions WHERE id = ?`, id)
	return err
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
```

## Rules
1. **WAL Mode**: MUST use `_journal_mode=WAL` for concurrent read access. Default rollback journal blocks readers during writes.
2. **MaxOpenConns = 1**: SQLite only supports one writer. Setting `MaxOpenConns > 1` causes `SQLITE_BUSY` errors.
3. **Duration as Milliseconds**: Store `time.Duration` as integer milliseconds in SQLite. SQLite has no native duration type.
4. **Pure Go Driver**: Use `modernc.org/sqlite` — no CGO dependency, compiles on all platforms without C toolchain.

## Pitfalls

### Pitfall 1: Concurrent writes with MaxOpenConns > 1
```go
// WRONG:
db.SetMaxOpenConns(10) // SQLite can't handle concurrent writes → SQLITE_BUSY panics

// CORRECT:
db.SetMaxOpenConns(1) // Single writer, WAL allows concurrent readers
```

### Pitfall 2: Forgetting foreign key enforcement
SQLite disables foreign keys by default. Always include `_foreign_keys=ON` in the DSN string.

## Verify
```bash
go build ./modules/mission/...
```

## Checklist
- [ ] File `modules/mission/store_sqlite.go` exists
- [ ] WAL mode enabled in DSN
- [ ] Auto-migration creates tables on first run
- [ ] Foreign key cascading delete for tasks
- [ ] Duration stored as milliseconds
- [ ] Pure Go SQLite driver (no CGO)
- [ ] `go build ./modules/mission/...` passes
