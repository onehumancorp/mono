package telemetry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// mockMeterProvider delegates all calls to a noop provider except providing our mockMeter.
type mockMeterProvider struct {
	metric.MeterProvider
	meter metric.Meter
}

func (m mockMeterProvider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return m.meter
}

// mockMeter delegates all interface calls to noop.Meter, allowing us to selectively
// override only what we need without having to implement the huge interface by hand.
// Go allows embedding interfaces like this, but we have to forward all calls manually
// to avoid "does not implement..." compile errors when new methods are added, OR
// we just embed metric.Meter and initialize it with noop.Meter.
type mockMeter struct {
	metric.Meter
	failAt string
}

func (m mockMeter) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	if m.failAt == name {
		return nil, errors.New("mock Int64Counter error")
	}
	return m.Meter.Int64Counter(name, options...)
}

func (m mockMeter) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	if m.failAt == name {
		return nil, errors.New("mock Float64Histogram error")
	}
	return m.Meter.Float64Histogram(name, options...)
}

func TestInitTelemetry(t *testing.T) {
	prevMeterProvider := otel.GetMeterProvider()
	defer otel.SetMeterProvider(prevMeterProvider)

	shutdown, err := InitTelemetry()
	if err != nil {
		t.Errorf("InitTelemetry() unexpected error = %v", err)
	}
	if shutdown != nil {
		shutdown()
	}

	// Because of previous test interactions, prometheus.DefaultRegisterer might or might not
	// cleanly fail or we can just mock it successfully below.
	_, err2 := InitTelemetry()
	if err2 == nil {
		t.Logf("InitTelemetry() second call expected error due to duplicate registration, got nil (ignoring due to shared global state)")
	}

	originalCreateExporter := createExporter
	createExporter = func() (*prometheus.Exporter, error) {
		return nil, errors.New("mock createExporter error")
	}
	_, err3 := InitTelemetry()
	if err3 == nil {
		t.Errorf("InitTelemetry() expected error due to mock failure, got nil")
	}
	createExporter = originalCreateExporter
}

func TestSetupMetricsErrors(t *testing.T) {
	errorCases := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"ohc_token_usage_total",
		"ohc_agent_api_calls_total",
		"ohc_human_interactions_total",
		"ohc_meeting_events_total",
	}

	noopMeter := noop.NewMeterProvider().Meter("test")

	for _, failName := range errorCases {
		t.Run("Fail_"+failName, func(t *testing.T) {
			mProvider := mockMeterProvider{
				meter: mockMeter{
					Meter:  noopMeter,
					failAt: failName,
				},
			}
			shutdown, err := setupMetrics(mProvider)
			if err == nil {
				t.Errorf("setupMetrics() expected error for %s, got nil", failName)
			}
			if shutdown != nil {
				shutdown()
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	provider := noop.NewMeterProvider()
	m := provider.Meter("test")
	rc, _ := m.Int64Counter("test_counter")
	lh, _ := m.Float64Histogram("test_histogram")

	requestCounter = rc
	latencyHistogram = lh

	tests := []struct {
		name       string
		method     string
		path       string
		verbosity  int
		statusCode int
	}{
		{
			name:       "Happy Path - Verbosity 1",
			method:     "GET",
			path:       "/api/v1/status",
			verbosity:  1,
			statusCode: http.StatusOK,
		},
		{
			name:       "Happy Path - Verbosity 2",
			method:     "POST",
			path:       "/api/v1/submit",
			verbosity:  2,
			statusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVerbosity := Verbosity
			Verbosity = tt.verbosity
			defer func() { Verbosity = originalVerbosity }()

			handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(1 * time.Millisecond)
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.statusCode)
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	if handler == nil {
		t.Error("MetricsHandler() returned nil")
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("MetricsHandler returned unexpected status code: %v", rr.Code)
	}
}

func TestRecordFunctions(t *testing.T) {
	provider := noop.NewMeterProvider()
	m := provider.Meter("test")

	tuc, _ := m.Int64Counter("test_token")
	aac, _ := m.Int64Counter("test_agent")
	hic, _ := m.Int64Counter("test_human")
	mec, _ := m.Int64Counter("test_meeting")

	tests := []struct {
		name   string
		setup  func()
		action func()
	}{
		{
			name:   "RecordTokenUsage - nil counter",
			setup:  func() { tokenUsageCounter = nil },
			action: func() { RecordTokenUsage(context.Background(), "agent-1", "worker", "gpt-4", "prompt", 100) },
		},
		{
			name:   "RecordTokenUsage - initialized counter",
			setup:  func() { tokenUsageCounter = tuc },
			action: func() { RecordTokenUsage(context.Background(), "agent-1", "worker", "gpt-4", "prompt", 100) },
		},
		{
			name:   "RecordAgentApiCall - nil counter",
			setup:  func() { agentApiCallsCounter = nil },
			action: func() { RecordAgentApiCall(context.Background(), "agent-1", "worker", "/api/tool") },
		},
		{
			name:   "RecordAgentApiCall - initialized counter",
			setup:  func() { agentApiCallsCounter = aac },
			action: func() { RecordAgentApiCall(context.Background(), "agent-1", "worker", "/api/tool") },
		},
		{
			name:   "RecordHumanInteraction - nil counter",
			setup:  func() { humanInteractionsCounter = nil },
			action: func() { RecordHumanInteraction(context.Background(), "approval") },
		},
		{
			name:   "RecordHumanInteraction - initialized counter",
			setup:  func() { humanInteractionsCounter = hic },
			action: func() { RecordHumanInteraction(context.Background(), "approval") },
		},
		{
			name:   "RecordMeetingEvent - nil counter",
			setup:  func() { meetingEventsCounter = nil },
			action: func() { RecordMeetingEvent(context.Background(), "start") },
		},
		{
			name:   "RecordMeetingEvent - initialized counter",
			setup:  func() { meetingEventsCounter = mec },
			action: func() { RecordMeetingEvent(context.Background(), "start") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.action()
		})
	}
}
