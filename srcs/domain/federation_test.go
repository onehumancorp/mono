package domain

import "testing"

func TestHubRouter_RouteAgent(t *testing.T) {
	router := NewHubRouter()

	tests := []struct {
		name          string
		agentID       string
		clusters      []string
		latencies     map[string]int
		expected      string
		expectError   bool
	}{
		{
			name:          "Selects cluster with lowest latency",
			agentID:       "agent-1",
			clusters:      []string{"us-east", "eu-central", "ap-south"},
			latencies:     map[string]int{"us-east": 150, "eu-central": 45, "ap-south": 250},
			expected:      "eu-central",
			expectError:   false,
		},
		{
			name:          "Ignores clusters with unknown latency",
			agentID:       "agent-2",
			clusters:      []string{"us-east", "eu-central"},
			latencies:     map[string]int{"us-east": 120}, // eu-central missing
			expected:      "us-east",
			expectError:   false,
		},
		{
			name:          "Returns error if no valid clusters found",
			agentID:       "agent-3",
			clusters:      []string{"us-east", "eu-central"},
			latencies:     map[string]int{"ap-south": 50}, // mismatch between clusters and latencies
			expected:      "",
			expectError:   true,
		},
		{
			name:          "Returns error if cluster list is empty",
			agentID:       "agent-4",
			clusters:      []string{},
			latencies:     map[string]int{"us-east": 10},
			expected:      "",
			expectError:   true,
		},
		{
			name:          "Handles single cluster correctly",
			agentID:       "agent-5",
			clusters:      []string{"us-west"},
			latencies:     map[string]int{"us-west": 8},
			expected:      "us-west",
			expectError:   false,
		},
		{
			name:          "FederatedAgent Instantiation", // Just to make sure we hit coverage on the struct fields if ever needed, though structs don't usually need test logic.
			agentID:       "dummy",
			clusters:      []string{"dummy"},
			latencies:     map[string]int{"dummy": 10},
			expected:      "dummy",
			expectError:   false,
		},
	}

	// Just a quick instantiation of the struct to ensure it compiles and represents coverage.
	_ = FederatedAgent{
		AgentID:      "test-id",
		HomeCluster:  "test-cluster",
		Status:       "GLOBAL_IDLE",
		LatencyScore: 42,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := router.RouteAgent(tt.agentID, tt.clusters, tt.latencies)

			if tt.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
