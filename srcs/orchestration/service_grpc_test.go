package orchestration

import (
	"context"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
)

func TestHubService_RegisterAgent(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	req := pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id:             "agent-1",
			Name:           "Test Agent",
			Role:           "Tester",
			OrganizationId: "org-1",
			Status:         "ACTIVE",
		}.Build(),
	}.Build()

	resp, err := srv.RegisterAgent(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Error("expected success to be true")
	}

	agent, ok := hub.Agent("agent-1")
	if !ok {
		t.Fatal("expected agent to be registered")
	}
	if agent.Name != "Test Agent" {
		t.Errorf("expected agent name Test Agent, got %s", agent.Name)
	}
}

func TestHubService_OpenMeeting(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	hub.RegisterAgent(Agent{ID: "agent-1"})

	req := pb.OpenMeetingRequest_builder{
		MeetingId:    "meet-1",
		Agenda:       "Discuss B2B",
		Participants: []string{"agent-1"},
	}.Build()

	resp, err := srv.OpenMeeting(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.GetId() != "meet-1" {
		t.Errorf("expected meeting id meet-1, got %s", resp.GetId())
	}
	if resp.GetAgenda() != "Discuss B2B" {
		t.Errorf("expected agenda Discuss B2B, got %s", resp.GetAgenda())
	}

	meeting, ok := hub.Meeting("meet-1")
	if !ok {
		t.Fatal("expected meeting to be created")
	}
	if meeting.Agenda != "Discuss B2B" {
		t.Errorf("expected agenda Discuss B2B, got %s", meeting.Agenda)
	}
}

func TestHubService_Publish(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	hub.RegisterAgent(Agent{ID: "sender-1"})
	hub.RegisterAgent(Agent{ID: "receiver-1"})

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			Id:             "msg-1",
			FromAgent:      "sender-1",
			ToAgent:        "receiver-1",
			Type:           EventTask,
			Content:        "Hello",
			OccurredAtUnix: time.Now().Unix(),
		}.Build(),
	}.Build()

	resp, err := srv.Publish(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Error("expected success to be true")
	}

	inbox := hub.Inbox("receiver-1")
	if len(inbox) != 1 {
		t.Fatalf("expected 1 message, got %d", len(inbox))
	}
	if inbox[0].Content != "Hello" {
		t.Errorf("expected content Hello, got %s", inbox[0].Content)
	}
}

func TestHubService_Publish_Error(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			FromAgent: "unknown-1", // Will fail validation
		}.Build(),
	}.Build()

	_, err := srv.Publish(context.Background(), req)
	if err == nil {
		t.Fatal("expected error publishing from unknown agent")
	}
}

// Dummy stream for testing
type mockStream struct {
	grpc.ServerStream
	ctx context.Context
	msgs []*pb.Message
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Send(msg *pb.Message) error {
	m.msgs = append(m.msgs, msg)
	return nil
}


func TestHubService_StreamMessages(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)
	hub.RegisterAgent(Agent{ID: "agent-1"})

	ctx, cancel := context.WithCancel(context.Background())

	stream := &mockStream{
		ctx: ctx,
	}

	// Publish a message
	hub.Publish(Message{
		ID:        "msg-1",
		FromAgent: "agent-1",
		ToAgent:   "agent-1", // Sending direct to self's inbox
		Content:   "Hello from stream",
	})

	// Stream messages in background
	go func() {
		time.Sleep(2 * time.Second)
		cancel() // Stop the streaming
	}()

	err := srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: "agent-1"}.Build(), stream)
	if err != nil {
		t.Fatalf("failed to open stream: %v", err)
	}

	if len(stream.msgs) < 1 {
		t.Fatalf("expected 1 message, got %d", len(stream.msgs))
	}

	if stream.msgs[0].GetContent() != "Hello from stream" {
		t.Errorf("expected Hello from stream, got %s", stream.msgs[0].GetContent())
	}
}
