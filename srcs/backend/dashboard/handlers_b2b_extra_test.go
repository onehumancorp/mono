package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/backend/auth"
	"github.com/onehumancorp/mono/srcs/backend/billing"
	"github.com/onehumancorp/mono/srcs/backend/domain"
	"github.com/onehumancorp/mono/srcs/backend/orchestration"
)

func TestHandleHandoffResolveExtraCoverage(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	_, _ = authStore.CreateUser("adminuser", "admin@test.com", "adminpass123", []string{"admin"})
	user, _ := authStore.Authenticate("adminuser", "adminpass123")
	token, _ := authStore.IssueToken(user)

	srv := &Server{
		org:       org,
		hub:       hub,
		tracker:   tracker,
		authStore: authStore,
		handoffs: []HandoffPackage{
			{
				ID:     "h-valid",
				Intent: "mock handoff",
				Status: "pending",
			},
		},
	}

	runTest := func(name string, method, path, body string, expectedCode int) {
		t.Run(name, func(t *testing.T) {
			var req *http.Request
			if body != "" {
				req = httptest.NewRequest(method, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(method, path, nil)
			}
			req.Header.Set("Authorization", "Bearer " + token)
			w := httptest.NewRecorder()
			handler := auth.Middleware(authStore)(http.HandlerFunc(srv.handleHandoffResolve))
			handler.ServeHTTP(w, req)
			if w.Code != expectedCode {
				t.Errorf("expected %d, got %d", expectedCode, w.Code)
			}
		})
	}

	runTest("invalid method", "GET", "/api/b2b/handoffs/resolve", "", http.StatusMethodNotAllowed)
	runTest("invalid json", "POST", "/api/b2b/handoffs/resolve", "{invalid}", http.StatusBadRequest)
	runTest("missing fields", "POST", "/api/b2b/handoffs/resolve", `{"handoffId": "h-1"}`, http.StatusBadRequest)
	runTest("invalid status", "POST", "/api/b2b/handoffs/resolve", `{"handoffId": "h-1", "status": "invalid_status"}`, http.StatusBadRequest)
	runTest("handoff not found", "POST", "/api/b2b/handoffs/resolve", `{"handoffId": "h-missing", "status": "resolved"}`, http.StatusNotFound)
	runTest("valid resolve", "POST", "/api/b2b/handoffs/resolve", `{"handoffId": "h-valid", "status": "resolved"}`, http.StatusOK)
}
