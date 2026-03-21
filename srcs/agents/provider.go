// Package agents defines the extensible provider interface for external AI agent implementations.
//
// It allows the platform to "hire" agents backed by well-known coding and assistant tools
// (Claude Code, Gemini CLI, OpenCode, OpenClaw, IronClaw) or by the built-in simple agent,
// while storing per-provider credentials so users authenticate once per provider and the
// platform forwards the auth to every agent of that type.
package agents

import (
	"errors"
	"sync"
)

// Summary: Defines the ProviderType type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ProviderType string

const (
	// Summary: Defines the ProviderTypeClaude type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeClaude ProviderType = "claude"

	// Summary: Defines the ProviderTypeGemini type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeGemini ProviderType = "gemini"

	// Summary: Defines the ProviderTypeOpenCode type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeOpenCode ProviderType = "opencode"

	// Summary: Defines the ProviderTypeOpenClaw type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeOpenClaw ProviderType = "openclaw"

	// Summary: Defines the ProviderTypeIronClaw type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeIronClaw ProviderType = "ironclaw"

	// Summary: Defines the ProviderTypeBuiltin type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeBuiltin ProviderType = "builtin"
)

// Summary: Defines the Credentials type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Credentials struct {
	APIKey     string            `json:"apiKey,omitempty"`
	OAuthToken string            `json:"oauthToken,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// Summary: IsEmpty functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (c Credentials) IsEmpty() bool {
	return c.APIKey == "" && c.OAuthToken == ""
}

// Summary: Defines the Provider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Provider interface {
	// Type returns the unique identifier for this provider.
	Type() ProviderType

	// Description returns a short human-readable explanation of the provider.
	Description() string

	// SupportedRoles lists the domain Role constants this provider is
	// optimised for.  The list is informational; the platform does not
	// prevent other roles from using the provider.
	SupportedRoles() []string

	// Authenticate validates and stores the supplied credentials.
	// Returns an error if the credentials are structurally invalid.
	Authenticate(creds Credentials) error

	// GetCredentials returns the currently stored credentials.
	// Secret fields (APIKey, OAuthToken) are returned as-is so that the
	// platform can forward them to the spawned agent process; callers that
	// render credentials to end-users should redact these fields.
	GetCredentials() Credentials

	// IsAuthenticated reports whether valid credentials are currently stored.
	IsAuthenticated() bool
}

// baseProvider is an embeddable helper that manages credential storage for
// any Provider implementation.
type baseProvider struct {
	mu   sync.RWMutex
	cred Credentials
}

func (b *baseProvider) store(cred Credentials) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cred = cred
}

func (b *baseProvider) load() Credentials {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cred
}

// ── Claude (Anthropic) ────────────────────────────────────────────────────────

// Summary: Defines the ClaudeProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ClaudeProvider struct{ baseProvider }

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *ClaudeProvider) Type() ProviderType { return ProviderTypeClaude }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *ClaudeProvider) Description() string {
	return "Anthropic Claude Code — advanced coding and reasoning agent backed by Claude Sonnet/Opus"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *ClaudeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "SECURITY_ENGINEER", "QA_TESTER", "ENGINEERING_DIRECTOR"}
}
// Summary: Authenticate functionality.
// Parameters: creds
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *ClaudeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("claude provider requires an API key (ANTHROPIC_API_KEY)")
	}
	p.store(creds)
	return nil
}
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *ClaudeProvider) GetCredentials() Credentials { return p.load() }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *ClaudeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Gemini (Google) ───────────────────────────────────────────────────────────

// Summary: Defines the GeminiProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type GeminiProvider struct{ baseProvider }

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *GeminiProvider) Type() ProviderType { return ProviderTypeGemini }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *GeminiProvider) Description() string {
	return "Google Gemini CLI — multimodal assistant agent backed by Gemini Pro/Ultra"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *GeminiProvider) SupportedRoles() []string {
	return []string{"PRODUCT_MANAGER", "ANALYTICS_ENGINEER", "MARKETING_MANAGER", "CEO"}
}
// Summary: Authenticate functionality.
// Parameters: creds
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *GeminiProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" && creds.OAuthToken == "" {
		return errors.New("gemini provider requires an API key (GEMINI_API_KEY) or an OAuth token")
	}
	p.store(creds)
	return nil
}
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *GeminiProvider) GetCredentials() Credentials { return p.load() }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *GeminiProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenCode ──────────────────────────────────────────────────────────────────

// Summary: Defines the OpenCodeProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type OpenCodeProvider struct{ baseProvider }

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *OpenCodeProvider) Type() ProviderType { return ProviderTypeOpenCode }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *OpenCodeProvider) Description() string {
	return "OpenCode — open-source software-engineering agent with full terminal and file-system access"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *OpenCodeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR", "QA_TESTER"}
}
// Summary: Authenticate functionality.
// Parameters: creds
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *OpenCodeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("opencode provider requires an API key")
	}
	p.store(creds)
	return nil
}
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *OpenCodeProvider) GetCredentials() Credentials { return p.load() }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *OpenCodeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenClaw ──────────────────────────────────────────────────────────────────

// Summary: Defines the OpenClawProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type OpenClawProvider struct{ baseProvider }

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *OpenClawProvider) Type() ProviderType { return ProviderTypeOpenClaw }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *OpenClawProvider) Description() string {
	return "OpenClaw — general-purpose assistant agent optimised for content strategy and growth tasks"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *OpenClawProvider) SupportedRoles() []string {
	return []string{"GROWTH_AGENT", "CONTENT_STRATEGIST", "MARKETING_MANAGER", "PRODUCT_MANAGER"}
}
// Summary: Authenticate functionality.
// Parameters: creds
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *OpenClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("openclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *OpenClawProvider) GetCredentials() Credentials { return p.load() }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *OpenClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── IronClaw ──────────────────────────────────────────────────────────────────

// Summary: Defines the IronClawProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type IronClawProvider struct{ baseProvider }

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *IronClawProvider) Type() ProviderType { return ProviderTypeIronClaw }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *IronClawProvider) Description() string {
	return "IronClaw — security and audit-focused agent with deep static-analysis capabilities"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *IronClawProvider) SupportedRoles() []string {
	return []string{"SECURITY_ENGINEER", "AUDIT_MANAGER", "QA_TESTER"}
}
// Summary: Authenticate functionality.
// Parameters: creds
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *IronClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("ironclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *IronClawProvider) GetCredentials() Credentials { return p.load() }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *IronClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Builtin ───────────────────────────────────────────────────────────────────

// Summary: Defines the BuiltinProvider type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type BuiltinProvider struct{}

// Summary: Type functionality.
// Parameters: None
// Returns: ProviderType
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) Type() ProviderType { return ProviderTypeBuiltin }
// Summary: Description functionality.
// Parameters: None
// Returns: string
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) Description() string {
	return "Built-in — platform-native agent; no external credentials required"
}
// Summary: SupportedRoles functionality.
// Parameters: None
// Returns: []string
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) SupportedRoles() []string {
	return []string{
		"CEO", "PRODUCT_MANAGER", "SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR",
		"QA_TESTER", "SECURITY_ENGINEER", "DESIGNER", "MARKETING_MANAGER",
		"GROWTH_AGENT", "CONTENT_STRATEGIST", "SEO_SPECIALIST", "PAID_MEDIA_MANAGER",
		"ANALYTICS_ENGINEER", "CFO", "BOOKKEEPER", "TAX_SPECIALIST",
		"AUDIT_MANAGER", "PAYROLL_MANAGER",
	}
}
// Summary: Authenticate functionality.
// Parameters: _
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (p *BuiltinProvider) Authenticate(_ Credentials) error { return nil }
// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) GetCredentials() Credentials      { return Credentials{} }
// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) IsAuthenticated() bool            { return true }
