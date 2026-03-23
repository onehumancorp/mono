package orchestration

import (
	"context"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
)

func TestPublish_ChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	hub.mu.Lock()
	ch := make(chan struct{}, 1)
	hub.subs["swe-1"] = append(hub.subs["swe-1"], ch)
	hub.mu.Unlock()

	ch <- struct{}{}

	err := hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       "task",
		Content:    "Implement the feature",
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("publish returned error: %v", err)
	}
}

func TestPublish_MeetingChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	hub.mu.Lock()
	ch := make(chan struct{}, 1)
	hub.subs["swe-1"] = append(hub.subs["swe-1"], ch)
	hub.mu.Unlock()

	ch <- struct{}{}

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
}

func TestMinimaxClient_Reason_NewRequestError(t *testing.T) {
	client := NewMinimaxClient("test")
    originalURL := minimaxAPIURL
	minimaxAPIURL = string([]byte{0x7f}) // Control character to fail http.NewRequestWithContext
	defer func() { minimaxAPIURL = originalURL }()

    _, err := client.Reason(context.Background(), "test")
	if err == nil {
		t.Fatalf("expected error from http.NewRequestWithContext")
	}
}

type mockStreamMessagesServerError struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockStreamMessagesServerError) Context() context.Context {
	return m.ctx
}

func (m *mockStreamMessagesServerError) Send(msg *pb.Message) error {
	return context.Canceled
}

func TestStreamMessages_SendErrorOnInitialSend(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	_ = hub.Publish(Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       EventTask,
		Content:    "Hello Streaming",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &mockStreamMessagesServerError{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{
		AgentId: "receiver",
	}.Build()

	err := server.StreamMessages(req, mockStream)
	if err == nil {
		t.Fatalf("expected error from Send(), got nil")
	}
}

func TestStreamMessages_ErrorOnLaterSend(t *testing.T) {
    hub := NewHub()
    hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
    hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
    server := NewHubServiceServer(hub)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    mockStream := &mockStreamMessagesServerError{ctx: ctx}
    req := pb.StreamMessagesRequest_builder{
        AgentId: "receiver",
    }.Build()

    errCh := make(chan error, 1)
    go func() {
        errCh <- server.StreamMessages(req, mockStream)
    }()

    time.Sleep(10 * time.Millisecond) // Let stream setup
    _ = hub.Publish(Message{
        ID:         "msg-2",
        FromAgent:  "sender",
        ToAgent:    "receiver",
        Type:       EventTask,
        Content:    "Hello Streaming",
        OccurredAt: time.Now(),
    })

    err := <-errCh
    if err == nil {
        t.Fatalf("expected error from Send(), got nil")
    }
}
