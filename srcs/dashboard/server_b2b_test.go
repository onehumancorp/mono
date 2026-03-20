package dashboard

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleB2BHandshake(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"partnerOrg":"globex.com","partnerJwksUrl":"https://ohc.globex.com/.well-known/jwks.json","allowedRoles":["SALES_AGENT"]}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodGet,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingFields",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			req := httptest.NewRequest(tt.method, "/api/b2b/handshake", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleB2BHandshake(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if len(app.trustAgreements) != 1 {
					t.Fatalf("expected 1 trust agreement, got %d", len(app.trustAgreements))
				}
				if app.trustAgreements[0].PartnerOrg != "globex.com" {
					t.Errorf("expected globex.com, got %s", app.trustAgreements[0].PartnerOrg)
				}
			}
		})
	}
}

func TestHandleB2BRevoke(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"agreementId":"ta-globex-com-1234"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodGet,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingAgreementId",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			method:       http.MethodPost,
			payload:      `{"agreementId":"missing"}`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			app.trustAgreements = append(app.trustAgreements, TrustAgreement{
				ID:     "ta-globex-com-1234",
				Status: TrustStatusActive,
			})

			req := httptest.NewRequest(tt.method, "/api/b2b/revoke", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleB2BRevoke(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if app.trustAgreements[0].Status != TrustStatusRevoked {
					t.Errorf("expected status revoked, got %s", app.trustAgreements[0].Status)
				}
			}
		})
	}
}

func TestHandleIncidents_Combined(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"severity":"P0","summary":"High error rate","rootCauseAnalysis":"investigating"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodDelete,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingFields",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			req := httptest.NewRequest(tt.method, "/api/incidents", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleIncidents(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if len(app.incidents) != 1 {
					t.Fatalf("expected 1 incident, got %d", len(app.incidents))
				}
				if app.incidents[0].Severity != SeverityP0 {
					t.Errorf("expected P0, got %s", app.incidents[0].Severity)
				}
			}
		})
	}
}

func TestHandleIncidentStatus(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"incidentId":"inc-1234","status":"PROPOSED","resolutionPlanId":"rollback","rootCauseAnalysis":"bad deploy"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodGet,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingFields",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			method:       http.MethodPost,
			payload:      `{"incidentId":"missing","status":"resolved"}`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			app.incidents = append(app.incidents, Incident{
				ID:     "inc-1234",
				Status: IncidentStatusInvestigating,
			})

			req := httptest.NewRequest(tt.method, "/api/incidents/status", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleIncidentStatus(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if app.incidents[0].Status != IncidentStatusProposed {
					t.Errorf("expected PROPOSED, got %s", app.incidents[0].Status)
				}
			}
		})
	}
}

func TestHandleComputeProfiles_Combined(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"roleId":"AUDIT_AGENT","minVramGb":40,"preferredGpuType":"h100","schedulingPriority":10}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodDelete,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingRoleId",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			req := httptest.NewRequest(tt.method, "/api/compute/profiles", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleComputeProfiles(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if len(app.computeProfiles) != 1 {
					t.Fatalf("expected 1 profile, got %d", len(app.computeProfiles))
				}
			}
		})
	}
}

func TestHandleBudgetAlerts_Combined(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"thresholdUsd":500,"notifyAtPct":0.8}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodDelete,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "InvalidThreshold",
			method:       http.MethodPost,
			payload:      `{"thresholdUsd":0}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			req := httptest.NewRequest(tt.method, "/api/billing/alerts", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handleBudgetAlerts(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if len(app.budgetAlerts) != 1 {
					t.Fatalf("expected 1 alert, got %d", len(app.budgetAlerts))
				}
			}
		})
	}
}

func TestHandlePipelines_Combined(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"name":"feat-analytics","branch":"feat/analytics","initiatedBy":"pm-1"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodDelete,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingName",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			req := httptest.NewRequest(tt.method, "/api/pipelines", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handlePipelines(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if len(app.pipelines) != 1 {
					t.Fatalf("expected 1 pipeline, got %d", len(app.pipelines))
				}
			}
		})
	}
}

func TestHandlePipelinePromote(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
		setupStatus  PipelineStatus
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"pipelineId":"pipe-123","approvedBy":"ceo"}`,
			expectedCode: http.StatusOK,
			setupStatus:  PipelineStatusStaging,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodGet,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
			setupStatus:  PipelineStatusStaging,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
			setupStatus:  PipelineStatusStaging,
		},
		{
			name:         "MissingPipelineId",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
			setupStatus:  PipelineStatusStaging,
		},
		{
			name:         "NotFound",
			method:       http.MethodPost,
			payload:      `{"pipelineId":"missing"}`,
			expectedCode: http.StatusNotFound,
			setupStatus:  PipelineStatusStaging,
		},
		{
			name:         "NotInStaging",
			method:       http.MethodPost,
			payload:      `{"pipelineId":"pipe-123","approvedBy":"ceo"}`,
			expectedCode: http.StatusBadRequest,
			setupStatus:  PipelineStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			app.pipelines = append(app.pipelines, Pipeline{
				ID:     "pipe-123",
				Status: tt.setupStatus,
			})

			req := httptest.NewRequest(tt.method, "/api/pipelines/promote", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handlePipelinePromote(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if app.pipelines[0].Status != PipelineStatusPromoted {
					t.Errorf("expected PROMOTED, got %s", app.pipelines[0].Status)
				}
			}
		})
	}
}

func TestHandlePipelineStatus(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		payload      string
		expectedCode int
	}{
		{
			name:         "Success",
			method:       http.MethodPost,
			payload:      `{"pipelineId":"pipe-123","status":"STAGING","stagingUrl":"https://staging.example.com"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodGet,
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			payload:      `{bad json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingFields",
			method:       http.MethodPost,
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			method:       http.MethodPost,
			payload:      `{"pipelineId":"missing","status":"STAGING"}`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestServer(t)
			app.pipelines = append(app.pipelines, Pipeline{
				ID:     "pipe-123",
				Status: PipelineStatusPending,
			})

			req := httptest.NewRequest(tt.method, "/api/pipelines/status", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.handlePipelineStatus(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d: %s", tt.expectedCode, rec.Code, rec.Body.String())
			}

			if tt.expectedCode == http.StatusOK {
				if app.pipelines[0].Status != PipelineStatusStaging {
					t.Errorf("expected STAGING, got %s", app.pipelines[0].Status)
				}
			}
		})
	}
}
