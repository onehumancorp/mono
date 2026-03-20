package interop

import "context"

// Summary: State represents shared agent state.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// Summary: UniversalAdapter defines the interface for interacting with different agent frameworks.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type UniversalAdapter interface {
	// SyncState synchronizes the agent's local state with the central shared state.
	SyncState(ctx context.Context, state *State) error

	// ExecuteCommand runs a generic command for the agent.
	ExecuteCommand(ctx context.Context, cmd string) (string, error)
}
