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

// Summary: ProviderType is the unique identifier for an external agent implementation.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ProviderType string

const (
	// Summary: ProviderTypeClaude targets Anthropic Claude Code (claude.ai/code). Best suited for software-engineering and security-review roles.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeClaude ProviderType = "claude"

	// Summary: ProviderTypeGemini targets Google Gemini CLI. Best suited for product-management, analytics, and assistant roles.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeGemini ProviderType = "gemini"

	// Summary: ProviderTypeOpenCode targets the open-source OpenCode SWE agent. Best suited for software-engineering roles.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeOpenCode ProviderType = "opencode"

	// Summary: ProviderTypeOpenClaw targets the OpenClaw assistant agent. Best suited for assistant and content-strategy roles.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeOpenClaw ProviderType = "openclaw"

	// Summary: ProviderTypeIronClaw targets the IronClaw agent. Best suited for security, audit, and QA roles.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeIronClaw ProviderType = "ironclaw"

	// Summary: ProviderTypeBuiltin is the platform's own lightweight agent implementation. Suitable for any role when no external provider is required.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	ProviderTypeBuiltin ProviderType = "builtin"
)

// Summary: Credentials holds the authentication material for an external agent provider. Providers may use an API key, an OAuth bearer token, or both alongside any additional provider-specific configuration entries.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Credentials struct {
	APIKey     string            `json:"apiKey,omitempty"`
	OAuthToken string            `json:"oauthToken,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// Summary: IsEmpty returns true when no authentication material has been set.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (c Credentials) IsEmpty() bool {
	return c.APIKey == "" && c.OAuthToken == ""
}

// Summary: Provider is the interface every external agent implementation must satisfy.  Implementations are registered with a Registry and selected by name when an agent is hired through the dashboard API.
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

// Summary: ClaudeProvider implements Provider for Anthropic Claude Code.
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
// Parameters: p *ClaudeProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
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
func (p *ClaudeProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── Gemini (Google) ───────────────────────────────────────────────────────────

// Summary: GeminiProvider implements Provider for Google Gemini CLI.
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
// Parameters: p *GeminiProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
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
func (p *GeminiProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── OpenCode ──────────────────────────────────────────────────────────────────

// Summary: OpenCodeProvider implements Provider for the open-source OpenCode SWE agent.
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
// Parameters: p *OpenCodeProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
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
func (p *OpenCodeProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── OpenClaw ──────────────────────────────────────────────────────────────────

// Summary: OpenClawProvider implements Provider for the OpenClaw assistant agent.
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
// Parameters: p *OpenClawProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
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
func (p *OpenClawProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── IronClaw ──────────────────────────────────────────────────────────────────

// Summary: IronClawProvider implements Provider for the IronClaw agent.
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
// Parameters: p *IronClawProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
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
func (p *IronClawProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── Builtin ───────────────────────────────────────────────────────────────────

// Summary: BuiltinProvider implements Provider for the platform's own lightweight agent. It requires no external credentials and is always considered authenticated.
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
// Parameters: p *BuiltinProvider (No Constraints)
// Returns: error
// Errors: Explicit error handling
// Side Effects: None
func (p *BuiltinProvider) Authenticate(_ Credentials) error { return nil }

// Summary: GetCredentials functionality.
// Parameters: None
// Returns: Credentials
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) GetCredentials() Credentials { return Credentials{} }

// Summary: IsAuthenticated functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func (p *BuiltinProvider) IsAuthenticated() bool { return true }
