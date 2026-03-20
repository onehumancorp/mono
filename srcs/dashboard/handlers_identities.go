package dashboard

import (
	"net/http"
	"time"
)

// AgentIdentity represents the SPIFFE SVID certificate issued to an agent workload.
type AgentIdentity struct {
	AgentID     string    `json:"agentId"`
	SVID        string    `json:"svid"`
	TrustDomain string    `json:"trustDomain"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

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
