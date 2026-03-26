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
