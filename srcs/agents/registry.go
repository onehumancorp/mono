package agents

import (
	"errors"
	"sync"
)

// Summary: Defines the Registry type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderType]Provider
}

// Summary: NewRegistry functionality.
// Parameters: None
// Returns: *Registry
// Errors: None
// Side Effects: None
func NewRegistry() *Registry {
	return &Registry{providers: make(map[ProviderType]Provider)}
}

// Summary: DefaultRegistry functionality.
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

// Summary: Register functionality.
// Parameters: p
// Returns: None
// Errors: None
// Side Effects: None
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Type()] = p
}

// Summary: Get functionality.
// Parameters: t
// Returns: (Provider, bool)
// Errors: None
// Side Effects: None
func (r *Registry) Get(t ProviderType) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[t]
	return p, ok
}

// Summary: All functionality.
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

// Summary: Authenticate functionality.
// Parameters: t, creds
// Returns: error
// Errors: Returns an error if applicable
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

// Summary: Defines the ProviderInfo type.
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

// Summary: Infos functionality.
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
