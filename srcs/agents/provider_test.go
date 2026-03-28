package agents

import (
	"testing"
)

func TestProviderGetCredentials(t *testing.T) {
	creds := Credentials{APIKey: "test-key"}

	tests := []struct {
		name     string
		provider Provider
	}{
		{"GeminiProvider", &GeminiProvider{}},
		{"OpenCodeProvider", &OpenCodeProvider{}},
		{"OpenClawProvider", &OpenClawProvider{}},
		{"IronClawProvider", &IronClawProvider{}},
		{"MiniMaxProvider", &MiniMaxProvider{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.provider.Authenticate(creds)
			if tt.provider.GetCredentials().APIKey != "test-key" {
				t.Errorf("Expected test-key for %s GetCredentials", tt.name)
			}
		})
	}
}

func TestMiniMaxProvider(t *testing.T) {
	p := &MiniMaxProvider{}

	if p.Type() != ProviderTypeMiniMax {
		t.Errorf("expected type %q, got %q", ProviderTypeMiniMax, p.Type())
	}
	if p.IsAuthenticated() {
		t.Error("expected unauthenticated before credentials set")
	}

	// Require API key
	if err := p.Authenticate(Credentials{}); err == nil {
		t.Error("expected error when no API key provided")
	}

	creds := Credentials{APIKey: "sk-minimax-test"}
	if err := p.Authenticate(creds); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsAuthenticated() {
		t.Error("expected authenticated after valid credentials")
	}
	if got := p.GetCredentials().APIKey; got != "sk-minimax-test" {
		t.Errorf("expected %q, got %q", "sk-minimax-test", got)
	}
	roles := p.SupportedRoles()
	if len(roles) == 0 {
		t.Error("expected non-empty supported roles")
	}
}

func TestDefaultRegistryIncludesMiniMax(t *testing.T) {
	r := DefaultRegistry()
	p, ok := r.Get(ProviderTypeMiniMax)
	if !ok {
		t.Fatal("MiniMaxProvider not found in default registry")
	}
	if p.Type() != ProviderTypeMiniMax {
		t.Errorf("expected type %q, got %q", ProviderTypeMiniMax, p.Type())
	}
}
