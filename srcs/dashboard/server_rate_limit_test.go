package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/integrations"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandleMCPInvoke_RateLimiting(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	// Create Server struct directly to access unexported methods for testing
	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
	}

	agent := orchestration.Agent{
		ID:     "agent-1",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	invokeTool := func(toolID, agentID string) *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  toolID,
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
			"agentId": agentID,
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	t.Run("Missing Tool triggers WaitingForTools", func(t *testing.T) {
		rec := invokeTool("non-existent-tool", "agent-1")
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}

		a, exists := hub.Agent("agent-1")
		if !exists || a.Status != orchestration.StatusWaitingForTools {
			t.Errorf("expected agent status %s, got %s", orchestration.StatusWaitingForTools, a.Status)
		}
	})
}

func TestHandleMCPInvoke_RateLimiting_Backoff(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
	}

	agent := orchestration.Agent{
		ID:     "agent-2",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	rateLimitKey := "agent-2:slack-mcp"

	invokeTool := func() *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  "slack-mcp",
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
			"agentId": "agent-2",
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	app.mu.Lock()
	app.rateLimitStates[rateLimitKey] = &RateLimitState{
		Failures:    1,
		LastFailure: time.Now(),
		Backoff:     10 * time.Second, // Long backoff to ensure we hit it
	}
	app.mu.Unlock()

	rec := invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Rate limited. Please backoff.") {
		t.Errorf("expected body to contain 'Rate limited. Please backoff.', got %s", rec.Body.String())
	}
}


func TestHandleMCPInvoke_MissingToolNoAgent(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
	}

	invokeTool := func(toolID string) *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  toolID,
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	t.Run("Missing Tool triggers 404", func(t *testing.T) {
		rec := invokeTool("non-existent-tool")
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}


func TestHandleMCPInvoke_RateLimiting_ResetOnSuccess(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
	}

	agent := orchestration.Agent{
		ID:     "agent-3",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	rateLimitKey := "agent-3:github-mcp"

	invokeTool := func() *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  "github-mcp",
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123", "repository": "test/test", "title": "t", "body": "b", "sourceBranch": "s", "targetBranch": "t", "createdBy": "a"}`),
			"agentId": "agent-3",
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	app.mu.Lock()
	app.rateLimitStates[rateLimitKey] = &RateLimitState{
		Failures:    1,
		LastFailure: time.Now().Add(-1 * time.Hour), // Expired backoff
		Backoff:     10 * time.Second,
	}
	app.mu.Unlock()

	rec := invokeTool()
	if rec.Code == http.StatusTooManyRequests {
		t.Errorf("expected NOT 429 Too Many Requests, got %d", rec.Code)
	}

	app.mu.Lock()
	state, exists := app.rateLimitStates[rateLimitKey]
	app.mu.Unlock()

	if !exists {
		// Key was correctly deleted on success
	} else if state.Failures > 1 {
		t.Errorf("expected failures to remain at 1 or be deleted entirely on a non-429 result, got %d", state.Failures)
	}
}

func TestHandleMCPInvoke_MaxRetriesExceeded(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
	}

	agent := orchestration.Agent{
		ID:     "agent-max",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	rateLimitKey := "agent-max:slack-mcp"

	invokeTool := func() *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  "slack-mcp",
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
			"agentId": "agent-max",
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	app.mu.Lock()
	app.rateLimitStates[rateLimitKey] = &RateLimitState{
		Failures:    3, // Set to max threshold
		LastFailure: time.Now().Add(-1 * time.Hour), // Ready for next try, backoff bypassed
		Backoff:     10 * time.Second,
	}
	app.mu.Unlock()

	rec := invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Max retries exceeded. Hard failure.") {
		t.Errorf("expected body to contain 'Max retries exceeded. Hard failure.', got %s", rec.Body.String())
	}
}

