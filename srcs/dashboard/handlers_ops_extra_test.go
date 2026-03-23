package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestHandleScaleOpsExtra(t *testing.T) {
	org := domain.Organization{ID: "test"}
	tracker := billing.NewTracker(nil)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Role",
			method:         http.MethodPost,
			body:           ScaleRequest{Count: 5},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Success",
			method:         http.MethodPost,
			body:           ScaleRequest{Role: "worker", Count: 5},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Scale Down",
			method:         http.MethodPost,
			body:           ScaleRequest{Role: "worker", Count: 2},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			// Local hub specifically for this test case
			localHub := orchestration.NewHub()
			if tt.name == "Scale Down" {
				localHub.RegisterAgent(orchestration.Agent{ID: "worker-active", Role: "worker", Status: orchestration.StatusActive})
				localHub.RegisterAgent(orchestration.Agent{ID: "worker-idle-1", Role: "worker", Status: orchestration.StatusIdle})
				localHub.RegisterAgent(orchestration.Agent{ID: "worker-idle-2", Role: "worker", Status: orchestration.StatusIdle})
			}
			localServer := &Server{
				org:     org,
				hub:     localHub,
				tracker: tracker,
			}

			req := httptest.NewRequest(tt.method, "/api/v1/scale", bytes.NewReader(body))
			w := httptest.NewRecorder()

			localServer.handleScale(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleScaleStreamOpsExtra(t *testing.T) {
	org := domain.Organization{ID: "test"}
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(nil)

	s := &Server{
		org:     org,
		hub:     hub,
		tracker: tracker,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/scale/stream", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	s.handleScaleStream(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "K8s Operator") {
		t.Errorf("expected stream to contain K8s Operator messages, got %q", body)
	}
}
