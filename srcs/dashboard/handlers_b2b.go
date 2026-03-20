package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// TrustAgreementStatus represents the lifecycle of a B2B trust agreement.
type TrustAgreementStatus string

const (
	TrustStatusPending TrustAgreementStatus = "PENDING"
	TrustStatusActive  TrustAgreementStatus = "ACTIVE"
	TrustStatusRevoked TrustAgreementStatus = "REVOKED"
)

// TrustAgreement is a federated trust relationship between two OHC organisations.
// It enables cross-org agent collaboration using SPIFFE-federated JWTs.
type TrustAgreement struct {
	ID           string               `json:"id"`
	PartnerOrg   string               `json:"partnerOrg"`
	PartnerJWKS  string               `json:"partnerJwksUrl"`
	AllowedRoles []string             `json:"allowedRoles"`
	Status       TrustAgreementStatus `json:"status"`
	CreatedAt    time.Time            `json:"createdAt"`
}

type b2bHandshakeRequest struct {
	PartnerOrg   string   `json:"partnerOrg"`
	PartnerJWKS  string   `json:"partnerJwksUrl"`
	AllowedRoles []string `json:"allowedRoles"`
}

func (s *Server) handleB2BAgreements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		agreements := append([]TrustAgreement(nil), s.trustAgreements...)
		s.mu.RUnlock()
		writeJSON(w, agreements)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleB2BHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req b2bHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PartnerOrg == "" || req.PartnerJWKS == "" {
		http.Error(w, "partnerOrg and partnerJwksUrl are required", http.StatusBadRequest)
		return
	}

	agreement := TrustAgreement{
		ID:           "ta-" + strings.ReplaceAll(req.PartnerOrg, ".", "-") + "-" + time.Now().Format("20060102150405"),
		PartnerOrg:   req.PartnerOrg,
		PartnerJWKS:  req.PartnerJWKS,
		AllowedRoles: req.AllowedRoles,
		Status:       TrustStatusActive,
		CreatedAt:    time.Now().UTC(),
	}

	s.mu.Lock()
	s.trustAgreements = append(s.trustAgreements, agreement)
	s.mu.Unlock()

	writeJSON(w, agreement)
}

func (s *Server) handleB2BRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		AgreementID string `json:"agreementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.AgreementID == "" {
		http.Error(w, "agreementId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, ag := range s.trustAgreements {
		if ag.ID == req.AgreementID {
			s.trustAgreements[i].Status = TrustStatusRevoked
			writeJSON(w, s.trustAgreements[i])
			return
		}
	}
	http.Error(w, "agreement not found", http.StatusNotFound)
}
