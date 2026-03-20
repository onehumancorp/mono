package agents

import (
	"testing"
)

// ── Provider tests ─────────────────────────────────────────────────────────────

func TestClaudeProvider_TypeAndDescription(t *testing.T) {
	p := &ClaudeProvider{}
	if p.Type() != ProviderTypeClaude {
		t.Fatalf("expected type %q, got %q", ProviderTypeClaude, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestClaudeProvider_AuthenticateRequiresAPIKey(t *testing.T) {
	p := &ClaudeProvider{}
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Fatal("expected error when authenticating with empty credentials")
	}
	if err := p.Authenticate(Credentials{APIKey: "sk-test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsAuthenticated() {
		t.Fatal("expected provider to be authenticated after valid Authenticate call")
	}
	if p.GetCredentials().APIKey != "sk-test" {
		t.Fatalf("expected stored API key %q, got %q", "sk-test", p.GetCredentials().APIKey)
	}
}

func TestGeminiProvider_AuthenticateAcceptsAPIKeyOrOAuth(t *testing.T) {
	p := &GeminiProvider{}
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Fatal("expected error when authenticating with empty credentials")
	}
	if err := p.Authenticate(Credentials{APIKey: "AIza-test"}); err != nil {
		t.Fatalf("unexpected error with API key: %v", err)
	}
	p2 := &GeminiProvider{}
	if err := p2.Authenticate(Credentials{OAuthToken: "ya29.token"}); err != nil {
		t.Fatalf("unexpected error with OAuth token: %v", err)
	}
	if !p2.IsAuthenticated() {
		t.Fatal("expected provider to be authenticated after OAuth token")
	}
}

func TestOpenCodeProvider_AuthenticateRequiresAPIKey(t *testing.T) {
	p := &OpenCodeProvider{}
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Fatal("expected error when no key supplied")
	}
	if err := p.Authenticate(Credentials{APIKey: "oc-key"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsAuthenticated() {
		t.Fatal("expected authenticated after key set")
	}
}

func TestOpenClawProvider_AuthenticateRequiresAPIKey(t *testing.T) {
	p := &OpenClawProvider{}
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Fatal("expected error when no key supplied")
	}
	if err := p.Authenticate(Credentials{APIKey: "claw-key"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIronClawProvider_AuthenticateRequiresAPIKey(t *testing.T) {
	p := &IronClawProvider{}
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Fatal("expected error when no key supplied")
	}
	if err := p.Authenticate(Credentials{APIKey: "iron-key"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuiltinProvider_AlwaysAuthenticated(t *testing.T) {
	p := &BuiltinProvider{}
	if !p.IsAuthenticated() {
		t.Fatal("builtin provider should always be authenticated")
	}
	if err := p.Authenticate(Credentials{}); err != nil {
		t.Fatalf("builtin provider should accept any credentials: %v", err)
	}
	creds := p.GetCredentials()
	if creds.APIKey != "" || creds.OAuthToken != "" {
		t.Fatal("builtin provider should return empty credentials")
	}
}

func TestProviderSupportedRolesNotEmpty(t *testing.T) {
	providers := []Provider{
		&ClaudeProvider{},
		&GeminiProvider{},
		&OpenCodeProvider{},
		&OpenClawProvider{},
		&IronClawProvider{},
		&BuiltinProvider{},
	}
	for _, p := range providers {
		if len(p.SupportedRoles()) == 0 {
			t.Fatalf("provider %q should have at least one supported role", p.Type())
		}
	}
}

func TestCredentials_IsEmpty(t *testing.T) {
	if !(Credentials{}).IsEmpty() {
		t.Fatal("empty credentials should return IsEmpty() == true")
	}
	if (Credentials{APIKey: "k"}).IsEmpty() {
		t.Fatal("credentials with API key should not be empty")
	}
	if (Credentials{OAuthToken: "t"}).IsEmpty() {
		t.Fatal("credentials with OAuth token should not be empty")
	}
}

// ── Registry tests ─────────────────────────────────────────────────────────────

func TestDefaultRegistryContainsAllProviders(t *testing.T) {
	r := DefaultRegistry()
	expected := []ProviderType{
		ProviderTypeClaude,
		ProviderTypeGemini,
		ProviderTypeOpenCode,
		ProviderTypeOpenClaw,
		ProviderTypeIronClaw,
		ProviderTypeBuiltin,
	}
	for _, pt := range expected {
		if _, ok := r.Get(pt); !ok {
			t.Fatalf("expected provider %q to be registered in default registry", pt)
		}
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	p := &ClaudeProvider{}
	r.Register(p)

	got, ok := r.Get(ProviderTypeClaude)
	if !ok {
		t.Fatal("expected registered provider to be found")
	}
	if got.Type() != ProviderTypeClaude {
		t.Fatalf("unexpected provider type: %s", got.Type())
	}
	if _, ok := r.Get(ProviderTypeGemini); ok {
		t.Fatal("expected unregistered provider to be missing")
	}
}

func TestRegistry_All_ReturnsStableOrder(t *testing.T) {
	r := DefaultRegistry()
	all := r.All()
	if len(all) < 6 {
		t.Fatalf("expected at least 6 providers, got %d", len(all))
	}
	// First provider should be Claude by stable ordering convention.
	if all[0].Type() != ProviderTypeClaude {
		t.Fatalf("expected first provider to be %q, got %q", ProviderTypeClaude, all[0].Type())
	}
}

func TestRegistry_Authenticate_UnknownProvider(t *testing.T) {
	r := NewRegistry()
	if err := r.Authenticate("unknown", Credentials{APIKey: "key"}); err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

func TestRegistry_Authenticate_ForwardsCredentials(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClaudeProvider{})

	if err := r.Authenticate(ProviderTypeClaude, Credentials{}); err == nil {
		t.Fatal("expected error when authenticating without API key")
	}
	if err := r.Authenticate(ProviderTypeClaude, Credentials{APIKey: "sk-abc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := r.Get(ProviderTypeClaude)
	if !p.IsAuthenticated() {
		t.Fatal("expected provider to be authenticated after successful call")
	}
}

func TestRegistry_Infos_ContainsAllFields(t *testing.T) {
	r := DefaultRegistry()
	infos := r.Infos()
	if len(infos) < 6 {
		t.Fatalf("expected at least 6 provider infos, got %d", len(infos))
	}
	for _, info := range infos {
		if info.Type == "" {
			t.Error("ProviderInfo.Type must not be empty")
		}
		if info.Description == "" {
			t.Errorf("ProviderInfo.Description must not be empty for %q", info.Type)
		}
		if len(info.SupportedRoles) == 0 {
			t.Errorf("ProviderInfo.SupportedRoles must not be empty for %q", info.Type)
		}
	}
	// Builtin provider is always authenticated.
	for _, info := range infos {
		if info.Type == ProviderTypeBuiltin && !info.IsAuthenticated {
			t.Fatal("builtin provider should always report IsAuthenticated = true")
		}
	}
}
