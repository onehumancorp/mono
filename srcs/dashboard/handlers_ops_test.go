package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandleScaleStreamOpsCoverage(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	_, err := authStore.CreateUser("adminuser", "admin@test.com", "adminpass123", []string{"admin"})
	if err != nil {
		t.Fatal("create user failed", err)
	}

	srv := &Server{
		org:       org,
		hub:       hub,
		tracker:   tracker,
		authStore: authStore,
	}

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/ops/scale/stream", nil)
		w := httptest.NewRecorder()
		srv.handleScaleStream(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})
}
