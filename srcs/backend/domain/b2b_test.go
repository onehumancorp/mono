package domain

import (
	"strings"
	"testing"
)

// TestTrustManager_ParseJWKS implements UT-01 from the test plan:
// "UT-01 | Trust Manager | Parse partner JWKS | Trust object set to ACTIVE"
func TestTrustManager_ParseJWKS(t *testing.T) {
	tm := &TrustManager{}

	tests := []struct {
		name         string
		partnerOrg   string
		jwksJSON     string
		allowedRoles []string
		wantErr      bool
		wantStatus   string
	}{
		{
			name:         "Valid JSON sets trust object to ACTIVE",
			partnerOrg:   "globex.com",
			jwksJSON:     `{"keys": [{"kty": "RSA"}]}`,
			allowedRoles: []string{"Sales Agent"},
			wantErr:      false,
			wantStatus:   "ACTIVE",
		},
		{
			name:         "Invalid JSON returns error",
			partnerOrg:   "bad-org.com",
			jwksJSON:     `{bad json}`,
			allowedRoles: []string{},
			wantErr:      true,
			wantStatus:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.ParseJWKS(tt.partnerOrg, tt.jwksJSON, tt.allowedRoles)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseJWKS() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got.Status != tt.wantStatus {
					t.Errorf("ParseJWKS() status = %v, want %v", got.Status, tt.wantStatus)
				}
				if got.PartnerOrg != tt.partnerOrg {
					t.Errorf("ParseJWKS() partnerOrg = %v, want %v", got.PartnerOrg, tt.partnerOrg)
				}
				if got.PartnerJWKS != tt.jwksJSON {
					t.Errorf("ParseJWKS() jwks = %v, want %v", got.PartnerJWKS, tt.jwksJSON)
				}
				if len(got.AllowedRoles) != len(tt.allowedRoles) {
					t.Errorf("ParseJWKS() allowed roles length = %v, want %v", len(got.AllowedRoles), len(tt.allowedRoles))
				}
				if !strings.HasPrefix(got.ID, "b2b-trust-") {
					t.Errorf("ParseJWKS() id = %v, want prefix 'b2b-trust-'", got.ID)
				}
			}
		})
	}
}

// TestEgressFilter_Scan implements UT-02 from the test plan:
// "UT-02 | Egress Filter | Scan outgoing message for internal keywords | Message blocked and flagged"
func TestEgressFilter_Scan(t *testing.T) {
	ef := &EgressFilter{}
	keywords := []string{"Internal Project X", "Confidential"}

	tests := []struct {
		name        string
		message     string
		wantBlocked bool
		wantCross   bool
	}{
		{
			name:        "Message with no internal keywords is not blocked",
			message:     "Hello, we would like to purchase 100 server racks.",
			wantBlocked: false,
			wantCross:   true,
		},
		{
			name:        "Message with exact internal keyword is blocked",
			message:     "We can offer a discount based on Internal Project X budget.",
			wantBlocked: true,
			wantCross:   true,
		},
		{
			name:        "Message with case-insensitive internal keyword is blocked",
			message:     "This is very CONFIDENTIAL information.",
			wantBlocked: true,
			wantCross:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ef.Scan(tt.message, keywords)

			if got.Blocked != tt.wantBlocked {
				t.Errorf("Scan() blocked = %v, want %v", got.Blocked, tt.wantBlocked)
			}
			if got.CrossOrg != tt.wantCross {
				t.Errorf("Scan() cross_org = %v, want %v", got.CrossOrg, tt.wantCross)
			}
			if got.Content != tt.message {
				t.Errorf("Scan() content = %v, want %v", got.Content, tt.message)
			}
		})
	}
}
