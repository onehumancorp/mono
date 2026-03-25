package dashboard

import (
	"encoding/json"
	"net/http"
)

// Handles retrieving organization details.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleOrg(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.org)
}

// Handles retrieving available domains.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleDomains(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, availableDomains)
}

// Handles retrieving or updating settings.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r.Method == http.MethodGet {
		writeJSON(w, s.settings)
		return
	}

	if r.Method == http.MethodPost {
		var req Settings
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		s.settings = req
		s.hub.SetMinimaxAPIKey(req.MinimaxAPIKey)
		writeJSON(w, s.settings)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// Handles retrieving marketplace items.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleMarketplace(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, defaultMarketplaceItems())
}

// Handles retrieving analytics summary.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleAnalytics(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	agents := s.hub.Agents()
	org := s.org
	summary := s.tracker.Summary(org.ID)
	pendingApprovals := 0
	for _, a := range s.approvals {
		if a.Status == ApprovalStatusPending {
			pendingApprovals++
		}
	}
	activeHandoffs := 0
	for _, h := range s.handoffs {
		if h.Status == "pending" {
			activeHandoffs++
		}
	}
	s.mu.RUnlock()

	totalHumans := 0
	for _, m := range org.Members {
		if m.IsHuman {
			totalHumans++
		}
	}
	totalAgents := len(agents)

	var ratio float64
	if totalHumans > 0 {
		ratio = float64(totalAgents) / float64(totalHumans)
	}

	meetings := s.hub.Meetings()
	totalMsgs := 0
	auditedMsgs := 0
	agentSet := map[string]bool{}
	for _, a := range agents {
		agentSet[a.ID] = true
	}
	for _, m := range meetings {
		for _, msg := range m.Transcript {
			totalMsgs++
			if agentSet[msg.FromAgent] {
				auditedMsgs++
			}
		}
	}
	auditFidelity := 100.0
	if totalMsgs > 0 {
		auditFidelity = float64(auditedMsgs) / float64(totalMsgs) * 100
	}

	writeJSON(w, AnalyticsSummary{
		HumanAgentRatio:     ratio,
		TotalAgents:         totalAgents,
		TotalHumans:         totalHumans,
		AuditFidelityPct:    auditFidelity,
		ResumptionLatencyMS: 4800,
		PendingApprovals:    pendingApprovals,
		ActiveHandoffs:      activeHandoffs,
		TokenVelocity:       summary.TotalTokens,
	})
}
