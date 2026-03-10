package orchestration

import (
	"testing"
	"time"
)

func TestPublishRoutesMessagesAndMeetingTranscript(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})
	err := hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       "task",
		Content:    "Implement the feature",
		MeetingID:  "kickoff",
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("publish returned error: %v", err)
	}

	inbox := hub.Inbox("swe-1")
	if len(inbox) != 1 || inbox[0].Content != "Implement the feature" {
		t.Fatalf("unexpected inbox contents: %+v", inbox)
	}

	meeting, ok := hub.Meeting("kickoff")
	if !ok {
		t.Fatalf("expected kickoff meeting to exist")
	}
	if len(meeting.Transcript) != 1 {
		t.Fatalf("expected transcript length 1, got %d", len(meeting.Transcript))
	}

	agent, ok := hub.Agent("pm-1")
	if !ok || agent.Status != StatusInMeeting {
		t.Fatalf("expected sender to be in meeting, got %+v", agent)
	}
}
