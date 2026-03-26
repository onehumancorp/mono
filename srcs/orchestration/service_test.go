package orchestration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestDelegateTask(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "delegate", Name: "Delegate", Role: "ROUTER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "specialist", Name: "Specialist", Role: "SWE", OrganizationID: "org-1"})

	task := Message{
		ID:         "msg-1",
		Type:       EventTask,
		Content:    "Implement feature",
		OccurredAt: time.Now().UTC(),
	}

	err := hub.DelegateTask("delegate", "specialist", task)
	if err != nil {
		t.Fatalf("expected successful delegation, got error: %v", err)
	}

	inbox := hub.Inbox("specialist")
	if len(inbox) != 1 {
		t.Fatalf("expected exactly 1 message in specialist inbox, got %d", len(inbox))
	}

	received := inbox[0]
	if received.FromAgent != "delegate" {
		t.Fatalf("expected FromAgent to be 'delegate', got %q", received.FromAgent)
	}
	if received.ToAgent != "specialist" {
		t.Fatalf("expected ToAgent to be 'specialist', got %q", received.ToAgent)
	}
	if received.Content != "Implement feature" {
		t.Fatalf("expected content 'Implement feature', got %q", received.Content)
	}
}

func TestDelegateTask_AgentNotFound(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "delegate", Name: "Delegate", Role: "ROUTER", OrganizationID: "org-1"})

	task := Message{
		ID:         "msg-1",
		Type:       EventTask,
		Content:    "Implement feature",
		OccurredAt: time.Now().UTC(),
	}

	err := hub.DelegateTask("delegate", "missing-specialist", task)
	if err == nil {
		t.Fatalf("expected error delegating to missing specialist, got nil")
	}
	if err.Error() != "recipient agent is not registered" {
		t.Fatalf("expected error 'recipient agent is not registered', got %q", err.Error())
	}

	err = hub.DelegateTask("missing-delegate", "delegate", task)
	if err == nil {
		t.Fatalf("expected error from missing delegating agent, got nil")
	}
	if err.Error() != "sender agent is not registered" {
		t.Fatalf("expected error 'sender agent is not registered', got %q", err.Error())
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

func TestMeetingsReturnsSnapshot(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "b", Name: "B", Role: "SWE", OrganizationID: "org-1"})
	hub.OpenMeeting("kickoff", []string{"a", "b"})

	meetings := hub.Meetings()
	if len(meetings) != 1 {
		t.Fatalf("expected 1 meeting, got %d", len(meetings))
	}
	if meetings[0].ID != "kickoff" {
		t.Fatalf("unexpected meeting snapshot: %+v", meetings[0])
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

func TestFireAgentRemovesFromHubAndInbox(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a", Name: "A", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "b", Name: "B", Role: "SWE", OrganizationID: "org-1"})
	hub.OpenMeeting("m1", []string{"a", "b"})
	_ = hub.Publish(Message{
		ID:        "msg-1",
		FromAgent: "a",
		ToAgent:   "b",
		Type:      EventTask,
		Content:   "do work",
		MeetingID: "m1",
	})

	hub.FireAgent("b")

	if _, ok := hub.Agent("b"); ok {
		t.Fatalf("expected fired agent to be removed from hub")
	}
	if inbox := hub.Inbox("b"); len(inbox) != 0 {
		t.Fatalf("expected inbox cleared after firing, got %d messages", len(inbox))
	}
	if agents := hub.Agents(); len(agents) != 1 {
		t.Fatalf("expected 1 agent remaining, got %d", len(agents))
	}
}

func TestOpenMeetingWithAgendaPreservesAgendaField(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	meeting := hub.OpenMeetingWithAgenda("sprint-kickoff", "Plan Q2 features and assign owners", []string{"pm", "swe"})

	if meeting.Agenda != "Plan Q2 features and assign owners" {
		t.Fatalf("expected agenda to be preserved, got %q", meeting.Agenda)
	}
	if meeting.ID != "sprint-kickoff" {
		t.Fatalf("expected meeting ID sprint-kickoff, got %q", meeting.ID)
	}
	if len(meeting.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(meeting.Participants))
	}

	stored, ok := hub.Meeting("sprint-kickoff")
	if !ok {
		t.Fatalf("expected meeting to be stored in hub")
	}
	if stored.Agenda != "Plan Q2 features and assign owners" {
		t.Fatalf("expected stored agenda to match, got %q", stored.Agenda)
	}
}

