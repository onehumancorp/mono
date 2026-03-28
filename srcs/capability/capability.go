package capability

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

// Plugin represents a registered capability plugin.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type Plugin struct {
	ID           string    `json:"plugin_id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	ManifestURL  string    `json:"manifest_url"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
}

// Store encapsulates the capability plugins database interactions.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type Store struct {
	db *sql.DB
}

// NewStore initializes a new capability plugins database connection and creates required tables.
// Accepts parameters: dbPath string (No Constraints).
// Returns (*Store, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initializeTables(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// NewStoreFromDB creates a capability store from an existing database connection.
// Accepts parameters: db *sql.DB (No Constraints).
// Returns *Store.
// Produces no errors.
// Has no side effects.
func NewStoreFromDB(db *sql.DB) *Store {
	_ = initializeTables(db)
	return &Store{db: db}
}

func initializeTables(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS capability_plugins (
		plugin_id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		manifest_url TEXT NOT NULL,
		status TEXT NOT NULL,
		registered_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	return err
}

// RegisterPlugin inserts or updates a capability plugin in the database.
// Accepts parameters: s *Store (No Constraints).
// Returns RegisterPlugin(ctx context.Context, p Plugin) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *Store) RegisterPlugin(ctx context.Context, p Plugin) error {
	query := `
		INSERT INTO capability_plugins (plugin_id, name, version, manifest_url, status, registered_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(plugin_id) DO UPDATE SET
			name=excluded.name,
			version=excluded.version,
			manifest_url=excluded.manifest_url,
			status=excluded.status
	`
	_, err := s.db.ExecContext(ctx, query, p.ID, p.Name, p.Version, p.ManifestURL, p.Status)
	return err
}

// GetPlugin retrieves a plugin by its ID.
// Accepts parameters: s *Store (No Constraints).
// Returns GetPlugin(ctx context.Context, id string) (Plugin, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *Store) GetPlugin(ctx context.Context, id string) (Plugin, error) {
	query := `SELECT plugin_id, name, version, manifest_url, status, registered_at FROM capability_plugins WHERE plugin_id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var p Plugin
	err := row.Scan(&p.ID, &p.Name, &p.Version, &p.ManifestURL, &p.Status, &p.RegisteredAt)
	if err == sql.ErrNoRows {
		return Plugin{}, errors.New("plugin not found")
	}
	return p, err
}

// ListPlugins returns all registered capability plugins.
// Accepts parameters: s *Store (No Constraints).
// Returns ListPlugins(ctx context.Context) ([]Plugin, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *Store) ListPlugins(ctx context.Context) ([]Plugin, error) {
	query := `SELECT plugin_id, name, version, manifest_url, status, registered_at FROM capability_plugins`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []Plugin
	for rows.Next() {
		var p Plugin
		if err := rows.Scan(&p.ID, &p.Name, &p.Version, &p.ManifestURL, &p.Status, &p.RegisteredAt); err != nil {
			return nil, err
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}
