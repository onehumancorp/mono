package orchestration

import (
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/sip"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestSIPDB_Chaos simulates high-concurrency ingestion and a simulated DB lock
// to verify the exponential backoff retry logic in withRetry.
func TestSIPDB_Chaos(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "chaos.db") + "?_journal_mode=WAL&_busy_timeout=15000&_txlock=immediate"

	db, err := sip.NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create sip.SIPDB: %v", err)
	}
	db.DB().SetMaxOpenConns(1)
	defer db.Close()

	ctx := context.Background()

	// 1. High-concurrency agent mission ingestion (Stress Test)
	var wg sync.WaitGroup
	numAgents := 50
	missionsPerAgent := 10

	errs := make(chan error, numAgents*missionsPerAgent)

	start := time.Now()
	for i := 0; i < numAgents; i++ {
		wg.Add(1)
		go func(agentIdx int) {
			defer wg.Done()
			for j := 0; j < missionsPerAgent; j++ {
				missionID := fmt.Sprintf("mission-%d-%d", agentIdx, j)
				task := domain.Message{
					ID:      missionID,
					Content: "Stress test task",
					Type:    domain.EventTask,
				}
				if err := db.DelegateMission(ctx, missionID, "SOFTWARE_ENGINEER", task); err != nil {
					errs <- fmt.Errorf("agent %d failed to delegate mission %d: %v", agentIdx, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Concurrency error: %v", err)
	}

	t.Logf("Ingested %d missions concurrently in %v", numAgents*missionsPerAgent, time.Since(start))

	// 2. Controlled failure (DB Lock simulation)
	// We will simulate a locked table by starting an exclusive transaction,
	// then we'll try to write to it from another goroutine which should trigger retries.

	// Open a raw connection to lock the database
	tx, err := db.DB().Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create an exclusive lock
	_, err = tx.Exec("BEGIN EXCLUSIVE")
	if err != nil {
		t.Logf("Expected or not: %v", err)
	} else {
		_, err = tx.Exec("UPDATE agent_missions SET status = 'LOCKED' WHERE 1=0")
		if err != nil {
			t.Fatalf("Failed to lock table: %v", err)
		}
	}

	var retryWg sync.WaitGroup
	retryWg.Add(1)

	startChaos := time.Now()

	// This should retry in the background
	go func() {
		defer retryWg.Done()
		task := domain.Message{
			ID:      "chaos-mission-1",
			Content: "Chaos test task",
			Type:    domain.EventTask,
		}

		// This will block and retry while the DB is locked
		err := db.DelegateMission(ctx, "chaos-mission-1", "SOFTWARE_ENGINEER", task)
		if err != nil {
			// It might ultimately fail if it exhausts retries before we unlock
			t.Logf("Mission delegation after chaos: %v", err)
		} else {
			t.Logf("Mission delegation succeeded after %v", time.Since(startChaos))
		}
	}()

	// Hold the lock for a short duration to trigger retries
	time.Sleep(200 * time.Millisecond)

	// Release the lock
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit and release lock: %v", err)
	}

	// Wait for the background retry to complete
	retryWg.Wait()

	// Verify the mission was actually added
	missions, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("Failed to get pending missions: %v", err)
	}

	found := false
	for _, m := range missions {
		if m.ID == "chaos-mission-1" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find chaos-mission-1 after recovery, but did not. It may have exhausted retries.")
	} else {
		t.Log("Successfully verified mission ingestion after DB lock recovery")
	}
}
