package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// HandoffPackage carries the structured context an agent sends to a human manager
// when escalating a task it cannot complete autonomously.
type HandoffPackage struct {
	ID             string    `json:"id"`
	FromAgentID    string    `json:"fromAgentId"`
	ToHumanRole    string    `json:"toHumanRole"`
	Intent         string    `json:"intent"`
	FailedAttempts int       `json:"failedAttempts"`
	CurrentState   string    `json:"currentState"`
	Status         string    `json:"status"` // pending | acknowledged | resolved
	CreatedAt      time.Time `json:"createdAt"`
}

type handoffCreateRequest struct {
	FromAgentID    string `json:"fromAgentId"`
	ToHumanRole    string `json:"toHumanRole"`
	Intent         string `json:"intent"`
	FailedAttempts int    `json:"failedAttempts"`
	CurrentState   string `json:"currentState"`
}

func (s *Server) handleHandoffs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		list := append([]HandoffPackage(nil), s.handoffs...)
		s.mu.RUnlock()
		writeJSON(w, list)
	case http.MethodPost:
		var req handoffCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
