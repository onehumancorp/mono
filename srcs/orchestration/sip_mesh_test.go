package orchestration

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSIPDB_CapabilityPlugins(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize SIPDB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Register Plugin
	plugin := CapabilityPlugin{
		PluginID:    "mcp-github",
		Name:        "GitHub MCP",
		Version:     "1.0.0",
		ManifestURL: "https://api.github.com/mcp",
		Status:      "ACTIVE",
	}

	err = db.RegisterCapabilityPlugin(ctx, plugin)
	if err != nil {
		t.Fatalf("RegisterCapabilityPlugin failed: %v", err)
	}

	// Get All Plugins
	plugins, err := db.GetCapabilityPlugins(ctx, "")
	if err != nil {
		t.Fatalf("GetCapabilityPlugins failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].PluginID != "mcp-github" {
		t.Fatalf("expected PluginID 'mcp-github', got '%s'", plugins[0].PluginID)
	}

	// Get Active Plugins
	plugins, err = db.GetCapabilityPlugins(ctx, "ACTIVE")
	if err != nil {
		t.Fatalf("GetCapabilityPlugins (ACTIVE) failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 ACTIVE plugin, got %d", len(plugins))
	}

	// Get Inactive Plugins
	plugins, err = db.GetCapabilityPlugins(ctx, "INACTIVE")
	if err != nil {
		t.Fatalf("GetCapabilityPlugins (INACTIVE) failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected 0 INACTIVE plugins, got %d", len(plugins))
	}

	// Test Update / Conflict
	plugin.Status = "INACTIVE"
	err = db.RegisterCapabilityPlugin(ctx, plugin)
	if err != nil {
		t.Fatalf("RegisterCapabilityPlugin (update) failed: %v", err)
	}

	plugins, err = db.GetCapabilityPlugins(ctx, "INACTIVE")
	if err != nil {
		t.Fatalf("GetCapabilityPlugins (INACTIVE) failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 INACTIVE plugins after update, got %d", len(plugins))
	}
}

func TestSIPDB_EpisodicMemory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize SIPDB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Store Memory
	memory := EpisodicMemory{
		MemoryID:        "mem-123",
		Context:         "User asked about K8s configuration.",
		VectorEmbedding: []byte{0x01, 0x02, 0x03, 0x04}, // dummy embedding
		SourcePlugin:    "mcp-github",
	}

	err = db.StoreEpisodicMemory(ctx, memory)
	if err != nil {
		t.Fatalf("StoreEpisodicMemory failed: %v", err)
	}

	// Retrieve All Memories
	memories, err := db.GetEpisodicMemoriesByPlugin(ctx, "")
	if err != nil {
		t.Fatalf("GetEpisodicMemoriesByPlugin failed: %v", err)
	}
	if len(memories) != 1 {
		t.Fatalf("expected 1 memory, got %d", len(memories))
	}
	if memories[0].MemoryID != "mem-123" {
		t.Fatalf("expected MemoryID 'mem-123', got '%s'", memories[0].MemoryID)
	}
	if len(memories[0].VectorEmbedding) != 4 {
		t.Fatalf("expected 4 bytes of vector embedding, got %d", len(memories[0].VectorEmbedding))
	}

	// Retrieve by Source Plugin
	memories, err = db.GetEpisodicMemoriesByPlugin(ctx, "mcp-github")
	if err != nil {
		t.Fatalf("GetEpisodicMemoriesByPlugin failed: %v", err)
	}
	if len(memories) != 1 {
		t.Fatalf("expected 1 memory for plugin 'mcp-github', got %d", len(memories))
	}

	// Retrieve by Non-Existent Source Plugin
	memories, err = db.GetEpisodicMemoriesByPlugin(ctx, "mcp-unknown")
	if err != nil {
		t.Fatalf("GetEpisodicMemoriesByPlugin failed: %v", err)
	}
	if len(memories) != 0 {
		t.Fatalf("expected 0 memories for plugin 'mcp-unknown', got %d", len(memories))
	}

	// Update memory
	memory.Context = "Updated context."
	err = db.StoreEpisodicMemory(ctx, memory)
	if err != nil {
		t.Fatalf("StoreEpisodicMemory (update) failed: %v", err)
	}
	memories, _ = db.GetEpisodicMemoriesByPlugin(ctx, "mcp-github")
	if memories[0].Context != "Updated context." {
		t.Fatalf("expected Context 'Updated context.', got '%s'", memories[0].Context)
	}
}

func TestSIPDB_CapabilityPlugins_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close() // close immediately to cause errors

	err = db.RegisterCapabilityPlugin(context.Background(), CapabilityPlugin{})
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}

	_, err = db.GetCapabilityPlugins(context.Background(), "")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}

func TestSIPDB_EpisodicMemory_DBError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSIPDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	db.Close() // close immediately to cause errors

	err = db.StoreEpisodicMemory(context.Background(), EpisodicMemory{})
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}

	_, err = db.GetEpisodicMemoriesByPlugin(context.Background(), "")
	if err == nil {
		t.Fatal("Expected error querying closed DB")
	}
}
