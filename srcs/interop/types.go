package interop

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/onehumancorp/mono/srcs/domain"
)

// State represents shared agent state.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type State struct {
	ID    string
	Data  map[string]interface{}
	Owner string
}

// LogCheckpoint logs a state synchronization checkpoint for LangGraph.
// Accepts parameters: state *State (No Constraints), identity string (No Constraints).
// Returns nothing.
// Produces no errors.
// Has side effects: Modifies state.Data by appending to the checkpoints list.
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
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type UniversalAdapter interface {
	// SyncState synchronizes the agent's local state with the central shared state.
	SyncState(ctx context.Context, state *State) error

	// ExecuteCommand runs a generic command for the agent.
	ExecuteCommand(ctx context.Context, cmd string) (string, error)
}

// ValidateSPIFFEID strictly validates SPIFFE IDs for agent identity.
// Accepts parameters: id string (No Constraints).
// Returns error.
// Produces errors: Returns an error if the SPIFFE ID is invalid or spoofed.
// Has no side effects.
func ValidateSPIFFEID(id string) error {
	if !strings.HasPrefix(id, "spiffe://") {
		return fmt.Errorf("invalid SPIFFE ID scheme: %s", id)
	}
	if strings.Contains(strings.ToLower(id), "%2f") {
		return fmt.Errorf("invalid SPIFFE ID format: contains url-encoded characters")
	}
	trimmed := strings.TrimPrefix(id, "spiffe://")
	if strings.Contains(trimmed, "..") || strings.Contains(trimmed, "//") {
		return fmt.Errorf("invalid SPIFFE ID format: contains path traversal or double slashes")
	}

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

	path := strings.TrimPrefix(u.Path, "/")
	idx := strings.IndexByte(path, '/')
	if idx == -1 || path[:idx] != "agent" || len(path[idx+1:]) == 0 {
		return fmt.Errorf("invalid SPIFFE ID path structure: %s", u.Path)
	}

	return nil
}


// ExecuteHandoff delegates a task from one agent to another by bridging their local states.
// Accepts parameters: adapter UniversalAdapter (No Constraints), ctx context.Context, req *domain.Message, targetID string.
// Returns (domain.Message, error).
func ExecuteHandoff(ctx context.Context, adapter UniversalAdapter, req *domain.Message, targetID string) (domain.Message, error) {
	if req == nil {
		return domain.Message{}, fmt.Errorf("handoff request cannot be nil")
	}
	if targetID == "" {
		return domain.Message{}, fmt.Errorf("target agent ID cannot be empty")
	}

	res := domain.Message{
		ID:         req.ID + "-handoff",
		FromAgent:  req.FromAgent,
		ToAgent:    targetID,
		Type:       domain.EventHandoff,
		Content:    req.Content,
		MeetingID:  req.MeetingID,
		OccurredAt: req.OccurredAt,
	}

	return res, nil
}
