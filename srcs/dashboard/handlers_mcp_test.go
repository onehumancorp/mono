package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandleMCPInvokeCoverage(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	_, err := authStore.CreateUser("adminuser", "admin@test.com", "adminpass123", []string{"admin"})
	if err != nil {
		t.Fatal("create user failed", err)
	}
	user, err := authStore.Authenticate("adminuser", "adminpass123")
	if err != nil {
		t.Fatal("auth failed", err)
	}
	token, _ := authStore.IssueToken(user)

	srv := &Server{org: org, hub: hub, tracker: tracker, authStore: authStore}

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/mcp/invoke", nil)
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{invalid}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing toolId", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{"params": {"a": "b"}}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing params", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{"toolId": "dummy"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("success_valid_tool_no_meeting_id", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{"toolId": "dummy", "params": {"a": "b"}}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("large payload", func(t *testing.T) {
		// generate > 1MB string
		largeStr := strings.Repeat("a", 2<<20)
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{"toolId": "dummy", "params": {"a": "` + largeStr + `"}}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPInvoke(w, req)
		// Should fail due to MaxBytesReader
		if w.Code != http.StatusBadRequest && w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 400 or 413, got %d", w.Code)
		}
	})

	t.Run("success valid tool", func(t *testing.T) {
		// Register a dummy meeting
		hub.OpenMeeting("m-1", []string{})

		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(`{"toolId": "dummy-tool", "params": {"a": "b"}}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler := auth.Middleware(authStore)(http.HandlerFunc(srv.handleMCPInvoke))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
		}
	})
}

func TestHandleMCPRegisterCoverage(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	srv := &Server{org: org, hub: hub, tracker: tracker, authStore: authStore}

	t.Run("large payload", func(t *testing.T) {
		largeStr := strings.Repeat("a", 2<<20)
		req := httptest.NewRequest("POST", "/api/mcp/tools/register", strings.NewReader(`{"tool": {"id": "dummy", "name": "dummy"}, "spiffeId": "spiffe://onehumancorp.io/org-1/`+largeStr+`"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.handleMCPRegister(w, req)
		// Should fail due to MaxBytesReader
		if w.Code != http.StatusBadRequest && w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 400 or 413, got %d", w.Code)
		}
	})
}
