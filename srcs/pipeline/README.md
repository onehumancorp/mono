# Pipeline Module

## Identity
The `pipeline` module orchestrates the complete Software Development Life Cycle (SDLC), advancing autonomous feature branches systematically from specification approval through to staged deployment.

## Architecture
This module implements the `Orchestrator` struct, an autonomous event-listener bound to the core `orchestration.Hub`. It dictates phase transitions (Implementing, Testing, Staging Ready, Deployed, Rollback) by responding directly to asynchronous messaging events (such as `EventSpecApproved` or `EventPRCreated`). Concurrency safety is maintained through private, thread-safe memory maps protected by localized `sync.RWMutex` locks, allowing agents to reliably invoke mock CI/CD jobs.

## Quick Start
Bind the pipeline orchestrator to the central event hub and simulate a spec approval:

```go
package main

import (
	"fmt"
	"time"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/pipeline"
)

func main() {
	hub := orchestration.NewHub()

	// Provision required baseline agents
	hub.RegisterAgent(orchestration.Agent{ID: "system-hub"})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1"})

	// Boot the pipeline listener
	orch := pipeline.NewOrchestrator(hub)

	// Disperse a specification sign-off event
	msg := orchestration.Message{
		ID:         "msg-event-1",
		FromAgent:  "pm-1",
		ToAgent:    "system-hub",
		Type:       orchestration.EventSpecApproved,
		Content:    "branch=feat-oauth,details=Implement federated login",
		OccurredAt: time.Now().UTC(),
	}

	if err := orch.HandleSpecApproved(msg); err != nil {
		fmt.Printf("Orchestrator fault: %v\n", err)
	}

	// Validate pipeline advancement
	state, _ := orch.GetPipelineState("feat-oauth")
	fmt.Printf("Initial pipeline progression state: %s\n", state)
}
```

## Developer Workflow
Compilation and unit execution depend purely on Bazel rules.

- **Build**: `bazelisk build //srcs/pipeline`
- **Test**: `bazelisk test //srcs/pipeline/...`

## Configuration
Zero runtime variables are strictly demanded. The package depends solely on the successful initialization of an accompanying `orchestration.Hub`.