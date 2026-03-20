package interop

import "context"

// State Intent: State represents shared agent state.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// UniversalAdapter Intent: UniversalAdapter defines the interface for interacting with different agent frameworks.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type UniversalAdapter interface {
	// SyncState synchronizes the agent's local state with the central shared state.
	SyncState(ctx context.Context, state *State) error

	// ExecuteCommand runs a generic command for the agent.
	ExecuteCommand(ctx context.Context, cmd string) (string, error)
}
