package integration

import (
	"github.com/onehumancorp/mono/srcs/sip"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/integrations/chatwoot"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// TestMultiAgentMeetingInteraction verifies that multiple agents can exchange
// messages within a shared meeting room and that the transcript is correctly
// accumulated in order.
func TestMultiAgentMeetingInteraction(t *testing.T) {
	hub := orchestration.NewHub()

	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Alice", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Bob", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})
	hub.RegisterAgent(orchestration.Agent{ID: "qa-1", Name: "Carol", Role: "QA_ENGINEER", OrganizationID: "org-1"})

	hub.OpenMeetingWithAgenda("sprint-1", "Plan Q3 sprint", []string{"pm-1", "swe-1", "qa-1"})

	messages := []struct {
		from    string
		content string
	}{
		{"pm-1", "Let's align on the sprint goals for Q3."},
		{"swe-1", "I can take the authentication refactor. Estimate: 3 days."},
		{"qa-1", "I'll write integration tests for the new auth flow once SWE is done."},
		{"pm-1", "Great. Let's target code-complete by end of week."},
	}

	for _, m := range messages {
		if err := hub.Publish(sip.Message{
			ID:        fmt.Sprintf("msg-%s-%d", m.from, time.Now().UnixNano()),
			FromAgent: m.from,
			Type:      orchestration.EventTask,
			Content:   m.content,
			MeetingID: "sprint-1",
		}); err != nil {
			t.Fatalf("Publish from %s: %v", m.from, err)
		}
	}

	meeting, ok := hub.Meeting("sprint-1")
	if !ok {
		t.Fatal("meeting sprint-1 not found")
	}

	if got, want := len(meeting.Transcript), len(messages); got != want {
		t.Fatalf("transcript length: got %d, want %d", got, want)
	}

	for i, m := range messages {
		if meeting.Transcript[i].FromAgent != m.from {
			t.Errorf("transcript[%d].FromAgent = %q, want %q", i, meeting.Transcript[i].FromAgent, m.from)
		}
		if meeting.Transcript[i].Content != m.content {
			t.Errorf("transcript[%d].Content = %q, want %q", i, meeting.Transcript[i].Content, m.content)
		}
		if meeting.Transcript[i].MeetingID != "sprint-1" {
			t.Errorf("transcript[%d].MeetingID = %q, want sprint-1", i, meeting.Transcript[i].MeetingID)
		}
	}

	// All three agents should be marked IN_MEETING.
	for _, id := range []string{"pm-1", "swe-1", "qa-1"} {
		agent, ok := hub.Agent(id)
		if !ok {
			t.Fatalf("agent %s not found", id)
		}
		if agent.Status != orchestration.StatusInMeeting {
			t.Errorf("agent %s status = %q, want IN_MEETING", id, agent.Status)
		}
	}
}

