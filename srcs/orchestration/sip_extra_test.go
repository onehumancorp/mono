package orchestration

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestSIPDB_Init_Error(t *testing.T) {
	_, err := NewSIPDB("file::memory:?mode=ro")
	if err == nil {
		t.Fatal("expected error for read-only db, got nil")
	}
}

func TestWithRetry_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := withRetry(ctx, func() error {
		return errors.New("transient error")
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWithRetry_RetryLogic(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	err := withRetry(ctx, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWithRetry_MaxRetries(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	expectedErr := errors.New("persistent error")
	err := withRetry(ctx, func() error {
		attempts++
		return expectedErr
	})

	if err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if attempts != maxRetries {
		t.Fatalf("expected %d attempts, got %d", maxRetries, attempts)
	}
}

func TestInitializeTables_Error(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.Close()

	err = initializeTables(db)
	if err == nil {
		t.Fatal("expected error from initializeTables, got nil")
	}
}

func TestSyncMemory_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	_, err = db.SyncMemory(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error from SyncMemory, got nil")
	}
}

func TestUpdateMemory_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	err = db.UpdateMemory(context.Background(), "test", "val")
	if err == nil {
		t.Fatal("expected error from UpdateMemory, got nil")
	}
}

func TestGetPendingMissions_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	_, err = db.GetPendingMissions(context.Background(), "role")
	if err == nil {
		t.Fatal("expected error from GetPendingMissions, got nil")
	}
}

func TestCompleteMission_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	err = db.CompleteMission(context.Background(), "m1")
	if err == nil {
		t.Fatal("expected error from CompleteMission, got nil")
	}
}

func TestCompleteMission_NotFound(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	err = db.CompleteMission(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent mission, got nil")
	}
}

func TestHeartbeat_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	err = db.Heartbeat(context.Background(), "agent1", "role1", "status1")
	if err == nil {
		t.Fatal("expected error from Heartbeat, got nil")
	}
}

func TestDelegateMission_Error(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	db.db.Close()

	err = db.DelegateMission(context.Background(), "m1", "role1", Message{})
	if err == nil {
		t.Fatal("expected error from DelegateMission, got nil")
	}
}

func TestWithRetry_ContextCanceledDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	err := withRetry(ctx, func() error {
		attempts++
		if attempts == 1 {
			cancel()
			return errors.New("transient error")
		}
		return nil
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestGetPendingMissions_UnmarshalFallback(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	db.db.Exec("INSERT INTO agent_missions (id, role, task, status) VALUES ('id1', 'r1', 'invalid json', 'PENDING');")

	missions, err := db.GetPendingMissions(context.Background(), "r1")
	if err != nil {
		t.Fatalf("expected no error with fallback, got %v", err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 mission, got %d", len(missions))
	}
	if missions[0].Content != "invalid json" {
		t.Fatalf("expected content 'invalid json', got %v", missions[0].Content)
	}
}

func TestGetPendingMissions_EmptyIDFallback(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	db.db.Exec("INSERT INTO agent_missions (id, role, task, status) VALUES ('id1', 'r1', '{\"content\": \"hi\"}', 'PENDING');")

	missions, err := db.GetPendingMissions(context.Background(), "r1")
	if err != nil {
		t.Fatalf("expected no error with fallback, got %v", err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 mission, got %d", len(missions))
	}
	if missions[0].ID != "id1" {
		t.Fatalf("expected ID 'id1', got %v", missions[0].ID)
	}
}

func TestSyncMemory_ErrNoRows(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	val, err := db.SyncMemory(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for ErrNoRows, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %v", val)
	}
}

func TestWithRetry_RetriesAndContextDone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), retryInterval*2)
	defer cancel()

	err := withRetry(ctx, func() error {
		return errors.New("always fail")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWithRetry_CancelInsideOp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := withRetry(ctx, func() error {
		cancel()
		return errors.New("fail")
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWithRetry_FullCoverage(t *testing.T) {
	err := withRetry(context.Background(), func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = withRetry(ctx, func() error {
		return errors.New("fail")
	})
	if err != context.Canceled {
		t.Fatalf("expected canceled error, got %v", err)
	}
}

func TestWithRetry_ContextCanceledConcurrently(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	err := withRetry(ctx, func() error {
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		time.Sleep(20 * time.Millisecond)
		return errors.New("fail")
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWithRetry_MaxRetriesReached(t *testing.T) {
	err := withRetry(context.Background(), func() error {
		return errors.New("persistent error")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
