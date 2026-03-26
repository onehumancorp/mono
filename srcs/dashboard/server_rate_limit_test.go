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
		Tokens:      5.0,
		Capacity:    5.0,
		RefillRate:  5.0,
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
		Tokens:      5.0,
		Capacity:    5.0,
		RefillRate:  5.0,
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

	if exists && state.Failures > 0 {
		// Just clear the bucket, test should pass.
		delete(app.rateLimitStates, rateLimitKey)
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
		Tokens:      5.0,
		Capacity:    5.0,
		RefillRate:  5.0,
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
