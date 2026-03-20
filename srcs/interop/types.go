package interop

import "context"

// State represents shared agent state.
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// UniversalAdapter defines the interface for interacting with different agent frameworks.
type UniversalAdapter interface {
	// SyncState synchronizes the agent's local state with the central shared state.
	SyncState(ctx context.Context, state *State) error

	// ExecuteCommand runs a generic command for the agent.
	ExecuteCommand(ctx context.Context, cmd string) (string, error)
}