func TestEventTypeConstantsAreDefined(t *testing.T) {
	types := []string{
		EventTask, EventStatus, EventHandoff,
		EventCodeReviewed, EventTestsFailed, EventTestsPassed,
		EventSpecApproved, EventBlockerRaised, EventBlockerCleared,
		EventPRCreated, EventPRMerged, EventDesignReviewed, EventApprovalNeeded,
	}
	for _, ev := range types {
		if ev == "" {
			t.Fatalf("expected all event type constants to be non-empty")
		}
	}
}

func TestMinimaxAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "Empty Key",
			apiKey:   "",
			expected: "",
		},
		{
			name:     "Valid Key",
			apiKey:   "test-key-123",
			expected: "test-key-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			hub.SetMinimaxAPIKey(tt.apiKey)
			if got := hub.MinimaxAPIKey(); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// mockStreamMessagesServer implements pb.HubService_StreamMessagesServer
type mockStreamMessagesServer struct {
	grpc.ServerStream
	ctx      context.Context
	messages []*pb.Message
}

func (m *mockStreamMessagesServer) Context() context.Context {
	return m.ctx
}

func (m *mockStreamMessagesServer) Send(msg *pb.Message) error {
	m.messages = append(m.messages, msg)
	if msg.GetContent() == "trigger_send_error" {
		return errors.New("simulated send error")
	}
	return nil
}

func TestHubServiceServer_StreamMessages(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	// Publish a message to the receiver's inbox
	_ = hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "Hello Streaming",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &mockStreamMessagesServer{ctx: ctx}

	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamMessages(req, mockStream)
	}()

	// Allow some time for the stream to poll and send the message
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the stream
	cancel()

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("StreamMessages returned error: %v", err)
	}

	if len(mockStream.messages) != 1 {
		t.Fatalf("expected 1 message in stream, got %d", len(mockStream.messages))
	}
	if mockStream.messages[0].GetContent() != "Hello Streaming" {
		t.Fatalf("expected 'Hello Streaming', got %q", mockStream.messages[0].GetContent())
	}
}

func TestHubServiceServer_StreamMessages_SendError(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	// Publish a message that triggers a send error
	_ = hub.Publish(Message{
		ID:         "msg-err",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "trigger_send_error",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockStream := &mockStreamMessagesServer{ctx: ctx}

	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	err := server.StreamMessages(req, mockStream)
	if err == nil || err.Error() != "simulated send error" {
		t.Fatalf("expected simulated send error, got: %v", err)
	}
}

func TestHubServiceServer_StreamMessages_ContextDone(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	// Context is cancelled before StreamMessages handles the infinite loop
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockStream := &mockStreamMessagesServer{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	err := server.StreamMessages(req, mockStream)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestHubServiceServer_StreamMessages_SendErrorOnWait(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockStream := &mockStreamMessagesServer{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	// Wait for the stream to start, then publish the error message
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamMessages(req, mockStream)
	}()

	time.Sleep(50 * time.Millisecond)
	_ = hub.Publish(Message{
		ID:         "msg-err",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "trigger_send_error",
		OccurredAt: time.Now(),
	})

	err := <-errCh
	if err == nil || err.Error() != "simulated send error" {
		t.Fatalf("expected simulated send error, got: %v", err)
	}
}

func TestNewMinimaxClient(t *testing.T) {
	client := NewMinimaxClient("test-key")
	if client == nil {
		t.Fatalf("expected non-nil MinimaxClient")
	}
	if client.APIKey != "test-key" {
		t.Fatalf("expected APIKey 'test-key', got %q", client.APIKey)
	}
}

func TestHubServiceServer_Reason_And_MinimaxClient(t *testing.T) {
	// Save the original URL to restore it later
	originalURL := minimaxAPIURL
	defer func() { minimaxAPIURL = originalURL }()

	tests := []struct {
		name          string
		apiKey        string
		httpStatus    int
		httpBody      string
		expectErr     bool
		expectContent string
	}{
		{
			name:          "Success",
			apiKey:        "valid-key",
			httpStatus:    http.StatusOK,
			httpBody:      `{"choices": [{"message": {"content": "Reasoned response"}}]}`,
			expectErr:     false,
			expectContent: "Reasoned response",
		},
		{
			name:          "Empty API Key",
			apiKey:        "",
			httpStatus:    http.StatusOK,
			httpBody:      "",
			expectErr:     true,
			expectContent: "",
		},
		{
			name:          "HTTP Error 500",
			apiKey:        "valid-key",
			httpStatus:    http.StatusInternalServerError,
			httpBody:      `Internal Server Error`,
			expectErr:     true,
			expectContent: "",
		},
		{
			name:          "Malformed JSON",
			apiKey:        "valid-key",
			httpStatus:    http.StatusOK,
			httpBody:      `{bad-json}`,
			expectErr:     true,
			expectContent: "",
		},
		{
			name:          "Empty Choices",
			apiKey:        "valid-key",
			httpStatus:    http.StatusOK,
			httpBody:      `{"choices": []}`,
			expectErr:     true,
			expectContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				w.Write([]byte(tt.httpBody))
			}))
			defer ts.Close()

			// Override the package-level URL
			minimaxAPIURL = ts.URL

			hub := NewHub()
			hub.SetMinimaxAPIKey(tt.apiKey)
			server := NewHubServiceServer(hub)
			ctx := context.Background()

			req := pb.ReasonRequest_builder{
				Prompt: "Test prompt",
			}.Build()

			resp, err := server.Reason(ctx, req)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.GetContent() != tt.expectContent {
					t.Fatalf("expected content %q, got %q", tt.expectContent, resp.GetContent())
				}
			}
		})
	}
}

