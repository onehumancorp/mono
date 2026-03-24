package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandleSharedMemoryBank(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	// Register an agent for testing
	hub.RegisterAgent(orchestration.Agent{
		ID:             "agent-1",
		Name:           "Test Agent",
		Role:           "SWE",
		OrganizationID: "test-org",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "memory-bank",
		Name:           "Memory Bank",
		Role:           "SYSTEM",
		OrganizationID: "test-org",
	})

	srv := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		authStore:       authStore,
		memoryBankState: make(map[string]bool),
	}

	tests := []struct {
		name           string
		method         string
		spiffeID       string
		body           string
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			spiffeID:       "",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			body:           "{invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown Fields Disallowed",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			body:           `{"event_id": "ev-1", "agent_id": "agent-1", "payload": "QUJD", "unknown_field": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing event_id",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			body:           `{"agent_id": "agent-1", "payload": "QUJD"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing agent_id",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			body:           `{"event_id": "ev-1", "payload": "QUJD"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing SPIFFE ID",
			method:         http.MethodPost,
			spiffeID:       "",
			body:           `{"event_id": "ev-1", "agent_id": "agent-1", "payload": "A"}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid SPIFFE ID Trust Domain",
			method:         http.MethodPost,
			spiffeID:       "spiffe://untrusted.com/agent/agent-1",
			body:           `{"event_id": "ev-1", "agent_id": "agent-1", "payload": "A"}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Payload Size Exceeds Bound",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			// A string of 1.4MB of 'A's (which decodes to > 1MB of bytes)
			body:           `{"event_id": "ev-big", "agent_id": "agent-1", "payload": "` + strings.Repeat("A", (1024*1024+10)*4/3) + `"}`,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Valid Request",
			method:         http.MethodPost,
			spiffeID:       "spiffe://onehumancorp.io/agent/agent-1",
			body:           `{"event_id": "ev-valid", "agent_id": "agent-1", "payload": "SGVsTG8="}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/memory_bank", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.spiffeID != "" {
				req.Header.Set("X-Spiffe-ID", tc.spiffeID)
			}
			w := httptest.NewRecorder()

			srv.handleSharedMemoryBank(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandleSharedMemoryBank_BoundedMemory(t *testing.T) {
	org := domain.NewSoftwareCompany("test-org", "Test", "CEO", time.Now())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	hub.RegisterAgent(orchestration.Agent{
		ID:             "agent-1",
		Name:           "Test Agent",
		Role:           "SWE",
		OrganizationID: "test-org",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "memory-bank",
		Name:           "Memory Bank",
		Role:           "SYSTEM",
		OrganizationID: "test-org",
	})

	srv := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		authStore:       authStore,
		memoryBankState: make(map[string]bool),
	}

	// High-frequency barrage of requests
	numRequests := 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()
			body := map[string]interface{}{
				"event_id": "ev-barrage",
				"agent_id": "agent-1",
				"payload":  "QUJD",
			}
			b, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/memory_bank", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Spiffe-ID", "spiffe://onehumancorp.io/agent/agent-1")
			w := httptest.NewRecorder()

			srv.handleSharedMemoryBank(w, req)
			// Status can be 200 OK or 409 Conflict if running concurrently with same event_id. Both are fine.
		}(i)
	}

	wg.Wait()

	// Verification of bounded memory growth (map entries should be explicitly deleted)
	srv.mu.RLock()
	mapLen := len(srv.memoryBankState)
	srv.mu.RUnlock()

	if mapLen != 0 {
		t.Errorf("expected memoryBankState map to be empty, got size %d", mapLen)
	}
}
