package domain

import (
	"errors"
	"sync"
)

// Summary: FederatedAgent defines the structure for an agent in a federated multi-cluster environment.
// Intent: FederatedAgent defines the structure for an agent in a federated multi-cluster environment.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type FederatedAgent struct {
	AgentID      string `json:"id"`
	HomeCluster  string `json:"home_cluster"`
	Status       string `json:"status"` // e.g., GLOBAL_IDLE, BUSY
	LatencyScore int    `json:"latency_ms"`
}

// Summary: FederatedRegistry holds the federated agents across the global organization.
// Intent: FederatedRegistry holds the federated agents across the global organization.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type FederatedRegistry struct {
	mu     sync.RWMutex
	agents map[string]FederatedAgent
}

// Summary: NewFederatedRegistry creates a new registry.
// Intent: NewFederatedRegistry creates a new registry.
// Params: None
// Returns: *FederatedRegistry
// Errors: None
// Side Effects: None
func NewFederatedRegistry() *FederatedRegistry {
	return &FederatedRegistry{
		agents: make(map[string]FederatedAgent),
	}
}

// Summary: RegisterAgent adds a new federated agent to the registry.
// Intent: RegisterAgent adds a new federated agent to the registry.
// Params: agent FederatedAgent
// Returns: error
// Errors: Returns error if agent already exists or home cluster is empty.
// Side Effects: Modifies the registry map.
func (r *FederatedRegistry) RegisterAgent(agent FederatedAgent) error {
	if agent.AgentID == "" {
		return errors.New("agent ID cannot be empty")
	}
	if agent.HomeCluster == "" {
		return errors.New("home cluster cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agent.AgentID]; exists {
		return errors.New("agent already registered")
	}
	r.agents[agent.AgentID] = agent
	return nil
}

// Summary: GetAgent retrieves a federated agent from the registry.
// Intent: GetAgent retrieves a federated agent from the registry.
// Params: agentID string
// Returns: FederatedAgent, bool
// Errors: None
// Side Effects: None
func (r *FederatedRegistry) GetAgent(agentID string) (FederatedAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentID]
	return agent, ok
}

// Summary: UpdateAgentStatus updates the status of an existing agent.
// Intent: UpdateAgentStatus updates the status of an existing agent.
// Params: agentID, status
// Returns: error
// Errors: Returns error if agent is not found.
// Side Effects: Modifies the agent's status in the registry.
func (r *FederatedRegistry) UpdateAgentStatus(agentID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return errors.New("agent not found")
	}
	agent.Status = status
	r.agents[agentID] = agent
	return nil
}
