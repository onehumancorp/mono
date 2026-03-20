package domain

// TrustAgreement represents a cross-organizational collaboration partnership
// used in B2B agent exchanges.
type TrustAgreement struct {
	ID           string   `json:"id"`
	PartnerOrg   string   `json:"partner_org"`
	PartnerJWKS  string   `json:"partner_jwks_url"`
	AllowedRoles []string `json:"allowed_roles"`
	Status       string   `json:"status"` // PENDING, ACTIVE, REVOKED
}

// B2BMessage represents an encapsulated message meant for tunneling
// across the Inter-Org Gateway.
type B2BMessage struct {
	ID        string `json:"id"`
	FromOrg   string `json:"from_org"`
	ToOrg     string `json:"to_org"`
	FromAgent string `json:"from_agent"`
	ToAgent   string `json:"to_agent"`
	Content   string `json:"content"`
	CrossOrg  bool   `json:"cross_org"`
	Signature string `json:"signature"`
}
