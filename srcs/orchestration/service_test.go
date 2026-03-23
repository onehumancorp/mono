package orchestration

import (
	"errors"
	"strings"
	"context"
	"net/http"
	"net/http/httptest"
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

	err = hub.DelegateTask("missing-delegate", "delegate", task)
	if err == nil {
		t.Fatalf("expected error from missing delegating agent, got nil")
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

func TestHubPublish_ChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "SWE", OrganizationID: "org-1"})

	// Subscribe to the receiver to get the channel
	hub.Subscribe("receiver")

	// Fill the channel (capacity is 100) using the unexported subs map
	hub.mu.Lock()
	subs := hub.subs["receiver"]
	for i := 0; i < cap(subs[0]); i++ {
		subs[0] <- struct{}{}
	}
	hub.mu.Unlock()

	// This should trigger the `default:` branch in Publish
	err := hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "overflow",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error publishing to full channel: %v", err)
	}
}

func TestHubPublish_MeetingChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "p1", Name: "Participant 1", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "p2", Name: "Participant 2", Role: "SWE", OrganizationID: "org-1"})

	hub.OpenMeetingWithAgenda("m1", "Test", []string{"p1", "p2"})

	hub.Subscribe("p1")
	hub.mu.Lock()
	subs := hub.subs["p1"]
	for i := 0; i < cap(subs[0]); i++ {
		subs[0] <- struct{}{}
	}
	hub.mu.Unlock()

	err := hub.Publish(Message{
		ID:         "msg-2",
		FromAgent:  "p2",
		MeetingID:  "m1",
		Type:       EventTask,
		Content:    "overflow meeting",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error publishing meeting message to full channel: %v", err)
	}
}

type failingStream struct {
	grpc.ServerStream
	ctx      context.Context
	messages []*pb.Message
	failOn   int
	sent     int
}

func (m *failingStream) Context() context.Context {
	return m.ctx
}

func (m *failingStream) Send(msg *pb.Message) error {
	m.sent++
	if m.sent == m.failOn {
		return errors.New("simulated stream send error")
	}
	m.messages = append(m.messages, msg)
	return nil
}

func TestHubServiceServer_StreamMessages_FailInitialSend(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	// Publish to inbox so initial sendNewMessages has something to send
	_ = hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "initial",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &failingStream{ctx: ctx, failOn: 1} // Fail immediately on first send

	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	err := server.StreamMessages(req, mockStream)
	if err == nil || !strings.Contains(err.Error(), "simulated stream send error") {
		t.Fatalf("expected stream send error, got %v", err)
	}
}

func TestHubServiceServer_StreamMessages_FailLaterSend(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &failingStream{ctx: ctx, failOn: 2} // Fail on second send

	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamMessages(req, mockStream)
	}()

	time.Sleep(50 * time.Millisecond) // wait for sub to happen

	// Publish first message (will succeed, count=1)
	_ = hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "first",
		OccurredAt: time.Now(),
	})

	time.Sleep(50 * time.Millisecond) // Wait for first send

	// Publish second message (will fail, count=2)
	_ = hub.Publish(Message{
		ID:         "msg-2",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "second",
		OccurredAt: time.Now(),
	})

	err := <-errCh
	if err == nil || !strings.Contains(err.Error(), "simulated stream send error") {
		t.Fatalf("expected stream send error on second send, got %v", err)
	}
}

func TestMinimaxClient_Reason_InvalidJSON(t *testing.T) {
	client := NewMinimaxClient("test-key")

	originalURL := minimaxAPIURL
	minimaxAPIURL = "http://\x7f unresolvable" // Invalid URL to trigger http.NewRequestWithContext error
	defer func() { minimaxAPIURL = originalURL }()

	_, err := client.Reason(context.Background(), "test")
	if err == nil {
		t.Fatalf("expected HTTP request creation error, got nil")
	}
}

func TestMinimaxClient_Reason_HTTPClientError(t *testing.T) {
	client := NewMinimaxClient("test-key")

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	originalURL := minimaxAPIURL
	minimaxAPIURL = server.URL
	defer func() { minimaxAPIURL = originalURL }()

	_, err := client.Reason(context.Background(), "test prompt")
	if err == nil {
		t.Fatalf("expected error from internal server error, got nil")
	}
	if !strings.Contains(err.Error(), "minimax API error (status 500)") {
		t.Fatalf("expected status 500 error, got %v", err)
	}
}
