package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TrustAgreementStatus represents the lifecycle of a B2B trust agreement.
type TrustAgreementStatus string

const (
	TrustStatusPending TrustAgreementStatus = "PENDING"
	TrustStatusActive  TrustAgreementStatus = "ACTIVE"
	TrustStatusRevoked TrustAgreementStatus = "REVOKED"
)

// TrustAgreement is a federated trust relationship between two OHC organisations.
type TrustAgreement struct {
	ID           string               `json:"id"`
	PartnerOrg   string               `json:"partnerOrg"`
	PartnerJWKS  string               `json:"partnerJwksUrl"`
	AllowedRoles []string             `json:"allowedRoles"`
	Status       TrustAgreementStatus `json:"status"`
	CreatedAt    time.Time            `json:"createdAt"`
}

// TrustManager handles parsing and verifying federated trust settings.
type TrustManager struct{}

// NewTrustManager constructs a new TrustManager.
func NewTrustManager() *TrustManager {
	return &TrustManager{}
}

// ParseJWKS parses a partner's JWKS URL, validates it, and returns an active TrustAgreement.
func (tm *TrustManager) ParseJWKS(jwks string) (*TrustAgreement, error) {
	if jwks == "" {
		return nil, errors.New("JWKS URL cannot be empty")
	}

	parsed, err := url.ParseRequestURI(jwks)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid JWKS URL format")
	}

	agreement := &TrustAgreement{
		ID:           uuid.New().String(),
		PartnerOrg:   parsed.Host,
		PartnerJWKS:  jwks,
		AllowedRoles: []string{},
		Status:       TrustStatusActive,
		CreatedAt:    time.Now().UTC(),
	}

	return agreement, nil
}

// EgressFilter enforces data perimeters by scanning outbound messages for restricted keywords.
type EgressFilter struct{}

// NewEgressFilter constructs a new EgressFilter.
func NewEgressFilter() *EgressFilter {
	return &EgressFilter{}
}

// Scan checks the given message for any of the internal keywords.
// Returns true (indicating blocked/flagged) if a match is found.
func (ef *EgressFilter) Scan(message string, keywords []string) (bool, error) {
	lowerMsg := strings.ToLower(message)
	for _, kw := range keywords {
		if kw == "" {
			continue
		}
		if strings.Contains(lowerMsg, strings.ToLower(kw)) {
			return true, nil
		}
	}
	return false, nil
}
