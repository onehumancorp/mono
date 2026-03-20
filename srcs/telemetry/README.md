# Telemetry Module

## Identity
The `telemetry` module provides observability and operational metrics for the One Human Corp platform, empowering the human CEO to monitor the real-time performance and resource consumption of the AI workforce.

## Architecture
The module integrates OpenTelemetry with Prometheus to capture, record, and expose application metrics. It offers a standardized set of instrumentations including HTTP request latency, total tokens used by AI agents, and specific event counters (e.g., meeting room interactions). The `Middleware` automatically instruments HTTP routes, while explicit package functions allow internal modules to record fine-grained operations.

## Quick Start
Initialize telemetry at the application start and register the Prometheus metrics endpoint:

```go
package main

import (
	"context"
	"net/http"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

func main() {
	// Initialize the OpenTelemetry provider
	shutdown, err := telemetry.InitTelemetry()
	if err != nil {
		panic(err)
	}
	defer shutdown()

	// Expose the /metrics endpoint for Prometheus scraping
	http.Handle("/metrics", telemetry.MetricsHandler())

	// Start your server
	http.ListenAndServe(":9090", nil)
}
```

To record custom agent activity within the application:

```go
telemetry.RecordAgentApiCall(context.Background(), "swe-1", "SOFTWARE_ENGINEER", "github.CreatePullRequest")
```

## Developer Workflow
This module is built and tested using the Bazel build system.

- **Build**: `bazelisk build //srcs/telemetry`
- **Test**: `bazelisk test //srcs/telemetry/...`

## Configuration
No mandatory environment variables are required. By default, telemetry data is exposed via standard Prometheus formats. The output verbosity can be adjusted programmatically using `telemetry.Verbosity`.