package domain

import (
	"testing"
)

func TestTrustManagerParseJWKS(t *testing.T) {
	tm := NewTrustManager()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "https://ohc.acme.corp/.well-known/jwks.json",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			url:     "ohc.acme.corp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.ParseJWKS(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("TrustManager.ParseJWKS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Fatal("expected non-nil TrustAgreement")
				}
				if got.Status != TrustStatusActive {
					t.Errorf("expected status %v, got %v", TrustStatusActive, got.Status)
				}
				if got.PartnerOrg != "ohc.acme.corp" {
					t.Errorf("expected partner org %q, got %q", "ohc.acme.corp", got.PartnerOrg)
				}
				if got.PartnerJWKS != tt.url {
					t.Errorf("expected partner JWKS %q, got %q", tt.url, got.PartnerJWKS)
				}
			}
		})
	}
}

func TestEgressFilterScan(t *testing.T) {
	ef := NewEgressFilter()
	keywords := []string{"Internal Project X", "Secret", "Confidential"}

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "safe message",
			message: "Here is the pricing for the 100 server racks.",
			want:    false,
		},
		{
			name:    "blocked message exact match",
			message: "We cannot agree because of Internal Project X constraints.",
			want:    true,
		},
		{
			name:    "blocked message case insensitive match",
			message: "The SECRET details are attached.",
			want:    true,
		},
		{
			name:    "blocked message another keyword",
			message: "This document is highly CoNfIdEnTiAl.",
			want:    true,
		},
		{
			name:    "empty message",
			message: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ef.Scan(tt.message, keywords)
			if err != nil {
				t.Errorf("EgressFilter.Scan() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("EgressFilter.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
