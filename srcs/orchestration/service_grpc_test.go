package orchestration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestRegisterAgentViaGRPC(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	req := pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id:             "test-agent",
			Name:           "Test Agent",
			Role:           "QA_ENGINEER",
			OrganizationId: "org-1",
			Status:         "ACTIVE",
		}.Build(),
	}.Build()

	res, err := srv.RegisterAgent(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.GetSuccess() {
		t.Errorf("expected success to be true")
	}

	agent, ok := hub.Agent("test-agent")
	if !ok {
		t.Fatalf("agent not registered in hub")
	}
	if agent.Name != "Test Agent" || agent.Role != "QA_ENGINEER" || agent.Status != "ACTIVE" {
		t.Errorf("agent fields mismatch: %+v", agent)
	}
}

func TestOpenMeetingViaGRPC(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "p1", Name: "P1", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "p2", Name: "P2", Role: "SWE", OrganizationID: "org-1"})
	srv := NewHubServiceServer(hub)

	req := pb.OpenMeetingRequest_builder{
		MeetingId:    "m-1",
		Agenda:       "Test Agenda",
		Participants: []string{"p1", "p2"},
	}.Build()

	res, err := srv.OpenMeeting(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GetId() != "m-1" || res.GetAgenda() != "Test Agenda" || len(res.GetParticipants()) != 2 {
		t.Errorf("meeting response mismatch: %+v", res)
	}

	meeting, ok := hub.Meeting("m-1")
	if !ok {
		t.Fatalf("meeting not registered in hub")
	}
	if meeting.Agenda != "Test Agenda" {
		t.Errorf("meeting agenda mismatch in hub")
	}
}

func TestPublishViaGRPC(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a1", Name: "A1", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "a2", Name: "A2", Role: "SWE", OrganizationID: "org-1"})
	srv := NewHubServiceServer(hub)

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			Id:             "msg-1",
			FromAgent:      "a1",
			ToAgent:        "a2",
			Type:           "task",
			Content:        "Do it",
			OccurredAtUnix: time.Now().Unix(),
		}.Build(),
	}.Build()

	res, err := srv.Publish(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.GetSuccess() {
		t.Errorf("expected success to be true")
	}

	inbox := hub.Inbox("a2")
	if len(inbox) != 1 || inbox[0].Content != "Do it" {
		t.Errorf("message not published to inbox correctly: %+v", inbox)
	}
}

// MockStreamServer implements pb.HubService_StreamMessagesServer
type MockStreamServer struct {
	ctx      context.Context
	messages []*pb.Message
}

func (m *MockStreamServer) Context() context.Context { return m.ctx }
func (m *MockStreamServer) Send(msg *pb.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}
func (m *MockStreamServer) SendHeader(metadata.MD) error { return nil }
func (m *MockStreamServer) SetTrailer(metadata.MD)       {}
func (m *MockStreamServer) SetHeader(metadata.MD) error  { return nil }
func (m *MockStreamServer) SendMsg(m_ interface{}) error { return nil }
func (m *MockStreamServer) RecvMsg(m_ interface{}) error { return nil }

func TestStreamMessagesViaGRPC(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "a1", Name: "A1", Role: "PM", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "a2", Name: "A2", Role: "SWE", OrganizationID: "org-1"})
	srv := NewHubServiceServer(hub)

	// Publish an initial message
	hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "a1",
		ToAgent:    "a2",
		Type:       "task",
		Content:    "initial task",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream := &MockStreamServer{ctx: ctx, messages: make([]*pb.Message, 0)}

	go func() {
		// Publish a message while streaming
		time.Sleep(100 * time.Millisecond)
		hub.Publish(Message{
			ID:         "msg-2",
			FromAgent:  "a1",
			ToAgent:    "a2",
			Type:       "task",
			Content:    "new task",
			OccurredAt: time.Now(),
		})
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: "a2"}.Build(), stream)
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Fatalf("StreamMessages failed: %v", err)
	}

	if len(stream.messages) != 2 {
		t.Errorf("expected 2 messages streamed, got %d", len(stream.messages))
	} else {
		if stream.messages[0].GetContent() != "initial task" {
			t.Errorf("expected msg-1, got %s", stream.messages[0].GetContent())
		}
		if stream.messages[1].GetContent() != "new task" {
			t.Errorf("expected msg-2, got %s", stream.messages[1].GetContent())
		}
	}
}

func TestHubMinimaxAPIKey(t *testing.T) {
	hub := NewHub()
	if hub.MinimaxAPIKey() != "" {
		t.Errorf("expected empty API key initially")
	}
	hub.SetMinimaxAPIKey("test-key")
	if hub.MinimaxAPIKey() != "test-key" {
		t.Errorf("expected 'test-key'")
	}
}

func TestReasonViaMinimaxEmptyKey(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	hub.SetMinimaxAPIKey("")
	req := pb.ReasonRequest_builder{Prompt: "test prompt"}.Build()
	_, err := srv.Reason(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error due to empty API key")
	}
}

func TestReasonViaMinimaxDummyKey(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	hub.SetMinimaxAPIKey("dummy-key")
	req := pb.ReasonRequest_builder{Prompt: "test prompt"}.Build()
	_, err := srv.Reason(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error due to invalid API key")
	}
}

func TestRegisterHubServiceCoverage(t *testing.T) {
	hub := NewHub()
	srv := grpc.NewServer()
	RegisterHubService(srv, hub)
}

func TestPublishViaGRPCError(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			Id:             "msg-1",
			FromAgent:      "missing",
			ToAgent:        "missing",
			Type:           "task",
			Content:        "Do it",
			OccurredAtUnix: time.Now().Unix(),
		}.Build(),
	}.Build()

	_, err := srv.Publish(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestStreamMessagesViaGRPCCancellation(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	stream := &MockStreamServer{ctx: ctx, messages: make([]*pb.Message, 0)}

	err := srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: "a2"}.Build(), stream)
	if err != nil && err != context.Canceled {
		t.Fatalf("expected graceful shutdown or context canceled, got %v", err)
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test read error")
}

func TestMinimaxClientReasonDecodeError(t *testing.T) {
	// Let's not test JSON decode error by writing to http.ResponseWriter
	// because httptest.Server encodes it. But we can send malformed json
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{ malformed json }"))
	}))
	defer ts.Close()

	originalURL := minimaxAPIURL
	minimaxAPIURL = ts.URL
	defer func() { minimaxAPIURL = originalURL }()

	client := NewMinimaxClient("valid-key")
	_, err := client.Reason(context.Background(), "test")
	if err == nil {
		t.Fatalf("expected error on malformed JSON")
	}
}

func TestMinimaxClientReasonInvalidRequest(t *testing.T) {
	client := NewMinimaxClient("valid-key")
	// Using a cancelled context to trigger a request creation or execution error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Reason(ctx, "test")
	if err == nil {
		t.Fatalf("expected error on cancelled context")
	}
}
