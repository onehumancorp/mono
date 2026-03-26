package settings

import (
	"os"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	store := NewStore()
	settings := store.Get()

	if settings.ListenAddr != "0.0.0.0:18789" {
		t.Errorf("expected default listen addr 0.0.0.0:18789, got %s", settings.ListenAddr)
	}
	if settings.DBPath == nil || *settings.DBPath != "ohc.db" {
		t.Errorf("expected default db path ohc.db, got %v", settings.DBPath)
	}
}

func TestSetExtra(t *testing.T) {
	store := NewStore()
	err := store.SetExtra("theme", "dark")
	if err != nil {
		t.Fatalf("SetExtra failed: %v", err)
	}

	settings := store.Get()
	if val, ok := settings.Extras["theme"]; !ok || val != "dark" {
		t.Errorf("expected extra theme=dark, got %v", settings.Extras["theme"])
	}
}

func TestPersistence(t *testing.T) {
	tmpFile := "test_settings.json"
	defer os.Remove(tmpFile)

	store, err := FromFile(tmpFile)
	if err != nil {
		t.Fatalf("FromFile failed: %v", err)
	}

	dbPath := "custom.db"
	settings := store.Get()
	settings.DBPath = &dbPath
	settings.Extras["foo"] = "bar"

	if err := store.Set(settings); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Load from same file
	store2, err := FromFile(tmpFile)
	if err != nil {
		t.Fatalf("second FromFile failed: %v", err)
	}

	settings2 := store2.Get()
	if settings2.DBPath == nil || *settings2.DBPath != "custom.db" {
		t.Errorf("expected persistent db path custom.db, got %v", settings2.DBPath)
	}
	if settings2.Extras["foo"] != "bar" {
		t.Errorf("expected persistent extra foo=bar, got %v", settings2.Extras["foo"])
	}
}
