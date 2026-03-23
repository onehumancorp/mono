# Pipeline Module

## Identity
The `pipeline` module is responsible for modelling and managing the Software Development Life Cycle (SDLC) progression for feature branches within the One Human Corp AI workforce.

## Architecture
The module implements an `Orchestrator` that oversees the automated pipeline phases: Implementing, Testing, Staging Ready, Deployed, and Rollback. It acts as an intermediary, consuming events from the `orchestration.Hub` (e.g., `EventSpecApproved`, `EventPRCreated`) and triggering corresponding CI/CD actions or assigning follow-up tasks to the AI agents (e.g., Software Engineers). State is maintained securely in memory using read-write mutexes.

## Quick Start
To initialize the orchestrator and process a pipeline event:

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

	// Register system and agents
	hub.RegisterAgent(orchestration.Agent{ID: "system-hub"})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1"})

	orch := pipeline.NewOrchestrator(hub)

	// Simulate an approved specification
	msg := orchestration.Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "system-hub",
		Type:       orchestration.EventSpecApproved,
		Content:    "branch=feat-login,details=Implement OAuth2",
		OccurredAt: time.Now().UTC(),
	}

	err := orch.HandleSpecApproved(msg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	state, _ := orch.GetPipelineState("feat-login")
	fmt.Printf("Pipeline state: %s\n", state)
}
```

## Developer Workflow
This module is built and tested using the Bazel build system.

- **Build**: `bazelisk build //srcs/pipeline`
- **Test**: `bazelisk test //srcs/pipeline/...`

## Configuration
No environment configuration is strictly required to run the mocked in-memory simulation. The orchestrator is tightly coupled to the `orchestration.Hub` and relies on correct agent registration within the Hub to successfully dispatch tasks.
