package interop

import (
	"context"
	"fmt"
)

// CrewAIAdapter implements UniversalAdapter for CrewAI.
type CrewAIAdapter struct {
	Identity string
}

// NewCrewAIAdapter creates a new CrewAIAdapter.
func NewCrewAIAdapter(identity string) *CrewAIAdapter {
	return &CrewAIAdapter{
		Identity: identity,
	}
}

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
	return nil
}

func (a *CrewAIAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("CrewAI executed: %s", cmd), nil
}
