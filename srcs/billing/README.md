# Billing Module

## Identity
The `billing` module provides a comprehensive Cost Estimation & Billing Engine for tracking Large Language Model (LLM) token usage and model-aware pricing across the entire One Human Corp AI workforce.

## Architecture
The module implements a thread-safe `Tracker` using `sync.RWMutex`, persisting token usage events in-memory. It computes USD costs dynamically by matching each LLM inference event against a `DefaultCatalog` of API pricing models (such as GPT-4o or Claude 3.5 Sonnet). Usage is aggregated hierarchically by `OrganizationID` and `AgentID` to support real-time token burn-rate forecasting and cost transparency.

## Quick Start
To instantiate a new billing tracker in your code, supply a pricing catalog:

```go
package main

import (
	"github.com/onehumancorp/mono/srcs/billing"
	"time"
)

func main() {
	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "agent-swe-1",
		OrganizationID:   "org-demo-1",
		Model:            "gpt-4o",
		PromptTokens:     1000,
		CompletionTokens: 500,
		OccurredAt:       time.Now().UTC(),
	})

	summary := tracker.Summary("org-demo-1")
	println(summary.TotalCostUSD)
}
```

## Developer Workflow
This module is built and tested using the Bazel build system.

- **Build**: `bazelisk build //srcs/billing`
- **Test**: `bazelisk test //srcs/billing/...`

## Configuration
No external environment variables are required. Pricing catalogs can be injected at initialization, enabling updates to model rates without source modifications. The module uses standard Go primitives and operates securely without accessing secrets.
