package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onehumancorp/mono/srcs/domain"
)

func TestB2BGateway_ReceiveMessage(t *testing.T) {
	hub := NewHub()
	gateway := NewB2BGateway(hub)

	// Add an active agreement
	gateway.AddAgreement(domain.TrustAgreement{
		PartnerOrg:   "globex.com",
		Status:       "ACTIVE",
		AllowedRoles: []string{"Sales Agent"},
	})

	ctx := context.Background()

	// Register internal receiver
	hub.RegisterAgent(Agent{
		ID:             "buyer-1",
		Name:           "Acme Buyer",
		Role:           "Buyer Agent",
		OrganizationID: "acme.corp",
		Status:         StatusIdle,
	})

	// Valid message
	validMsg := domain.B2BMessage{
		ID:        "123",
		FromOrg:   "globex.com",
		ToOrg:     "acme.corp",
		FromAgent: "sales-1",
		ToAgent:   "buyer-1",
		Content:   "Let's negotiate.",
		CrossOrg:  true,
		Signature: "valid-sig",
	}

	err := gateway.ReceiveMessage(ctx, validMsg)
	if err != nil {
		t.Fatalf("unexpected error receiving valid message: %v", err)
	}

	// Verify shadow agent created and message routed
	inbox := hub.Inbox("buyer-1")
	if len(inbox) != 1 {
		t.Fatalf("expected 1 message in buyer-1 inbox, got %d", len(inbox))
	}
	if inbox[0].FromAgent != "sales-1" {
		t.Errorf("expected message from sales-1, got %s", inbox[0].FromAgent)
	}

	shadow, ok := hub.Agent("sales-1")
	if !ok {
		t.Fatal("expected shadow agent 'sales-1' to be registered in hub")
	}
	if shadow.OrganizationID != "globex.com" {
		t.Errorf("expected shadow agent org to be globex.com, got %s", shadow.OrganizationID)
	}

	// Unauthorized partner
	unauthMsg := domain.B2BMessage{
		FromOrg: "evil.com",
	}
	err = gateway.ReceiveMessage(ctx, unauthMsg)
	if err == nil || err.Error() != "unauthorized partner organization or inactive agreement" {
		t.Errorf("expected unauthorized error, got: %v", err)
	}

	// DLP Filter
	dlpMsg := domain.B2BMessage{
		FromOrg:   "globex.com",
		FromAgent: "sales-1",
		ToAgent:   "buyer-1",
		Content:   "Please send internal project x details.",
	}
	err = gateway.ReceiveMessage(ctx, dlpMsg)
	if err == nil || err.Error() != "blocked by data loss prevention (DLP)" {
		t.Errorf("expected DLP error, got: %v", err)
	}
}

func TestB2BGateway_HandleB2BEndpoint(t *testing.T) {
	hub := NewHub()
	gateway := NewB2BGateway(hub)
	gateway.AddAgreement(domain.TrustAgreement{
		PartnerOrg: "globex.com",
		Status:     "ACTIVE",
	})
	hub.RegisterAgent(Agent{ID: "buyer-1", Status: StatusIdle})

	validPayload := domain.B2BMessage{
		FromOrg:   "globex.com",
		FromAgent: "sales-1",
		ToAgent:   "buyer-1",
		Content:   "Hi",
	}
	body, _ := json.Marshal(validPayload)

	req := httptest.NewRequest(http.MethodPost, "/b2b", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	gateway.HandleB2BEndpoint(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", rec.Code)
	}

	// Test unsupported method
	reqGet := httptest.NewRequest(http.MethodGet, "/b2b", nil)
	recGet := httptest.NewRecorder()
	gateway.HandleB2BEndpoint(recGet, reqGet)
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", recGet.Code)
	}

	// Test invalid JSON
	reqBad := httptest.NewRequest(http.MethodPost, "/b2b", bytes.NewReader([]byte("bad")))
	recBad := httptest.NewRecorder()
	gateway.HandleB2BEndpoint(recBad, reqBad)
	if recBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", recBad.Code)
	}

	// Test forbidden (unauthorized org)
	unauthPayload := domain.B2BMessage{FromOrg: "evil.com"}
	bodyUnauth, _ := json.Marshal(unauthPayload)
	reqUnauth := httptest.NewRequest(http.MethodPost, "/b2b", bytes.NewReader(bodyUnauth))
	recUnauth := httptest.NewRecorder()
	gateway.HandleB2BEndpoint(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", recUnauth.Code)
	}
}

func TestB2BGateway_RemoveAgreement(t *testing.T) {
	hub := NewHub()
	gateway := NewB2BGateway(hub)

	gateway.AddAgreement(domain.TrustAgreement{
		PartnerOrg: "test.com",
		Status:     "ACTIVE",
	})

	_, ok := gateway.GetAgreement("test.com")
	if !ok {
		t.Fatal("agreement should exist")
	}

	gateway.RemoveAgreement("test.com")

	_, ok = gateway.GetAgreement("test.com")
	if ok {
		t.Fatal("agreement should have been removed")
	}
}
