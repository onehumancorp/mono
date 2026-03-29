package orchestration

import (
	"github.com/onehumancorp/mono/srcs/domain"

	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type benchStreamMessagesServer struct {
	grpc.ServerStream
	ctx context.Context
	ch  chan *pb.Message
}

func (m *benchStreamMessagesServer) Context() context.Context {
	return m.ctx
}

func (m *benchStreamMessagesServer) Send(msg *pb.Message) error {
	select {
	case m.ch <- msg:
	default:
	}
	return nil
}

func BenchmarkStreamLatency(b *testing.B) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	hub.RegisterAgent(Agent{ID: "agent1", Status: StatusIdle})
	hub.RegisterAgent(Agent{ID: "agent2", Status: StatusIdle})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &benchStreamMessagesServer{
		ctx: ctx,
		ch:  make(chan *pb.Message, 100),
	}

	go func() {
		_ = srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: proto.String("agent2")}.Build(), stream)
	}()

	// wait for stream to start
	time.Sleep(10 * time.Millisecond)

	msg := domain.Message{
		ID:         "msg1",
		FromAgent:  "agent1",
		ToAgent:    "agent2",
		Type:       "test",
		Content:    "hello",
		OccurredAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hub.Publish(msg)
		<-stream.ch
	}
}

func BenchmarkPublish_Concurrent(b *testing.B) {
	hub := NewHub()
	numAgents := 100
	for i := 0; i < numAgents; i++ {
		hub.RegisterAgent(Agent{ID: fmt.Sprintf("agent%d", i), Status: StatusIdle})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		msg := domain.Message{
			ID:         "msg1",
			FromAgent:  "agent1",
			ToAgent:    "agent2",
			Type:       "test",
			Content:    "hello",
			OccurredAt: time.Now(),
		}
		for pb.Next() {
			_ = hub.Publish(msg)
		}
	})
}

func BenchmarkInbox(b *testing.B) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "agent1", Status: StatusIdle})

	msg := domain.Message{
		ID:         "msg1",
		FromAgent:  "agent1",
		ToAgent:    "agent1",
		Type:       "test",
		Content:    "hello",
		OccurredAt: time.Now(),
	}

	for i := 0; i < 1000; i++ {
		hub.Publish(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hub.Inbox("agent1")
	}
}

func BenchmarkReason(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"benchmark"}}]}`))
	}))
	defer ts.Close()

	originalURL := minimaxAPIURL
	minimaxAPIURL = ts.URL
	defer func() { minimaxAPIURL = originalURL }()

	client := NewMinimaxClient("test-key")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Reason(ctx, "test prompt")
	}
}
