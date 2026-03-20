package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// IncidentSeverity classifies the urgency of an operational incident.
type IncidentSeverity string

const (
	SeverityP0 IncidentSeverity = "P0"
	SeverityP1 IncidentSeverity = "P1"
	SeverityP2 IncidentSeverity = "P2"
)

// IncidentStatus reflects the investigation lifecycle state.
type IncidentStatus string

const (
	IncidentStatusInvestigating IncidentStatus = "INVESTIGATING"
	IncidentStatusProposed      IncidentStatus = "PROPOSED"
	IncidentStatusResolved      IncidentStatus = "RESOLVED"
)

// Incident represents an operational event requiring SRE attention.
type Incident struct {
	ID               string           `json:"id"`
	Severity         IncidentSeverity `json:"severity"`
	Summary          string           `json:"summary"`
	RCA              string           `json:"rootCauseAnalysis"`
	ResolutionPlanID string           `json:"resolutionPlanId,omitempty"`
	Status           IncidentStatus   `json:"status"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

type incidentCreateRequest struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	RCA      string `json:"rootCauseAnalysis,omitempty"`
}

type incidentStatusRequest struct {
	IncidentID       string `json:"incidentId"`
	Status           string `json:"status"`
	ResolutionPlanID string `json:"resolutionPlanId,omitempty"`
	RCA              string `json:"rootCauseAnalysis,omitempty"`
}

func (s *Server) handleIncidents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		incidents := append([]Incident(nil), s.incidents...)
		s.mu.RUnlock()
		writeJSON(w, incidents)
	case http.MethodPost:
		var req incidentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Severity == "" || req.Summary == "" {
			http.Error(w, "severity and summary are required", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		incident := Incident{
			ID:        "inc-" + now.Format("20060102150405"),
			Severity:  IncidentSeverity(req.Severity),
			Summary:   req.Summary,
			RCA:       req.RCA,
			Status:    IncidentStatusInvestigating,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.mu.Lock()
		s.incidents = append(s.incidents, incident)
		s.mu.Unlock()
		writeJSON(w, incident)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleIncidentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req incidentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IncidentID == "" || req.Status == "" {
		http.Error(w, "incidentId and status are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, inc := range s.incidents {
		if inc.ID == req.IncidentID {
			s.incidents[i].Status = IncidentStatus(req.Status)
			s.incidents[i].UpdatedAt = time.Now().UTC()
			if req.ResolutionPlanID != "" {
				s.incidents[i].ResolutionPlanID = req.ResolutionPlanID
			}
			if req.RCA != "" {
				s.incidents[i].RCA = req.RCA
			}
			writeJSON(w, s.incidents[i])
			return
		}
	}
	http.Error(w, "incident not found", http.StatusNotFound)
}
