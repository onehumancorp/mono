# Dashboard Module

## Identity
The `dashboard` module serves as the primary HTTP API and Gateway for the One Human Corp "Agentic OS", allowing the human CEO to direct teams, monitor ongoing operations, and orchestrate the broader AI workforce.

## Architecture
The Dashboard relies on `net/http` to serve a React frontend application alongside JSON-based REST APIs. It acts as the orchestration proxy, linking the `domain` (organizational structure), `orchestration` (agent communication pub/sub hub), and `billing` (cost tracker) components together. The server maintains unified operational state within a thread-safe registry (`sync.RWMutex`). Functionality includes snapshot and recovery, confidence gating (approvals), and B2B federated handshake processing.

## Quick Start
To initialize the HTTP server in an application context:

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
This module is built and tested using Bazel.

- **Build**: `bazelisk build //srcs/dashboard`
- **Test**: `bazelisk test //srcs/dashboard/...`

## Configuration
- `MONO_FRONTEND_DIST`: (Optional) Setting this environment variable directs the HTTP handler to serve statically built frontend files from the provided path instead of returning a generic fallback UI.