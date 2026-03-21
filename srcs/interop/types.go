package interop

import "context"

// Summary: Defines the State type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// Summary: Defines the UniversalAdapter type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type UniversalAdapter interface {
	// SyncState synchronizes the agent's local state with the central shared state.
	SyncState(ctx context.Context, state *State) error

	// ExecuteCommand runs a generic command for the agent.
	ExecuteCommand(ctx context.Context, cmd string) (string, error)
}
