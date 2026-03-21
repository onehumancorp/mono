package interop

import (
	"context"
	"fmt"
)

// Summary: CrewAIAdapter implements UniversalAdapter for CrewAI.
// Intent: CrewAIAdapter implements UniversalAdapter for CrewAI.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type CrewAIAdapter struct {
	Identity string
}

// Summary: NewCrewAIAdapter creates a new CrewAIAdapter.
// Intent: NewCrewAIAdapter creates a new CrewAIAdapter.
// Params: identity
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

// Summary: SyncState functionality.
// Intent: SyncState functionality.
// Params: ctx, state
// Returns: error
// Errors: Returns an error if applicable
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

// Summary: ExecuteCommand functionality.
// Intent: ExecuteCommand functionality.
// Params: ctx, cmd
// Returns: (string, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (a *CrewAIAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("CrewAI executed: %s", cmd), nil
}