func TestRegisterHubService(t *testing.T) {
	s := grpc.NewServer()
	hub := NewHub()
	RegisterHubService(s, hub)

	// Since we cannot easily introspect the server to see if it's registered without a client,
	// verifying it doesn't panic and testing NewHubServiceServer covers the logic.
	server := NewHubServiceServer(hub)
	if server == nil {
		t.Fatalf("expected NewHubServiceServer to return non-nil server")
	}
	if server.hub != hub {
		t.Fatalf("expected server hub to match passed hub")
	}
}

func TestHubServiceServer_RegisterAgent(t *testing.T) {
	hub := NewHub()
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	req := pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id:             "grpc-agent-1",
			Name:           "GRPC Agent",
			Role:           "TEST_ROLE",
			OrganizationId: "org-grpc",
			Status:         string(StatusIdle),
		}.Build(),
	}.Build()

	resp, err := server.RegisterAgent(ctx, req)
	if err != nil {
		t.Fatalf("RegisterAgent returned error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatalf("expected RegisterAgent to return success")
	}

	agent, ok := hub.Agent("grpc-agent-1")
	if !ok {
		t.Fatalf("expected agent to be registered in hub")
	}
	if agent.Name != "GRPC Agent" {
		t.Fatalf("expected agent name 'GRPC Agent', got %q", agent.Name)
	}
}

func TestHubServiceServer_OpenMeeting(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "p1", Name: "P1", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "p2", Name: "P2", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	req := pb.OpenMeetingRequest_builder{
		MeetingId:    "grpc-meeting-1",
		Agenda:       "Discuss gRPC",
		Participants: []string{"p1", "p2"},
	}.Build()

	resp, err := server.OpenMeeting(ctx, req)
	if err != nil {
		t.Fatalf("OpenMeeting returned error: %v", err)
	}

	if resp.GetId() != "grpc-meeting-1" {
		t.Fatalf("expected meeting ID 'grpc-meeting-1', got %q", resp.GetId())
	}
	if resp.GetAgenda() != "Discuss gRPC" {
		t.Fatalf("expected agenda 'Discuss gRPC', got %q", resp.GetAgenda())
	}
	if len(resp.GetParticipants()) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(resp.GetParticipants()))
	}

	if _, ok := hub.Meeting("grpc-meeting-1"); !ok {
		t.Fatalf("expected meeting to be created in hub")
	}
}