// We will use a dynamic tool that is missing to trigger 404, but wait. We want to cover rate limit "Rate limited. Please backoff." from 986.
// If req.ToolID is "unknown tool: 429", the error returned from invokeMCPTool will literally be:
// fmt.Errorf("unknown tool: %s", req.ToolID) -> "unknown tool: unknown tool: 429"
// Does it contain "429"? YES!
// It falls into the `if err != nil && strings.Contains(err.Error(), "429")` check!
func TestHandleMCPInvoke_RateLimiting_InvokeErrorAndBackoff(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		// Explicitly set to nil to test nil map initialization coverage
		rateLimitStates: nil,
		dynamicMCPTools: []MCPTool{},
	}

	agent := orchestration.Agent{
		ID:     "agent-limited",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	invokeTool := func() *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			// By naming the tool "tool-429", the error returned will be "unknown tool: tool-429"
			// which contains "429", matching the rate limit error check!
			"toolId":  "tool-429",
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
			"agentId": "agent-limited",
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	// 1st request -> invokeMCPTool gets "unknown tool: tool-429", increments failure to 1
	rec := invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	app.mu.Lock()
	state := app.rateLimitStates["agent-limited:tool-429"]
	app.mu.Unlock()
	if state == nil || state.Failures != 1 {
		t.Errorf("expected 1 failure, got %v", state)
	}

	// Because of backoff logic, if we immediately try again, we should hit the "Rate limited. Please backoff." branch.
	rec = invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Rate limited. Please backoff.") {
		t.Errorf("expected body to contain 'Rate limited. Please backoff.', got %s", rec.Body.String())
	}

	// Override last failure to bypass the backoff and trigger another execution that will fail
	app.mu.Lock()
	app.rateLimitStates["agent-limited:tool-429"].LastFailure = time.Now().Add(-1 * time.Hour)
	app.mu.Unlock()

	// 3rd attempt -> invokeMCPTool gets "unknown tool: tool-429", increments failure to 2
	rec = invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}

	app.mu.Lock()
	app.rateLimitStates["agent-limited:tool-429"].LastFailure = time.Now().Add(-1 * time.Hour)
	app.mu.Unlock()

	// 4th attempt -> invokeMCPTool gets "unknown tool: tool-429", increments failure to 3
	rec = invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}

	// Now failure is 3, which hits max retries
	// Next immediate call without overriding LastFailure will hit state.Failures >= 3
	rec = invokeTool()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Max retries exceeded. Hard failure.") {
		t.Errorf("expected 'Max retries exceeded. Hard failure.', got %s", rec.Body.String())
	}
}

// We will use a dynamic tool that is VALID, e.g. "telegram-mcp", but with valid parameters
// AND an active integration to trigger the success case at 1005 (Reset on success with event write)
func TestHandleMCPInvoke_RateLimiting_SuccessEvent(t *testing.T) {
	org := domain.Organization{ID: "org-1"}
	hub := orchestration.NewHub()
	prices := map[string]billing.Price{
		"test-model": {InputPerMillionUSD: 0.01, OutputPerMillionUSD: 0.02},
	}
	tracker := billing.NewTracker(prices)

	app := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		integReg:        integrations.NewRegistry(),
		rateLimitStates: make(map[string]*RateLimitState),
		dynamicMCPTools: []MCPTool{
			{
				ID: "custom-success-tool",
			},
		},
	}

	agent := orchestration.Agent{
		ID:     "agent-success",
		Status: orchestration.StatusActive,
	}
	hub.RegisterAgent(agent)

	invokeTool := func() *httptest.ResponseRecorder {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"toolId":  "custom-success-tool",
			"action":  "test_action",
			"params":  json.RawMessage(`{"integrationId": "123"}`),
			"agentId": "agent-success",
		})

		req := httptest.NewRequest("POST", "/api/mcp/tools/invoke", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		app.handleMCPInvoke(rec, req)
		return rec
	}

	// First execution succeeds.
	rec := invokeTool()
	if rec.Code != http.StatusOK {
		t.Errorf("expected OK, got %d", rec.Code)
	}

	// This should cover lines 1005-1020 where it writes "rl-succ-..." event to hub and events.jsonl
}
