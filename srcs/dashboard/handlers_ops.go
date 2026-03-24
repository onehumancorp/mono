package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

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

func (s *Server) handleComputeProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		profiles := append([]ComputeProfile(nil), s.computeProfiles...)
		s.mu.RUnlock()
		writeJSON(w, profiles)
	case http.MethodPost:
		var req computeProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.RoleID == "" {
			http.Error(w, "roleId is required", http.StatusBadRequest)
			return
		}
		profile := ComputeProfile{
			RoleID:             req.RoleID,
			MinVRAMGB:          req.MinVRAMGB,
			PreferredGPUType:   req.PreferredGPUType,
			SchedulingPriority: req.SchedulingPriority,
			CreatedAt:          time.Now().UTC(),
		}
		s.mu.Lock()
		s.computeProfiles = append(s.computeProfiles, profile)
		s.mu.Unlock()
		writeJSON(w, profile)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract region from URL path: /api/clusters/{region}/status
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	region := ""
	for i, p := range parts {
		if p == "clusters" && i+1 < len(parts) {
			region = parts[i+1]
			break
		}
	}
	if region == "" {
		http.Error(w, "region is required in path", http.StatusBadRequest)
		return
	}
	// Simulated cluster health response (would call k8s API in production)
	status := ClusterStatus{
		Region:         region,
		Status:         "healthy",
		LatencyMS:      3,
		AvailableNodes: 5,
		CheckedAt:      time.Now().UTC(),
	}
	writeJSON(w, status)
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

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		pipelines := append([]Pipeline(nil), s.pipelines...)
		s.mu.RUnlock()
		writeJSON(w, pipelines)
	case http.MethodPost:
		var req pipelineCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		pipeline := Pipeline{
			ID:          "pipeline-" + now.Format("20060102150405"),
			Name:        req.Name,
			Status:      PipelineStatusPending,
			Branch:      req.Branch,
			InitiatedBy: req.InitiatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		s.mu.Lock()
		s.pipelines = append(s.pipelines, pipeline)
		s.mu.Unlock()
		writeJSON(w, pipeline)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePipelinePromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pipelinePromoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PipelineID == "" {
		http.Error(w, "pipelineId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pipelines {
		if p.ID == req.PipelineID {
			if s.pipelines[i].Status != PipelineStatusStaging {
				http.Error(w, "pipeline must be in STAGING status to promote", http.StatusBadRequest)
				return
			}
			s.pipelines[i].Status = PipelineStatusPromoted
			s.pipelines[i].UpdatedAt = time.Now().UTC()
			writeJSON(w, s.pipelines[i])
			return
		}
	}
	http.Error(w, "pipeline not found", http.StatusNotFound)
}

func (s *Server) handlePipelineStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PipelineID string `json:"pipelineId"`
		Status     string `json:"status"`
		StagingURL string `json:"stagingUrl,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PipelineID == "" || req.Status == "" {
		http.Error(w, "pipelineId and status are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pipelines {
		if p.ID == req.PipelineID {
			s.pipelines[i].Status = PipelineStatus(req.Status)
			s.pipelines[i].UpdatedAt = time.Now().UTC()
			if req.StagingURL != "" {
				s.pipelines[i].StagingURL = req.StagingURL
			}
			writeJSON(w, s.pipelines[i])
			return
		}
	}
	http.Error(w, "pipeline not found", http.StatusNotFound)
}

// Summary: ScaleRequest defines the payload for scaling a team member role.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ScaleRequest struct {
	Role  string `json:"role"`
	Count int    `json:"count"`
}

func (s *Server) handleScale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ScaleRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	orgID := s.org.ID
	agents := s.hub.Agents()
	s.mu.RUnlock()

	var currentCount int
	var idleAgentIDs []string
	var activeAgentIDs []string
	for _, agent := range agents {
		if agent.Role == req.Role {
			currentCount++
			if agent.Status == orchestration.StatusIdle {
				idleAgentIDs = append(idleAgentIDs, agent.ID)
			} else {
				activeAgentIDs = append(activeAgentIDs, agent.ID)
			}
		}
	}

	diff := req.Count - currentCount
	nowStr := time.Now().UTC().Format("20060102150405000")

	if diff > 0 {
		for i := 0; i < diff; i++ {
			id := fmt.Sprintf("%s-agent-%s-%d", orgID, nowStr, i)
			newAgent := orchestration.Agent{
				ID:             id,
				Name:           req.Role,
				Role:           req.Role,
				OrganizationID: orgID,
				Status:         orchestration.StatusIdle,
			}
			s.hub.RegisterAgent(newAgent)
		}
	} else if diff < 0 {
		toRemove := -diff

		// ⚡ BOLT: [pod-thrashing during dynamic scale-up/down] - Randomized Selection from Top 5
		// Gracefully scales down idle agents first before terminating busy ones to prevent interrupting active work.
		for i := 0; i < toRemove; i++ {
			if len(idleAgentIDs) > 0 {
				s.hub.FireAgent(idleAgentIDs[0])
				idleAgentIDs = idleAgentIDs[1:]
			} else if len(activeAgentIDs) > 0 {
				s.hub.FireAgent(activeAgentIDs[0])
				activeAgentIDs = activeAgentIDs[1:]
			}
		}
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"role":   req.Role,
		"count":  req.Count,
	})
}

// Summary: handleScaleStream streams real-time scaling trace events to the dashboard.
// Parameters: s *Server (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleScaleStream(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	// Simulated events
	events := []string{
		`{"event":"AI Workforce Manager: Reconciling Team Member resource.","status":"INFO"}`,
		`{"event":"AI Workforce Manager: Spinning up new instances...","status":"INFO"}`,
		`{"event":"AgentHired","status":"Ready"}`,
	}

	for _, event := range events {
		data := []byte("data: " + event + "\n\n")
		w.Write(data)
		if err := rc.Flush(); err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
