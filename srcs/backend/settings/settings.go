package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AiProvider configuration.
type AiProvider struct {
	Name    string  `json:"name"`
	APIKey  *string `json:"api_key,omitempty"`
	BaseURL *string `json:"base_url,omitempty"`
	Model   string  `json:"model"`
	Enabled bool    `json:"enabled"`
}

// AppSettings for a single OHC deployment.
type AppSettings struct {
	ListenAddr  string            `json:"listen_addr"`
	DBPath      *string           `json:"db_path,omitempty"`
	PostgresURL *string           `json:"postgres_url,omitempty"`
	RedisURL    *string           `json:"redis_url,omitempty"`
	AIProviders []AiProvider      `json:"ai_providers"`
	Extras      map[string]string `json:"extras"`
}

// DefaultSettings returns the default configuration.
func DefaultSettings() AppSettings {
	dbPath := "ohc.db"
	return AppSettings{
		ListenAddr:  "0.0.0.0:18789",
		DBPath:      &dbPath,
		AIProviders: []AiProvider{},
		Extras:      make(map[string]string),
	}
}

// SettingsStore manages thread-safe access to AppSettings.
type SettingsStore struct {
	mu       sync.RWMutex
	settings AppSettings
	path     string
}

// NewStore creates an in-memory store with default settings.
func NewStore() *SettingsStore {
	return &SettingsStore{
		settings: DefaultSettings(),
	}
}

// FromFile loads settings from a JSON file, falling back to defaults if not found.
func FromFile(path string) (*SettingsStore, error) {
	store := &SettingsStore{
		path:     path,
		settings: DefaultSettings(),
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return store, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	if err := json.Unmarshal(data, &store.settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	return store, nil
}

// Save persists the current settings to disk if a path is configured.
func (s *SettingsStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.path == "" {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// Get returns a snapshot of the current settings.
func (s *SettingsStore) Get() AppSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Deep copy extras map
	extras := make(map[string]string, len(s.settings.Extras))
	for k, v := range s.settings.Extras {
		extras[k] = v
	}

	// Deep copy providers slice
	providers := make([]AiProvider, len(s.settings.AIProviders))
	copy(providers, s.settings.AIProviders)

	snapshot := s.settings
	snapshot.Extras = extras
	snapshot.AIProviders = providers

	return snapshot
}

// Set replaces the entire settings object and persists it.
func (s *SettingsStore) Set(settings AppSettings) error {
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()

	return s.Save()
}

// SetExtra updates a single extra key/value pair and persists.
func (s *SettingsStore) SetExtra(key, value string) error {
	s.mu.Lock()
	if s.settings.Extras == nil {
		s.settings.Extras = make(map[string]string)
	}
	s.settings.Extras[key] = value
	s.mu.Unlock()

	return s.Save()
}
