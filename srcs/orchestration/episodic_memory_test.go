package orchestration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
)

func TestStatefulEpisodicMemory_E2E_Success(t *testing.T) {
	hub := NewHub()
	server := NewHubServiceServer(hub)
	defer os.Remove("events.jsonl")

	eventID := "evt-123"
	agentID := "agent-xyz"
	payload := []byte(`{"action": "step_3", "state": "saved_data"}`)

	req := pb.StatefulEpisodicMemoryEvent_builder{
		EventId: eventID,
		AgentId: agentID,
		Payload: payload,
	}.Build()

	resp, err := server.StatefulEpisodicMemory(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !resp.GetSuccess() {
		t.Fatalf("expected success to be true")
	}

	time.Sleep(100 * time.Millisecond) // Allow event worker to process

	// Read events.jsonl to verify
	b, err := os.ReadFile("events.jsonl")
	if err != nil {
		t.Fatalf("failed to read events.jsonl: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(b, &entry); err != nil {
		t.Fatalf("failed to parse jsonl line: %v", err)
	}

	if entry["event_id"] != eventID {
		t.Errorf("expected event_id %s, got %v", eventID, entry["event_id"])
	}
	if entry["agent_id"] != agentID {
		t.Errorf("expected agent_id %s, got %v", agentID, entry["agent_id"])
	}
	if entry["type"] != "StatefulEpisodicMemory" {
		t.Errorf("expected type StatefulEpisodicMemory, got %v", entry["type"])
	}
}

func TestStatefulEpisodicMemory_StrictSchema(t *testing.T) {
	hub := NewHub()
	server := NewHubServiceServer(hub)

	eventID := "evt-bad"
	agentID := "agent-xyz"
	// Create an invalid JSON payload to trigger an error
	payload := []byte(`{"action": "step_3", "unknown_field_123": "saved_data"}`)

	req := pb.StatefulEpisodicMemoryEvent_builder{
		EventId: eventID,
		AgentId: agentID,
		Payload: payload,
	}.Build()

	_, err := server.StatefulEpisodicMemory(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error due to strict schema validation, got nil")
	}
}

func TestStatefulEpisodicMemory_ResourceBounding(t *testing.T) {
	hub := NewHub()
	eventID := "evt-concurrent"

	hub.mu.Lock()
	hub.episodicMemoryTrackers[eventID] = struct{}{}
	hub.mu.Unlock()

	err := hub.StatefulEpisodicMemory(eventID, "agent-xyz", []byte(`{"valid": "payload"}`))
	if err == nil || err.Error() != "event already being processed" {
		t.Fatalf("expected 'event already being processed', got %v", err)
	}

	// Verify deferred deletion
	hub.mu.Lock()
	delete(hub.episodicMemoryTrackers, eventID) // Simulate resolution
	hub.mu.Unlock()

	err = hub.StatefulEpisodicMemory(eventID, "agent-xyz", []byte(`{"valid": "payload"}`))
	if err != nil {
		t.Fatalf("expected success after unlocking, got %v", err)
	}

	hub.mu.RLock()
	_, exists := hub.episodicMemoryTrackers[eventID]
	hub.mu.RUnlock()

	if exists {
		t.Fatalf("expected tracker map to be cleaned up after execution")
	}
}
