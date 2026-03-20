package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// Summary: TrustAgreement represents a mutual trust relationship between two organizations.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type TrustAgreement struct {
	ID           string   `json:"id"`
	PartnerOrg   string   `json:"partner_org"`
	PartnerJWKS  string   `json:"partner_jwks_url"`
	AllowedRoles []string `json:"allowed_roles"`
	Status       string   `json:"status"` // PENDING, ACTIVE, REVOKED
}

// Summary: TrustManager handles the creation and management of TrustAgreements.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type TrustManager struct{}

// Summary: ParseJWKS parses a partner's JWKS data (JSON) and creates an active TrustAgreement. This implements UT-01 from the test plan.
// Params: partnerOrg, jwksJSON, allowedRoles
// Returns: Returns the computed value
// Errors: Returns an error if the operation fails
// Side Effects: None
func (tm *TrustManager) ParseJWKS(partnerOrg, jwksJSON string, allowedRoles []string) (TrustAgreement, error) {
	// Simple validation to simulate parsing the JWKS JSON string.
	// In a real implementation, this would validate the keys cryptographically.
	var jwks map[string]interface{}
	if err := json.Unmarshal([]byte(jwksJSON), &jwks); err != nil {
		return TrustAgreement{}, err
	}

	return TrustAgreement{
		ID:           "b2b-trust-" + time.Now().UTC().Format("20060102150405.000"),
		PartnerOrg:   partnerOrg,
		PartnerJWKS:  jwksJSON, // Store raw JSON string or URL
		AllowedRoles: allowedRoles,
		Status:       "ACTIVE",
	}, nil
}

// Summary: B2BMessage represents an encapsulated agent message for cross-org tunneling.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type B2BMessage struct {
	Content  string `json:"content"`
	CrossOrg bool   `json:"cross_org"`
	Blocked  bool   `json:"blocked"`
}

// Summary: EgressFilter enforces the data perimeter for B2B collaboration.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type EgressFilter struct{}

// Summary: Scan checks the outgoing message content for internal keywords. If a keyword is found, the message is blocked and flagged. Messages entering/leaving an Inter-Org Room are flagged with CrossOrg: true. This implements UT-02 from the test plan.
// Params: message, keywords
// Returns: Returns the computed value
// Errors: None
// Side Effects: None
func (ef *EgressFilter) Scan(message string, keywords []string) B2BMessage {
	blocked := false
	lowerMsg := strings.ToLower(message)

	for _, kw := range keywords {
		if strings.Contains(lowerMsg, strings.ToLower(kw)) {
			blocked = true
			break
		}
	}

	return B2BMessage{
		Content:  message,
		CrossOrg: true,
		Blocked:  blocked,
	}
}