// TestChatwootAgentChat verifies agent-to-agent communication brokered through
// a Chatwoot inbox: two agents open a conversation and exchange messages, each
// visible in the conversation's message list.
func TestChatwootAgentChat(t *testing.T) {
	srv := newMockChatwootServer(t)
	defer srv.Close()

	// Agent A signs in and sets up the inbox.
	agentA := chatwoot.NewClient(srv.URL)
	if err := agentA.SignIn("admin@ohc.local", "changeme"); err != nil {
		t.Fatalf("agentA sign-in: %v", err)
	}

	inbox, err := agentA.CreateAPIInbox("agent-chat")
	if err != nil {
		t.Fatalf("create inbox: %v", err)
	}
	if inbox.ID == 0 {
		t.Fatal("expected non-zero inbox ID")
	}

	// Create a contact representing the human / counterpart.
	contact, err := agentA.CreateContact("Agent-B", "agent-b@ohc.local")
	if err != nil {
		t.Fatalf("create contact: %v", err)
	}

	// Open a conversation.
	conv, err := agentA.CreateConversation(inbox.ID, contact.ID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if conv.ID == 0 {
		t.Fatal("expected non-zero conversation ID")
	}

	// Agent A sends the opening message.
	msg1, err := agentA.SendMessage(conv.ID, "Hello from Agent A — ready for the handoff.", "outgoing")
	if err != nil {
		t.Fatalf("agentA send message: %v", err)
	}
	if msg1.Content == "" {
		t.Fatal("expected non-empty message content in response")
	}

	// Agent B (same client credentials for the mock) sends a reply.
	agentB := chatwoot.NewClient(srv.URL)
	if err := agentB.SignIn("admin@ohc.local", "changeme"); err != nil {
		t.Fatalf("agentB sign-in: %v", err)
	}
	agentB.AccountID = agentA.AccountID

	msg2, err := agentB.SendMessage(conv.ID, "Acknowledged. Taking over the conversation.", "incoming")
	if err != nil {
		t.Fatalf("agentB send message: %v", err)
	}
	if msg2.Content == "" {
		t.Fatal("expected non-empty message content in agentB response")
	}

	// Verify both messages are retrievable.
	msgs, err := agentA.ListMessages(conv.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello from Agent A — ready for the handoff." {
		t.Errorf("unexpected first message: %q", msgs[0].Content)
	}
	if msgs[1].Content != "Acknowledged. Taking over the conversation." {
		t.Errorf("unexpected second message: %q", msgs[1].Content)
	}
}

// TestHumanChatWithAgent simulates a human user starting a conversation in
// Chatwoot and an AI agent (from the orchestration Hub) replying.  The flow:
//  1. Human sends an "incoming" message via the Chatwoot API.
//  2. The agent detects the message and publishes a reply via the Hub.
//  3. The agent posts the reply back to the conversation as an "outgoing" message.
//  4. The conversation history contains both messages.
func TestHumanChatWithAgent(t *testing.T) {
	srv := newMockChatwootServer(t)
	defer srv.Close()

	// Set up the orchestration hub with a router and a support agent.
	hub := orchestration.NewHub()
	// The "router" is the system that dispatches inbound human messages to agents.
	hub.RegisterAgent(orchestration.Agent{
		ID:             "router",
		Name:           "Router",
		Role:           "ROUTER",
		OrganizationID: "org-1",
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "support-1",
		Name:           "Support Agent",
		Role:           "CUSTOMER_SUPPORT",
		OrganizationID: "org-1",
	})

	// The support agent signs in to Chatwoot.
	ct := chatwoot.NewClient(srv.URL)
	if err := ct.SignIn("admin@ohc.local", "changeme"); err != nil {
		t.Fatalf("chatwoot sign-in: %v", err)
	}

	inbox, err := ct.CreateAPIInbox("support")
	if err != nil {
		t.Fatalf("create inbox: %v", err)
	}

	// Create a contact for the human user.
	human, err := ct.CreateContact("Human User", "human@example.com")
	if err != nil {
		t.Fatalf("create contact: %v", err)
	}

	// Human opens a conversation.
	conv, err := ct.CreateConversation(inbox.ID, human.ID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	// Human sends an incoming message.
	humanMsg, err := ct.SendMessage(conv.ID, "Hi! I need help with my account.", "incoming")
	if err != nil {
		t.Fatalf("human send message: %v", err)
	}
	if humanMsg.Content != "Hi! I need help with my account." {
		t.Fatalf("unexpected human message content: %q", humanMsg.Content)
	}

	// Agent reads the conversation messages (simulating a webhook or polling).
	incomingMsgs, err := ct.ListMessages(conv.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(incomingMsgs) == 0 {
		t.Fatal("agent found no messages in the conversation")
	}

	// The router dispatches the human message to the support agent's inbox.
	if err := hub.Publish(sip.Message{
		ID:        "human-chat-1",
		FromAgent: "router",
		ToAgent:   "support-1",
		Type:      orchestration.EventTask,
		Content:   fmt.Sprintf("Inbound customer query: %s", incomingMsgs[0].Content),
	}); err != nil {
		t.Fatalf("hub dispatch to support agent: %v", err)
	}

	// Agent sends its reply back through Chatwoot.
	agentReply := "Hello! I'm happy to help you with your account. Could you please provide your account email?"
	replyMsg, err := ct.SendMessage(conv.ID, agentReply, "outgoing")
	if err != nil {
		t.Fatalf("agent reply: %v", err)
	}
	if replyMsg.Content != agentReply {
		t.Fatalf("unexpected agent reply content: %q", replyMsg.Content)
	}

	// Verify the full conversation history: human message + agent reply.
	allMsgs, err := ct.ListMessages(conv.ID)
	if err != nil {
		t.Fatalf("list all messages: %v", err)
	}
	if len(allMsgs) != 2 {
		t.Fatalf("expected 2 messages in conversation, got %d", len(allMsgs))
	}

	// Confirm the agent's inbox on the hub is populated.
	inbox2 := hub.Inbox("support-1")
	if len(inbox2) == 0 {
		t.Fatal("expected at least one message in agent's hub inbox")
	}
}

// ── Mock Chatwoot server ──────────────────────────────────────────────────────

// newMockChatwootServer creates a lightweight httptest.Server that implements
// the Chatwoot REST API endpoints used by the integration tests.
func newMockChatwootServer(t *testing.T) *httptest.Server {
	t.Helper()

	type state struct {
		nextInboxID int
		nextConvID  int
		nextMsgID   int
		nextContact int
		inboxes     []map[string]any
		convMsgs    map[int][]map[string]any
	}

	s := &state{
		nextInboxID: 1,
		nextConvID:  100,
		nextMsgID:   1000,
		nextContact: 10,
		convMsgs:    map[int][]map[string]any{},
	}

	mux := http.NewServeMux()

	// Auth.
	mux.HandleFunc("/auth/sign_in", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		respondJSON(w, map[string]any{
			"data": map[string]any{
				"access_token": "test-token",
				"account_id":   42,
			},
		})
	})

	// Inboxes.
	mux.HandleFunc("/api/v1/accounts/42/inboxes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, map[string]any{"payload": s.inboxes})
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inbox := map[string]any{"id": s.nextInboxID, "name": body["name"]}
			s.inboxes = append(s.inboxes, inbox)
			s.nextInboxID++
			respondJSON(w, inbox)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Contacts.
	mux.HandleFunc("/api/v1/accounts/42/contacts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		contact := map[string]any{"id": s.nextContact, "name": body["name"], "email": body["email"]}
		s.nextContact++
		respondJSON(w, contact)
	})

	// Conversations (create).
	mux.HandleFunc("/api/v1/accounts/42/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		id := s.nextConvID
		s.nextConvID++
		s.convMsgs[id] = []map[string]any{}
		respondJSON(w, map[string]any{
			"id":         id,
			"inbox_id":   body["inbox_id"],
			"account_id": 42,
			"display_id": id,
		})
	})

	// Messages (send / list) – path: /api/v1/accounts/42/conversations/{id}/messages
	mux.HandleFunc("/api/v1/accounts/42/conversations/", func(w http.ResponseWriter, r *http.Request) {
		var convID int
		if _, err := fmt.Sscanf(r.URL.Path, "/api/v1/accounts/42/conversations/%d/messages", &convID); err != nil {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			msg := map[string]any{
				"id":              s.nextMsgID,
				"content":         body["content"],
				"message_type":    0,
				"conversation_id": convID,
			}
			s.nextMsgID++
			s.convMsgs[convID] = append(s.convMsgs[convID], msg)
			respondJSON(w, msg)
		case http.MethodGet:
			respondJSON(w, map[string]any{"payload": s.convMsgs[convID]})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func respondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
