package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// ApprovalStatus represents the lifecycle state of a guardian-gate request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

// ApprovalRequest is created by the Guardian Agent when a high-risk action
// requires explicit human sign-off.
type ApprovalRequest struct {
	ID               string         `json:"id"`
	AgentID          string         `json:"agentId"`
	Action           string         `json:"action"`
	Reason           string         `json:"reason"`
	EstimatedCostUSD float64        `json:"estimatedCostUsd"`
	RiskLevel        string         `json:"riskLevel"` // low | medium | high | critical
	Status           ApprovalStatus `json:"status"`
	CreatedAt        time.Time      `json:"createdAt"`
	DecidedAt        *time.Time     `json:"decidedAt,omitempty"`
	DecidedBy        string         `json:"decidedBy,omitempty"`
}

type approvalCreateRequest struct {
	AgentID          string  `json:"agentId"`
	Action           string  `json:"action"`
	Reason           string  `json:"reason"`
	EstimatedCostUSD float64 `json:"estimatedCostUsd"`
	RiskLevel        string  `json:"riskLevel"`
}

type approvalDecideRequest struct {
	ApprovalID string `json:"approvalId"`
	Decision   string `json:"decision"` // approve | reject
	DecidedBy  string `json:"decidedBy"`
}

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

func (s *Server) handleApprovalRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req approvalCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

func (s *Server) handleApprovalDecide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req approvalDecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
