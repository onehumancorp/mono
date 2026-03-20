package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// PipelineStatus reflects the lifecycle of an autonomous CI/CD pipeline.
type PipelineStatus string

const (
	PipelineStatusPending      PipelineStatus = "PENDING"
	PipelineStatusImplementing PipelineStatus = "IMPLEMENTING"
	PipelineStatusTesting      PipelineStatus = "TESTING"
	PipelineStatusStaging      PipelineStatus = "STAGING"
	PipelineStatusPromoted     PipelineStatus = "PROMOTED"
	PipelineStatusFailed       PipelineStatus = "FAILED"
)

// Pipeline represents an autonomous implementation pipeline from spec to production.
type Pipeline struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Status      PipelineStatus `json:"status"`
	Branch      string         `json:"branch"`
	StagingURL  string         `json:"stagingUrl,omitempty"`
	InitiatedBy string         `json:"initiatedBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type pipelineCreateRequest struct {
	Name        string `json:"name"`
	Branch      string `json:"branch"`
	InitiatedBy string `json:"initiatedBy"`
}

type pipelinePromoteRequest struct {
	PipelineID string `json:"pipelineId"`
	ApprovedBy string `json:"approvedBy"`
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
