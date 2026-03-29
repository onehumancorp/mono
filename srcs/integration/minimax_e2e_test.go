package integration

// minimax_e2e_test.go provides end-to-end and integration tests that exercise
// the Minimax AI backend through real API calls.  Tests are skipped automatically
// when MINIMAX_API_KEY is not set in the environment, keeping CI green without a
// live key while still enabling full verification when the key is available.
//
// Covered scenarios:
//   - Agent receives a task and uses Minimax to reason about it, then replies.
//   - Multiple agents collaborate inside a meeting room; each turn is driven by
//     a Minimax reasoning call so the conversation is fully autonomous.

import (
	"github.com/onehumancorp/mono/srcs/domain"

	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

// minimaxAPIKey returns the Minimax API key from the environment, or an empty
// string when the variable is unset.
func minimaxAPIKey() string {
	return os.Getenv("MINIMAX_API_KEY")
}

// TestMinimaxAgentTaskE2E verifies the full task-assignment flow when a real
// Minimax API key is available:
//  1. A Product Manager agent assigns a coding task to a Software Engineer agent
//     via the Hub's Publish mechanism.
//  2. The Software Engineer reads the task from its inbox, calls the Minimax
//     reasoning API to generate an implementation plan, and replies to the PM.
//  3. The Product Manager's inbox contains the non-empty reply, confirming that
//     the round-trip from task dispatch → Minimax reasoning → acknowledgment
//     works end-to-end.
func TestMinimaxAgentTaskE2E(t *testing.T) {
	key := minimaxAPIKey()
	if key == "" {
		t.Skip("MINIMAX_API_KEY not set; skipping live Minimax E2E test")
	}

	hub := orchestration.NewHub()
	hub.SetMinimaxAPIKey(key)

	hub.RegisterAgent(orchestration.Agent{
		ID:             "pm-e2e",
		Name:           "Product Manager",
		Role:           "PRODUCT_MANAGER",
		OrganizationID: "org-e2e",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "swe-e2e",
		Name:           "Software Engineer",
		Role:           "SOFTWARE_ENGINEER",
		OrganizationID: "org-e2e",
	})

	// PM assigns a concrete task to the SWE.
	taskContent := "Implement a simple HTTP health-check endpoint that returns 200 OK."
	if err := hub.Publish(domain.Message{
		ID:         "task-e2e-1",
		FromAgent:  "pm-e2e",
		ToAgent:    "swe-e2e",
		Type:       orchestration.EventTask,
		Content:    taskContent,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("PM→SWE task publish: %v", err)
	}

	// SWE reads the task from its inbox.
	sweInbox := hub.Inbox("swe-e2e")
	if len(sweInbox) == 0 {
		t.Fatal("SWE inbox is empty after task assignment")
	}
	receivedTask := sweInbox[0].Content
	if receivedTask != taskContent {
		t.Fatalf("SWE received unexpected task content: %q", receivedTask)
	}

	// SWE uses Minimax to reason about the task and produce an implementation plan.
	client := orchestration.NewMinimaxClient(key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(
		"You are a Software Engineer in a collaborative AI workforce. "+
			"Briefly describe (2-3 sentences) how you would implement the following task: %s",
		receivedTask,
	)
	reasoningResult, err := client.Reason(ctx, prompt)
	if err != nil {
		t.Fatalf("Minimax reasoning call failed: %v", err)
	}
	if strings.TrimSpace(reasoningResult) == "" {
		t.Fatal("expected a non-empty reasoning response from Minimax")
	}

	// SWE publishes the implementation plan back to the PM.
	if err := hub.Publish(domain.Message{
		ID:         "reply-e2e-1",
		FromAgent:  "swe-e2e",
		ToAgent:    "pm-e2e",
		Type:       orchestration.EventHandoff,
		Content:    reasoningResult,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SWE→PM reply publish: %v", err)
	}

	// Wait a moment for background goroutines to finish (e.g. telemetry, summarization)
	time.Sleep(100 * time.Millisecond)

	// PM inbox should now contain the SWE's reply.
	pmInbox := hub.Inbox("pm-e2e")
	if len(pmInbox) == 0 {
		t.Fatal("PM inbox is empty after SWE reply")
	}
	if strings.TrimSpace(pmInbox[0].Content) == "" {
		t.Fatal("PM inbox message has empty content")
	}
}

// TestMinimaxAgentMeetingRoomE2E verifies that multiple agents can hold a fully
// autonomous conversation inside a meeting room where every turn is powered by
// a real Minimax API call.  The test opens a sprint-planning meeting with three
// agents (PM, SWE, QA) and drives a short, structured discussion:
//
//  1. PM sets the meeting agenda via Minimax.
//  2. SWE estimates effort via Minimax, responding to the PM's message.
//  3. QA describes its testing strategy via Minimax, responding to the SWE.
//
// After the three-turn exchange the test asserts that the meeting transcript
// contains exactly three messages in the correct order.
func TestMinimaxAgentMeetingRoomE2E(t *testing.T) {
	key := minimaxAPIKey()
	if key == "" {
		t.Skip("MINIMAX_API_KEY not set; skipping live Minimax meeting-room E2E test")
	}

	hub := orchestration.NewHub()
	hub.SetMinimaxAPIKey(key)

	hub.RegisterAgent(orchestration.Agent{
		ID:             "pm-meet",
		Name:           "Alice (PM)",
		Role:           "PRODUCT_MANAGER",
		OrganizationID: "org-meet",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "swe-meet",
		Name:           "Bob (SWE)",
		Role:           "SOFTWARE_ENGINEER",
		OrganizationID: "org-meet",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "qa-meet",
		Name:           "Carol (QA)",
		Role:           "QA_TESTER",
		OrganizationID: "org-meet",
	})

	hub.OpenMeetingWithAgenda(
		"sprint-e2e",
		"Q3 Sprint Planning — align on goals and estimate tasks",
		[]string{"pm-meet", "swe-meet", "qa-meet"},
	)

	client := orchestration.NewMinimaxClient(key)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	type turn struct {
		fromAgent string
		role      string
		prompt    string
	}

	turns := []turn{
		{
			fromAgent: "pm-meet",
			role:      "Product Manager",
			prompt: "You are a Product Manager running a sprint planning meeting. " +
				"In 1-2 sentences, state the top goal for this sprint.",
		},
		{
			fromAgent: "swe-meet",
			role:      "Software Engineer",
			prompt: "You are a Software Engineer in a sprint planning meeting. " +
				"In 1-2 sentences, give a brief effort estimate for building a user authentication module.",
		},
		{
			fromAgent: "qa-meet",
			role:      "QA Tester",
			prompt: "You are a QA Tester in a sprint planning meeting. " +
				"In 1-2 sentences, describe your testing strategy for the user authentication module.",
		},
	}

	for i, turn := range turns {
		content, err := client.Reason(ctx, turn.prompt)
		if err != nil {
			t.Fatalf("turn %d (%s) Minimax reasoning failed: %v", i+1, turn.role, err)
		}
		if strings.TrimSpace(content) == "" {
			t.Fatalf("turn %d (%s) returned empty Minimax response", i+1, turn.role)
		}

		if err := hub.Publish(domain.Message{
			ID:         fmt.Sprintf("meet-msg-%d", i+1),
			FromAgent:  turn.fromAgent,
			Type:       orchestration.EventTask,
			Content:    content,
			MeetingID:  "sprint-e2e",
			OccurredAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("turn %d publish to meeting: %v", i+1, err)
		}
	}

	// Wait a moment for background goroutines to finish (e.g. telemetry, summarization)
	time.Sleep(100 * time.Millisecond)

	// Verify the meeting transcript captured all three turns in order.
	meeting, ok := hub.Meeting("sprint-e2e")
	if !ok {
		t.Fatal("meeting sprint-e2e not found after conversation")
	}

	if got, want := len(meeting.Transcript), len(turns); got != want {
		t.Fatalf("transcript length: got %d, want %d", got, want)
	}

	expectedAgents := []string{"pm-meet", "swe-meet", "qa-meet"}
	for i, agentID := range expectedAgents {
		entry := meeting.Transcript[i]
		if entry.FromAgent != agentID {
			t.Errorf("transcript[%d].FromAgent = %q, want %q", i, entry.FromAgent, agentID)
		}
		if strings.TrimSpace(entry.Content) == "" {
			t.Errorf("transcript[%d] has empty content", i)
		}
		if entry.MeetingID != "sprint-e2e" {
			t.Errorf("transcript[%d].MeetingID = %q, want sprint-e2e", i, entry.MeetingID)
		}
	}

	// All participants must still be marked IN_MEETING.
	for _, id := range []string{"pm-meet", "swe-meet", "qa-meet"} {
		agent, ok := hub.Agent(id)
		if !ok {
			t.Fatalf("agent %s not found after meeting", id)
		}
		if agent.Status != orchestration.StatusInMeeting {
			t.Errorf("agent %s status = %q, want IN_MEETING", id, agent.Status)
		}
	}
}

// TestMinimaxHubAPIKeyFromEnv is a focused integration test that verifies the
// Hub correctly picks up the API key from the environment.  It does not make a
// live Minimax call — only the hub bookkeeping is verified.
func TestMinimaxHubAPIKeyFromEnv(t *testing.T) {
	const testKey = "test-minimax-key-from-env"
	t.Setenv("MINIMAX_API_KEY", testKey)

	hub := orchestration.NewHub()

	// Simulate what the dashboard server does on startup.
	if key := os.Getenv("MINIMAX_API_KEY"); key != "" {
		hub.SetMinimaxAPIKey(key)
	}

	if got := hub.MinimaxAPIKey(); got != testKey {
		t.Fatalf("hub.MinimaxAPIKey() = %q, want %q", got, testKey)
	}
}

// TestMinimaxClientInitializedWithEnvKey verifies that a MinimaxClient
// constructed with the environment key carries the key correctly without
// making a live API call.
func TestMinimaxClientInitializedWithEnvKey(t *testing.T) {
	const testKey = "sk-test-env-key"
	t.Setenv("MINIMAX_API_KEY", testKey)

	key := os.Getenv("MINIMAX_API_KEY")
	if key == "" {
		t.Fatal("expected MINIMAX_API_KEY to be set by t.Setenv")
	}

	client := orchestration.NewMinimaxClient(key)
	if client == nil {
		t.Fatal("NewMinimaxClient returned nil")
	}
	if client.APIKey != key {
		t.Fatalf("client.APIKey = %q, want %q", client.APIKey, key)
	}
}
