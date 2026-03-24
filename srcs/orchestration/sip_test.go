package orchestration

import (
	"context"
	"testing"
)

func TestSIPDB_Init(t *testing.T) {
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize SIPDB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test Memory
	err = db.UpdateMemory(ctx, "architecture", "microservices")
	if err != nil {
		t.Fatalf("UpdateMemory failed: %v", err)
	}

	val, err := db.SyncMemory(ctx, "architecture")
	if err != nil {
		t.Fatalf("SyncMemory failed: %v", err)
	}
	if val != "microservices" {
		t.Fatalf("expected 'microservices', got '%s'", val)
	}

	// Test Heartbeat
	err = db.Heartbeat(ctx, "agent-1", "SOFTWARE_ENGINEER", "ACTIVE")
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}

	// Test Delegation & Mission
	msg := Message{ID: "m1", Content: "Build a feature", Type: EventTask}
	err = db.DelegateMission(ctx, "m1", "SOFTWARE_ENGINEER", msg)
	if err != nil {
		t.Fatalf("DelegateMission failed: %v", err)
	}

	missions, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("GetPendingMissions failed: %v", err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 mission, got %d", len(missions))
	}
	if missions[0].ID != "m1" {
		t.Fatalf("expected mission ID 'm1', got '%s'", missions[0].ID)
	}

	// Test Completion
	err = db.CompleteMission(ctx, "m1")
	if err != nil {
		t.Fatalf("CompleteMission failed: %v", err)
	}

	missions, err = db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
	if err != nil {
		t.Fatalf("GetPendingMissions failed: %v", err)
	}
	if len(missions) != 0 {
		t.Fatalf("expected 0 missions, got %d", len(missions))
	}
}
