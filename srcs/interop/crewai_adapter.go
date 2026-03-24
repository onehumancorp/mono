package interop

import (
	"context"
	"fmt"
)

// CrewAIAdapter implements UniversalAdapter for CrewAI.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type CrewAIAdapter struct {
	Identity string
}

// NewCrewAIAdapter creates a new CrewAIAdapter.
// Parameters: identity string (No Constraints)
// Returns: *CrewAIAdapter, error
// Errors: Returns error if identity is invalid
// Side Effects: None
func NewCrewAIAdapter(identity string) (*CrewAIAdapter, error) {
	if err := ValidateSPIFFEID(identity); err != nil {
		return nil, fmt.Errorf("invalid identity for CrewAIAdapter: %w", err)
	}
	return &CrewAIAdapter{
		Identity: identity,
	}, nil
}

// SyncState functionality.
// Parameters: a *CrewAIAdapter (No Constraints)
// Returns: error
// Errors: Explicit error handling
// Side Effects: None
func (a *CrewAIAdapter) SyncState(ctx context.Context, state *State) error {
	// Mock K8s/LangGraph state sync
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	// Simulate adding CrewAI specific metadata
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data["crewai_synced"] = true
	state.Data["last_identity"] = a.Identity

	// Ensure shared state via LangGraph is synchronized
	LogCheckpoint(state, a.Identity)
	return nil
}

// ExecuteCommand functionality.
// Parameters: a *CrewAIAdapter (No Constraints)
// Returns: (string, error)
// Errors: Explicit error handling
// Side Effects: None
func (a *CrewAIAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("CrewAI executed: %s", cmd), nil
}
