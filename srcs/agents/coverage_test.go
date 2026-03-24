package agents

import (
	"sync"
	"testing"
)

// ── Additional provider coverage ──────────────────────────────────────────────

func TestGeminiProvider_TypeAndDescription(t *testing.T) {
	p := &GeminiProvider{}
	if p.Type() != ProviderTypeGemini {
		t.Fatalf("expected type %q, got %q", ProviderTypeGemini, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestOpenCodeProvider_TypeAndDescription(t *testing.T) {
	p := &OpenCodeProvider{}
	if p.Type() != ProviderTypeOpenCode {
		t.Fatalf("expected type %q, got %q", ProviderTypeOpenCode, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestOpenClawProvider_TypeAndDescription(t *testing.T) {
	p := &OpenClawProvider{}
	if p.Type() != ProviderTypeOpenClaw {
		t.Fatalf("expected type %q, got %q", ProviderTypeOpenClaw, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestIronClawProvider_TypeAndDescription(t *testing.T) {
	p := &IronClawProvider{}
	if p.Type() != ProviderTypeIronClaw {
		t.Fatalf("expected type %q, got %q", ProviderTypeIronClaw, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestBuiltinProvider_TypeAndDescription(t *testing.T) {
	p := &BuiltinProvider{}
	if p.Type() != ProviderTypeBuiltin {
		t.Fatalf("expected type %q, got %q", ProviderTypeBuiltin, p.Type())
	}
	if p.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

// ── isAuthenticated / GetCredentials before auth ──────────────────────────────

func TestProviders_NotAuthenticatedBeforeAuthenticate(t *testing.T) {
	providers := []Provider{
		&ClaudeProvider{},
		&GeminiProvider{},
		&OpenCodeProvider{},
		&OpenClawProvider{},
		&IronClawProvider{},
	}
	for _, p := range providers {
		if p.IsAuthenticated() {
			t.Errorf("provider %q should not be authenticated before Authenticate()", p.Type())
		}
		creds := p.GetCredentials()
		if creds.APIKey != "" || creds.OAuthToken != "" {
			t.Errorf("provider %q should return empty credentials before Authenticate()", p.Type())
		}
	}
}

// ── Extra / OAuthToken field ──────────────────────────────────────────────────

func TestCredentials_IsEmpty_ExtraFieldIgnored(t *testing.T) {
	// Extra map should not make credentials non-empty.
	c := Credentials{Extra: map[string]string{"key": "value"}}
	if !c.IsEmpty() {
		t.Fatal("credentials with only Extra map should still be IsEmpty()")
	}
}

func TestGeminiProvider_OAuthOnlyAuthentication(t *testing.T) {
	p := &GeminiProvider{}
	if err := p.Authenticate(Credentials{OAuthToken: "ya29.token"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.GetCredentials().OAuthToken != "ya29.token" {
		t.Error("expected OAuth token to be stored")
	}
}

// ── Registry.All with empty registry ──────────────────────────────────────────

func TestRegistry_All_EmptyRegistry(t *testing.T) {
	r := NewRegistry()
	if all := r.All(); len(all) != 0 {
		t.Fatalf("expected empty slice from empty registry, got %d", len(all))
	}
}

// ── Registry.All includes providers not in the ordered list ───────────────────

type stubProvider struct {
	pt ProviderType
}

func (s *stubProvider) Type() ProviderType        { return s.pt }
func (s *stubProvider) Description() string       { return "stub" }
func (s *stubProvider) SupportedRoles() []string  { return []string{"STUB"} }
func (s *stubProvider) Authenticate(_ Credentials) error { return nil }
func (s *stubProvider) GetCredentials() Credentials      { return Credentials{} }
func (s *stubProvider) IsAuthenticated() bool            { return true }

func TestRegistry_All_CustomProvider(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{pt: "custom"})

	all := r.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(all))
	}
	if all[0].Type() != "custom" {
		t.Fatalf("expected type 'custom', got %q", all[0].Type())
	}
}

// ── Registry concurrency ──────────────────────────────────────────────────────

func TestRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register(&ClaudeProvider{})
			_, _ = r.Get(ProviderTypeClaude)
			_ = r.All()
		}()
	}
	wg.Wait()
}

// ── Registry.Register overwrites existing ─────────────────────────────────────

func TestRegistry_Register_Overwrite(t *testing.T) {
	r := NewRegistry()
	p1 := &ClaudeProvider{}
	_ = p1.Authenticate(Credentials{APIKey: "old"})
	r.Register(p1)

	p2 := &ClaudeProvider{}
	_ = p2.Authenticate(Credentials{APIKey: "new"})
	r.Register(p2)

	got, ok := r.Get(ProviderTypeClaude)
	if !ok {
		t.Fatal("provider not found")
	}
	if got.GetCredentials().APIKey != "new" {
		t.Errorf("expected 'new' API key after overwrite, got %q", got.GetCredentials().APIKey)
	}
}

// ── Infos on empty registry ───────────────────────────────────────────────────

func TestRegistry_Infos_EmptyRegistry(t *testing.T) {
	r := NewRegistry()
	if infos := r.Infos(); len(infos) != 0 {
		t.Fatalf("expected empty infos from empty registry, got %d", len(infos))
	}
}

// ── ProviderType constants ────────────────────────────────────────────────────

func TestProviderTypeConstants(t *testing.T) {
	types := map[ProviderType]string{
		ProviderTypeClaude:    "claude",
		ProviderTypeGemini:    "gemini",
		ProviderTypeOpenCode:  "opencode",
		ProviderTypeOpenClaw:  "openclaw",
		ProviderTypeIronClaw:  "ironclaw",
		ProviderTypeBuiltin:   "builtin",
	}
	for pt, expected := range types {
		if string(pt) != expected {
			t.Errorf("expected ProviderType value %q, got %q", expected, pt)
		}
	}
}

// ── IronClaw supports security-specific roles ─────────────────────────────────

func TestIronClawProvider_SecurityRoles(t *testing.T) {
	p := &IronClawProvider{}
	roles := p.SupportedRoles()
	wantRoles := map[string]bool{
		"SECURITY_ENGINEER": false,
		"AUDIT_MANAGER":     false,
		"QA_TESTER":         false,
	}
	for _, r := range roles {
		if _, ok := wantRoles[r]; ok {
			wantRoles[r] = true
		}
	}
	for role, found := range wantRoles {
		if !found {
			t.Errorf("IronClaw should support role %q", role)
		}
	}
}

// ── OpenClaw supports growth/content roles ────────────────────────────────────

func TestOpenClawProvider_GrowthRoles(t *testing.T) {
	p := &OpenClawProvider{}
	roles := p.SupportedRoles()
	wantRoles := map[string]bool{
		"GROWTH_AGENT":       false,
		"CONTENT_STRATEGIST": false,
	}
	for _, r := range roles {
		if _, ok := wantRoles[r]; ok {
			wantRoles[r] = true
		}
	}
	for role, found := range wantRoles {
		if !found {
			t.Errorf("OpenClaw should support role %q", role)
		}
	}
}
