package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// SkillPackRole pairs a role name with its override base prompt.
type SkillPackRole struct {
	Role       string `json:"role"`
	BasePrompt string `json:"basePrompt"`
}

// SkillPack is an importable module that extends or overrides agent capabilities.
type SkillPack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Domain      string          `json:"domain"`
	Description string          `json:"description"`
	Source      string          `json:"source"` // builtin | custom | marketplace
	Author      string          `json:"author,omitempty"`
	Roles       []SkillPackRole `json:"roles"`
	ImportedAt  time.Time       `json:"importedAt"`
}

type skillImportRequest struct {
	Name        string          `json:"name"`
	Domain      string          `json:"domain"`
	Description string          `json:"description"`
	Source      string          `json:"source"`
	Author      string          `json:"author,omitempty"`
	Roles       []SkillPackRole `json:"roles"`
}

func (s *Server) handleSkills(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	list := append([]SkillPack(nil), s.skills...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

func (s *Server) handleSkillImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req skillImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Domain == "" {
		http.Error(w, "name and domain are required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	source := req.Source
	if source == "" {
		source = "custom"
	}
	pack := SkillPack{
		ID:          s.org.ID + "-skill-" + now.Format("20060102150405000"),
		Name:        req.Name,
		Domain:      req.Domain,
		Description: req.Description,
		Source:      source,
		Author:      req.Author,
		Roles:       req.Roles,
		ImportedAt:  now,
	}
	if pack.Roles == nil {
		pack.Roles = []SkillPackRole{}
	}

	s.mu.Lock()
	s.skills = append(s.skills, pack)
	s.mu.Unlock()

	writeJSON(w, pack)
}
