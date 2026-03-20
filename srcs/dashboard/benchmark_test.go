package dashboard

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/auth"
)

func BenchmarkDashboardSnapshot(b *testing.B) {
	org := domain.Organization{ID: "test-org"}
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	srv := &Server{
		org:       org,
		hub:       hub,
		tracker:   tracker,
		authStore: authStore,
		snapshots: make([]OrgSnapshot, 0),
	}

	for i := 0; i < 5; i++ {
		srv.snapshots = append(srv.snapshots, OrgSnapshot{
			ID:    "snap-",
			Label: "Snapshot",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqBody := `{"label": "Test Snapshot"}`
		req := httptest.NewRequest("POST", "/api/snapshots/create", strings.NewReader(reqBody))
		w := httptest.NewRecorder()
		srv.handleSnapshotCreate(w, req)
	}
}
