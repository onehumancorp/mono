package orchestration

import (
	"context"
	"testing"
	"time"
)

func TestSIPDB_PruneStaleMissions(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now()

	// Insert some test missions
	// 1. Pending mission (should not be pruned)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, updated_at) VALUES ('1', 'ROLE', 'task1', 'PENDING', ?)", now.Add(-2*time.Hour))
	if err != nil { t.Fatalf("failed to insert: %v", err) }

	// 2. Completed mission, recently updated (should not be pruned)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, updated_at) VALUES ('2', 'ROLE', 'task2', 'COMPLETED', ?)", now.Add(-30*time.Minute))
	if err != nil { t.Fatalf("failed to insert: %v", err) }

	// 3. Completed mission, stale (should be pruned)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, updated_at) VALUES ('3', 'ROLE', 'task3', 'COMPLETED', ?)", now.Add(-2*time.Hour))
	if err != nil { t.Fatalf("failed to insert: %v", err) }

	// 4. Completed mission, very stale (should be pruned)
	_, err = db.db.ExecContext(ctx, "INSERT INTO agent_missions (id, role, task, status, updated_at) VALUES ('4', 'ROLE', 'task4', 'COMPLETED', ?)", now.Add(-24*time.Hour))
	if err != nil { t.Fatalf("failed to insert: %v", err) }

	affected, err := db.PruneStaleMissions(ctx, 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if affected != 2 {
		t.Fatalf("expected 2 missions to be pruned, got %d", affected)
	}

	// Verify remaining rows
	var count int
	err = db.db.QueryRowContext(ctx, "SELECT count(*) FROM agent_missions").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 remaining rows, got %d", count)
	}
}

func TestSIPDB_PruneStaleMissions_DBError(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	db.Close()

	_, err = db.PruneStaleMissions(context.Background(), 1*time.Hour)
	if err == nil {
		t.Fatal("Expected error when pruning on closed DB")
	}
}
