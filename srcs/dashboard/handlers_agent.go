package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/onehumancorp/mono/srcs/agents"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// hireRequest carries agent creation parameters.
type hireRequest struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Model        string `json:"model,omitempty"`
	ProviderType string `json:"providerType,omitempty"`
}

// fireRequest carries the ID of the agent to remove.
type fireRequest struct {
	AgentID string `json:"agentId"`
}

func (s *Server) handleHireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req hireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Role == "" {
		http.Error(w, "name and role are required", http.StatusBadRequest)
		return
	}

	// Resolve provider type: default to "builtin" when unspecified.
	providerType := req.ProviderType
	if providerType == "" {
		providerType = string(agents.ProviderTypeBuiltin)
	}

	// Validate that the requested provider is registered.
	if _, ok := s.agentProviderRegistry.Get(agents.ProviderType(providerType)); !ok {
		http.Error(w, "unknown provider type: "+providerType, http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	id := s.org.ID + "-agent-" + time.Now().UTC().Format("20060102150405000")
	agent := orchestration.Agent{
		ID:             id,
		Name:           req.Name,
		Role:           req.Role,
		OrganizationID: s.org.ID,
		Status:         orchestration.StatusIdle,
		ProviderType:   providerType,
	}
	s.hub.RegisterAgent(agent)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

func (s *Server) handleFireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req fireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.AgentID == "" {
		http.Error(w, "agentId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.hub.FireAgent(req.AgentID)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

// handleAgentProviders lists all registered external agent providers and their
// authentication status.  Responds to GET only.
func (s *Server) handleAgentProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.agentProviderRegistry.Infos())
}

// providerAuthRequest carries credentials for authenticating with an agent provider.
type providerAuthRequest struct {
	ProviderType string            `json:"providerType"`
	APIKey       string            `json:"apiKey,omitempty"`
	OAuthToken   string            `json:"oauthToken,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// handleAgentProviderAuth accepts POST requests to authenticate with an external
// agent provider.  Credentials are stored in memory and forwarded to any
// subsequently hired agent of that provider type.
func (s *Server) handleAgentProviderAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req providerAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.ProviderType == "" {
		http.Error(w, "providerType is required", http.StatusBadRequest)
		return
	}

	creds := agents.Credentials{
		APIKey:     req.APIKey,
		OAuthToken: req.OAuthToken,
		Extra:      req.Extra,
	}
	if err := s.agentProviderRegistry.Authenticate(agents.ProviderType(req.ProviderType), creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	infos := s.agentProviderRegistry.Infos()
	writeJSON(w, infos)
}
