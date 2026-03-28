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
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type ProviderType string

const (
	// ProviderTypeClaude targets Anthropic Claude Code (claude.ai/code). Best suited for software-engineering and security-review roles.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeClaude ProviderType = "claude"

	// ProviderTypeGemini targets Google Gemini CLI. Best suited for product-management, analytics, and assistant roles.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeGemini ProviderType = "gemini"

	// ProviderTypeOpenCode targets the open-source OpenCode SWE agent. Best suited for software-engineering roles.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeOpenCode ProviderType = "opencode"

	// ProviderTypeOpenClaw targets the OpenClaw assistant agent. Best suited for assistant and content-strategy roles.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeOpenClaw ProviderType = "openclaw"

	// ProviderTypeIronClaw targets the IronClaw agent. Best suited for security, audit, and QA roles.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeIronClaw ProviderType = "ironclaw"

	// ProviderTypeMiniMax targets the MiniMax AI platform. Best suited for all roles; uses the MiniMax abab series models.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeMiniMax ProviderType = "minimax"

	// ProviderTypeBuiltin is the platform's own lightweight agent implementation. Suitable for any role when no external provider is required.
	// Accepts no parameters.
	// Returns nothing.
	// Produces no errors.
	// Has no side effects.
	ProviderTypeBuiltin ProviderType = "builtin"
)

