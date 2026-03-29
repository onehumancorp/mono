package chaos

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
	_ "modernc.org/sqlite"
)

func TestChaos_SwarmRecoveryE2E(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "chaos_e2e.db")

	db, err := orchestration.NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init SIP DB: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	numAgents := 20
	missionsPerAgent := 5
	errs := make(chan error, numAgents*missionsPerAgent)

	slog.Info("Starting stress test")
	start := time.Now()
	for i := 0; i < numAgents; i++ {
		wg.Add(1)
		go func(agentIdx int) {
			defer wg.Done()
			for j := 0; j < missionsPerAgent; j++ {
				missionID := fmt.Sprintf("stress-%d-%d", agentIdx, j)
				task := orchestration.Message{
					ID:      missionID,
					Content: "Stress test task",
					Type:    orchestration.EventTask,
				}
				if err := db.DelegateMission(ctx, missionID, "SOFTWARE_ENGINEER", task); err != nil {
					if err.Error() != "database is locked (5) (SQLITE_BUSY)" && !strings.Contains(err.Error(), "database is locked") {
						errs <- fmt.Errorf("agent %d failed to delegate mission %d: %v", agentIdx, j, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Concurrency error: %v", err)
	}

	slog.Info("Stress test complete", "duration", time.Since(start))
	slog.Info("Starting Chaos DB Lock")

	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open raw DB: %v", err)
	}
	defer rawDB.Close()

	tx, err := rawDB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("BEGIN EXCLUSIVE")
	if err != nil {
		slog.Warn("BEGIN EXCLUSIVE error", "err", err)
	} else {
		_, err = tx.Exec("UPDATE agent_missions SET status = 'LOCKED' WHERE 1=0")
		if err != nil {
			t.Fatalf("Failed to lock table: %v", err)
		}
	}

	var retryWg sync.WaitGroup
	retryWg.Add(1)

	startChaos := time.Now()

	go func() {
		defer retryWg.Done()
		task := orchestration.Message{
			ID:      "chaos-mission-1",
			Content: "Chaos test task",
			Type:    orchestration.EventTask,
		}

		err := db.DelegateMission(ctx, "chaos-mission-1", "SOFTWARE_ENGINEER", task)
		if err != nil {
			slog.Error("Mission delegation after chaos failed", "err", err)
		} else {
			slog.Info("Mission delegation succeeded after", "duration", time.Since(startChaos))
		}
	}()

	time.Sleep(200 * time.Millisecond)

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit and release lock: %v", err)
	}

	retryWg.Wait()

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
		slog.Info("Successfully verified mission ingestion after DB lock recovery")
	}
}
