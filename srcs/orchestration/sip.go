package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

// SIPDB encapsulates the Swarm Intelligence Protocol database interactions.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type SIPDB struct {
	db *sql.DB
}

const (
	maxRetries    = 3
	retryInterval = 100 * time.Millisecond
)

// withRetry executes a database operation with exponential backoff for transient errors (e.g. database is locked).
func withRetry(ctx context.Context, op func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = op()
		if err == nil {
			return nil
		}

		// If context is done, abort retries
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		slog.Warn("sipdb: operation failed, retrying", "attempt", i+1, "error", err)
		time.Sleep(retryInterval * time.Duration(1<<i))
	}
	return err
}

// NewSIPDB initializes a new database connection and creates required tables.
// Accepts parameters: dbPath string (No Constraints).
// Returns (*SIPDB, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func NewSIPDB(dbPath string) (*SIPDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initializeTables(db); err != nil {
		return nil, err
	}

	return &SIPDB{db: db}, nil
}

func initializeTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS swarm_memory (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS agent_missions (
			id TEXT PRIMARY KEY,
			role TEXT NOT NULL,
			task TEXT NOT NULL,
			status TEXT NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS agent_status (
			agent_id TEXT PRIMARY KEY,
			role TEXT NOT NULL,
			status TEXT NOT NULL,
			last_heartbeat DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS capability_plugins (
			plugin_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			version TEXT NOT NULL,
			manifest_url TEXT NOT NULL,
			status TEXT NOT NULL,
			registered_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// SyncMemory retrieves the global state for architectural alignment.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns SyncMemory(ctx context.Context, key string) (string, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) SyncMemory(ctx context.Context, key string) (string, error) {
	var value string
	err := withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx, "SELECT value FROM swarm_memory WHERE key = ?", key).Scan(&value)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	return value, err
}

// UpdateMemory updates the global state.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns UpdateMemory(ctx context.Context, key, value string) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) UpdateMemory(ctx context.Context, key, value string) error {
	return withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO swarm_memory (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP",
			key, value,
		)
		return err
	})
}

// GetPendingMissions proactively seeks tasks assigned to the role.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns GetPendingMissions(ctx context.Context, role string) ([]Message, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) GetPendingMissions(ctx context.Context, role string) ([]Message, error) {
	var missions []Message
	err := withRetry(ctx, func() error {
		missions = nil
		rows, err := s.db.QueryContext(ctx, "SELECT id, task FROM agent_missions WHERE role = ? AND status = 'PENDING'", role)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id, taskStr string
			if err := rows.Scan(&id, &taskStr); err != nil {
				return err
			}

			var msg Message
			if err := json.Unmarshal([]byte(taskStr), &msg); err != nil {
				// fallback
				msg = Message{ID: id, Content: taskStr, Type: EventTask}
			} else {
				if msg.ID == "" {
					msg.ID = id
				}
			}
			missions = append(missions, msg)
		}
		return nil
	})
	return missions, err
}

// CompleteMission updates the mission status to COMPLETED.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns CompleteMission(ctx context.Context, missionID string) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) CompleteMission(ctx context.Context, missionID string) error {
	return withRetry(ctx, func() error {
		res, err := s.db.ExecContext(ctx, "UPDATE agent_missions SET status = 'COMPLETED', updated_at = CURRENT_TIMESTAMP WHERE id = ?", missionID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return errors.New("mission not found")
		}
		return nil
	})
}

// Heartbeat maintains the agent's heartbeat and domain-health metrics.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns Heartbeat(ctx context.Context, agentID, role, status string) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) Heartbeat(ctx context.Context, agentID, role, status string) error {
	return withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO agent_status (agent_id, role, status, last_heartbeat) VALUES (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(agent_id) DO UPDATE SET role=excluded.role, status=excluded.status, last_heartbeat=CURRENT_TIMESTAMP",
			agentID, role, status,
		)
		return err
	})
}

// GetDB exposes the underlying SQL DB connection, useful for creating dependent stores.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns *sql.DB.
// Produces no errors.
// Has no side effects.
func (s *SIPDB) GetDB() *sql.DB {
	return s.db
}

// DelegateMission delegates specialized tasks via the agent_missions table.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns DelegateMission(ctx context.Context, missionID, role string, task Message) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) DelegateMission(ctx context.Context, missionID, role string, task Message) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO agent_missions (id, role, task, status, created_at, updated_at) VALUES (?, ?, ?, 'PENDING', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			missionID, role, string(taskBytes),
		)
		return err
	})
}

// Close closes the database connection.
// Accepts parameters: s *SIPDB (No Constraints).
// Returns Close() error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *SIPDB) Close() error {
	return s.db.Close()
}
