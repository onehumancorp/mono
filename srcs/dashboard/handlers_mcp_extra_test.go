package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestTokenEfficientContextSummarization(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	_, err := authStore.CreateUser("adminuser", "admin@test.com", "adminpass123", []string{"admin"})
	if err != nil {
		t.Fatal("create user failed", err)
	}
	user, err := authStore.Authenticate("adminuser", "adminpass123")
	if err != nil {
		t.Fatal("auth failed", err)
	}
	token, _ := authStore.IssueToken(user)

	srv := &Server{org: org, hub: hub, tracker: tracker, authStore: authStore}

	// Make sure events.jsonl does not exist before tests start.
	os.Remove("events.jsonl")

	t.Run("Standard Execution Flow", func(t *testing.T) {
		reqBody := `{"toolId": "token-efficient-context-summarization", "agentId": "agent-1", "params": {"event_id": "e-1", "agent_id": "a-1", "payload": {"foo": "bar"}}}`
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := auth.Middleware(authStore)(http.HandlerFunc(srv.handleMCPInvoke))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp["status"] != "summarized" {
			t.Errorf("expected status 'summarized', got %v", resp["status"])
		}

		// Verify event was logged.
		data, err := os.ReadFile("events.jsonl")
		if err != nil {
			t.Fatalf("failed to read events.jsonl: %v", err)
		}

		if !strings.Contains(string(data), "token-efficient-context-summarization") {
			t.Errorf("expected event log to contain action, got: %s", string(data))
		}
	})

	t.Run("Strict Schema and Payload Validation", func(t *testing.T) {
		reqBody := `{"toolId": "token-efficient-context-summarization", "agentId": "agent-1", "params": {"event_id": "e-1", "agent_id": "a-1", "payload": {"foo": "bar"}, "unknown_field": "123"}}`
		req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := auth.Middleware(authStore)(http.HandlerFunc(srv.handleMCPInvoke))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
			t.Errorf("expected error, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "invalid token-efficient context summarization parameters") && !strings.Contains(w.Body.String(), "invalid JSON payload") {
			t.Errorf("expected invalid params error, got: %v", w.Body.String())
		}
	})

	t.Run("Memory and Resource Bounding", func(t *testing.T) {
		// Send 100 fast requests. The rate limiter entry must be created and deleted correctly to keep memory bounded.
		for i := 0; i < 100; i++ {
			reqBody := `{"toolId": "token-efficient-context-summarization", "agentId": "agent-1", "params": {"event_id": "e-1", "agent_id": "a-1", "payload": {"foo": "bar"}}}`
			req := httptest.NewRequest("POST", "/api/mcp/invoke", strings.NewReader(reqBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler := auth.Middleware(authStore)(http.HandlerFunc(srv.handleMCPInvoke))
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("request %d failed: expected 200, got %d", i, w.Code)
			}
		}

		// Validate memory bounding.
		srv.mu.Lock()
		state, exists := srv.rateLimitStates["agent-1:token-efficient-context-summarization"]
		srv.mu.Unlock()

		if exists {
			t.Errorf("expected rateLimitStates entry to be deleted, but it exists: %v", state)
		}
	})

	os.Remove("events.jsonl")
}