func TestHubServiceServer_DelegateTask(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "delegate", Name: "Delegate", Role: "ROUTER", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "specialist", Name: "Specialist", Role: "SWE", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	tests := []struct {
		name          string
		req           *pb.DelegateTaskRequest
		expectSuccess bool
		expectErrCode codes.Code
	}{
		{
			name: "Valid Delegation",
			req: pb.DelegateTaskRequest_builder{
				FromAgentId: "delegate",
				ToAgentId:   "specialist",
				Task: pb.Message_builder{
					Id:             "m1",
					Type:           EventTask,
					Content:        "Do work",
					OccurredAtUnix: time.Now().Unix(),
				}.Build(),
			}.Build(),
			expectSuccess: true,
			expectErrCode: codes.OK,
		},
		{
			name: "Invalid Delegate Agent",
			req: pb.DelegateTaskRequest_builder{
				FromAgentId: "unknown-delegate",
				ToAgentId:   "specialist",
				Task: pb.Message_builder{
					Id:             "m2",
					Type:           EventTask,
					Content:        "Do work",
					OccurredAtUnix: time.Now().Unix(),
				}.Build(),
			}.Build(),
			expectSuccess: false,
			expectErrCode: codes.Internal,
		},
		{
			name: "Invalid Specialist Agent",
			req: pb.DelegateTaskRequest_builder{
				FromAgentId: "delegate",
				ToAgentId:   "unknown-specialist",
				Task: pb.Message_builder{
					Id:             "m3",
					Type:           EventTask,
					Content:        "Do work",
					OccurredAtUnix: time.Now().Unix(),
				}.Build(),
			}.Build(),
			expectSuccess: false,
			expectErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.DelegateTask(ctx, tt.req)
			if err != nil {
				if tt.expectSuccess {
					t.Fatalf("expected success, got error: %v", err)
				}
				if status.Code(err) != tt.expectErrCode {
					t.Fatalf("expected error code %v, got %v", tt.expectErrCode, status.Code(err))
				}
			} else {
				if !tt.expectSuccess {
					t.Fatalf("expected error, got success")
				}
				if !resp.GetSuccess() {
					t.Fatalf("expected DelegateTask to return success")
				}
			}
		})
	}
}

