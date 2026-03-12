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

func TestNewHubStartsEmpty(t *testing.T) {
	hub := NewHub()

	if meetings := hub.Meetings(); len(meetings) != 0 {
		t.Fatalf("expected no meetings, got %d", len(meetings))
	}
	if inbox := hub.Inbox("missing"); len(inbox) != 0 {
		t.Fatalf("expected empty inbox, got %+v", inbox)
	}
}

func TestRegisterAgentDefaultsStatusAndLookupMiss(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "agent-1", Name: "Agent", Role: "SWE", OrganizationID: "org-1"})

	agent, ok := hub.Agent("agent-1")
	if !ok {
		t.Fatalf("expected registered agent lookup to succeed")
	}
	if agent.Status != StatusIdle {
		t.Fatalf("expected default idle status, got %s", agent.Status)
	}
	if _, ok := hub.Agent("missing"); ok {
		t.Fatalf("expected missing agent lookup to fail")
	}
}

func TestOpenMeetingMarksParticipantsInMeeting(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "b", Name: "B", Role: "SWE", OrganizationID: "org-1"})

	meeting := hub.OpenMeeting("m1", []string{"a", "b"})
	if len(meeting.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(meeting.Participants))
	}

	agent, _ := hub.Agent("a")
	if agent.Status != StatusInMeeting {
		t.Fatalf("expected participant to be in meeting, got %s", agent.Status)
	}
}

func TestPublishValidationErrors(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})

	if err := hub.Publish(Message{FromAgent: "missing"}); err == nil {
		t.Fatalf("expected sender validation error")
	}
	if err := hub.Publish(Message{FromAgent: "a", ToAgent: "missing"}); err == nil {
		t.Fatalf("expected recipient validation error")
	}
	if err := hub.Publish(Message{FromAgent: "a", MeetingID: "missing"}); err == nil {
		t.Fatalf("expected meeting validation error")
	}
}

func TestPublishWithoutMeetingMarksSenderActive(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})

	if err := hub.Publish(Message{
		ID:         "m1",
		FromAgent:  "a",
		Type:       "status",
		Content:    "done",
		OccurredAt: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("publish returned error: %v", err)
	}

	agent, _ := hub.Agent("a")
	if agent.Status != StatusActive {
		t.Fatalf("expected sender to become active, got %s", agent.Status)
	}
}

func TestMeetingLookupMiss(t *testing.T) {
	hub := NewHub()
	if _, ok := hub.Meeting("missing"); ok {
		t.Fatalf("expected missing meeting lookup to fail")
	}
}

func TestAgentsReturnsSortedSnapshot(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "b", Name: "B", Role: "SWE", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})

	agents := hub.Agents()
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
	if agents[0].ID != "a" || agents[1].ID != "b" {
		t.Fatalf("expected sorted agent IDs, got %+v", agents)
	}

	agents[0].Name = "mutated"
	original, _ := hub.Agent("a")
	if original.Name != "A" {
		t.Fatalf("expected agent snapshot mutation not to affect hub, got %+v", original)
	}
}
