package telemetry

import (
	"context"
	"log"
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

	tokenUsageCounter       metric.Int64Counter
	agentApiCallsCounter    metric.Int64Counter
	humanInteractionsCounter metric.Int64Counter
	meetingEventsCounter    metric.Int64Counter
)

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
			log.Printf("[telemetry] INFO(2) recorded request: %s %s %.3fs", r.Method, r.URL.Path, duration)
		}
	})
}

var Verbosity = 1 // Default level

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

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

func RecordHumanInteraction(ctx context.Context, interactionType string) {
	if humanInteractionsCounter == nil {
		return
	}
	humanInteractionsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", interactionType),
	))
}

func RecordMeetingEvent(ctx context.Context, eventType string) {
	if meetingEventsCounter == nil {
		return
	}
	meetingEventsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", eventType),
	))
}
