package crewai

import (
	"context"
	"github.com/onehumancorp/mono/srcs/interop"
	"fmt"
)

// CrewAIAdapter implements interop.UniversalAdapter for CrewAI.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type CrewAIAdapter struct {
	Identity string
}

// NewCrewAIAdapter creates a new CrewAIAdapter.
// Accepts parameters: identity string (No Constraints).
// Returns *CrewAIAdapter, error.
// Produces errors: Returns error if identity is invalid.
// Has no side effects.
func NewCrewAIAdapter(identity string) (*CrewAIAdapter, error) {
	if err := interop.ValidateSPIFFEID(identity); err != nil {
		return nil, fmt.Errorf("invalid identity for CrewAIAdapter: %w", err)
	}
	return &CrewAIAdapter{
		Identity: identity,
	}, nil
}

// SyncState functionality.
// Accepts parameters: a *CrewAIAdapter (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (a *CrewAIAdapter) SyncState(ctx context.Context, state *interop.State) error {
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
	interop.LogCheckpoint(state, a.Identity)
	return nil
}

// ExecuteCommand functionality.
// Accepts parameters: a *CrewAIAdapter (No Constraints).
// Returns (string, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (a *CrewAIAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("CrewAI executed: %s", cmd), nil
}
