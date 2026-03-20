package orchestration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/domain"
)

// B2BGateway handles cross-organizational communications via HTTP.
type B2BGateway struct {
	mu         sync.RWMutex
	agreements map[string]domain.TrustAgreement
	hub        *Hub
}

// NewB2BGateway constructs a B2B gateway.
func NewB2BGateway(hub *Hub) *B2BGateway {
	return &B2BGateway{
		agreements: make(map[string]domain.TrustAgreement),
		hub:        hub,
	}
}

// AddAgreement stores a trust agreement allowing incoming connections
// from the specified partner.
func (g *B2BGateway) AddAgreement(agreement domain.TrustAgreement) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.agreements[agreement.PartnerOrg] = agreement
}

// GetAgreement returns the trust agreement for a partner org.
func (g *B2BGateway) GetAgreement(partnerOrg string) (domain.TrustAgreement, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	agr, ok := g.agreements[partnerOrg]
	return agr, ok
}

// RemoveAgreement revokes trust for a partner organization.
func (g *B2BGateway) RemoveAgreement(partnerOrg string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.agreements, partnerOrg)
}

// ReceiveMessage processes an incoming B2B message and injects it into the Hub.
func (g *B2BGateway) ReceiveMessage(ctx context.Context, b2bMsg domain.B2BMessage) error {
	// Validate trust
	agr, ok := g.GetAgreement(b2bMsg.FromOrg)
	if !ok || agr.Status != "ACTIVE" {
		return errors.New("unauthorized partner organization or inactive agreement")
	}

	// Validate roles (simplified mock check)
	// Real implementation would verify the role via JWT claims against the JWKS endpoint
	if len(agr.AllowedRoles) > 0 {
		// In a real scenario we'd decode the token and get the role.
		// For now we accept it to bridge the systems.
	}

	// Sanitize content (Egress filter logic conceptually runs here or on the sender side)
	if strings.Contains(strings.ToLower(b2bMsg.Content), "internal project x") {
		return errors.New("blocked by data loss prevention (DLP)")
	}

	// Map into a standard Hub Message
	internalMsg := Message{
		ID:         fmt.Sprintf("b2b-%s", b2bMsg.ID),
		FromAgent:  b2bMsg.FromAgent, // Needs to be registered dynamically as external
		ToAgent:    b2bMsg.ToAgent,
		Type:       EventTask,
		Content:    b2bMsg.Content,
		OccurredAt: time.Now().UTC(),
	}

	// For an external agent sending a message, we must temporarily register them
	// in the Hub as a shadow agent if they don't exist
	if _, exists := g.hub.Agent(b2bMsg.FromAgent); !exists {
		shadow := Agent{
			ID:             b2bMsg.FromAgent,
			Name:           fmt.Sprintf("External (%s)", b2bMsg.FromOrg),
			Role:           "Partner Agent",
			OrganizationID: b2bMsg.FromOrg,
			Status:         StatusActive,
		}
		g.hub.RegisterAgent(shadow)
	}

	return g.hub.Publish(internalMsg)
}

// HandleB2BEndpoint provides a standard HTTP handler for receiving B2B messages.
func (g *B2BGateway) HandleB2BEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg domain.B2BMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := g.ReceiveMessage(r.Context(), msg); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"delivered"}`))
}
