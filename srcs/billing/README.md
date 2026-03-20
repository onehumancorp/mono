# Billing Module

## Identity
The `billing` module offers a high-performance Cost Estimation & Usage Tracker precisely geared toward granular LLM token accounting, enabling precise financial transparency of AI workforce utilization across the entire platform.

## Architecture
Leveraging a sophisticated, multi-sharded lock design (bypassing global mutex contention), the `Tracker` safely ingests concurrent asynchronous AI events. It rapidly derives real-time USD estimations using a dynamically injected `DefaultCatalog` of API pricing strategies (ranging from Anthropic Claude to OpenAI GPT implementations). Ultimately, costs are hierarchically aggregated down to the explicit `AgentID` and `OrganizationID`, surfacing actionable real-time burn-rate forecasts.

## Quick Start
Initialize a threaded pricing tracking engine:

```go
package main

import (
	"fmt"
	"time"
	"github.com/onehumancorp/mono/srcs/billing"
)

func main() {
	tracker := billing.NewTracker(billing.DefaultCatalog)

	// Simulate an AI workflow consumption event
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "agent-swe-1",
		OrganizationID:   "org-demo-1",
		Model:            "gpt-4o",
		PromptTokens:     1500,
		CompletionTokens: 850,
		OccurredAt:       time.Now().UTC(),
	})

	// Retrieve the aggregated organizational dashboard snapshot
	summary := tracker.Summary("org-demo-1")
	fmt.Printf("Estimated Org Spend: $%v\n", summary.TotalCostUSD)
}
```

## Developer Workflow
This module requires the Bazel build ecosystem.

- **Build**: `bazelisk build //srcs/billing`
- **Test**: `bazelisk test //srcs/billing/...`

*Note: Ensure performance optimizations within this package retain the signature `BOLT` comment format to justify architectural improvements.*

## Configuration
No explicit environment variables are mandated. Catalog registries are passed via initialization bounds, intentionally severing dependency loops to allow dynamic, external updates to foundational model rate adjustments.