package dashboard

import (
	"encoding/json"
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

func TestHandleScale(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Wrong method",
			method:     http.MethodGet,
			body:       `{}`,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed\n",
		},
		{
			name:       "Invalid JSON body",
			method:     http.MethodPost,
			body:       `{invalid json`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid JSON payload\n",
		},
		{
			name:       "Missing role",
			method:     http.MethodPost,
			body:       `{"count": 2}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "role is required\n",
		},
		{
			name:       "Valid request",
			method:     http.MethodPost,
			body:       `{"role": "agent-1", "count": 2}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"count":2,"role":"agent-1","status":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, server, _ := newTestServer(t)
			defer server.Close()

			req := httptest.NewRequest(tt.method, "/api/v1/scale", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			app.handleScale(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantBody != "" {
				body := rec.Body.String()
				if !strings.Contains(body, tt.wantBody) {
					t.Errorf("expected body to contain %q, got %q", tt.wantBody, body)
				}
			}
		})
	}
}

func TestHandleScaleStream(t *testing.T) {
	app, server, _ := newTestServer(t)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scale/stream", nil)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		app.handleScaleStream(rec, req)
		close(done)
	}()

	<-done

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %q", contentType)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %q", cacheControl)
	}

	connection := rec.Header().Get("Connection")
	if connection != "keep-alive" {
		t.Errorf("expected Connection keep-alive, got %q", connection)
	}

	body := rec.Body.String()
	expectedEvents := []string{
		"data: {\"event\":\"K8s Operator: Reconciling TeamMember resource.\",\"status\":\"INFO\"}\n\n",
		"data: {\"event\":\"K8s Operator: Spinning up new pods...\",\"status\":\"INFO\"}\n\n",
		"data: {\"event\":\"AgentHired\",\"status\":\"Ready\"}\n\n",
	}

	for _, event := range expectedEvents {
		if !strings.Contains(body, event) {
			t.Errorf("expected body to contain %q, got:\n%s", event, body)
		}
	}
}

// ── Additional coverage: handleChatTest ──────────────────────────────────────

func TestHandleChatTestMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/chat/test", nil)
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleChatTestInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleChatTestMissingIntegrationID(t *testing.T) {
	app, _, _ := newTestServer(t)

	reqBody := `{"botToken":"foo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing integrationId, got %d", rec.Code)
	}
}

func TestHandleChatTestFailure(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Since we are not doing a real connect, testing with a fake integration might fail depending on the mock.
	// But it should return 400.
	reqBody := `{"integrationId":"unknown","botToken":"foo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for test connection failure, got %d", rec.Code)
	}
}

func TestHandleChatTestSuccess(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Builtin slack/discord mocks in integrations should succeed connection test.
	reqBody := `{"integrationId":"slack","botToken":"foo","webhookUrl":"https://hooks.slack.com/test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/chat/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleChatTest(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for test connection success, got %d", rec.Code)
	}

	var result map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode test connection response: %v", err)
	}
	if !result["success"] {
		t.Errorf("expected success: true, got %v", result)
	}
}

// ── Additional coverage: handleMCPInvoke ─────────────────────────────────────

func TestHandleMCPInvokeMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/tools/invoke", nil)
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleMCPInvokeInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleMCPInvokeMissingToolID(t *testing.T) {
	app, _, _ := newTestServer(t)

	reqBody := `{"params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing toolId, got %d", rec.Code)
	}
}

func TestHandleMCPInvokeUnknownTool(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Since we fall back to default case which acknowledges unknown tools, it should return 200
	reqBody := `{"toolId":"unknown-tool","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode mcp invoke response: %v", err)
	}
	if result["status"] != "invoked" {
		t.Errorf("expected status 'invoked', got %v", result)
	}
}

func TestHandleMCPInvokeWithNilParams(t *testing.T) {
	app, _, _ := newTestServer(t)

	reqBody := `{"toolId":"unknown-tool"}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/invoke", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleMCPInvoke(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode mcp invoke response: %v", err)
	}
	if result["status"] != "invoked" {
		t.Errorf("expected status 'invoked', got %v", result)
	}
}

// ── Additional coverage: invokeMCPTool ───────────────────────────────────────

