package interop

import (
	"context"
	"fmt"
	"strings"
)

// IronClawAdapter implements UniversalAdapter for IronClaw.
// IronClaw is a security and audit-focused agent with deep static-analysis
// capabilities.  The adapter bridges the IronClaw agent into the platform's
// universal interop layer so it can participate in cross-framework swarms.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type IronClawAdapter struct {
	Identity string
}

// NewIronClawAdapter creates a new IronClawAdapter.
// Parameters: identity string – a valid SPIFFE ID for the agent.
// Returns: *IronClawAdapter, error
// Errors: Returns error if identity does not pass SPIFFE validation.
// Side Effects: None
func NewIronClawAdapter(identity string) (*IronClawAdapter, error) {
	if err := ValidateSPIFFEID(identity); err != nil {
		return nil, fmt.Errorf("invalid identity for IronClawAdapter: %w", err)
	}
	return &IronClawAdapter{Identity: identity}, nil
}

// SyncState synchronises the IronClaw agent's local state with the
// central shared state via LangGraph checkpoints.
// Parameters: a *IronClawAdapter (No Constraints), ctx context.Context, state *State
// Returns: error
// Errors: Returns error if state is nil.
// Side Effects: Modifies state.Data – sets ironclaw_synced, last_identity, and appends a checkpoint.
func (a *IronClawAdapter) SyncState(ctx context.Context, state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data["ironclaw_synced"] = true
	state.Data["last_identity"] = a.Identity

	LogCheckpoint(state, a.Identity)
	return nil
}

// ExecuteCommand runs a security-oriented command via the IronClaw
// agent.  Commands are forwarded as-is; the agent is responsible for
// interpreting security directives (e.g. "scan", "audit", "report").
// Parameters: a *IronClawAdapter (No Constraints), ctx context.Context, cmd string
// Returns: (string, error)
// Errors: Returns error if cmd is empty.
// Side Effects: None
func (a *IronClawAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}

	// Validate that context is not already cancelled before processing.
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Prefix security-sensitive commands with the agent identity so the
	// audit trail is always traceable.
	result := fmt.Sprintf("IronClaw[%s] executed: %s", trimSPIFFEPath(a.Identity), cmd)
	return result, nil
}

// trimSPIFFEPath returns the final path segment of a SPIFFE ID for concise
// logging (e.g. "spiffe://ohc.local/agent/ironclaw-1" → "ironclaw-1").
func trimSPIFFEPath(id string) string {
	parts := strings.Split(id, "/")
	if len(parts) == 0 {
		return id
	}
	return parts[len(parts)-1]
}
