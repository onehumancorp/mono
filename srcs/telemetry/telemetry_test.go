package telemetry

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
)

type mockRegisterer struct {
	prometheus.Registerer
}

func (m *mockRegisterer) Register(prometheus.Collector) error {
	return errors.New("mock register error")
}

func TestInitTelemetry(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Happy path: initialization succeeds
	cleanup, err := InitTelemetry()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cleanup == nil {
		t.Fatal("expected cleanup function, got nil")
	}

	// Verify that the globals are set
	if requestCounter == nil {
		t.Error("expected requestCounter to be initialized")
	}
	if latencyHistogram == nil {
		t.Error("expected latencyHistogram to be initialized")
	}
	if tokenUsageCounter == nil {
		t.Error("expected tokenUsageCounter to be initialized")
	}
	if agentApiCallsCounter == nil {
		t.Error("expected agentApiCallsCounter to be initialized")
	}
	if humanInteractionsCounter == nil {
		t.Error("expected humanInteractionsCounter to be initialized")
	}
	if meetingEventsCounter == nil {
		t.Error("expected meetingEventsCounter to be initialized")
	}

	cleanup() // Clean up resources
}

func TestInitTelemetryError(t *testing.T) {
	originalRegisterer := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = &mockRegisterer{}
	defer func() { prometheus.DefaultRegisterer = originalRegisterer }()

	cleanup, err := InitTelemetry()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if cleanup != nil {
		t.Fatal("expected nil cleanup, got non-nil")
	}
}

type errMeter struct {
	metric.Meter
	failAt string
}

func (m *errMeter) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	if m.failAt == name {
		return nil, errors.New("mock meter error for " + name)
	}
	return m.Meter.Int64Counter(name, options...)
}

func (m *errMeter) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	if m.failAt == name {
		return nil, errors.New("mock meter error for " + name)
	}
	return m.Meter.Float64Histogram(name, options...)
}

func TestInitTelemetryMeterErrors(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	cleanup, _ := InitTelemetry()
	if cleanup != nil {
		cleanup()
	}
	originalMeter := meter
	defer func() { meter = originalMeter }()

	tests := []struct {
		failAt string
	}{
		{"http_requests_total"},
		{"http_request_duration_seconds"},
		{"ohc_token_usage_total"},
		{"ohc_agent_api_calls_total"},
		{"ohc_human_interactions_total"},
		{"ohc_meeting_events_total"},
	}

	for _, tt := range tests {
		t.Run("Fail "+tt.failAt, func(t *testing.T) {
			meter = &errMeter{Meter: originalMeter, failAt: tt.failAt}
			prometheus.DefaultRegisterer = prometheus.NewRegistry()
			c, err := InitTelemetry()
			if err == nil {
				t.Errorf("expected error when failing %s, got nil", tt.failAt)
			}
			if c != nil {
				c()
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		verbosity int
	}{
		{
			name:      "default verbosity",
			verbosity: 1,
		},
		{
			name:      "high verbosity",
			verbosity: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVerbosity := Verbosity
			Verbosity = tt.verbosity
			defer func() { Verbosity = originalVerbosity }()

			prometheus.DefaultRegisterer = prometheus.NewRegistry()

			cleanup, err := InitTelemetry()
			if err != nil {
				t.Fatalf("failed to init telemetry: %v", err)
			}
			defer cleanup()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(5 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			})

			handlerToTest := Middleware(nextHandler)

			req := httptest.NewRequest("GET", "/test/path", nil)
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	if handler == nil {
		t.Fatal("expected handler, got nil")
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestRecordFunctions(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	cleanup, err := InitTelemetry()
	if err != nil {
		t.Fatalf("failed to init telemetry: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	t.Run("RecordTokenUsage", func(t *testing.T) {
		RecordTokenUsage(ctx, "agent-1", "developer", "gpt-4", "prompt", 100)
	})

	t.Run("RecordAgentApiCall", func(t *testing.T) {
		RecordAgentApiCall(ctx, "agent-1", "developer", "get_file")
	})

	t.Run("RecordHumanInteraction", func(t *testing.T) {
		RecordHumanInteraction(ctx, "approval")
	})

	t.Run("RecordMeetingEvent", func(t *testing.T) {
		RecordMeetingEvent(ctx, "start")
	})
}

func TestRecordFunctionsUninitialized(t *testing.T) {
	originalTokenUsageCounter := tokenUsageCounter
	originalAgentApiCallsCounter := agentApiCallsCounter
	originalHumanInteractionsCounter := humanInteractionsCounter
	originalMeetingEventsCounter := meetingEventsCounter

	tokenUsageCounter = nil
	agentApiCallsCounter = nil
	humanInteractionsCounter = nil
	meetingEventsCounter = nil

	defer func() {
		tokenUsageCounter = originalTokenUsageCounter
		agentApiCallsCounter = originalAgentApiCallsCounter
		humanInteractionsCounter = originalHumanInteractionsCounter
		meetingEventsCounter = originalMeetingEventsCounter
	}()

	ctx := context.Background()

	t.Run("RecordTokenUsage Uninitialized", func(t *testing.T) {
		RecordTokenUsage(ctx, "agent-1", "developer", "gpt-4", "prompt", 100)
	})

	t.Run("RecordAgentApiCall Uninitialized", func(t *testing.T) {
		RecordAgentApiCall(ctx, "agent-1", "developer", "get_file")
	})

	t.Run("RecordHumanInteraction Uninitialized", func(t *testing.T) {
		RecordHumanInteraction(ctx, "approval")
	})

	t.Run("RecordMeetingEvent Uninitialized", func(t *testing.T) {
		RecordMeetingEvent(ctx, "start")
	})
}

func TestLogAgentExecution(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(originalLogger)

	ctx := context.Background()
	LogAgentExecution(ctx, "agent-123", "analyst", "fetch_data", "task", "processed 10 records")

	output := buf.String()
	if !strings.Contains(output, "agent execution trace") {
		t.Errorf("expected log to contain 'agent execution trace', got %q", output)
	}
	if !strings.Contains(output, "agent_id=agent-123") {
		t.Errorf("expected log to contain 'agent_id=agent-123', got %q", output)
	}
	if !strings.Contains(output, "role=analyst") {
		t.Errorf("expected log to contain 'role=analyst', got %q", output)
	}
	if !strings.Contains(output, "api=fetch_data") {
		t.Errorf("expected log to contain 'api=fetch_data', got %q", output)
	}
	if !strings.Contains(output, "event_type=task") {
		t.Errorf("expected log to contain 'event_type=task', got %q", output)
	}
	if !strings.Contains(output, "content=\"processed 10 records\"") {
		t.Errorf("expected log to contain 'content=\"processed 10 records\"', got %q", output)
	}
}
