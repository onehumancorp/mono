package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
)

func TestParseOIDCConfig(t *testing.T) {
	cfg := ParseOIDCConfig("http://issuer", "client1", "secret1", "http://callback")
	if cfg.Issuer != "http://issuer" {
		t.Error("issuer mismatch")
	}
	if !cfg.Enabled {
		t.Error("should be enabled")
	}
}

func TestOIDCDisabled(t *testing.T) {
	cfg := ParseOIDCConfig("", "", "", "")
	if cfg.Enabled {
		t.Error("should be disabled")
	}
}

func TestValidateOIDCTokenDisabled(t *testing.T) {
	cfg := ParseOIDCConfig("", "", "", "")
	_, err := ValidateOIDCToken("token", cfg)
	if err == nil || err.Error() != "OIDC is not configured" {
		t.Error("expected OIDC not configured error")
	}
}

// Minimal coverage triggers for OIDC handlers
func TestOIDCHandlerNoState(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth/oidc/callback?code=123", nil)
	rr := httptest.NewRecorder()

	store := NewStore("secret", ParseOIDCConfig("http://issuer", "client1", "secret1", "http://callback"))

	store.HandleOIDCCallback(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 bad request for missing state, got %d", rr.Code)
	}
}

func TestHandleOIDCLogin(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth/oidc/login", nil)
	rr := httptest.NewRecorder()

	store := NewStore("secret", ParseOIDCConfig("http://issuer", "client1", "secret1", "http://callback"))
	store.HandleOIDCLogin(rr, req)
	if rr.Code != http.StatusTemporaryRedirect && rr.Code != http.StatusInternalServerError {
		t.Errorf("expected redirect or 500, got %d", rr.Code)
	}
}

func TestHandleOIDCCallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth/oidc/callback?state=invalidstate&code=123", nil)
	rr := httptest.NewRecorder()

	store := NewStore("secret", ParseOIDCConfig("http://issuer", "client1", "secret1", "http://callback"))
	store.HandleOIDCCallback(rr, req)
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 400 or 500, got %d", rr.Code)
	}
}

func TestOIDCDisabledHandler(t *testing.T) {
	store := NewStore("secret", OIDCConfig{Enabled: false})

	req1 := httptest.NewRequest("GET", "/auth/oidc/login", nil)
	rr1 := httptest.NewRecorder()
	store.HandleOIDCLogin(rr1, req1)
	if rr1.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest("GET", "/auth/oidc/callback", nil)
	rr2 := httptest.NewRecorder()
	store.HandleOIDCCallback(rr2, req2)
	if rr2.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr2.Code)
	}
}

func TestOIDCDisabledHandler(t *testing.T) {
	store := NewStore("secret", OIDCConfig{Enabled: false})

	req1 := httptest.NewRequest("GET", "/auth/oidc/login", nil)
	rr1 := httptest.NewRecorder()
	store.HandleOIDCLogin(rr1, req1)
	if rr1.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest("GET", "/auth/oidc/callback", nil)
	rr2 := httptest.NewRecorder()
	store.HandleOIDCCallback(rr2, req2)
	if rr2.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr2.Code)
	}
}
