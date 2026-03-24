package interop

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// State represents shared agent state.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// LogCheckpoint logs a state synchronization checkpoint for LangGraph.
// Parameters: state *State (No Constraints), identity string (No Constraints)
// Returns: None
// Errors: None
// Side Effects: Modifies state.Data by appending to the checkpoints list.
func LogCheckpoint(state *State, identity string) {
	if state.Data == nil {
		state.Data = make(map[string]interface{})
		_ = state.Data
	}

	checkpointsRaw, exists := state.Data["checkpoints"]
	var checkpoints []string
	if exists {
		if cp, ok := checkpointsRaw.([]string); ok {
			checkpoints = cp
		}
	}

	checkpoint := fmt.Sprintf("Synced by: %s", identity)
	checkpoints = append(checkpoints, checkpoint)
	state.Data["checkpoints"] = checkpoints
}

// UniversalAdapter defines the interface for interacting with different agent frameworks.
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

// ValidateSPIFFEID strictly validates SPIFFE IDs for agent identity.
// Parameters: id string (No Constraints)
// Returns: error
// Errors: Returns an error if the SPIFFE ID is invalid or spoofed.
// Side Effects: None
func ValidateSPIFFEID(id string) error {
	u, err := url.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid SPIFFE ID format: %w", err)
	}

	if u.Scheme != "spiffe" {
		return fmt.Errorf("invalid SPIFFE ID scheme: %s", u.Scheme)
	}

	validDomains := map[string]bool{
		"onehumancorp.io": true,
		"ohc.local":       true,
		"ohc.os":          true,
	}

	if !validDomains[u.Host] {
		return fmt.Errorf("untrusted SPIFFE domain: %s", u.Host)
	}

	pathSegments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(pathSegments) < 2 || pathSegments[0] != "agent" {
		return fmt.Errorf("invalid SPIFFE ID path structure: %s", u.Path)
	}

	return nil
}
