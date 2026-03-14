package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/dashboard"
	"github.com/onehumancorp/mono/srcs/domain"
	frontend "github.com/onehumancorp/mono/srcs/frontend/server"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestFrontendCanReachBackendAPI(t *testing.T) {
	org := domain.NewSoftwareCompany("org-1", "Acme", "CEO", time.Now().UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})
	tracker := billing.NewTracker(billing.DefaultCatalog)

	backendServer := httptest.NewServer(dashboard.NewServer(org, hub, tracker))
	defer backendServer.Close()

	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<html>frontend</html>"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}

	t.Setenv("BACKEND_URL", backendServer.URL)
	t.Setenv("FRONTEND_STATIC_DIR", staticDir)

	frontendServer, err := frontend.New()
	if err != nil {
		t.Fatalf("frontend.New error: %v", err)
	}

	proxyServer := httptest.NewServer(frontendServer.Handler())
	defer proxyServer.Close()

	resp, err := http.Get(proxyServer.URL + "/api/org")
	if err != nil {
		t.Fatalf("GET /api/org through frontend server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d body=%s", resp.StatusCode, string(b))
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode /api/org response: %v", err)
	}

	if got["id"] != org.ID {
		t.Fatalf("expected org id %s, got %v", org.ID, got["id"])
	}
}