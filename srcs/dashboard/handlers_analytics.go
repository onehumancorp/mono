package dashboard

import (
	"net/http"
)

// AnalyticsSummary surfaces operational health metrics.
type AnalyticsSummary struct {
	HumanAgentRatio     float64 `json:"humanAgentRatio"`
	TotalAgents         int     `json:"totalAgents"`
	TotalHumans         int     `json:"totalHumans"`
	AuditFidelityPct    float64 `json:"auditFidelityPct"`
	ResumptionLatencyMS int     `json:"resumptionLatencyMs"`
	PendingApprovals    int     `json:"pendingApprovals"`
	ActiveHandoffs      int     `json:"activeHandoffs"`
	TokenVelocity       int64   `json:"tokenVelocity"`
}

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
