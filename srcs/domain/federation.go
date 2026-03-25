package domain

import (
	"errors"
	"sync"
)

// FederatedAgent defines the structure for an agent in a federated multi-cluster environment.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type FederatedAgent struct {
	AgentID      string `json:"id"`
	HomeCluster  string `json:"home_cluster"`
	Status       string `json:"status"` // e.g., GLOBAL_IDLE, BUSY
	LatencyScore int    `json:"latency_ms"`
}

// FederatedRegistry holds the federated agents across the global organization.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type FederatedRegistry struct {
	mu     sync.RWMutex
	agents map[string]FederatedAgent
}

// NewFederatedRegistry creates a new registry.
// Accepts no parameters.
// Returns *FederatedRegistry.
// Produces no errors.
// Has no side effects.
func NewFederatedRegistry() *FederatedRegistry {
	return &FederatedRegistry{
		agents: make(map[string]FederatedAgent),
	}
}

// RegisterAgent adds a new federated agent to the registry.
// Accepts parameters: r *FederatedRegistry (No Constraints).
// Returns error.
// Produces errors: Returns error if agent already exists or home cluster is empty.
// Has side effects: Modifies the registry map.
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

// GetAgent retrieves a federated agent from the registry.
// Accepts parameters: r *FederatedRegistry (No Constraints).
// Returns FederatedAgent, bool.
// Produces no errors.
// Has no side effects.
func (r *FederatedRegistry) GetAgent(agentID string) (FederatedAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentID]
	return agent, ok
}

// UpdateAgentStatus updates the status of an existing agent.
// Accepts parameters: r *FederatedRegistry (No Constraints).
// Returns error.
// Produces errors: Returns error if agent is not found.
// Has side effects: Modifies the agent's status in the registry.
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
