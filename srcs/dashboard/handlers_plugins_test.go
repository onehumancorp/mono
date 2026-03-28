package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandlePlugins(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	srv := &Server{org: org, hub: hub, tracker: tracker, authStore: authStore}

	t.Run("import invalid schema", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/plugins/import", bytes.NewBufferString(`invalid yaml`))
		w := httptest.NewRecorder()
		srv.handleImportPlugin(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("import missing fields", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/plugins/import", bytes.NewBufferString(`plugin_id: valid_id`))
		w := httptest.NewRecorder()
		srv.handleImportPlugin(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("import valid plugin", func(t *testing.T) {
		yamlData := `plugin_id: test-plugin
name: Test Plugin
version: v1.0.0
manifest_url: http://test-plugin/manifest.yaml`
		req := httptest.NewRequest("POST", "/api/plugins/import", bytes.NewBufferString(yamlData))
		w := httptest.NewRecorder()
		srv.handleImportPlugin(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal("failed to unmarshal response", err)
		}
		if resp["status"] != "ACTIVE" {
			t.Errorf("expected ACTIVE status, got %s", resp["status"])
		}
	})

	t.Run("get plugins", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/plugins", nil)
		w := httptest.NewRecorder()
		srv.handleGetPlugins(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid methods", func(t *testing.T) {
		req1 := httptest.NewRequest("GET", "/api/plugins/import", nil)
		w1 := httptest.NewRecorder()
		srv.handleImportPlugin(w1, req1)
		if w1.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w1.Code)
		}

		req2 := httptest.NewRequest("POST", "/api/plugins", nil)
		w2 := httptest.NewRecorder()
		srv.handleGetPlugins(w2, req2)
		if w2.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w2.Code)
		}
	})
}
