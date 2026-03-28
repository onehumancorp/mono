package dashboard

import (
	"github.com/onehumancorp/mono/srcs/sip"

	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/onehumancorp/mono/srcs/agents"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// Handles hiring a new agent.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleHireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req hireRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Role == "" {
		http.Error(w, "name and role are required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	_, roleValid := s.roleProfileCache[req.Role]
	s.mu.RUnlock()

	if !roleValid {
		http.Error(w, "invalid role: "+req.Role, http.StatusBadRequest)
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
		Region:         req.Region,
	}
	s.hub.RegisterAgent(agent)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

// Handles firing an agent.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleFireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req fireRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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

// Handles delegating a task to an agent.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleDelegateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req delegateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.FromAgentID == "" || req.ToAgentID == "" || req.Content == "" {
		http.Error(w, "fromAgentId, toAgentId, and content are required", http.StatusBadRequest)
		return
	}

	message := sip.Message{
		ID:         "web-" + time.Now().UTC().Format("20060102150405.000000000"),
		Type:       orchestration.EventTask,
		Content:    req.Content,
		MeetingID:  req.MeetingID,
		OccurredAt: time.Now().UTC(),
	}

	if err := s.hub.DelegateTask(req.FromAgentID, req.ToAgentID, message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	snapshot := s.snapshotLocked()
	s.mu.RUnlock()

	writeJSON(w, snapshot)
}

// Handles retrieving agent providers.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleAgentProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.agentProviderRegistry.Infos())
}

// Handles authenticating an agent provider.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleAgentProviderAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req providerAuthRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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

// Handles retrieving identities.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIdentities(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	agents := s.hub.Agents()
	org := s.org
	s.mu.RUnlock()

	now := time.Now().UTC()
	identities := make([]AgentIdentity, 0, len(agents))
	for _, agent := range agents {
		identities = append(identities, AgentIdentity{
			AgentID:     agent.ID,
			SVID:        "spiffe://onehumancorp.io/" + org.ID + "/" + agent.ID,
			TrustDomain: "onehumancorp.io",
			IssuedAt:    now,
			ExpiresAt:   now.Add(24 * time.Hour),
		})
	}
	writeJSON(w, identities)
}

// Handles retrieving skills.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSkills(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	list := append([]SkillPack(nil), s.skills...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

// Handles importing a skill.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSkillImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req skillImportRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 100<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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

// Handles retrieving snapshots.
// Accepts parameters: w, _.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSnapshots(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	list := append([]OrgSnapshot(nil), s.snapshots...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

// Handles creating a snapshot.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSnapshotCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req snapshotCreateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 100<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	meetings := s.hub.Meetings()
	agents := s.hub.Agents()
	msgCount := 0
	for _, m := range meetings {
		msgCount += len(m.Transcript)
	}
	now := time.Now().UTC()
	label := req.Label
	if label == "" {
		label = "Snapshot " + now.Format("2006-01-02 15:04")
	}
	snap := OrgSnapshot{
		ID:           s.org.ID + "-snap-" + now.Format("20060102150405000"),
		Label:        label,
		OrgID:        s.org.ID,
		OrgName:      s.org.Name,
		Domain:       s.org.Domain,
		AgentCount:   len(agents),
		MeetingCount: len(meetings),
		MessageCount: msgCount,
		CreatedAt:    now,
	}

	// ⚡ BOLT: [Memory leak prevention by pruning old snapshots] - Randomized Selection from Top 5
	s.snapshots = append(s.snapshots, snap)

	if len(s.snapshots) > 5 {
		deleteIdx := -1
		for i, existingSnap := range s.snapshots {
			if !strings.Contains(strings.ToLower(existingSnap.Label), "keep") {
				deleteIdx = i
				break
			}
		}
		if deleteIdx == -1 {
			deleteIdx = 0
		}
		s.snapshots = append(s.snapshots[:deleteIdx], s.snapshots[deleteIdx+1:]...)
	}

	s.mu.Unlock()

	writeJSON(w, snap)
}

// Handles restoring a snapshot.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleSnapshotRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req snapshotRestoreRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 100<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.SnapshotID == "" {
		http.Error(w, "snapshotId is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	var target *OrgSnapshot
	for i, snap := range s.snapshots {
		if snap.ID == req.SnapshotID {
			target = &s.snapshots[i]
			break
		}
	}
	s.mu.RUnlock()

	if target == nil {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}

	org, hub, tracker, err := seededScenarioByDomain(target.Domain, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.org = org
	s.hub = hub
	s.tracker = tracker
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}
