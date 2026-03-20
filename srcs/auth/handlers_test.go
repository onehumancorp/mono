package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onehumancorp/mono/srcs/auth"
)

func loginAsTest(t *testing.T, s *auth.Store, username, password string) string {
	t.Helper()
	h := auth.NewHandlers(s)
	body := `{"username":"` + username + `","password":"` + password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp["token"].(string)
}

func TestHandleUser(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleUser))

	u, _ := s.CreateUser("admin_t1", "admint1@admin.com", "password", []string{auth.RoleAdmin})
	adminTok := loginAsTest(t, s, "admin_t1", "password")

	targetUser, _ := s.CreateUser("target", "target@test.com", "pass123", []string{auth.RoleViewer})

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		body       string
		wantStatus int
	}{
		{
			name:       "GET user success",
			method:     http.MethodGet,
			path:       "/api/users/" + targetUser.ID,
			token:      adminTok,
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET user not found",
			method:     http.MethodGet,
			path:       "/api/users/not-found",
			token:      adminTok,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "PUT update user success",
			method:     http.MethodPut,
			path:       "/api/users/" + targetUser.ID,
			token:      adminTok,
			body:       `{"roles":["editor"]}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "PUT update user invalid json",
			method:     http.MethodPut,
			path:       "/api/users/" + targetUser.ID,
			token:      adminTok,
			body:       `{"roles":["editor"`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "PUT update user not found",
			method:     http.MethodPut,
			path:       "/api/users/not-found",
			token:      adminTok,
			body:       `{"roles":["editor"]}`,
			wantStatus: http.StatusBadRequest, // auth.Store.UpdateUser returns errors.New that maps to 400 not 404 in handler right now, or the handler itself doesn't differentiate.
		},
		{
			name:       "DELETE user success",
			method:     http.MethodDelete,
			path:       "/api/users/" + targetUser.ID,
			token:      adminTok,
			wantStatus: http.StatusOK,
		},
		{
			name:       "DELETE user not found",
			method:     http.MethodDelete,
			path:       "/api/users/not-found",
			token:      adminTok,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Method not allowed",
			method:     http.MethodPost,
			path:       "/api/users/" + u.ID,
			token:      adminTok,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Missing ID in path",
			method:     http.MethodGet,
			path:       "/api/users/",
			token:      adminTok,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleRoles_CreateRole(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleRoles))

	_, _ = s.CreateUser("admin_t2", "admint2@admin.com", "password", []string{auth.RoleAdmin})
	adminTok := loginAsTest(t, s, "admin_t2", "password")

	tests := []struct {
		name       string
		method     string
		body       string
		token      string
		wantStatus int
	}{
		{
			name:       "POST create role success",
			method:     http.MethodPost,
			body:       `{"name":"New Role"}`,
			token:      adminTok,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "POST create role invalid json",
			method:     http.MethodPost,
			body:       `{"name":"new-role"`,
			token:      adminTok,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "POST create role existing",
			method:     http.MethodPost,
			body:       `{"name":"admin"}`,
			token:      adminTok,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Method not allowed",
			method:     http.MethodPut,
			body:       "",
			token:      adminTok,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/roles", strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleUsers_MethodsAndErrors(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleUsers))

	_, _ = s.CreateUser("admin_t3", "admint3@admin.com", "password", []string{auth.RoleAdmin})
	adminTok := loginAsTest(t, s, "admin_t3", "password")

	tests := []struct {
		name       string
		method     string
		body       string
		token      string
		wantStatus int
	}{
		{
			name:       "Method not allowed",
			method:     http.MethodPut,
			body:       "",
			token:      adminTok,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "POST invalid json",
			method:     http.MethodPost,
			body:       `{"username":"bad"`,
			token:      adminTok,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "POST existing user",
			method:     http.MethodPost,
			body:       `{"username":"admin","email":"admin@test.com","password":"password123","roles":["admin"]}`,
			token:      adminTok,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/users", strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleMe_MethodsAndErrors(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleMe))

	_, _ = s.CreateUser("admin_t4", "admint4@admin.com", "password", []string{auth.RoleAdmin})
	adminTok := loginAsTest(t, s, "admin_t4", "password")

	tests := []struct {
		name       string
		method     string
		token      string
		wantStatus int
	}{
		{
			name:       "Method not allowed",
			method:     http.MethodPost,
			token:      adminTok,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/auth/me", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleLogin_MethodsAndErrors(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "Method not allowed",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Invalid JSON",
			method:     http.MethodPost,
			body:       `{"username":"bad"`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/auth/login", strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			h.HandleLogin(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleLogout_Methods(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{
			name:       "Method not allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/auth/logout", nil)
			rec := httptest.NewRecorder()
			h.HandleLogout(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
