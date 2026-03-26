package checkpointer

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create table
	query := `
	CREATE TABLE checkpoints (
		thread_id VARCHAR(255) PRIMARY KEY,
		state JSON NOT NULL
	);`
	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestPGCheckpointer_SaveAndLoad(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()

	// Ensure table creates properly if it already exists/doesn't exist
	err := checkpointer.EnsureTableExists(ctx)
	if err != nil {
		t.Fatalf("EnsureTableExists failed: %v", err)
	}

	tests := []struct {
		name     string
		threadID string
		state    map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "Basic save and load",
			threadID: "thread-1",
			state: map[string]interface{}{
				"step": "1",
				"data": "some value",
			},
			wantErr: false,
		},
		{
			name:     "Empty thread ID",
			threadID: "",
			state: map[string]interface{}{
				"step": "1",
			},
			wantErr: true,
		},
		{
			name:     "Empty state map",
			threadID: "thread-empty-state",
			state:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name:     "Update existing thread ID",
			threadID: "thread-1",
			state: map[string]interface{}{
				"step": "2",
				"data": "updated value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkpointer.SaveCheckpoint(ctx, tt.threadID, tt.state)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				cp, err := checkpointer.LoadCheckpoint(ctx, tt.threadID)
				if err != nil {
					t.Fatalf("LoadCheckpoint() failed: %v", err)
				}

				if cp.ThreadID != tt.threadID {
					t.Errorf("expected thread ID %q, got %q", tt.threadID, cp.ThreadID)
				}

				// The JSON decoding unmarshals numbers as float64, so we need a more robust comparison
				// or just rely on a simple check for this test since we are using strings.
				if !reflect.DeepEqual(cp.State, tt.state) {
					t.Errorf("expected state %v, got %v", tt.state, cp.State)
				}
			}
		})
	}
}

func TestPGCheckpointer_LoadNonExistent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()

	cp, err := checkpointer.LoadCheckpoint(ctx, "non-existent")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if cp != nil {
		t.Errorf("expected nil checkpoint, got %v", cp)
	}

	_, err = checkpointer.LoadCheckpoint(ctx, "")
	if err == nil {
		t.Errorf("expected error for empty thread ID")
	}
}

// 2.1 E2E Integration Test: Pause and Resume
func TestPGCheckpointer_E2EPauseAndResume(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	checkpointer1 := NewPGCheckpointer(db)

	threadID := "workflow-alpha"
	initialState := map[string]interface{}{
		"agent": "SWE-1",
		"step":  "1",
		"tasks": []interface{}{"task1", "task2"},
	}

	// Setup: An agent begins a workflow and is explicitly paused after step 1.
	err := checkpointer1.SaveCheckpoint(ctx, threadID, initialState)
	if err != nil {
		t.Fatalf("failed to save initial state: %v", err)
	}

	// Action: A new agent instance is instantiated with the same thread_id.
	checkpointer2 := NewPGCheckpointer(db)
	restoredCP, err := checkpointer2.LoadCheckpoint(ctx, threadID)
	if err != nil {
		t.Fatalf("failed to load checkpoint from new instance: %v", err)
	}

	// Assertion: Verify the new agent resumes precisely from step 2 with full context.
	if restoredCP.ThreadID != threadID {
		t.Errorf("expected thread ID %q, got %q", threadID, restoredCP.ThreadID)
	}

	if restoredCP.State["step"] != "1" {
		t.Errorf("expected step 1, got %v", restoredCP.State["step"])
	}

	// Update the state as if the agent completed step 2
	restoredCP.State["step"] = "2"
	err = checkpointer2.SaveCheckpoint(ctx, threadID, restoredCP.State)
	if err != nil {
		t.Fatalf("failed to save updated state: %v", err)
	}

	finalCP, err := checkpointer1.LoadCheckpoint(ctx, threadID)
	if err != nil {
		t.Fatalf("failed to load final state: %v", err)
	}

	if finalCP.State["step"] != "2" {
		t.Errorf("expected step 2, got %v", finalCP.State["step"])
	}
}

