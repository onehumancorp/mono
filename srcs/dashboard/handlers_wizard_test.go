package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/settings"
)

// newWizardTestServer creates a minimal test server for wizard endpoint tests.
func newWizardTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	hub := orchestration.NewHub()
	s := &Server{
		hub:      hub,
		settings: settings.DefaultSettings(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/wizard/status", s.handleWizardStatus)
	mux.HandleFunc("/api/wizard/configure", s.handleWizardConfigure)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return s, ts
}

func TestHandleWizardStatus_Default(t *testing.T) {
	s, ts := newWizardTestServer(t)
	_ = s

	resp, err := http.Get(ts.URL + "/api/wizard/status")
	if err != nil {
		t.Fatalf("GET /api/wizard/status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result wizardStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Default settings have listen_addr and db_path set.
	if !result.Steps.Server {
		t.Error("expected Steps.Server to be true with default settings")
	}
	// Default settings have no enabled AI providers.
	if result.Steps.AiProvider {
		t.Error("expected Steps.AiProvider to be false with default settings")
	}
	// Default settings have centrifuge URL set.
	if !result.Steps.Centrifuge {
		t.Error("expected Steps.Centrifuge to be true with default settings (CentrifugeURL is set)")
	}
}

func TestHandleWizardConfigure(t *testing.T) {
	_, ts := newWizardTestServer(t)

	payload, _ := json.Marshal(wizardConfigureRequest{
		CentrifugeURL: "ws://centrifuge.example.com/connection/websocket",
		MinimaxAPIKey: "sk-test-minimax",
		AiProviders: []settings.AiProvider{
			{Name: "minimax", Model: "abab6.5s", Enabled: true},
		},
	})

	resp, err := http.Post(ts.URL+"/api/wizard/configure", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /api/wizard/configure: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result wizardStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !result.Steps.Centrifuge {
		t.Error("expected Steps.Centrifuge = true after configure")
	}
	if !result.Steps.AiProvider {
		t.Error("expected Steps.AiProvider = true after configure with enabled provider")
	}
	if !result.Configured {
		t.Error("expected Configured = true after all steps complete")
	}
}

func TestHandleWizardStatus_WrongMethod(t *testing.T) {
	_, ts := newWizardTestServer(t)
	resp, err := http.Post(ts.URL+"/api/wizard/status", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/wizard/status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandleWizardConfigure_WrongMethod(t *testing.T) {
	_, ts := newWizardTestServer(t)
	resp, err := http.Get(ts.URL + "/api/wizard/configure")
	if err != nil {
		t.Fatalf("GET /api/wizard/configure: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
