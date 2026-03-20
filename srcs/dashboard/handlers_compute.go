package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ComputeProfile defines the hardware requirements for a given agent role.
type ComputeProfile struct {
	RoleID             string    `json:"roleId"`
	MinVRAMGB          int       `json:"minVramGb"`
	PreferredGPUType   string    `json:"preferredGpuType"` // "h100", "a10g", "cpu"
	SchedulingPriority int       `json:"schedulingPriority"`
	CreatedAt          time.Time `json:"createdAt"`
}

type computeProfileRequest struct {
	RoleID             string `json:"roleId"`
	MinVRAMGB          int    `json:"minVramGb"`
	PreferredGPUType   string `json:"preferredGpuType"`
	SchedulingPriority int    `json:"schedulingPriority"`
}

// ClusterStatus reflects the health of a remote Kubernetes cluster region.
type ClusterStatus struct {
	Region         string    `json:"region"`
	Status         string    `json:"status"` // healthy, degraded, offline
	LatencyMS      int       `json:"latencyMs"`
	AvailableNodes int       `json:"availableNodes"`
	CheckedAt      time.Time `json:"checkedAt"`
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
