package dashboard

import (
	"net/http"
)

// handlePruneMissions triggers the background cleanup of stale or completed tasks
func (s *Server) handlePruneMissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Execute pruning task
	if s.hub.SIPDB() != nil {
		_ = s.hub.SIPDB().PruneStaleMissions(r.Context(), 0) // Prune all completed or stale missions immediately
	}
	writeJSON(w, map[string]string{"status": "success", "message": "agent missions pruned"})
}
