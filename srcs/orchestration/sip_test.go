package orchestration

import (
	"context"
	"testing"
	"path/filepath"
	"time"
)

func TestSIPDB_Init(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize SIPDB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test Memory
	err = db.UpdateMemory(ctx, "architecture", "microservices")
	if err != nil {
		t.Fatalf("UpdateMemory failed: %v", err)
	}

	val, err := db.SyncMemory(ctx, "architecture")
	if err != nil {
		t.Fatalf("SyncMemory failed: %v", err)
	}
	if val != "microservices" {
		t.Fatalf("expected 'microservices', got '%s'", val)
	}

	// Test Heartbeat
	err = db.Heartbeat(ctx, "agent-1", "SOFTWARE_ENGINEER", "ACTIVE")
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}

	// Test Delegation & Mission
	msg := Message{ID: "m1", Content: "Build a feature", Type: EventTask}
	err = db.DelegateMission(ctx, "m1", "SOFTWARE_ENGINEER", msg)
	if err != nil {
		t.Fatalf("DelegateMission failed: %v", err)
	}

	missions, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("GetPendingMissions failed: %v", err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 mission, got %d", len(missions))
	}
	if missions[0].ID != "m1" {
		t.Fatalf("expected mission ID 'm1', got '%s'", missions[0].ID)
	}

	// Test Completion
	err = db.CompleteMission(ctx, "m1")
	if err != nil {
		t.Fatalf("CompleteMission failed: %v", err)
	}

	missions, err = db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("GetPendingMissions failed: %v", err)
	}
	if len(missions) != 0 {
		t.Fatalf("expected 0 missions, got %d", len(missions))
	}
}

func TestSIPDB_NewSIPDB_Fail(t *testing.T) {
	// Attempt to create a database on a read-only directory to trigger an error.
	// We'll just provide a path we know will fail SQLite open.
	_, err := NewSIPDB("/root/illegal/path/db.sqlite")
	if err == nil {
		t.Fatal("Expected error when opening DB in illegal path")
	}
}

func TestSIPDB_PollMissions_ScanError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Insert invalid schema data manually to cause a row scan error.
	// Since we can't easily break the type in sqlite (it's dynamically typed),
	// this one is hard to hit purely through SQLite without mocking the DB connection.
	// Instead, we focus on the Unmarshal error line 150-151.

	// Manually insert bad JSON
	ctx := context.Background()
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status) VALUES ('123', 'SOFTWARE_ENGINEER', 'invalid-json', 'PENDING')")
	if err != nil {
		t.Fatalf("Failed to insert bad json: %v", err)
	}

	missions, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("Expected fallback to message string on JSON unmarshal error, got error: %v", err)
	}

	if len(missions) != 1 {
		t.Fatalf("Expected 1 mission, got %d", len(missions))
	}

	if missions[0].Content != "invalid-json" {
		t.Fatalf("Expected content 'invalid-json', got %s", missions[0].Content)
	}
}

func TestSIPDB_CompleteMission_RowsAffectedError(t *testing.T) {
	// We can't easily trigger RowsAffected() error with go-sqlite3 normally,
	// but let's at least test the "mission not found" path.
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	err = db.CompleteMission(context.Background(), "non-existent")
	if err == nil || err.Error() != "mission not found" {
		t.Fatalf("Expected 'mission not found' error, got %v", err)
	}
}

func TestSIPDB_GetPendingMissions_BadData(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status) VALUES ('123', 'SOFTWARE_ENGINEER', 'invalid-json', 'PENDING')")
	if err != nil {
		t.Fatalf("Failed to insert bad json: %v", err)
	}

	// Ensure we handle invalid JSON in GetPendingMissions without blowing up completely
	missions, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(missions) != 1 {
		t.Fatalf("Expected 1 mission, got %d", len(missions))
	}

	if missions[0].Content != "invalid-json" {
		t.Fatalf("Expected content to be 'invalid-json' fallback, got: %s", missions[0].Content)
	}
}

func TestSIPDB_CompleteMission_ExecError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// drop table to cause error
	_, err = db.db.ExecContext(context.Background(), "DROP TABLE agent_missions")
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	err = db.CompleteMission(context.Background(), "some-id")
	if err == nil {
		t.Fatal("Expected error updating missing table")
	}
}

func TestSIPDB_PruneStaleMissions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Insert missions:
	// 1. Pending and new (should not be deleted)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, created_at) VALUES ('1', 'ROLE', 'task', 'PENDING', datetime('now'))")
	if err != nil { t.Fatal(err) }

	// 2. Completed but new (should NOT be deleted)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, created_at) VALUES ('2', 'ROLE', 'task', 'COMPLETED', datetime('now'))")
	if err != nil { t.Fatal(err) }

	// 3. Pending but old (should NOT be deleted, only completed ones get deleted)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, created_at) VALUES ('3', 'ROLE', 'task', 'PENDING', datetime('now', '-2 days'))")
	if err != nil { t.Fatal(err) }

	// 4. Completed and old (should be deleted)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, created_at) VALUES ('4', 'ROLE', 'task', 'COMPLETED', datetime('now', '-2 days'))")
	if err != nil { t.Fatal(err) }

	// Prune missions older than 24 hours
	err = db.PruneStaleMissions(ctx, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to prune stale missions: %v", err)
	}

	var count int
	err = db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM agent_missions").Scan(&count)
	if err != nil { t.Fatal(err) }

	if count != 3 {
		t.Fatalf("Expected 3 missions remaining, got %d", count)
	}

	// Verify the remaining missions are correct
	rows, err := db.db.QueryContext(ctx, "SELECT id FROM agent_missions ORDER BY id ASC")
	if err != nil { t.Fatal(err) }
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil { t.Fatal(err) }
		ids = append(ids, id)
	}

	if len(ids) != 3 || ids[0] != "1" || ids[1] != "2" || ids[2] != "3" {
		t.Fatalf("Expected remaining missions to be [1, 2, 3], got %v", ids)
	}
}

func TestSIPDB_PruneStaleMissions_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	err = db.PruneStaleMissions(context.Background(), 24*time.Hour)
	if err == nil {
		t.Fatal("Expected error when pruning on closed DB")
	}
}

func TestSIPDB_CompleteMission_ExecErrorAgain(t *testing.T) {
	// Let's create a test that calls CompleteMission on a closed DB
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	err = db.CompleteMission(context.Background(), "some-id")
	if err == nil {
		t.Fatal("Expected error updating on closed DB")
	}
}

func TestSIPDB_GetPendingMissions_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	_, err = db.GetPendingMissions(context.Background(), "role")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}

func TestSIPDB_UpdateMemory_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	err = db.UpdateMemory(context.Background(), "key", "val")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}

func TestSIPDB_SyncMemory_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	_, err = db.SyncMemory(context.Background(), "key")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}

func TestSIPDB_Heartbeat_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	err = db.Heartbeat(context.Background(), "agent", "role", "status")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}

func TestSIPDB_DelegateMission_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close()

	err = db.DelegateMission(context.Background(), "mission", "role", Message{})
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}


func TestSIPDB_InitTables_InvalidDBDir(t *testing.T) {
	dbPath := t.TempDir()
	_, err := NewSIPDB(dbPath)
	if err == nil {
		t.Fatal("Expected error initializing tables when path is a directory")
	}
}
