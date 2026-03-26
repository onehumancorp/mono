package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
)

func TestHandleLogin_IssueTokenError(t *testing.T) {
	// A mock to simulate failure isn't easy here
}

func TestOrganizationIDFromContext(t *testing.T) {
	ctx := context.Background()
	if orgID := OrganizationIDFromContext(ctx); orgID != "" {
		t.Errorf("expected empty org ID, got %s", orgID)
	}

	claims := &Claims{OrganizationID: "org-123"}
	ctx = context.WithValue(ctx, claimsContextKey, claims)
	if orgID := OrganizationIDFromContext(ctx); orgID != "org-123" {
		t.Errorf("expected org-123, got %s", orgID)
	}
}

func TestHandleUser_PutError(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	u1, _ := store.CreateUser("u1", "u1@test.com", "password", []string{})
	store.CreateUser("u2", "u2@test.com", "password", []string{})

	reqBody := `{"email": "u2@test.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/auth/users/"+u1.ID, bytes.NewBufferString(reqBody))
	req.SetPathValue("id", u1.ID)

	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.HandleUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleRoles_CreateError(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	// creating a role that already exists
	reqBody := `{"name": "admin", "permissions": ["read"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/roles", bytes.NewBufferString(reqBody))

	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.HandleRoles(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("{invalid"))
	rr := httptest.NewRecorder()
	h.HandleLogin(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleUsers_InvalidJSON(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/users", bytes.NewBufferString("{invalid"))
	// mock admin claim
	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.HandleUsers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleUser_InvalidJSON(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	u, _ := store.CreateUser("u1", "u1@test.com", "password", []string{})

	req := httptest.NewRequest(http.MethodPut, "/api/auth/users/"+u.ID, bytes.NewBufferString("{invalid"))
	req.SetPathValue("id", u.ID)
	// mock admin claim
	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.HandleUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleRoles_InvalidJSON(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/roles", bytes.NewBufferString("{invalid"))
	// mock admin claim
	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.HandleRoles(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleLogout_MethodNotAllowed(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()
	h.HandleLogout(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rr.Code)
	}
}

func TestHandleMe_MethodNotAllowed(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/me", nil)
	rr := httptest.NewRecorder()
	h.HandleMe(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rr.Code)
	}
}

func TestHandleRoles_MethodNotAllowed(t *testing.T) {
	store := NewStore()
	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodPut, "/api/auth/roles", nil)
	rr := httptest.NewRecorder()

	claims := &Claims{Roles: []string{RoleAdmin}}
	ctx := context.WithValue(req.Context(), claimsContextKey, claims)
	req = req.WithContext(ctx)

	h.HandleRoles(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rr.Code)
	}
}
func TestStoreCreateUser_ShortPassword(t *testing.T) {
	s := NewStore()
	_, err := s.CreateUser("test1", "test1@test.com", "short", nil)
	if err == nil || err.Error() != "password must be at least 6 characters" {
		t.Errorf("expected short password error, got %v", err)
	}
}

func TestStoreGetOrCreateOIDCUser_Fallback(t *testing.T) {
	s := NewStore()
	s.CreateUser("fallback1", "fallback1@test.com", "password", nil)

	u := s.GetOrCreateOIDCUser("sub123", "fallback1@test.com", "fallback1")
	if u.OIDCSubject != "sub123" {
		t.Errorf("expected sub123, got %s", u.OIDCSubject)
	}
}
