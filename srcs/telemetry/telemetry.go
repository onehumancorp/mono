package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	meter            metric.Meter
	requestCounter   metric.Int64Counter
	latencyHistogram metric.Float64Histogram

	tokenUsageCounter        metric.Int64Counter
	agentApiCallsCounter     metric.Int64Counter
	humanInteractionsCounter metric.Int64Counter
	meetingEventsCounter     metric.Int64Counter
)

// Summary: InitTelemetry configures and starts the OpenTelemetry metrics provider with a Prometheus exporter.
// Parameters: None
// Returns: (func(), error)
// Errors: Returns an error if applicable
// Side Effects: None
func InitTelemetry() (func(), error) {
	exporter, err := otelprom.New(otelprom.WithRegisterer(prometheus.DefaultRegisterer))
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	meter = provider.Meter("github.com/onehumancorp/mono/ohc")

	requestCounter, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	latencyHistogram, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request latency in seconds"),
	)
	if err != nil {
		return nil, err
	}

	tokenUsageCounter, err = meter.Int64Counter(
		"ohc_token_usage_total",
		metric.WithDescription("Total tokens used by agents"),
	)
	if err != nil {
		return nil, err
	}

	agentApiCallsCounter, err = meter.Int64Counter(
		"ohc_agent_api_calls_total",
		metric.WithDescription("Total API calls made by or for agents"),
	)
	if err != nil {
		return nil, err
	}

	humanInteractionsCounter, err = meter.Int64Counter(
		"ohc_human_interactions_total",
		metric.WithDescription("Total human-agent interactions"),
	)
	if err != nil {
		return nil, err
	}

	meetingEventsCounter, err = meter.Int64Counter(
		"ohc_meeting_events_total",
		metric.WithDescription("Total meeting room events"),
	)
	if err != nil {
		return nil, err
	}

	return func() {
		_ = provider.Shutdown(context.Background())
	}, nil
}

// Summary: Middleware injects telemetry instrumentation into an HTTP handler chain.    - next: http.Handler; The next HTTP handler in the request pipeline.
// Parameters: next
// Returns: http.Handler
// Errors: None
// Side Effects: None
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()

		if requestCounter != nil && latencyHistogram != nil {
			attributes := metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
			)
			requestCounter.Add(r.Context(), 1, attributes)
			latencyHistogram.Record(r.Context(), duration, attributes)
		}
		if Verbosity >= 2 {
			slog.Info("recorded request", "component", "telemetry", "method", r.Method, "path", r.URL.Path, "duration", duration)
		}
	})
}

// Summary: Defines the Verbosity type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
var Verbosity = 1 // Default level

// Summary: MetricsHandler provides an HTTP handler that exposes the collected Prometheus metrics.
// Parameters: None
// Returns: http.Handler
// Errors: None
// Side Effects: None
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// Summary: RecordTokenUsage increments the global counter for LLM tokens consumed by the workforce.    - ctx: context.Context; The context of the active trace or request.   - agentID: string; The identifier of the agent consuming the tokens.   - role: string; The role of the agent.   - model: string; The specific AI model being inferred (e.g., gpt-4o).   - tokenType: string; The type of tokens (e.g., prompt or completion).   - count: int64; The number of tokens consumed.
// Parameters: ctx, agentID, role, model, tokenType, count
// Returns: None
// Errors: None
// Side Effects: None
func RecordTokenUsage(ctx context.Context, agentID, role, model, tokenType string, count int64) {
	if tokenUsageCounter == nil {
		return
	}
	tokenUsageCounter.Add(ctx, count, metric.WithAttributes(
		attribute.String("agent_id", agentID),
		attribute.String("role", role),
		attribute.String("model", model),
		attribute.String("type", tokenType),
	))
}

// Summary: RecordAgentApiCall increments the global counter for external tool or API invocations made by agents.    - ctx: context.Context; The context of the active trace or request.   - agentID: string; The identifier of the agent making the call.   - role: string; The role of the agent.   - api: string; The name or route of the invoked API/tool.
// Parameters: ctx, agentID, role, api
// Returns: None
// Errors: None
// Side Effects: None
func RecordAgentApiCall(ctx context.Context, agentID, role, api string) {
	if agentApiCallsCounter == nil {
		return
	}
	agentApiCallsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("agent_id", agentID),
		attribute.String("role", role),
		attribute.String("api", api),
	))
}

// Summary: RecordHumanInteraction increments the global counter for events involving direct human oversight.    - ctx: context.Context; The context of the active trace or request.   - interactionType: string; The category of interaction (e.g., approval, handoff).
// Parameters: ctx, interactionType
// Returns: None
// Errors: None
// Side Effects: None
func RecordHumanInteraction(ctx context.Context, interactionType string) {
	if humanInteractionsCounter == nil {
		return
	}
	humanInteractionsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", interactionType),
	))
}

// Summary: RecordMeetingEvent increments the global counter for collaborative meeting room actions.    - ctx: context.Context; The context of the active trace or request.   - eventType: string; The nature of the meeting event (e.g., start, message, end).
// Parameters: ctx, eventType
// Returns: None
// Errors: None
// Side Effects: None
func RecordMeetingEvent(ctx context.Context, eventType string) {
	if meetingEventsCounter == nil {
		return
	}
	meetingEventsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", eventType),
	))
}

// Summary: LogAgentExecution provides structured JSON logging for agent execution traces.    - ctx: context.Context; The context of the active trace or request.   - agentID: string; The identifier of the agent.   - role: string; The role of the agent.   - api: string; The API or tool being executed.   - eventType: string; The specific type of the event (e.g. task, status).   - content: string; The content or message payload associated with the execution.
// Parameters: ctx, agentID, role, api, eventType, content
// Returns: None
// Errors: None
// Side Effects: None
func LogAgentExecution(ctx context.Context, agentID, role, api, eventType, content string) {
	slog.InfoContext(ctx, "agent execution trace",
		"component", "telemetry",
		"agent_id", agentID,
		"role", role,
		"api", api,
		"event_type", eventType,
		"content", content,
	)
}