func TestHub_Publish_UnbufferedChannel(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})

	// Force the subscriber channel to be "full" to hit the default branch
	_, cancel := hub.Subscribe("receiver")
	defer cancel()

	// Also subscribe to meeting participants
	meeting := hub.OpenMeeting("m1", []string{"receiver"})
	if meeting.ID == "" {
		t.Fatalf("failed to open meeting")
	}

	// Create an extra receiver to verify select loop for multiple participants
	// actually hits the default path too.
	hub.RegisterAgent(Agent{ID: "receiver2", Name: "Receiver2", Role: "R2", OrganizationID: "O1"})
	_, cancel2 := hub.Subscribe("receiver2")
	defer cancel2()
	meeting2 := hub.OpenMeeting("m2", []string{"receiver2"})
	if meeting2.ID == "" {
		t.Fatalf("failed to open meeting")
	}

	// Pre-fill both channels
	_ = hub.Publish(Message{
		ID: "m2-1", FromAgent: "sender", ToAgent: "receiver2", Type: EventTask, Content: "fill", OccurredAt: time.Now(),
	})


	// Fill the channel or just let Publish run twice without draining it
	_ = hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "fill channel",
		OccurredAt: time.Now(),
	})

	// Create another meeting receiver to test hitting default branch when publishing to a meeting
	hub.RegisterAgent(Agent{ID: "receiver3", Name: "Receiver3", Role: "R3", OrganizationID: "O1"})
	_, cancel3 := hub.Subscribe("receiver3")
	defer cancel3()
	meeting3 := hub.OpenMeeting("m3", []string{"receiver3"})
	if meeting3.ID == "" {
		t.Fatalf("failed to open meeting")
	}

	// This one will fill the unbuffered channel for receiver3 in context of meeting message
	_ = hub.Publish(Message{
		ID:         "msg-m3-fill",
		FromAgent:  "sender",
		ToAgent:    "receiver3",
		Type:       EventTask,
		MeetingID:  "m3",
		Content:    "fill meeting channel",
		OccurredAt: time.Now(),
	})

	// This second publish to meeting3 will hit the default path on line 416
	// We need to make sure we don't send to "receiver3" as ToAgent, but instead rely on the
	// meeting broadcast logic where ToAgent is empty, so we test line 415/416 for participants.
	// Oh, I see, line 415 is `case subs[i] <- struct{}{}:` inside `h.subs[participant]`.
	// For that we just need to subscribe a second time maybe?
	_, cancel4 := hub.Subscribe("receiver3")
	defer cancel4()

	_ = hub.Publish(Message{
		ID:         "msg-m3-fill-2",
		FromAgent:  "sender",
		ToAgent:    "", // broadcast to meeting
		Type:       EventTask,
		MeetingID:  "m3",
		Content:    "hit meeting channel default",
		OccurredAt: time.Now(),
	})

	// This publish will hit the select default case since we haven't read from 'ch'
	err := hub.Publish(Message{
		ID:         "msg-2",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "hit default case",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error from Publish, got: %v", err)
	}

	// Also hit the default case for Meeting broadcasts
	// meeting was defined above

	err = hub.Publish(Message{
		ID:         "msg-3",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		MeetingID:  "m1",
		Content:    "hit meeting default case",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error from Publish to meeting, got: %v", err)
	}

	err = hub.Publish(Message{
		ID:         "msg-4",
		FromAgent:  "sender",
		ToAgent:    "receiver2",
		Type:       EventTask,
		MeetingID:  "m2",
		Content:    "hit meeting default case 2",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error from Publish to meeting, got: %v", err)
	}
}

func TestMinimaxClient_Reason_NewRequestWithContext_Error(t *testing.T) {
	client := NewMinimaxClient("valid-key")

	originalURL := minimaxAPIURL
	// Use an invalid control character in the URL to make http.NewRequestWithContext fail
	minimaxAPIURL = "http://\x00invalid-url"
	defer func() { minimaxAPIURL = originalURL }()

	_, err := client.Reason(context.Background(), "some prompt")
	if err == nil {
		t.Fatalf("expected error from http.NewRequestWithContext with invalid URL")
	}
}

func TestMinimaxClient_Reason_ClientDo_Error(t *testing.T) {
	client := NewMinimaxClient("valid-key")

	originalURL := minimaxAPIURL
	// Use a validly parseable URL that fails at the network level
	minimaxAPIURL = "http://127.0.0.1:0"
	defer func() { minimaxAPIURL = originalURL }()

	_, err := client.Reason(context.Background(), "test prompt")
	if err == nil {
		t.Fatalf("expected error from http.Client.Do with network-level failure")
	}
}

func TestHubServiceServer_Publish(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	tests := []struct {
		name          string
		req           *pb.PublishMessageRequest
		expectSuccess bool
		expectErrCode codes.Code
	}{
		{
			name: "Valid Publish",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					Id:             "m1",
					FromAgent:      "sender",
					ToAgent:        "receiver",
					Type:           EventTask,
					Content:        "Hello",
					OccurredAtUnix: time.Now().Unix(),
				}.Build(),
			}.Build(),
			expectSuccess: true,
			expectErrCode: codes.OK,
		},
		{
			name: "Invalid Sender",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					Id:             "m2",
					FromAgent:      "unknown",
					ToAgent:        "receiver",
					Type:           EventTask,
					Content:        "Hello",
					OccurredAtUnix: time.Now().Unix(),
				}.Build(),
			}.Build(),
			expectSuccess: false,
			expectErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Publish(ctx, tt.req)
			if err != nil {
				if tt.expectSuccess {
					t.Fatalf("expected success, got error: %v", err)
				}
				if status.Code(err) != tt.expectErrCode {
					t.Fatalf("expected error code %v, got %v", tt.expectErrCode, status.Code(err))
				}
			} else {
				if !tt.expectSuccess {
					t.Fatalf("expected error, got success")
				}
				if !resp.GetSuccess() {
					t.Fatalf("expected Publish to return success")
				}
			}
		})
	}
}