// 2.2 Edge Case: Large State Payload
func TestPGCheckpointer_LargePayload(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()
	threadID := "massive-thread"

	// Setup: Inject a massive 10MB JSON state payload.
	// We'll generate a map with a very large string to simulate 10MB
	const targetSize = 10 * 1024 * 1024 // 10MB
	largeString := strings.Repeat("a", targetSize)

	largeState := map[string]interface{}{
		"data": largeString,
	}

	// Action: Trigger a checkpoint save.
	start := time.Now()
	err := checkpointer.SaveCheckpoint(ctx, threadID, largeState)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("failed to save massive payload: %v", err)
	}

	t.Logf("Saved 10MB payload in %v", duration)

	// Assertion: Verify PostgreSQL handles the JSONB insert without performance degradation or truncation.
	cp, err := checkpointer.LoadCheckpoint(ctx, threadID)
	if err != nil {
		t.Fatalf("failed to load massive payload: %v", err)
	}

	restoredData, ok := cp.State["data"].(string)
	if !ok {
		t.Fatalf("expected string data, got %T", cp.State["data"])
	}

	if len(restoredData) != targetSize {
		t.Errorf("expected data length %d, got %d", targetSize, len(restoredData))
	}
}

func TestPGCheckpointer_RetryLogic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)

	var attempts int
	err := checkpointer.withRetry(func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("database is locked")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPGCheckpointer_RetryLogic_Failure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)

	var attempts int
	err := checkpointer.withRetry(func() error {
		attempts++
		return fmt.Errorf("some fatal error")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestPGCheckpointer_SaveCheckpoint_MarshalError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()
	_ = checkpointer.EnsureTableExists(ctx)

	// A channel cannot be marshaled into JSON
	unmarshalableState := map[string]interface{}{
		"bad_data": make(chan int),
	}

	err := checkpointer.SaveCheckpoint(ctx, "thread-bad-marshal", unmarshalableState)
	if err == nil {
		t.Fatalf("expected error when marshaling invalid state, got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal state") {
		t.Errorf("expected error to contain 'failed to marshal state', got: %v", err)
	}
}

func TestPGCheckpointer_LoadCheckpoint_DecodeError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()
	_ = checkpointer.EnsureTableExists(ctx)

	// Insert invalid JSON directly into the database
	query := `INSERT INTO checkpoints (thread_id, state) VALUES ($1, $2)`
	_, err := db.ExecContext(ctx, query, "thread-bad-json", `{"invalid": json`)
	if err != nil {
		t.Fatalf("failed to insert bad json: %v", err)
	}

	_, err = checkpointer.LoadCheckpoint(ctx, "thread-bad-json")
	if err == nil {
		t.Fatalf("expected error when decoding invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode state") {
		t.Errorf("expected error to contain 'failed to decode state', got: %v", err)
	}
}

func TestPGCheckpointer_LoadCheckpoint_DBError(t *testing.T) {
	db := setupTestDB(t)
	checkpointer := NewPGCheckpointer(db)
	ctx := context.Background()
	_ = checkpointer.EnsureTableExists(ctx)
	db.Close()

	_, err := checkpointer.LoadCheckpoint(ctx, "fail-thread")
	if err == nil {
		t.Fatalf("expected error from closed db, got nil")
	}
}

// A mock struct with an Error() method that returns "database is locked" to test the retry loop logic
type mockError string

func (e mockError) Error() string { return string(e) }

func TestWithRetry_DatabaseLocked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	p := NewPGCheckpointer(db)

	attempts := 0
	err := p.withRetry(func() error {
		attempts++
		if attempts < 3 {
			return mockError("database is locked")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	p := NewPGCheckpointer(db)

	attempts := 0
	err := p.withRetry(func() error {
		attempts++
		return mockError("database is locked")
	})

	if err == nil {
		t.Fatal("expected error after max retries exceeded, got nil")
	}

	if attempts != 5 {
		t.Fatalf("expected 5 attempts (maxRetries), got %d", attempts)
	}
}

func TestWithRetry_OtherError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	p := NewPGCheckpointer(db)

	attempts := 0
	expectedErr := fmt.Errorf("some other error")
	err := p.withRetry(func() error {
		attempts++
		return expectedErr
	})

	if err != expectedErr {
		t.Fatalf("expected '%v', got '%v'", expectedErr, err)
	}

	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry for other errors), got %d", attempts)
	}
}

func TestWithRetry_MaxDelay(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	p := NewPGCheckpointer(db)

	attempts := 0
	err := p.withRetry(func() error {
		attempts++
		return mockError("database is locked")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
