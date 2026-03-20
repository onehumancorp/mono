package domain

import "testing"

func TestTrustAgreement(t *testing.T) {
	agreement := TrustAgreement{
		ID:           "test-1",
		PartnerOrg:   "globex.com",
		PartnerJWKS:  "https://globex.com/.well-known/jwks.json",
		AllowedRoles: []string{"Sales Agent", "Buyer Agent"},
		Status:       "ACTIVE",
	}

	if agreement.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", agreement.ID)
	}
	if agreement.PartnerOrg != "globex.com" {
		t.Errorf("Expected PartnerOrg 'globex.com', got '%s'", agreement.PartnerOrg)
	}
	if agreement.PartnerJWKS != "https://globex.com/.well-known/jwks.json" {
		t.Errorf("Expected PartnerJWKS 'https://globex.com/.well-known/jwks.json', got '%s'", agreement.PartnerJWKS)
	}
	if len(agreement.AllowedRoles) != 2 {
		t.Errorf("Expected 2 AllowedRoles, got %d", len(agreement.AllowedRoles))
	}
	if agreement.Status != "ACTIVE" {
		t.Errorf("Expected Status 'ACTIVE', got '%s'", agreement.Status)
	}
}

func TestB2BMessage(t *testing.T) {
	msg := B2BMessage{
		ID:        "msg-1",
		FromOrg:   "acme.corp",
		ToOrg:     "globex.com",
		FromAgent: "buyer-1",
		ToAgent:   "sales-1",
		Content:   "Hello",
		CrossOrg:  true,
		Signature: "sig",
	}

	if msg.ID != "msg-1" {
		t.Errorf("Expected ID 'msg-1', got '%s'", msg.ID)
	}
	if msg.FromOrg != "acme.corp" {
		t.Errorf("Expected FromOrg 'acme.corp', got '%s'", msg.FromOrg)
	}
	if msg.ToOrg != "globex.com" {
		t.Errorf("Expected ToOrg 'globex.com', got '%s'", msg.ToOrg)
	}
	if msg.FromAgent != "buyer-1" {
		t.Errorf("Expected FromAgent 'buyer-1', got '%s'", msg.FromAgent)
	}
	if msg.ToAgent != "sales-1" {
		t.Errorf("Expected ToAgent 'sales-1', got '%s'", msg.ToAgent)
	}
	if msg.Content != "Hello" {
		t.Errorf("Expected Content 'Hello', got '%s'", msg.Content)
	}
	if msg.CrossOrg != true {
		t.Errorf("Expected CrossOrg true, got %v", msg.CrossOrg)
	}
	if msg.Signature != "sig" {
		t.Errorf("Expected Signature 'sig', got '%s'", msg.Signature)
	}
}
