package domain

import "errors"

// Summary: FederatedAgent represents an agent operating within a multi-cluster federation.
// Intent: FederatedAgent represents an agent operating within a multi-cluster federation.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type FederatedAgent struct {
	AgentID      string `json:"id"`
	HomeCluster  string `json:"home_cluster"`
	Status       string `json:"status"` // GLOBAL_IDLE, BUSY
	LatencyScore int    `json:"latency_ms"`
}

// Summary: HubRouter is responsible for routing tasks across clusters in a multi-cluster federation.
// Intent: HubRouter is responsible for routing tasks across clusters in a multi-cluster federation.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type HubRouter struct{}

// Summary: NewHubRouter constructs a new instance of HubRouter.
// Intent: NewHubRouter constructs a new instance of HubRouter.
// Params: None
// Returns: *HubRouter
// Errors: None
// Side Effects: None
func NewHubRouter() *HubRouter {
	return &HubRouter{}
}

// Summary: RouteAgent selects the optimal cluster for a given agent based on latency scores.
// Intent: RouteAgent selects the optimal cluster for a given agent based on latency scores.
// Params: agentID, clusters, latencies
// Returns: (string, error)
// Errors: Returns an error if no suitable cluster is found or parameters are invalid.
// Side Effects: None
func (hr *HubRouter) RouteAgent(agentID string, clusters []string, latencies map[string]int) (string, error) {
	if len(clusters) == 0 {
		return "", errors.New("no clusters provided")
	}

	bestCluster := ""
	bestLatency := -1

	for _, cluster := range clusters {
		latency, ok := latencies[cluster]
		if !ok {
			// Skip clusters without a known latency score
			continue
		}

		if bestLatency == -1 || latency < bestLatency {
			bestLatency = latency
			bestCluster = cluster
		}
	}

	if bestCluster == "" {
		return "", errors.New("no cluster with valid latency found")
	}

	return bestCluster, nil
}
