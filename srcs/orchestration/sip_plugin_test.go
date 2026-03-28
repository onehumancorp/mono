package orchestration

import (
	"context"
	"testing"
	"path/filepath"
	"os"
)

func TestSIPDB_Plugins(t *testing.T) {
	dbPath := filepath.Join(os.TempDir(), "sipdb_plugin_test.db")
	os.Remove(dbPath)

	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	ctx := context.Background()

	err = db.RegisterCapabilityPlugin(ctx, "plugin_1", "Marketing Automation", "v1.0.0", "http://marketing-plugin.local/manifest.yaml")
	if err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	plugins, err := db.QueryCapabilityPlugins(ctx)
	if err != nil {
		t.Fatalf("failed to query plugins: %v", err)
	}

	if len(plugins) != 1 || plugins[0] != "http://marketing-plugin.local/manifest.yaml" {
		t.Fatalf("unexpected plugins returned: %v", plugins)
	}

	err = db.RecordSwarmMemoryEmbedding(ctx, "mem_1", "Agent completed marketing campaign", "plugin_1", []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("failed to record embedding: %v", err)
	}
}
