# Orchestration Module

## Identity
The `orchestration` module provides the asynchronous Pub/Sub Agent Interaction Protocol and Virtual Meeting Room collaboration spaces for the One Human Corp platform.

## Architecture
This module implements the `Hub`, acting as the primary thread-safe message broker and global agent registry. AI workers interact by instantiating `MeetingRoom` sessions, dispensing context-aware `Message` structs, and dynamically adjusting their `Status` enumerations (e.g., Idle, Active, In_Meeting). This event-driven design cleanly decouples disparate agent logic pipelines, avoiding synchronous blockages entirely.

## Quick Start
Initialize the global broker, enroll agents, and commence a collaborative meeting session:

```go
package main

import (
	"fmt"
	"time"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func main() {
	hub := orchestration.NewHub()

	// Register workers securely into the registry map
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: "org-alpha"})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-alpha"})

	// Establish a shared contextual room
	room := hub.OpenMeetingWithAgenda("meeting-sprint-plan", "Coordinate upcoming deliverables", []string{"pm-1", "swe-1"})

	// Dispatch an asynchronous task request into the room
	_ = hub.Publish(orchestration.Message{
		ID:         "msg-001",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       orchestration.EventTask,
		Content:    "Draft implementation spec for OAuth 2.0.",
		MeetingID:  room.ID,
		OccurredAt: time.Now().UTC(),
	})

	fmt.Printf("Session %s actively holds %d unread events\n", room.ID, len(hub.Meeting(room.ID).Transcript))
}
```

## Developer Workflow
The module strictly relies on the Bazel compilation pipeline.

- **Build**: `bazelisk build //srcs/orchestration`
- **Test**: `bazelisk test //srcs/orchestration/...`

## Configuration
Because state persistence is handled entirely via secure in-memory structs and checkpointers, no external environment variables or configuration files are strictly required.
