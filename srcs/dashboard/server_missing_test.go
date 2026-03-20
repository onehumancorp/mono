package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests for missing coverage in handlers

func TestHandleB2BAgreements_MethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/b2b/agreements", nil)
	rec := httptest.NewRecorder()
	app.handleB2BAgreements(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleB2BHandshake_Errors(t *testing.T) {
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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
	app, _, _ := newTestServer(t)

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


func TestHandleHealthzReadyz(t *testing.T) {
	_, server, _ := newTestServer(t)
	defer server.Close()

	tests := []struct {
		name string
		path string
	}{
		{"Healthz", "/healthz"},
		{"Readyz", "/readyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			if err != nil {
				t.Fatalf("GET %s returned error: %v", tt.path, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHandleIncidentStatus_UpdateRCA(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantRCA string
	}{
		{
			name:    "Update RCA",
			payload: `{"incidentId":"inc-1", "status":"RESOLVED", "rootCauseAnalysis":"It was DNS"}`,
			wantRCA: "It was DNS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, server, _ := newTestServer(t)
			defer server.Close()

			app.incidents = append(app.incidents, Incident{
				ID:     "inc-1",
				Status: IncidentStatusInvestigating,
			})

			req := httptest.NewRequest(http.MethodPost, "/api/incidents/status", strings.NewReader(tt.payload))
			rec := httptest.NewRecorder()
			app.handleIncidentStatus(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			if len(app.incidents) != 1 {
				t.Fatalf("expected 1 incident, got %d", len(app.incidents))
			}

			if app.incidents[0].RCA != tt.wantRCA {
				t.Errorf("expected RCA to be %q, got %q", tt.wantRCA, app.incidents[0].RCA)
			}
		})
	}
}

func TestHandleBudgetAlerts_NotifyAtPctHandling(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantPct float64
	}{
		{
			name:    "Invalid pct defaults",
			payload: `{"thresholdUsd":100, "notifyAtPct": 1.5}`,
			wantPct: defaultBudgetAlertNotifyPct,
		},
		{
			name:    "Valid pct",
			payload: `{"thresholdUsd":100, "notifyAtPct": 0.5}`,
			wantPct: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, server, _ := newTestServer(t)
			defer server.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/billing/alerts", strings.NewReader(tt.payload))
			rec := httptest.NewRecorder()
			app.handleBudgetAlerts(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			if len(app.budgetAlerts) != 1 {
				t.Fatalf("expected 1 budget alert, got %d", len(app.budgetAlerts))
			}

			if app.budgetAlerts[0].NotifyAtPct != tt.wantPct {
				t.Errorf("expected NotifyAtPct to be %v, got %v", tt.wantPct, app.budgetAlerts[0].NotifyAtPct)
			}
		})
	}
}

func TestHandleSettings(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test GET method
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Test PUT method (unsupported)
	req = httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	app.handleSettings(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSettingsPostValid(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test POST method with valid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(`{"minimaxApiKey":"test"}`))
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleSettingsPostInvalid(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test POST method with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(`bad json`))
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleChatTest(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test GET method (unsupported)
	req := httptest.NewRequest(http.MethodGet, "/api/integrations/chat/test", nil)
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

    // Test Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(`bad json`))
	rec = httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleChatTestValid(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test POST method with valid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(`{"integrationId":"telegram"}`))
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest { // Need credentials logic
		t.Errorf("expected 400, got %d", rec.Code)
	}
}


func TestHandleMCPInvoke(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test GET method (unsupported)
	req := httptest.NewRequest(http.MethodGet, "/api/mcp/tools/invoke", nil)
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

    // Test Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(`bad json`))
	rec = httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleMCPInvokeValid(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test POST method with valid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(`{"toolId":"slack-mcp", "params":{"integrationId":"slack"}}`))
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusBadRequest { // Not found tool
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleChatTestValid3(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Valid POST but missing parameters logic
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(`{"integrationId":""}`))
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleMCPInvokeValid3(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Valid POST but missing parameters logic
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(`{"toolId":""}`))
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestInvokeMCPTool(t *testing.T) {
	app, _, _ := newTestServer(t)

	tests := []struct {
		name    string
		req     mcpInvokeRequest
		wantErr bool
	}{
		{
			name: "telegram-mcp missing content",
			req: mcpInvokeRequest{
				ToolID: "telegram-mcp",
				Params: map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "slack-mcp missing channel",
			req: mcpInvokeRequest{
				ToolID: "slack-mcp",
				Params: map[string]any{
					"content": "hello",
				},
			},
			wantErr: true,
		},
        {
			name: "git-mcp",
			req: mcpInvokeRequest{
				ToolID: "git-mcp",
				Params: map[string]any{
                    "integrationId": "github",
                    "repository": "repo",
                    "title": "title",
                    "body": "body",
                    "sourceBranch": "source",
				},
			},
			wantErr: false, // we don't have git integrations setup so it doesn't fail due to mock
		},
        {
			name: "jira-mcp",
			req: mcpInvokeRequest{
				ToolID: "jira-mcp",
				Params: map[string]any{
                    "integrationId": "jira",
                    "project": "proj",
                    "title": "title",
                    "description": "desc",
				},
			},
			wantErr: false, // mock returns nil
		},
        {
			name: "linear-mcp",
			req: mcpInvokeRequest{
				ToolID: "linear-mcp",
				Params: map[string]any{
                    "integrationId": "linear",
                    "project": "proj",
                    "title": "title",
                    "description": "desc",
				},
			},
			wantErr: false, // mock returns nil
		},
		{
			name: "unknown tool",
			req: mcpInvokeRequest{
				ToolID: "unknown-mcp",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.invokeMCPTool(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("invokeMCPTool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleB2BHandshakeValid(t *testing.T) {
	app, server, _ := newTestServer(t)
    defer server.Close()

	// Valid payload
	payload := `{"partnerOrg":"acme.com", "partnerJwksUrl":"https://acme.com/.well-known/jwks.json"}`
	req := httptest.NewRequest(http.MethodPost, "/api/b2b/handshake", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	app.handleB2BHandshake(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

    // Verify it updated internal array
    app.mu.RLock()
    if len(app.trustAgreements) != 1 {
        t.Errorf("expected 1 agreement, got %d", len(app.trustAgreements))
    }
    app.mu.RUnlock()

    // Verify it updated the gateway
    if _, ok := app.b2bGateway.GetAgreement("acme.com"); !ok {
        t.Errorf("expected gateway to have acme.com")
    }
}

func TestHandleB2BRevokeValid(t *testing.T) {
	app, server, _ := newTestServer(t)
    defer server.Close()

	// Valid payload
	payload := `{"partnerOrg":"acme.com", "partnerJwksUrl":"https://acme.com/.well-known/jwks.json"}`
	req := httptest.NewRequest(http.MethodPost, "/api/b2b/handshake", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	app.handleB2BHandshake(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

    app.mu.RLock()
    id := app.trustAgreements[0].ID
    app.mu.RUnlock()

    revokePayload := `{"agreementId":"` + id + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/b2b/revoke", strings.NewReader(revokePayload))
	rec = httptest.NewRecorder()
	app.handleB2BRevoke(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

    // Verify status is REVOKED
    app.mu.RLock()
    if app.trustAgreements[0].Status != TrustStatusRevoked {
        t.Errorf("expected status revoked")
    }
    app.mu.RUnlock()

    // Verify gateway no longer has it
    if _, ok := app.b2bGateway.GetAgreement("acme.com"); ok {
        t.Errorf("expected gateway to not have acme.com anymore")
    }
}
