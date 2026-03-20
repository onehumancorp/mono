package interop

import (
	"context"
	"fmt"
)

// AutoGenAdapter implements UniversalAdapter for AutoGen.
type AutoGenAdapter struct {
	Identity string
}

// NewAutoGenAdapter creates a new AutoGenAdapter.
func NewAutoGenAdapter(identity string) *AutoGenAdapter {
	return &AutoGenAdapter{
		Identity: identity,
	}
}

func (a *AutoGenAdapter) SyncState(ctx context.Context, state *State) error {
	// Mock K8s/LangGraph state sync for AutoGen
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data["autogen_synced"] = true
	state.Data["last_identity"] = a.Identity
	return nil
}

func (a *AutoGenAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("AutoGen executed: %s", cmd), nil
}
