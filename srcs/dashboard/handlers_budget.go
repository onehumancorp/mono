package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// defaultBudgetAlertNotifyPct is the default notification threshold (80 %).
const defaultBudgetAlertNotifyPct = 0.8

// BudgetAlert defines a spending threshold with notification behaviour.
type BudgetAlert struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ThresholdUSD   float64   `json:"thresholdUsd"`
	NotifyAtPct    float64   `json:"notifyAtPct"` // e.g. 0.8 → notify at 80 %
	Triggered      bool      `json:"triggered"`
	CreatedAt      time.Time `json:"createdAt"`
}

type budgetAlertRequest struct {
	OrganizationID string  `json:"organizationId"`
	ThresholdUSD   float64 `json:"thresholdUsd"`
	NotifyAtPct    float64 `json:"notifyAtPct"`
}

func (s *Server) handleBudgetAlerts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		alerts := append([]BudgetAlert(nil), s.budgetAlerts...)
		s.mu.RUnlock()
		// Evaluate triggered state against current spend.
		summary := s.tracker.Summary(s.org.ID)
		for i, a := range alerts {
			alerts[i].Triggered = summary.TotalCostUSD >= a.ThresholdUSD*a.NotifyAtPct
		}
		writeJSON(w, alerts)
	case http.MethodPost:
		var req budgetAlertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.ThresholdUSD <= 0 {
			http.Error(w, "thresholdUsd must be greater than zero", http.StatusBadRequest)
			return
		}
		if req.NotifyAtPct <= 0 || req.NotifyAtPct > 1 {
			req.NotifyAtPct = defaultBudgetAlertNotifyPct // default 80 %
		}
		orgID := req.OrganizationID
		if orgID == "" {
			s.mu.RLock()
			orgID = s.org.ID
			s.mu.RUnlock()
		}
		alert := BudgetAlert{
			ID:             "alert-" + time.Now().Format("20060102150405"),
			OrganizationID: orgID,
			ThresholdUSD:   req.ThresholdUSD,
			NotifyAtPct:    req.NotifyAtPct,
			Triggered:      false,
			CreatedAt:      time.Now().UTC(),
		}
		s.mu.Lock()
		s.budgetAlerts = append(s.budgetAlerts, alert)
		s.mu.Unlock()
		writeJSON(w, alert)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
