package orchestration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
)

type mockStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) Send(msg *pb.Message) error {
	return nil
}

func TestHubServiceServer(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	// RegisterAgent
	_, _ = srv.RegisterAgent(context.Background(), pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id:             "test-agent",
			Name:           "Test Agent",
			Role:           "Tester",
			OrganizationId: "org1",
			Status:         string(StatusIdle),
		}.Build(),
	}.Build())

	// OpenMeeting
	_, _ = srv.OpenMeeting(context.Background(), pb.OpenMeetingRequest_builder{
		MeetingId:    "m1",
		Agenda:       "Test",
		Participants: []string{"test-agent"},
	}.Build())

	// Publish
	_, _ = srv.Publish(context.Background(), pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			Id:             "msg1",
			FromAgent:      "test-agent",
			ToAgent:        "test-agent",
			Type:           "test",
			Content:        "hello",
			MeetingId:      "m1",
			OccurredAtUnix: time.Now().Unix(),
		}.Build(),
	}.Build())

    // Publish error
	_, _ = srv.Publish(context.Background(), pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			Id:             "msg2",
			FromAgent:      "not-found",
		}.Build(),
	}.Build())

	// StreamMessages
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: "test-agent"}.Build(), &mockStream{ctx: ctx})

	// Reason
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer ts.Close()

	hub.SetMinimaxAPIKey("test-key")

    old := MinimaxAPIURL
    defer func() { MinimaxAPIURL = old }()
    MinimaxAPIURL = ts.URL

    client := NewMinimaxClient("test-key")
    _, _ = srv.Reason(context.Background(), pb.ReasonRequest_builder{Prompt: "test"}.Build())
    _, _ = client.Reason(context.Background(), "test")
}

func TestMinimaxClientErrors(t *testing.T) {
    client := NewMinimaxClient("")
    _, _ = client.Reason(context.Background(), "test")
}

func TestMinimaxClientReasonSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer ts.Close()

    old := MinimaxAPIURL
    defer func() { MinimaxAPIURL = old }()
    MinimaxAPIURL = ts.URL

    client := NewMinimaxClient("test-key")
    res, err := client.Reason(context.Background(), "test")
    if err != nil {
        t.Fatalf("unexpected err: %v", err)
    }
    if res != "ok" {
        t.Errorf("expected ok, got %v", res)
    }
}

func TestMinimaxClientReasonError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

    old := MinimaxAPIURL
    defer func() { MinimaxAPIURL = old }()
    MinimaxAPIURL = ts.URL

    client := NewMinimaxClient("test-key")
    _, err := client.Reason(context.Background(), "test")
    if err == nil {
        t.Fatal("expected err")
    }
}

func TestMinimaxClientReasonEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer ts.Close()

    old := MinimaxAPIURL
    defer func() { MinimaxAPIURL = old }()
    MinimaxAPIURL = ts.URL

    client := NewMinimaxClient("test-key")
    _, err := client.Reason(context.Background(), "test")
    if err == nil {
        t.Fatal("expected err")
    }
}

func TestMinimaxClientReasonBadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{bad}`))
	}))
	defer ts.Close()

    old := MinimaxAPIURL
    defer func() { MinimaxAPIURL = old }()
    MinimaxAPIURL = ts.URL

    client := NewMinimaxClient("test-key")
    _, err := client.Reason(context.Background(), "test")
    if err == nil {
        t.Fatal("expected err")
    }
}
