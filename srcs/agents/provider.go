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

// ProviderType Intent: ProviderType is the unique identifier for an external agent implementation.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
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

// Credentials Intent: Credentials holds the authentication material for an external agent provider. Providers may use an API key, an OAuth bearer token, or both alongside any additional provider-specific configuration entries.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Credentials struct {
	APIKey     string            `json:"apiKey,omitempty"`
	OAuthToken string            `json:"oauthToken,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// IsEmpty Intent: IsEmpty returns true when no authentication material has been set.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (c Credentials) IsEmpty() bool {
	return c.APIKey == "" && c.OAuthToken == ""
}

// Provider Intent: Provider is the interface every external agent implementation must satisfy.  Implementations are registered with a Registry and selected by name when an agent is hired through the dashboard API.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
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

// ClaudeProvider Intent: ClaudeProvider implements Provider for Anthropic Claude Code.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type ClaudeProvider struct{ baseProvider }
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) Type() ProviderType { return ProviderTypeClaude }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) Description() string {
	return "Anthropic Claude Code — advanced coding and reasoning agent backed by Claude Sonnet/Opus"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "SECURITY_ENGINEER", "QA_TESTER", "ENGINEERING_DIRECTOR"}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("claude provider requires an API key (ANTHROPIC_API_KEY)")
	}
	p.store(creds)
	return nil
}
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *ClaudeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Gemini (Google) ───────────────────────────────────────────────────────────

// GeminiProvider Intent: GeminiProvider implements Provider for Google Gemini CLI.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type GeminiProvider struct{ baseProvider }
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) Type() ProviderType { return ProviderTypeGemini }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) Description() string {
	return "Google Gemini CLI — multimodal assistant agent backed by Gemini Pro/Ultra"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) SupportedRoles() []string {
	return []string{"PRODUCT_MANAGER", "ANALYTICS_ENGINEER", "MARKETING_MANAGER", "CEO"}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" && creds.OAuthToken == "" {
		return errors.New("gemini provider requires an API key (GEMINI_API_KEY) or an OAuth token")
	}
	p.store(creds)
	return nil
}
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *GeminiProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenCode ──────────────────────────────────────────────────────────────────

// OpenCodeProvider Intent: OpenCodeProvider implements Provider for the open-source OpenCode SWE agent.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type OpenCodeProvider struct{ baseProvider }
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) Type() ProviderType { return ProviderTypeOpenCode }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) Description() string {
	return "OpenCode — open-source software-engineering agent with full terminal and file-system access"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) SupportedRoles() []string {
	return []string{"SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR", "QA_TESTER"}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("opencode provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenCodeProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── OpenClaw ──────────────────────────────────────────────────────────────────

// OpenClawProvider Intent: OpenClawProvider implements Provider for the OpenClaw assistant agent.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type OpenClawProvider struct{ baseProvider }
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) Type() ProviderType { return ProviderTypeOpenClaw }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) Description() string {
	return "OpenClaw — general-purpose assistant agent optimised for content strategy and growth tasks"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) SupportedRoles() []string {
	return []string{"GROWTH_AGENT", "CONTENT_STRATEGIST", "MARKETING_MANAGER", "PRODUCT_MANAGER"}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("openclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *OpenClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── IronClaw ──────────────────────────────────────────────────────────────────

// IronClawProvider Intent: IronClawProvider implements Provider for the IronClaw agent.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type IronClawProvider struct{ baseProvider }
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) Type() ProviderType { return ProviderTypeIronClaw }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) Description() string {
	return "IronClaw — security and audit-focused agent with deep static-analysis capabilities"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) SupportedRoles() []string {
	return []string{"SECURITY_ENGINEER", "AUDIT_MANAGER", "QA_TESTER"}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - creds: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) Authenticate(creds Credentials) error {
	if creds.APIKey == "" {
		return errors.New("ironclaw provider requires an API key")
	}
	p.store(creds)
	return nil
}
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) GetCredentials() Credentials { return p.load() }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *IronClawProvider) IsAuthenticated() bool       { return !p.load().IsEmpty() }

// ── Builtin ───────────────────────────────────────────────────────────────────

// BuiltinProvider Intent: BuiltinProvider implements Provider for the platform's own lightweight agent. It requires no external credentials and is always considered authenticated.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type BuiltinProvider struct{}
// Type Intent: Handles operations related to Type.
//
// Params: None.
//
// Returns:
//   - ProviderType: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) Type() ProviderType { return ProviderTypeBuiltin }
// Description Intent: Handles operations related to Description.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) Description() string {
	return "Built-in — platform-native agent; no external credentials required"
}
// SupportedRoles Intent: Handles operations related to SupportedRoles.
//
// Params: None.
//
// Returns:
//   - []string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) SupportedRoles() []string {
	return []string{
		"CEO", "PRODUCT_MANAGER", "SOFTWARE_ENGINEER", "ENGINEERING_DIRECTOR",
		"QA_TESTER", "SECURITY_ENGINEER", "DESIGNER", "MARKETING_MANAGER",
		"GROWTH_AGENT", "CONTENT_STRATEGIST", "SEO_SPECIALIST", "PAID_MEDIA_MANAGER",
		"ANALYTICS_ENGINEER", "CFO", "BOOKKEEPER", "TAX_SPECIALIST",
		"AUDIT_MANAGER", "PAYROLL_MANAGER",
	}
}
// Authenticate Intent: Handles operations related to Authenticate.
//
// Params:
//   - _: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) Authenticate(_ Credentials) error { return nil }
// GetCredentials Intent: Handles operations related to GetCredentials.
//
// Params: None.
//
// Returns:
//   - Credentials: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) GetCredentials() Credentials      { return Credentials{} }
// IsAuthenticated Intent: Handles operations related to IsAuthenticated.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (p *BuiltinProvider) IsAuthenticated() bool            { return true }
