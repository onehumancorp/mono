package agents

import (
	"errors"
	"sync"
)

// Registry Intent: Registry manages the set of available agent providers and their stored credentials.  Providers are registered once (typically at startup via DefaultRegistry) and then authenticated on-demand through the dashboard API.  Multiple goroutines may call Registry methods concurrently.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderType]Provider
}

// NewRegistry Intent: NewRegistry returns an empty Registry.  Use DefaultRegistry to get a pre-populated instance with all built-in providers.
//
// Params: None.
//
// Returns:
//   - *Registry: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[ProviderType]Provider)}
}

// DefaultRegistry Intent: DefaultRegistry returns a Registry pre-populated with every built-in provider.  This is the standard factory used by the dashboard server at startup.
//
// Params: None.
//
// Returns:
//   - *Registry: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(&ClaudeProvider{})
	r.Register(&GeminiProvider{})
	r.Register(&OpenCodeProvider{})
	r.Register(&OpenClawProvider{})
	r.Register(&IronClawProvider{})
	r.Register(&BuiltinProvider{})
	return r
}

// Register Intent: Register adds a Provider to the Registry, overwriting any previously registered provider with the same type.
//
// Params:
//   - p: parameter inferred from signature.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Type()] = p
}

// Get Intent: Get returns the Provider for the given type, or false if not found.
//
// Params:
//   - t: parameter inferred from signature.
//
// Returns:
//   - Provider: return value inferred from signature.
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (r *Registry) Get(t ProviderType) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[t]
	return p, ok
}

// All Intent: All returns a snapshot of all registered providers, ordered by type string.
//
// Params: None.
//
// Returns:
//   - []Provider: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Provider, 0, len(r.providers))
	// Stable ordering: iterate known ordering list; unknown types appended last.
	ordered := []ProviderType{
		ProviderTypeClaude,
		ProviderTypeGemini,
		ProviderTypeOpenCode,
		ProviderTypeOpenClaw,
		ProviderTypeIronClaw,
		ProviderTypeBuiltin,
	}
	seen := map[ProviderType]bool{}
	for _, t := range ordered {
		if p, ok := r.providers[t]; ok {
			out = append(out, p)
			seen[t] = true
		}
	}
	for t, p := range r.providers {
		if !seen[t] {
			out = append(out, p)
		}
	}
	return out
}

// Authenticate Intent: Authenticate forwards credentials to the named provider.  Returns an error if the provider type is unknown or if the provider rejects the supplied credentials.
//
// Params:
//   - t: parameter inferred from signature.
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (r *Registry) Authenticate(t ProviderType, creds Credentials) error {
	r.mu.RLock()
	p, ok := r.providers[t]
	r.mu.RUnlock()
	if !ok {
		return errors.New("unknown provider type: " + string(t))
	}
	return p.Authenticate(creds)
}

// ProviderInfo Intent: ProviderInfo is a serializable summary of a provider used by the dashboard API.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type ProviderInfo struct {
	Type            ProviderType `json:"type"`
	Description     string       `json:"description"`
	SupportedRoles  []string     `json:"supportedRoles"`
	IsAuthenticated bool         `json:"isAuthenticated"`
}

// Infos Intent: Infos returns a ProviderInfo summary for every registered provider.
//
// Params: None.
//
// Returns:
//   - []ProviderInfo: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (r *Registry) Infos() []ProviderInfo {
	providers := r.All()
	out := make([]ProviderInfo, len(providers))
	for i, p := range providers {
		out[i] = ProviderInfo{
			Type:            p.Type(),
			Description:     p.Description(),
			SupportedRoles:  p.SupportedRoles(),
			IsAuthenticated: p.IsAuthenticated(),
		}
	}
	return out
}
