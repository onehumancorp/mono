package domain

import (
	"testing"
)

func TestFederatedRegistry(t *testing.T) {
	registry := NewFederatedRegistry()

	t.Run("Register Valid Agent", func(t *testing.T) {
		agent := FederatedAgent{
			AgentID:      "agent-1",
			HomeCluster:  "eu-central-1",
			Status:       "GLOBAL_IDLE",
			LatencyScore: 10,
		}
		err := registry.RegisterAgent(agent)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Register Missing ID", func(t *testing.T) {
		agent := FederatedAgent{
			HomeCluster: "us-east-1",
		}
		err := registry.RegisterAgent(agent)
		if err == nil {
			t.Fatal("expected error for missing ID")
		}
	})

	t.Run("Register Missing HomeCluster", func(t *testing.T) {
		agent := FederatedAgent{
			AgentID: "agent-2",
		}
		err := registry.RegisterAgent(agent)
		if err == nil {
			t.Fatal("expected error for missing home cluster")
		}
	})

	t.Run("Register Duplicate Agent", func(t *testing.T) {
		agent := FederatedAgent{
			AgentID:      "agent-1",
			HomeCluster:  "eu-central-1",
		}
		err := registry.RegisterAgent(agent)
		if err == nil {
			t.Fatal("expected error for duplicate agent")
		}
	})

	t.Run("Get Existing Agent", func(t *testing.T) {
		agent, ok := registry.GetAgent("agent-1")
		if !ok {
			t.Fatal("expected to find agent-1")
		}
		if agent.HomeCluster != "eu-central-1" {
			t.Errorf("expected home cluster eu-central-1, got %s", agent.HomeCluster)
		}
	})

	t.Run("Get Missing Agent", func(t *testing.T) {
		_, ok := registry.GetAgent("agent-not-found")
		if ok {
			t.Fatal("expected not to find agent")
		}
	})

	t.Run("Update Existing Agent Status", func(t *testing.T) {
		err := registry.UpdateAgentStatus("agent-1", "BUSY")
		if err != nil {
			t.Fatalf("expected no error updating status, got %v", err)
		}
		agent, _ := registry.GetAgent("agent-1")
		if agent.Status != "BUSY" {
			t.Errorf("expected status BUSY, got %s", agent.Status)
		}
	})

	t.Run("Update Missing Agent Status", func(t *testing.T) {
		err := registry.UpdateAgentStatus("agent-not-found", "BUSY")
		if err == nil {
			t.Fatal("expected error updating status of missing agent")
		}
	})
}
