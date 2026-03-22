package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

type mockRegisterer struct{}

func (m mockRegisterer) Register(c prometheus.Collector) error {
	return prometheus.AlreadyRegisteredError{}
}
func (m mockRegisterer) MustRegister(c ...prometheus.Collector) {}
func (m mockRegisterer) Unregister(c prometheus.Collector) bool { return false }

func TestInitTelemetryError(t *testing.T) {
	// Attempting to register again on the same registry will cause the exporter to fail setup sometimes
	originalRegisterer := prometheus.DefaultRegisterer
	defer func() { prometheus.DefaultRegisterer = originalRegisterer }()

	prometheus.DefaultRegisterer = mockRegisterer{}

	cleanup, err := InitTelemetry()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cleanup != nil {
		t.Fatalf("expected nil cleanup, got non-nil")
	}
}

// Add dummy test to trigger otelprom error for better coverage

func TestLogAgentExecution(t *testing.T) {
	ctx := context.Background()

	// Capture log output
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, nil)

	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)
	slog.SetDefault(slog.New(handler))

	agentID := "agent-123"
	role := "developer"
	api := "run_in_bash_session"
	eventType := "task"
	content := "executing command ls"

	LogAgentExecution(ctx, agentID, role, api, eventType, content)

	logOutput := buf.String()

	if !strings.Contains(logOutput, "agent execution trace") {
		t.Errorf("expected log to contain message, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "component=telemetry") {
		t.Errorf("expected log to contain component, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "agent_id="+agentID) {
		t.Errorf("expected log to contain agent_id, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "role="+role) {
		t.Errorf("expected log to contain role, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "api="+api) {
		t.Errorf("expected log to contain api, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "event_type="+eventType) {
		t.Errorf("expected log to contain event_type, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "content=\"executing command ls\"") {
		t.Errorf("expected log to contain content, got %s", logOutput)
	}
}
