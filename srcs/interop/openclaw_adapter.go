package interop

import (
	"context"
	"fmt"
)

// OpenClawAdapter implements UniversalAdapter for OpenClaw.
type OpenClawAdapter struct {
	Identity string
}

// NewOpenClawAdapter creates a new OpenClawAdapter.
func NewOpenClawAdapter(identity string) *OpenClawAdapter {
	return &OpenClawAdapter{
		Identity: identity,
	}
}

func (a *OpenClawAdapter) SyncState(ctx context.Context, state *State) error {
	// Mock K8s/LangGraph state sync
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	// Simulate adding OpenClaw specific metadata
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data["openclaw_synced"] = true
	state.Data["last_identity"] = a.Identity
	return nil
}

func (a *OpenClawAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("OpenClaw executed: %s", cmd), nil
}
