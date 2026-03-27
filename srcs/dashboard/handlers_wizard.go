package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/onehumancorp/mono/srcs/settings"
)

// wizardStatusResponse describes the current setup state of the platform.
type wizardStatusResponse struct {
	// Configured is true when all required fields have been set.
	Configured bool `json:"configured"`
	// Steps holds per-step completion status.
	Steps wizardSteps `json:"steps"`
}

type wizardSteps struct {
	Server     bool `json:"server"`      // listen_addr and db_path set
	AiProvider bool `json:"ai_provider"` // at least one AI provider enabled
	Centrifuge bool `json:"centrifuge"`  // centrifuge_url set
}

// wizardConfigureRequest carries a partial or complete settings update from
// the wizard UI.
type wizardConfigureRequest struct {
	ListenAddr    string                `json:"listen_addr,omitempty"`
	DBPath        string                `json:"db_path,omitempty"`
	PostgresURL   string                `json:"postgres_url,omitempty"`
	RedisURL      string                `json:"redis_url,omitempty"`
	CentrifugeURL string                `json:"centrifuge_url,omitempty"`
	MinimaxAPIKey string                `json:"minimax_api_key,omitempty"`
	AiProviders   []settings.AiProvider `json:"ai_providers,omitempty"`
}

// handleWizardStatus returns a JSON summary of whether each wizard step has
// been completed so the Flutter wizard UI can determine which steps to show.
func (s *Server) handleWizardStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	cfg := s.settings
	s.mu.RUnlock()

	steps := wizardSteps{
		Server:     cfg.ListenAddr != "" && cfg.DBPath != "",
		AiProvider: hasEnabledProvider(cfg.AiProviders),
		Centrifuge: cfg.CentrifugeURL != "",
	}
	resp := wizardStatusResponse{
		Configured: steps.Server && steps.AiProvider && steps.Centrifuge,
		Steps:      steps,
	}
	writeJSON(w, resp)
}

// handleWizardConfigure applies a partial settings update from the wizard and
// persists it via the settings store.
func (s *Server) handleWizardConfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req wizardConfigureRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	cfg := s.settings
	if req.ListenAddr != "" {
		cfg.ListenAddr = req.ListenAddr
	}
	if req.DBPath != "" {
		cfg.DBPath = req.DBPath
	}
	if req.PostgresURL != "" {
		cfg.PostgresURL = req.PostgresURL
	}
	if req.RedisURL != "" {
		cfg.RedisURL = req.RedisURL
	}
	if req.CentrifugeURL != "" {
		cfg.CentrifugeURL = req.CentrifugeURL
	}
	if req.MinimaxAPIKey != "" {
		cfg.MinimaxAPIKey = req.MinimaxAPIKey
		s.hub.SetMinimaxAPIKey(req.MinimaxAPIKey)
	}
	if len(req.AiProviders) > 0 {
		cfg.AiProviders = req.AiProviders
	}
	s.settings = cfg
	s.mu.Unlock()

	_ = s.hub.SettingsStore().Update(cfg)

	steps := wizardSteps{
		Server:     cfg.ListenAddr != "" && cfg.DBPath != "",
		AiProvider: hasEnabledProvider(cfg.AiProviders),
		Centrifuge: cfg.CentrifugeURL != "",
	}
	writeJSON(w, wizardStatusResponse{
		Configured: steps.Server && steps.AiProvider && steps.Centrifuge,
		Steps:      steps,
	})
}

// hasEnabledProvider returns true if at least one AiProvider is enabled.
func hasEnabledProvider(providers []settings.AiProvider) bool {
	for _, p := range providers {
		if p.Enabled {
			return true
		}
	}
	return false
}
