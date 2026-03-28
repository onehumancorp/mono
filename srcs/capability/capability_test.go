package capability

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCapabilityStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "capability_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.db.Close()

	ctx := context.Background()
	plugin := Plugin{
		ID:          "plugin-1",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		ManifestURL: "http://example.com/manifest.json",
		Status:      "ACTIVE",
	}

	err = store.RegisterPlugin(ctx, plugin)
	if err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	fetched, err := store.GetPlugin(ctx, "plugin-1")
	if err != nil {
		t.Fatalf("failed to fetch plugin: %v", err)
	}

	if fetched.Name != plugin.Name {
		t.Errorf("expected name %s, got %s", plugin.Name, fetched.Name)
	}
	if fetched.Version != plugin.Version {
		t.Errorf("expected version %s, got %s", plugin.Version, fetched.Version)
	}

	plugins, err := store.ListPlugins(ctx)
	if err != nil {
		t.Fatalf("failed to list plugins: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	plugin.Status = "INACTIVE"
	err = store.RegisterPlugin(ctx, plugin)
	if err != nil {
		t.Fatalf("failed to update plugin: %v", err)
	}

	fetched, err = store.GetPlugin(ctx, "plugin-1")
	if err != nil {
		t.Fatalf("failed to fetch plugin: %v", err)
	}

	if fetched.Status != "INACTIVE" {
		t.Errorf("expected status INACTIVE, got %s", fetched.Status)
	}

	// negative test
	_, err = store.GetPlugin(ctx, "nonexistent")
	if err == nil {
		t.Errorf("expected error fetching nonexistent plugin, got nil")
	}
}
