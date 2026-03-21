package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// Summary: Defines the TrustAgreement type.
// Parameters: None
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

// Summary: Defines the TrustManager type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type TrustManager struct{}

// Summary: ParseJWKS functionality.
// Parameters: partnerOrg, jwksJSON, allowedRoles
// Returns: (TrustAgreement, error)
// Errors: Returns an error if applicable
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

// Summary: Defines the B2BMessage type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type B2BMessage struct {
	Content  string `json:"content"`
	CrossOrg bool   `json:"cross_org"`
	Blocked  bool   `json:"blocked"`
}

// Summary: Defines the EgressFilter type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type EgressFilter struct{}

// Summary: Scan functionality.
// Parameters: message, keywords
// Returns: B2BMessage
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
