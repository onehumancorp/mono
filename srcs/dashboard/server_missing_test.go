package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests for missing coverage in handlers

func TestHandleB2BAgreements_MethodNotAllowed(t *testing.T) {
	app, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/b2b/agreements", nil)
	rec := httptest.NewRecorder()
	app.handleB2BAgreements(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleB2BHandshake_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodGet, "/api/b2b/handshake", nil)
	rec := httptest.NewRecorder()
	app.handleB2BHandshake(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/handshake", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleB2BHandshake(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing fields
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/handshake", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleB2BHandshake(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleB2BRevoke_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodGet, "/api/b2b/revoke", nil)
	rec := httptest.NewRecorder()
	app.handleB2BRevoke(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/revoke", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleB2BRevoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing agreementId
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/revoke", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleB2BRevoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Not found
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/revoke", strings.NewReader(`{"agreementId":"missing"}`))
	rec = httptest.NewRecorder()
	app.handleB2BRevoke(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleIncidents_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodDelete, "/api/incidents", nil)
	rec := httptest.NewRecorder()
	app.handleIncidents(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/incidents", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleIncidents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing fields
	req = httptest.NewRequest(http.MethodPost, "/api/incidents", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleIncidents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleIncidentStatus_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodGet, "/api/incidents/status", nil)
	rec := httptest.NewRecorder()
	app.handleIncidentStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/incidents/status", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleIncidentStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing fields
	req = httptest.NewRequest(http.MethodPost, "/api/incidents/status", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleIncidentStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Not found
	req = httptest.NewRequest(http.MethodPost, "/api/incidents/status", strings.NewReader(`{"incidentId":"missing", "status":"resolved"}`))
	rec = httptest.NewRecorder()
	app.handleIncidentStatus(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleComputeProfiles_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodDelete, "/api/compute/profiles", nil)
	rec := httptest.NewRecorder()
	app.handleComputeProfiles(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/compute/profiles", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleComputeProfiles(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing roleId
	req = httptest.NewRequest(http.MethodPost, "/api/compute/profiles", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleComputeProfiles(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleClusterStatus_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodPost, "/api/clusters/eu/status", nil)
	rec := httptest.NewRecorder()
	app.handleClusterStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Missing region (path parsing)
	req = httptest.NewRequest(http.MethodGet, "/api/clusters", nil)
	rec = httptest.NewRecorder()
	app.handleClusterStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleBudgetAlerts_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodDelete, "/api/billing/alerts", nil)
	rec := httptest.NewRecorder()
	app.handleBudgetAlerts(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/billing/alerts", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handleBudgetAlerts(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Invalid threshold
	req = httptest.NewRequest(http.MethodPost, "/api/billing/alerts", strings.NewReader(`{"thresholdUsd":0}`))
	rec = httptest.NewRecorder()
	app.handleBudgetAlerts(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePipelines_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodDelete, "/api/pipelines", nil)
	rec := httptest.NewRecorder()
	app.handlePipelines(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handlePipelines(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing name
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handlePipelines(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePipelinePromote_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodGet, "/api/pipelines/promote", nil)
	rec := httptest.NewRecorder()
	app.handlePipelinePromote(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/promote", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handlePipelinePromote(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing pipelineId
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/promote", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handlePipelinePromote(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Not found
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/promote", strings.NewReader(`{"pipelineId":"missing"}`))
	rec = httptest.NewRecorder()
	app.handlePipelinePromote(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	// Not in STAGING
	app.pipelines = append(app.pipelines, Pipeline{ID: "pipe-1", Status: PipelineStatusPending})
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/promote", strings.NewReader(`{"pipelineId":"pipe-1"}`))
	rec = httptest.NewRecorder()
	app.handlePipelinePromote(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePipelineStatus_Errors(t *testing.T) {
	app, _ := newTestServer(t)

	// Wrong method
	req := httptest.NewRequest(http.MethodGet, "/api/pipelines/status", nil)
	rec := httptest.NewRecorder()
	app.handlePipelineStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/status", strings.NewReader("bad json"))
	rec = httptest.NewRecorder()
	app.handlePipelineStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Missing fields
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/status", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handlePipelineStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Not found
	req = httptest.NewRequest(http.MethodPost, "/api/pipelines/status", strings.NewReader(`{"pipelineId":"missing", "status":"STAGING"}`))
	rec = httptest.NewRecorder()
	app.handlePipelineStatus(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