func TestInvokeMCPToolTelegram(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Telegram missing content
	req := mcpInvokeRequest{
		ToolID: "telegram-mcp",
		Params: map[string]any{},
	}
	_, err := app.invokeMCPTool(req)
	if err == nil {
		t.Fatalf("expected error for missing content")
	}

	// Telegram missing channel (no chatspace configured fallback testing)
	req = mcpInvokeRequest{
		ToolID: "telegram-mcp",
		Params: map[string]any{
			"content": "hello",
		},
	}
	_, err = app.invokeMCPTool(req)
	if err == nil {
		t.Fatalf("expected error for missing channel")
	}

	// Telegram success
	req = mcpInvokeRequest{
		ToolID: "telegram-mcp",
		Params: map[string]any{
			"content": "hello",
			"channel": "test-channel",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if !res["delivered"].(bool) {
		t.Errorf("expected delivered: true")
	}
}

func TestInvokeMCPToolSlack(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Slack missing channel (no chatspace configured fallback testing)
	req := mcpInvokeRequest{
		ToolID: "slack-mcp",
		Params: map[string]any{
			"content": "hello",
		},
	}
	_, err := app.invokeMCPTool(req)
	if err == nil {
		t.Fatalf("expected error for missing channel")
	}

	// Slack success
	req = mcpInvokeRequest{
		ToolID: "slack-mcp",
		Params: map[string]any{
			"content": "hello",
			"channel": "test-channel",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if !res["delivered"].(bool) {
		t.Errorf("expected delivered: true")
	}
}

func TestInvokeMCPToolTeams(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Teams missing channel (no chatspace configured fallback testing)
	req := mcpInvokeRequest{
		ToolID: "teams-mcp",
		Params: map[string]any{
			"content": "hello",
		},
	}
	_, err := app.invokeMCPTool(req)
	if err == nil {
		t.Fatalf("expected error for missing channel")
	}

	// Teams success
	req = mcpInvokeRequest{
		ToolID: "teams-mcp",
		Params: map[string]any{
			"content": "hello",
			"channel": "test-channel",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if !res["delivered"].(bool) {
		t.Errorf("expected delivered: true")
	}
}

func TestInvokeMCPToolGit(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Git success
	req := mcpInvokeRequest{
		ToolID: "git-mcp",
		Params: map[string]any{
			"repository":   "test-repo",
			"title":        "test-title",
			"body":         "test-body",
			"sourceBranch": "feat-branch",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if res["pullRequest"] == nil {
		t.Errorf("expected pullRequest in response")
	}
}

func TestInvokeMCPToolJira(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Jira success
	req := mcpInvokeRequest{
		ToolID: "jira-mcp",
		Params: map[string]any{
			"project": "test-project",
			"title":   "test-title",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if res["issue"] == nil {
		t.Errorf("expected issue in response")
	}
}

func TestInvokeMCPToolLinear(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Linear success
	req := mcpInvokeRequest{
		ToolID: "linear-mcp",
		Params: map[string]any{
			"project": "test-project",
			"title":   "test-title",
		},
	}
	res, err := app.invokeMCPTool(req)
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if res["issue"] == nil {
		t.Errorf("expected issue in response")
	}
}

// ── Additional coverage: handleSettings ──────────────────────────────────────

func TestHandleSettingsMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/settings", nil)
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSettingsGetAndPost(t *testing.T) {
	app, _, _ := newTestServer(t)

	// POST settings
	reqBody := `{"minimaxApiKey":"test-minimax-key","theme":"dark"}`
	req := httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var postResp Settings
	if err := json.NewDecoder(rec.Body).Decode(&postResp); err != nil {
		t.Fatalf("decode post response: %v", err)
	}
	if postResp.MinimaxAPIKey != "test-minimax-key" {
		t.Errorf("expected minimaxApiKey 'test-minimax-key', got %q", postResp.MinimaxAPIKey)
	}

	// GET settings
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec = httptest.NewRecorder()
	app.handleSettings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var getResp Settings
	if err := json.NewDecoder(rec.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.MinimaxAPIKey != "test-minimax-key" {
		t.Errorf("expected minimaxApiKey 'test-minimax-key', got %q", getResp.MinimaxAPIKey)
	}
}

func TestHandleSettingsPostInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSettings(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
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

func TestHandleScale(t *testing.T) {
	app, _, _ := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "Wrong method",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Invalid JSON",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing role",
			method:     http.MethodPost,
			body:       `{"count": 5}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Valid request",
			method:     http.MethodPost,
			body:       `{"role": "worker", "count": 5}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/scale", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			app.handleScale(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestHandleScaleStream(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scale/stream", nil)
	rec := httptest.NewRecorder()

	app.handleScaleStream(rec, req)

	// Check headers
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %s", cc)
	}
	if conn := rec.Header().Get("Connection"); conn != "keep-alive" {
		t.Errorf("expected Connection keep-alive, got %s", conn)
	}

	body := rec.Body.String()

	// Check that expected events are flushed
	expectedEvents := []string{
		`"event":"K8s Operator: Reconciling TeamMember resource."`,
		`"event":"K8s Operator: Spinning up new pods..."`,
		`"event":"AgentHired"`,
	}

	for _, expected := range expectedEvents {
		if !strings.Contains(body, expected) {
			t.Errorf("expected stream body to contain %s, got: %s", expected, body)
		}
	}
}