// Credentials holds the authentication material for an external agent provider. Providers may use an API key, an OAuth bearer token, or both alongside any additional provider-specific configuration entries.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type Credentials struct {
	APIKey     string            `json:"apiKey,omitempty"`
	OAuthToken string            `json:"oauthToken,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// IsEmpty returns true when no authentication material has been set.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (c Credentials) IsEmpty() bool {
	return c.APIKey == "" && c.OAuthToken == ""
}

// Provider is the interface every external agent implementation must satisfy.  Implementations are registered with a Registry and selected by name when an agent is hired through the dashboard API.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
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
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type ClaudeProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *ClaudeProvider) Type() ProviderType { return ProviderTypeClaude }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *ClaudeProvider) Description() string {
	return "Anthropic Claude Code — advanced coding and reasoning agent backed by Claude Sonnet/Opus"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *ClaudeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "SECURITY_ENGINEER", "QA_TESTER", "ENGINEERING_DIRECTOR"}
}

// Authenticate functionality.
// Accepts parameters: p *ClaudeProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *ClaudeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("claude provider requires an API key (ANTHROPIC_API_KEY)")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *ClaudeProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *ClaudeProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── Gemini (Google) ───────────────────────────────────────────────────────────

// GeminiProvider implements Provider for Google Gemini CLI.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type GeminiProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *GeminiProvider) Type() ProviderType { return ProviderTypeGemini }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *GeminiProvider) Description() string {
	return "Google Gemini CLI — multimodal assistant agent backed by Gemini Pro/Ultra"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *GeminiProvider) SupportedRoles() []string {
	return []string{"PRODUCT_MANAGER", "ANALYTICS_ENGINEER", "MARKETING_MANAGER", "CEO"}
}

// Authenticate functionality.
// Accepts parameters: p *GeminiProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *GeminiProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" && creds.OAuthToken == "" {
		return errors.New("gemini provider requires an API key (GEMINI_API_KEY) or an OAuth token")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *GeminiProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *GeminiProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── OpenCode ──────────────────────────────────────────────────────────────────

// OpenCodeProvider implements Provider for the open-source OpenCode SWE agent.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type OpenCodeProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *OpenCodeProvider) Type() ProviderType { return ProviderTypeOpenCode }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *OpenCodeProvider) Description() string {
	return "OpenCode — open-source software-engineering agent with full terminal and file-system access"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *OpenCodeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR", "QA_TESTER"}
}

// Authenticate functionality.
// Accepts parameters: p *OpenCodeProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *OpenCodeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("opencode provider requires an API key")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *OpenCodeProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *OpenCodeProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── OpenClaw ──────────────────────────────────────────────────────────────────

// OpenClawProvider implements Provider for the OpenClaw assistant agent.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type OpenClawProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *OpenClawProvider) Type() ProviderType { return ProviderTypeOpenClaw }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *OpenClawProvider) Description() string {
	return "OpenClaw — general-purpose assistant agent optimised for content strategy and growth tasks"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *OpenClawProvider) SupportedRoles() []string {
	return []string{"GROWTH_AGENT", "CONTENT_STRATEGIST", "MARKETING_MANAGER", "PRODUCT_MANAGER"}
}

// Authenticate functionality.
// Accepts parameters: p *OpenClawProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *OpenClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("openclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *OpenClawProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *OpenClawProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── IronClaw ──────────────────────────────────────────────────────────────────

// IronClawProvider implements Provider for the IronClaw agent.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type IronClawProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *IronClawProvider) Type() ProviderType { return ProviderTypeIronClaw }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *IronClawProvider) Description() string {
	return "IronClaw — security and audit-focused agent with deep static-analysis capabilities"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *IronClawProvider) SupportedRoles() []string {
	return []string{"SECURITY_ENGINEER", "AUDIT_MANAGER", "QA_TESTER"}
}

// Authenticate functionality.
// Accepts parameters: p *IronClawProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *IronClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("ironclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *IronClawProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *IronClawProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── MiniMax ───────────────────────────────────────────────────────────────────

// MiniMaxProvider implements Provider for the MiniMax AI platform.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type MiniMaxProvider struct{ baseProvider }

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *MiniMaxProvider) Type() ProviderType { return ProviderTypeMiniMax }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *MiniMaxProvider) Description() string {
	return "MiniMax — Chinese multimodal AI platform providing abab series models for all agent roles"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *MiniMaxProvider) SupportedRoles() []string {
	return []string{
		"CEO", "PRODUCT_MANAGER", "SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR",
		"QA_TESTER", "SECURITY_ENGINEER", "DESIGNER", "MARKETING_MANAGER",
		"GROWTH_AGENT", "CONTENT_STRATEGIST", "SEO_SPECIALIST", "PAID_MEDIA_MANAGER",
		"ANALYTICS_ENGINEER", "CFO", "BOOKKEEPER", "TAX_SPECIALIST",
		"AUDIT_MANAGER", "PAYROLL_MANAGER",
	}
}

// Authenticate functionality.
// Accepts parameters: p *MiniMaxProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *MiniMaxProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("minimax provider requires an API key (MINIMAX_API_KEY)")
	}
	p.store(creds)
	return nil
}

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *MiniMaxProvider) GetCredentials() Credentials { return p.load() }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *MiniMaxProvider) IsAuthenticated() bool { return !p.load().IsEmpty() }

// ── Builtin ───────────────────────────────────────────────────────────────────

// BuiltinProvider implements Provider for the platform's own lightweight agent. It requires no external credentials and is always considered authenticated.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type BuiltinProvider struct{}

// Type functionality.
// Accepts no parameters.
// Returns ProviderType.
// Produces no errors.
// Has no side effects.
func (p *BuiltinProvider) Type() ProviderType { return ProviderTypeBuiltin }

// Description functionality.
// Accepts no parameters.
// Returns string.
// Produces no errors.
// Has no side effects.
func (p *BuiltinProvider) Description() string {
	return "Built-in — platform-native agent; no external credentials required"
}

// SupportedRoles functionality.
// Accepts no parameters.
// Returns []string.
// Produces no errors.
// Has no side effects.
func (p *BuiltinProvider) SupportedRoles() []string {
	return []string{
		"CEO", "PRODUCT_MANAGER", "SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR",
		"QA_TESTER", "SECURITY_ENGINEER", "DESIGNER", "MARKETING_MANAGER",
		"GROWTH_AGENT", "CONTENT_STRATEGIST", "SEO_SPECIALIST", "PAID_MEDIA_MANAGER",
		"ANALYTICS_ENGINEER", "CFO", "BOOKKEEPER", "TAX_SPECIALIST",
		"AUDIT_MANAGER", "PAYROLL_MANAGER",
	}
}

// Authenticate functionality.
// Accepts parameters: p *BuiltinProvider (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (p *BuiltinProvider) Authenticate(_ Credentials) error { return nil }

// GetCredentials functionality.
// Accepts no parameters.
// Returns Credentials.
// Produces no errors.
// Has no side effects.
func (p *BuiltinProvider) GetCredentials() Credentials { return Credentials{} }

// IsAuthenticated functionality.
// Accepts no parameters.
// Returns bool.
// Produces no errors.
// Has no side effects.
func (p *BuiltinProvider) IsAuthenticated() bool { return true }