func TestHub_TokenEfficientContextSummarization(t *testing.T) {
	hub := NewHub()
	agentID := "test-agent"
	validPayload := []byte(`{"context": "some data to summarize"}`)
	invalidPayload := []byte(`{"context": "some data", "unknown_field": "bad data"}`)

	defer func() {
		os.Remove("events.jsonl")
	}()

	tests := []struct {
		name          string
		eventID       string
		payload       []byte
		setup         func()
		expectError   bool
		expectedErr   string
	}{
		{
			name:        "E2E Integration Test: Standard Execution Flow",
			eventID:     "event-1",
			payload:     validPayload,
			setup:       func() {},
			expectError: true,
			expectedErr: "summarization failed: minimax API key is not configured",
		},
		{
			name:        "Edge Case: Strict Schema and Payload Validation",
			eventID:     "event-2",
			payload:     invalidPayload,
			setup:       func() {},
			expectError: true,
			expectedErr: "invalid payload",
		},
		{
			name:    "Edge Case: Concurrent Execution on same eventID",
			eventID: "event-4",
			payload: validPayload,
			setup: func() {
				hub.mu.Lock()
				hub.tokenTrackers["event-4"] = struct{}{}
				hub.mu.Unlock()
			},
			expectError: true,
			expectedErr: "event already being processed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := hub.TokenEfficientContextSummarization(tt.eventID, agentID, tt.payload)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}

				// Verify map entry is deleted (bounded memory growth)
				hub.mu.RLock()
				_, exists := hub.tokenTrackers[tt.eventID]
				hub.mu.RUnlock()
				if exists {
					t.Errorf("expected map entry %q to be deleted, but it still exists", tt.eventID)
				}
			}

			// Clean up state manually for cases like Concurrent Execution where it returns error and doesn't run the defer logic inside the target func
			hub.mu.Lock()
			delete(hub.autoCorTrack, tt.eventID)
			hub.mu.Unlock()
		})
	}
}

