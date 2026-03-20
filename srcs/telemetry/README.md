# Telemetry Module

## Identity
The `telemetry` module aggregates deep observability metrics and structured instrumentation events across the One Human Corp platform, giving operators an auditable, real-time dashboard for performance and agent AI token usage.

## Architecture
Leveraging the industry-standard OpenTelemetry library bridged to a Prometheus pull-exporter, this package natively captures `float64` latency distributions and `int64` system counters. `Middleware` seamlessly envelops HTTP handlers to auto-inject dimension tagging for routes, methods, and status codes. Companion functions (e.g. `RecordAgentApiCall`) grant granular insight into autonomous AI workflows. All exports map securely to a unified, thread-safe meter provider.

## Quick Start
To attach telemetry reporting to your application service initialization:

```go
package main

import (
	"context"
	"net/http"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

func main() {
	// Initialize the central OpenTelemetry metrics provider
	shutdown, err := telemetry.InitTelemetry()
	if err != nil {
		panic(err)
	}
	defer shutdown()

	// Provision a pull-endpoint for standard Prometheus scrapers
	http.Handle("/metrics", telemetry.MetricsHandler())

	// Start your HTTP service
	http.ListenAndServe(":9090", nil)
}
```

To explicitly track agent interactions deep within business logic:

```go
telemetry.RecordAgentApiCall(context.Background(), "swe-1", "SOFTWARE_ENGINEER", "github.CreatePullRequest")
```

## Developer Workflow
This module mandates the Bazel build system.

- **Build**: `bazelisk build //srcs/telemetry`
- **Test**: `bazelisk test //srcs/telemetry/...`

*Note: All unit tests must be explicitly executed locally to confirm the OpenTelemetry global provider is thoroughly exercised.*

## Configuration
No external environment variables are strictly mandated. Metrics are served in standardized Prometheus formats at the configured `/metrics` URI. System logging verbosity may be scaled dynamically via `telemetry.Verbosity`.