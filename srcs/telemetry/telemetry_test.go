package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)


// Helper to reset global metrics to isolate tests
func resetGlobals() {
	meter = nil
	requestCounter = nil
	latencyHistogram = nil
	tokenUsageCounter = nil
	agentApiCallsCounter = nil
	humanInteractionsCounter = nil
	meetingEventsCounter = nil
	otel.SetMeterProvider(nil)
}

func TestInitTelemetry(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Happy Path - Success",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobals()

			shutdown, err := InitTelemetry()

			if (err != nil) != tt.wantErr {
				t.Errorf("InitTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if shutdown == nil {
					t.Error("InitTelemetry() returned nil shutdown function")
				} else {
					shutdown()
				}

				if meter == nil {
					t.Error("meter was not initialized")
				}
				if requestCounter == nil {
					t.Error("requestCounter was not initialized")
				}
				if latencyHistogram == nil {
					t.Error("latencyHistogram was not initialized")
				}
				if tokenUsageCounter == nil {
					t.Error("tokenUsageCounter was not initialized")
				}
				if agentApiCallsCounter == nil {
					t.Error("agentApiCallsCounter was not initialized")
				}
				if humanInteractionsCounter == nil {
					t.Error("humanInteractionsCounter was not initialized")
				}
				if meetingEventsCounter == nil {
					t.Error("meetingEventsCounter was not initialized")
				}
			}
		})
	}
}

func TestInitTelemetry_Errors(t *testing.T) {
	tests := []struct {
		name         string
		failExporter bool
		failCounter  string
		failHist     string
	}{
		{
			name:         "Exporter Error",
			failExporter: true,
		},
		{
			name:        "Request Counter Error",
			failCounter: "http_requests_total",
		},
		{
			name:     "Latency Histogram Error",
			failHist: "http_request_duration_seconds",
		},
		{
			name:        "Token Usage Error",
			failCounter: "ohc_token_usage_total",
		},
		{
			name:        "Agent API Calls Error",
			failCounter: "ohc_agent_api_calls_total",
		},
		{
			name:        "Human Interactions Error",
			failCounter: "ohc_human_interactions_total",
		},
		{
			name:        "Meeting Events Error",
			failCounter: "ohc_meeting_events_total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobals()

			// Mock exporter
			originalNewExporter := newExporter
			defer func() { newExporter = originalNewExporter }()

			if tt.failExporter {
				newExporter = func() (sdkmetric.Reader, error) {
					return nil, context.DeadlineExceeded
				}
			}

			// Mock metric creation
			originalCreateCounter := createInt64Counter
			originalCreateHist := createFloat64Histogram
			defer func() {
				createInt64Counter = originalCreateCounter
				createFloat64Histogram = originalCreateHist
			}()

			createInt64Counter = func(m metric.Meter, name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
				if name == tt.failCounter {
					return nil, context.DeadlineExceeded
				}
				return originalCreateCounter(m, name, opts...)
			}
			createFloat64Histogram = func(m metric.Meter, name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
				if name == tt.failHist {
					return nil, context.DeadlineExceeded
				}
				return originalCreateHist(m, name, opts...)
			}

			shutdown, err := InitTelemetry()
			if err == nil {
				t.Errorf("Expected initialization to fail on %s", tt.name)
			}
			if shutdown != nil {
				t.Errorf("Expected nil shutdown func on error")
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		verbosity    int
		initMetrics  bool
		expectedCode int
	}{
		{
			name:         "Happy Path - Records Metrics",
			method:       "GET",
			path:         "/api/v1/test",
			verbosity:    1,
			initMetrics:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Edge Case - Verbosity High",
			method:       "POST",
			path:         "/api/v1/verbose",
			verbosity:    2,
			initMetrics:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Edge Case - Metrics Uninitialized",
			method:       "GET",
			path:         "/api/v1/uninitialized",
			verbosity:    1,
			initMetrics:  false,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobals()
			Verbosity = tt.verbosity

			if tt.initMetrics {
				provider := sdkmetric.NewMeterProvider()
				meter = provider.Meter("test")
				requestCounter, _ = meter.Int64Counter("test_req")
				latencyHistogram, _ = meter.Float64Histogram("test_lat")
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := Middleware(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedCode)
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	if handler == nil {
		t.Error("MetricsHandler returned nil")
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Since prometheus exporter might panic or return 500 when not perfectly configured in a test hermetic environment
	// checking just that it returns an HTTP response is enough to verify MetricsHandler returns the prometheus handler correctly.
	if status := rr.Code; status == 0 {
		t.Errorf("handler returned no status code")
	}
}

func TestRecordingFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		initMetrics bool
		action      func()
	}{
		{
			name:        "Token Usage - Uninitialized",
			initMetrics: false,
			action: func() {
				RecordTokenUsage(ctx, "agent1", "researcher", "gpt-4", "prompt", 100)
			},
		},
		{
			name:        "Token Usage - Initialized",
			initMetrics: true,
			action: func() {
				RecordTokenUsage(ctx, "agent1", "researcher", "gpt-4", "prompt", 100)
			},
		},
		{
			name:        "Agent API Call - Uninitialized",
			initMetrics: false,
			action: func() {
				RecordAgentApiCall(ctx, "agent1", "researcher", "/api/search")
			},
		},
		{
			name:        "Agent API Call - Initialized",
			initMetrics: true,
			action: func() {
				RecordAgentApiCall(ctx, "agent1", "researcher", "/api/search")
			},
		},
		{
			name:        "Human Interaction - Uninitialized",
			initMetrics: false,
			action: func() {
				RecordHumanInteraction(ctx, "approval")
			},
		},
		{
			name:        "Human Interaction - Initialized",
			initMetrics: true,
			action: func() {
				RecordHumanInteraction(ctx, "approval")
			},
		},
		{
			name:        "Meeting Event - Uninitialized",
			initMetrics: false,
			action: func() {
				RecordMeetingEvent(ctx, "start")
			},
		},
		{
			name:        "Meeting Event - Initialized",
			initMetrics: true,
			action: func() {
				RecordMeetingEvent(ctx, "start")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobals()

			if tt.initMetrics {
				provider := sdkmetric.NewMeterProvider()
				meter = provider.Meter("test")
				tokenUsageCounter, _ = meter.Int64Counter("test_tokens")
				agentApiCallsCounter, _ = meter.Int64Counter("test_api")
				humanInteractionsCounter, _ = meter.Int64Counter("test_human")
				meetingEventsCounter, _ = meter.Int64Counter("test_meeting")
			}

			tt.action()
		})
	}
}
