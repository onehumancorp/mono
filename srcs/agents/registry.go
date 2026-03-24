package agents

import (
	"errors"
	"sync"
)

// Registry manages the set of available agent providers and their stored credentials.  Providers are registered once (typically at startup via DefaultRegistry) and then authenticated on-demand through the dashboard API.  Multiple goroutines may call Registry methods concurrently.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderType]Provider
}

// NewRegistry returns an empty Registry.  Use DefaultRegistry to get a pre-populated instance with all built-in providers.
// Parameters: None
// Returns: *Registry
// Errors: None
// Side Effects: None
func NewRegistry() *Registry {
	return &Registry{providers: make(map[ProviderType]Provider)}
}

// DefaultRegistry returns a Registry pre-populated with every built-in provider.  This is the standard factory used by the dashboard server at startup.
// Parameters: None
// Returns: *Registry
// Errors: None
// Side Effects: None
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

// Register adds a Provider to the Registry, overwriting any previously registered provider with the same type.
// Parameters: r *Registry (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Type()] = p
}

// Get returns the Provider for the given type, or false if not found.
// Parameters: r *Registry (No Constraints)
// Returns: (Provider, bool)
// Errors: None
// Side Effects: None
func (r *Registry) Get(t ProviderType) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[t]
	return p, ok
}

// All returns a snapshot of all registered providers, ordered by type string.
// Parameters: None
// Returns: []Provider
// Errors: None
// Side Effects: None
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

// Authenticate forwards credentials to the named provider.  Returns an error if the provider type is unknown or if the provider rejects the supplied credentials.
// Parameters: r *Registry (No Constraints)
// Returns: error
// Errors: Explicit error handling
// Side Effects: None
func (r *Registry) Authenticate(t ProviderType, creds Credentials) error {
	r.mu.RLock()
	p, ok := r.providers[t]
	r.mu.RUnlock()
	if !ok {
		return errors.New("unknown provider type: " + string(t))
	}
	return p.Authenticate(creds)
}

// ProviderInfo is a serializable summary of a provider used by the dashboard API.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ProviderInfo struct {
	Type            ProviderType `json:"type"`
	Description     string       `json:"description"`
	SupportedRoles  []string     `json:"supportedRoles"`
	IsAuthenticated bool         `json:"isAuthenticated"`
}

// Infos returns a ProviderInfo summary for every registered provider.
// Parameters: None
// Returns: []ProviderInfo
// Errors: None
// Side Effects: None
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
