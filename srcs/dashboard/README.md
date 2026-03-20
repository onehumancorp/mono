# Dashboard Module

## Identity
The `dashboard` module serves as the primary HTTP API and REST Gateway for the One Human Corp "Agentic OS". It enables the human CEO to securely direct teams, monitor real-time operations, and oversee the AI workforce via the React frontend.

## Architecture
The Dashboard relies on the standard `net/http` library to serve a unified REST API and proxy static Next.js/React frontend assets. It functions as the central orchestration hub, integrating the `domain` (organizational hierarchy), `orchestration` (Pub/Sub messaging), and `billing` (Cost Estimation Engine) modules. Operational state is safely maintained in memory utilizing a granular sharded lock design (`sync.RWMutex`) to prevent global mutex contention. Core capabilities include data snapshots, confidence gating (HITL approvals), and B2B federated handshake routing.

## Quick Start
To initialize the HTTP server and bind it to an application context:

```go
package main

import (
    "net/http"
    "time"

    "github.com/onehumancorp/mono/srcs/billing"
    "github.com/onehumancorp/mono/srcs/dashboard"
    "github.com/onehumancorp/mono/srcs/domain"
    "github.com/onehumancorp/mono/srcs/orchestration"
)

func main() {
    now := time.Now().UTC()
    org := domain.NewSoftwareCompany("org-1", "My Org", "Alice", now)
    hub := orchestration.NewHub()
    tracker := billing.NewTracker(billing.DefaultCatalog)

    // Wire the Dashboard Handler
    server := dashboard.NewServer(org, hub, tracker)
    http.ListenAndServe(":8080", server)
}
```

## Developer Workflow
This module rigidly mandates Bazel for deterministic builds and tests.

- **Build**: `bazelisk build //srcs/dashboard`
- **Test**: `bazelisk test //srcs/dashboard/...`

*Note: Ensure coverage strictly exceeds 95%. Use `go test -coverprofile=coverage.out ./srcs/dashboard/... && go tool cover -func=coverage.out` to verify line-by-line coverage if needed.*

## Configuration
- `MONO_FRONTEND_DIST`: *(Optional)* Specifying this environment variable directs the HTTP handler to serve statically compiled Next.js/React frontend files from the provided directory, suppressing the generic fallback UI.