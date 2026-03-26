# Orchestration Module

## Identity
The `orchestration` module provides the asynchronous Pub/Sub Agent Interaction Protocol and Virtual Meeting Room infrastructure for the One Human Corp platform.

## Architecture
This module implements the `Hub`, a thread-safe message broker and agent registry. Agents collaborate by opening `MeetingRoom` sessions, passing context via `Message` events, and shifting their `Status` (Idle, Active, In_Meeting, Blocked) dynamically. This entirely decouples agent logic from direct synchronous calls, matching the platform's distributed event-driven design.

## Quick Start
Initialise the hub, register agents, and create a meeting room:

```go
package main

import (
	"fmt"
	"time"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func main() {
	hub := orchestration.NewHub()

	// Register agents
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	// Open a meeting
	room := hub.OpenMeetingWithAgenda("sprint-planning", "Plan the next sprint", []string{"pm-1", "swe-1"})

	// Publish a message
	_ = hub.Publish(orchestration.Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       orchestration.EventTask,
		Content:    "Please implement the new login page.",
		MeetingID:  room.ID,
		OccurredAt: time.Now().UTC(),
	})

	fmt.Printf("Meeting %s has %d messages\n", room.ID, len(hub.Meeting(room.ID).Transcript))
}
```

## Developer Workflow
This module is built and tested using Bazel.

- **Build**: `bazelisk build //srcs/orchestration`
- **Test**: `bazelisk test //srcs/orchestration/...`

## Configuration
No environment configuration is required. State is held entirely in memory in this implementation.