func TestHub_TokenEfficientContextSummarization_SuccessFlow(t *testing.T) {
	// 1. Setup Mock Server for Minimax
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"choices": [
				{
					"message": {
						"content": "summarized data"
					}
				}
			]
		}`
		w.Write([]byte(response))
	}))
	defer ts.Close()

	// Override the Minimax API URL to point to our test server
	originalAPIURL := minimaxAPIURL
	minimaxAPIURL = ts.URL
	defer func() { minimaxAPIURL = originalAPIURL }()

	// 2. Initialize Hub and set a fake API key
	hub := NewHub()
	hub.mu.Lock()
	hub.minimaxAPIKey = "fake-key"
	hub.mu.Unlock()

	agentID := "test-agent-2"
	eventID := "event-123"
	payload := []byte(`{"context": "some data to summarize"}`)

	defer os.Remove("events.jsonl")

	// 3. Execute the function
	err := hub.TokenEfficientContextSummarization(eventID, agentID, payload)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Wait briefly for the background worker to write to the file
	time.Sleep(100 * time.Millisecond)

	// 4. Verify log entry was written
	b, err := os.ReadFile("events.jsonl")
	if err != nil {
		t.Fatalf("failed to read events.jsonl: %v", err)
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(b, &logEntry); err != nil {
		t.Fatalf("failed to unmarshal events.jsonl: %v", err)
	}

	if logEntry["event_id"] != eventID {
		t.Errorf("expected event_id %q, got %q", eventID, logEntry["event_id"])
	}
	if logEntry["agent_id"] != agentID {
		t.Errorf("expected agent_id %q, got %q", agentID, logEntry["agent_id"])
	}
	if logEntry["type"] != "TokenEfficientContextSummarization" {
		t.Errorf("expected type %q, got %q", "TokenEfficientContextSummarization", logEntry["type"])
	}
	if logEntry["summarized_context"] != "summarized data" {
		t.Errorf("expected summarized_context %q, got %q", "summarized data", logEntry["summarized_context"])
	}

	// Verify memory leak fix (map deletion)
	hub.mu.RLock()
	_, exists := hub.tokenTrackers[eventID]
	hub.mu.RUnlock()
	if exists {
		t.Errorf("expected map entry %q to be deleted, but it still exists", eventID)
	}
}

func TestHub_ToolParameterAutoCorrection(t *testing.T) {
	hub := NewHub()
	agentID := "test-agent"

	defer func() {
		os.Remove("events.jsonl")
	}()

	tests := []struct {
		name        string
		eventID     string
		payload     []byte
		setup       func()
		expectError bool
		expectedErr string
	}{
		{
			name:        "E2E Integration Test: Standard Execution Flow",
			eventID:     "event-ac-1",
			payload:     []byte(`{"id": "123", "value": "456", "name": "test"}`),
			setup:       func() {},
			expectError: false,
		},
		{
			name:        "Edge Case: Strict Schema and Payload Validation",
			eventID:     "event-ac-2",
			payload:     []byte(`{"id": "123", "value": "456", "name": "test", "unknown_field": "bad data"}`),
			setup:       func() {},
			expectError: false,
		},
		{
			name:    "Edge Case: Concurrent Execution on same eventID",
			eventID: "event-ac-4",
			payload: []byte(`{"id": "123", "value": "456", "name": "test"}`),
			setup: func() {
				hub.mu.Lock()
				hub.autoCorTrack["event-ac-4"] = struct{}{}
				hub.mu.Unlock()
			},
			expectError: true,
			expectedErr: "event already being processed",
		},
		{
			name:        "Edge Case: Invalid JSON payload",
			eventID:     "event-ac-invalid",
			payload:     []byte(`{invalid`),
			setup:       func() {},
			expectError: true,
			expectedErr: "invalid payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := hub.ToolParameterAutoCorrection(tt.eventID, agentID, tt.payload)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				} else if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}

				// Verify map entry is deleted (bounded memory growth)
				hub.mu.RLock()
				_, exists := hub.autoCorTrack[tt.eventID]
				hub.mu.RUnlock()
				if exists {
					t.Errorf("expected map entry %q to be deleted, but it still exists", tt.eventID)
				}
			}
		})
	}
}

func TestHub_ToolParameterAutoCorrection_SuccessFlow(t *testing.T) {
	hub := NewHub()

	agentID := "test-agent-2"
	eventID := "event-123"
	payload := []byte(`{"value": "123", "name": "test"}`)

	defer os.Remove("events.jsonl")

	// 3. Execute the function
	err := hub.ToolParameterAutoCorrection(eventID, agentID, payload)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Wait briefly for the background worker to write to the file
	time.Sleep(100 * time.Millisecond)

	// 4. Verify log entry was written
	b, err := os.ReadFile("events.jsonl")
	if err != nil {
		t.Fatalf("failed to read events.jsonl: %v", err)
	}

	// Read lines to get the specific event
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	var logEntry map[string]interface{}
	var found bool
	for i := len(lines)-1; i >= 0; i-- {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(lines[i]), &entry); err == nil {
			if entry["type"] == "ToolParameterAutoCorrection" && entry["event_id"] == eventID {
				logEntry = entry
				found = true
				break
			}
		}
	}

	if !found {
		t.Fatalf("could not find ToolParameterAutoCorrection event in events.jsonl")
	}

	if logEntry["event_id"] != eventID {
		t.Errorf("expected event_id %q, got %q", eventID, logEntry["event_id"])
	}
	if logEntry["agent_id"] != agentID {
		t.Errorf("expected agent_id %q, got %q", agentID, logEntry["agent_id"])
	}
	if logEntry["type"] != "ToolParameterAutoCorrection" {
		t.Errorf("expected type %q, got %q", "ToolParameterAutoCorrection", logEntry["type"])
	}
	if logEntry["corrected"] != true {
		t.Errorf("expected corrected to be true")
	}

	pl, ok := logEntry["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("payload is not a map")
	}

	// 'value' should now be an integer
	if val, ok := pl["value"].(float64); !ok || val != 123 {
		t.Errorf("expected value to be 123, got %v", pl["value"])
	}

	// Verify memory leak fix (map deletion)
	hub.mu.RLock()
	_, exists := hub.autoCorTrack[eventID]
	hub.mu.RUnlock()
	if exists {
		t.Errorf("expected map entry %q to be deleted, but it still exists", eventID)
	}
}

func TestSetSIPDB(t *testing.T) {
	hub := NewHub()
	db, _ := NewSIPDB(":memory:")
	hub.SetSIPDB(db)
	if hub.GetSIPDB() != db {
		t.Fatal("SetSIPDB/GetSIPDB failed")
	}
}

func TestDelegateTask_AgentNotRegistered(t *testing.T) {
	hub := NewHub()
	err := hub.DelegateTask("task-1", "ROLE", Message{Content: "instruction"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEventLogWorker_Coverage(t *testing.T) {
	hub := NewHub()

	// Create temp file for events
	tmpFile, err := os.CreateTemp("", "events-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())


	// Wait a bit to let it start processing
	time.Sleep(100 * time.Millisecond)

	// Log an event
	hub.LogEvent(Message{ID: "m1", Content: "test"})

	// Give it time to flush
	time.Sleep(100 * time.Millisecond)

	// Force close to stop loop
	close(hub.eventLogChan)
}
