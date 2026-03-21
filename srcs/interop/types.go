package interop

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Summary: State represents shared agent state.
// Intent: State represents shared agent state.
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
// Intent: UniversalAdapter defines the interface for interacting with different agent frameworks.
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

// Summary: ValidateSPIFFEID validates an agent's SPIFFE identity.
// Intent: ValidateSPIFFEID validates an agent's SPIFFE identity.
// Params: identity string
// Returns: error
// Errors: Returns error for invalid format, scheme, domain, or path length.
// Side Effects: None
func ValidateSPIFFEID(identity string) error {
	if identity == "" {
		return fmt.Errorf("identity cannot be empty")
	}

	u, err := url.Parse(identity)
	if err != nil {
		return fmt.Errorf("invalid SPIFFE ID format: %w", err)
	}

	if u.Scheme != "spiffe" {
		return fmt.Errorf("invalid SPIFFE ID scheme: expected 'spiffe', got %q", u.Scheme)
	}

	validDomains := map[string]bool{
		"onehumancorp.io": true,
		"ohc.local":       true,
		"ohc.os":          true,
	}

	if !validDomains[u.Host] {
		return fmt.Errorf("invalid SPIFFE ID trust domain: %q", u.Host)
	}

	if u.Path == "" || u.Path == "/" {
		return fmt.Errorf("invalid SPIFFE ID: missing path")
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) < 2 {
		return fmt.Errorf("invalid SPIFFE ID path: expected at least 2 segments (e.g., /agent/name)")
	}

    for _, segment := range segments {
        if len(segment) > 255 {
            return fmt.Errorf("invalid SPIFFE ID path: segment too long")
        }
    }

	return nil
}
