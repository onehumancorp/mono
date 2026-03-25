package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Handles retrieving approvals.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleApprovals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		list := append([]ApprovalRequest(nil), s.approvals...)
		s.mu.RUnlock()
		writeJSON(w, list)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handles creating an approval request.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleApprovalRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req approvalCreateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.AgentID == "" || req.Action == "" {
		http.Error(w, "agentId and action are required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	approval := ApprovalRequest{
		ID:               s.org.ID + "-approval-" + now.Format("20060102150405000"),
		AgentID:          req.AgentID,
		Action:           req.Action,
		Reason:           req.Reason,
		EstimatedCostUSD: req.EstimatedCostUSD,
		RiskLevel:        req.RiskLevel,
		Status:           ApprovalStatusPending,
		CreatedAt:        now,
	}
	if approval.RiskLevel == "" {
		if approval.EstimatedCostUSD > 500 {
			approval.RiskLevel = "critical"
		} else if approval.EstimatedCostUSD > 100 {
			approval.RiskLevel = "high"
		} else {
			approval.RiskLevel = "medium"
		}
	}

	s.mu.Lock()
	s.approvals = append(s.approvals, approval)
	s.mu.Unlock()

	writeJSON(w, approval)
}

// Handles deciding on an approval request.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleApprovalDecide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req approvalDecideRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.ApprovalID == "" || req.Decision == "" {
		http.Error(w, "approvalId and decision are required", http.StatusBadRequest)
		return
	}

	var newStatus ApprovalStatus
	switch req.Decision {
	case "approve":
		newStatus = ApprovalStatusApproved
	case "reject":
		newStatus = ApprovalStatusRejected
	default:
		http.Error(w, "decision must be 'approve' or 'reject'", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	s.mu.Lock()
	found := false
	for i, a := range s.approvals {
		if a.ID == req.ApprovalID {
			s.approvals[i].Status = newStatus
			s.approvals[i].DecidedAt = &now
			s.approvals[i].DecidedBy = req.DecidedBy
			found = true
			break
		}
	}
	s.mu.Unlock()

	if !found {
		http.Error(w, "approval not found", http.StatusNotFound)
		return
	}

	s.mu.RLock()
	list := append([]ApprovalRequest(nil), s.approvals...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

// Handles retrieving or creating handoffs.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleHandoffs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		list := append([]HandoffPackage(nil), s.handoffs...)
		s.mu.RUnlock()
		writeJSON(w, list)
	case http.MethodPost:
		var req handoffCreateRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.FromAgentID == "" || req.Intent == "" {
			http.Error(w, "fromAgentId and intent are required", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		handoff := HandoffPackage{
			ID:             s.org.ID + "-handoff-" + now.Format("20060102150405000"),
			FromAgentID:    req.FromAgentID,
			ToHumanRole:    req.ToHumanRole,
			Intent:         req.Intent,
			FailedAttempts: req.FailedAttempts,
			CurrentState:   req.CurrentState,
			Status:         "pending",
			CreatedAt:      now,
		}
		s.mu.Lock()
		s.handoffs = append(s.handoffs, handoff)
		s.mu.Unlock()
		writeJSON(w, handoff)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handles resolving a handoff.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleHandoffResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		HandoffID string `json:"handoffId"`
		Status    string `json:"status"` // acknowledged | resolved
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.HandoffID == "" || req.Status == "" {
		http.Error(w, "handoffId and status are required", http.StatusBadRequest)
		return
	}

	if req.Status != "acknowledged" && req.Status != "resolved" {
		http.Error(w, "status must be 'acknowledged' or 'resolved'", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	found := false
	alreadyResolved := false
	for i, h := range s.handoffs {
		if h.ID == req.HandoffID {
			found = true
			if h.Status != "pending" {
				alreadyResolved = true
			} else {
				s.handoffs[i].Status = req.Status
			}
			break
		}
	}
	s.mu.Unlock()

	if !found {
		http.Error(w, "handoff not found", http.StatusNotFound)
		return
	}

	if alreadyResolved {
		http.Error(w, "State Changed: This handoff has already been acknowledged or resolved.", http.StatusConflict)
		return
	}

	s.mu.RLock()
	list := append([]HandoffPackage(nil), s.handoffs...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

// Handles retrieving B2B agreements.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
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

// Handles B2B handshake.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleB2BHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req b2bHandshakeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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

// Handles revoking a B2B agreement.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleB2BRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		AgreementID string `json:"agreementId"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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
