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

// ProviderType is the unique identifier for an external agent implementation.
// Summary: ProviderType functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type ProviderType string

const (
	// ProviderTypeClaude targets Anthropic Claude Code (claude.ai/code).
	// Best suited for software-engineering and security-review roles.
	ProviderTypeClaude ProviderType = "claude"

	// ProviderTypeGemini targets Google Gemini CLI.
	// Best suited for product-management, analytics, and assistant roles.
	ProviderTypeGemini ProviderType = "gemini"

	// ProviderTypeOpenCode targets the open-source OpenCode SWE agent.
	// Best suited for software-engineering roles.
	ProviderTypeOpenCode ProviderType = "opencode"

	// ProviderTypeOpenClaw targets the OpenClaw assistant agent.
	// Best suited for assistant and content-strategy roles.
	ProviderTypeOpenClaw ProviderType = "openclaw"

	// ProviderTypeIronClaw targets the IronClaw agent.
	// Best suited for security, audit, and QA roles.
	ProviderTypeIronClaw ProviderType = "ironclaw"

	// ProviderTypeBuiltin is the platform's own lightweight agent implementation.
	// Suitable for any role when no external provider is required.
	ProviderTypeBuiltin ProviderType = "builtin"
)

// Credentials holds the authentication material for an external agent provider.
// Providers may use an API key, an OAuth bearer token, or both alongside
// any additional provider-specific configuration entries.
// Summary: Credentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type Credentials struct {
	APIKey     string            `json:"apiKey,omitempty"`
	OAuthToken string            `json:"oauthToken,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// IsEmpty returns true when no authentication material has been set.
// Summary: IsEmpty functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (c Credentials) IsEmpty() bool {
	return c.APIKey == "" && c.OAuthToken == ""
}

// Provider is the interface every external agent implementation must satisfy.
//
// Implementations are registered with a Registry and selected by name when
// an agent is hired through the dashboard API.
// Summary: Provider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// ClaudeProvider implements Provider for Anthropic Claude Code.
// Summary: ClaudeProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type ClaudeProvider struct{ baseProvider }

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) Type() ProviderType { return ProviderTypeClaude }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) Description() string {
	return "Anthropic Claude Code — advanced coding and reasoning agent backed by Claude Sonnet/Opus"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "SECURITY_ENGINEER", "QA_TESTER", "ENGINEERING_DIRECTOR"}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("claude provider requires an API key (ANTHROPIC_API_KEY)")
	}
	p.store(creds)
	return nil
}
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *ClaudeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Gemini (Google) ───────────────────────────────────────────────────────────

// GeminiProvider implements Provider for Google Gemini CLI.
// Summary: GeminiProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type GeminiProvider struct{ baseProvider }

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) Type() ProviderType { return ProviderTypeGemini }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) Description() string {
	return "Google Gemini CLI — multimodal assistant agent backed by Gemini Pro/Ultra"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) SupportedRoles() []string {
	return []string{"PRODUCT_MANAGER", "ANALYTICS_ENGINEER", "MARKETING_MANAGER", "CEO"}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" && creds.OAuthToken == "" {
		return errors.New("gemini provider requires an API key (GEMINI_API_KEY) or an OAuth token")
	}
	p.store(creds)
	return nil
}
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *GeminiProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenCode ──────────────────────────────────────────────────────────────────

// OpenCodeProvider implements Provider for the open-source OpenCode SWE agent.
// Summary: OpenCodeProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type OpenCodeProvider struct{ baseProvider }

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) Type() ProviderType { return ProviderTypeOpenCode }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) Description() string {
	return "OpenCode — open-source software-engineering agent with full terminal and file-system access"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR", "QA_TESTER"}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("opencode provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenCodeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenClaw ──────────────────────────────────────────────────────────────────

// OpenClawProvider implements Provider for the OpenClaw assistant agent.
// Summary: OpenClawProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type OpenClawProvider struct{ baseProvider }

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) Type() ProviderType { return ProviderTypeOpenClaw }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) Description() string {
	return "OpenClaw — general-purpose assistant agent optimised for content strategy and growth tasks"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) SupportedRoles() []string {
	return []string{"GROWTH_AGENT", "CONTENT_STRATEGIST", "MARKETING_MANAGER", "PRODUCT_MANAGER"}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("openclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *OpenClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── IronClaw ──────────────────────────────────────────────────────────────────

// IronClawProvider implements Provider for the IronClaw agent.
// Summary: IronClawProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type IronClawProvider struct{ baseProvider }

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) Type() ProviderType { return ProviderTypeIronClaw }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) Description() string {
	return "IronClaw — security and audit-focused agent with deep static-analysis capabilities"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) SupportedRoles() []string {
	return []string{"SECURITY_ENGINEER", "AUDIT_MANAGER", "QA_TESTER"}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("ironclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *IronClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Builtin ───────────────────────────────────────────────────────────────────

// BuiltinProvider implements Provider for the platform's own lightweight agent.
// It requires no external credentials and is always considered authenticated.
// Summary: BuiltinProvider functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type BuiltinProvider struct{}

// Type provides functionality for Type.
// Summary: Type functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) Type() ProviderType { return ProviderTypeBuiltin }
// Description provides functionality for Description.
// Summary: Description functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) Description() string {
	return "Built-in — platform-native agent; no external credentials required"
}
// SupportedRoles provides functionality for SupportedRoles.
// Summary: SupportedRoles functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) SupportedRoles() []string {
	return []string{
		"CEO", "PRODUCT_MANAGER", "SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR",
		"QA_TESTER", "SECURITY_ENGINEER", "DESIGNER", "MARKETING_MANAGER",
		"GROWTH_AGENT", "CONTENT_STRATEGIST", "SEO_SPECIALIST", "PAID_MEDIA_MANAGER",
		"ANALYTICS_ENGINEER", "CFO", "BOOKKEEPER", "TAX_SPECIALIST",
		"AUDIT_MANAGER", "PAYROLL_MANAGER",
	}
}
// Authenticate provides functionality for Authenticate.
// Summary: Authenticate functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) Authenticate(_ Credentials) error { return nil }
// GetCredentials provides functionality for GetCredentials.
// Summary: GetCredentials functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) GetCredentials() Credentials      { return Credentials{} }
// IsAuthenticated provides functionality for IsAuthenticated.
// Summary: IsAuthenticated functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (p *BuiltinProvider) IsAuthenticated() bool            { return true }
