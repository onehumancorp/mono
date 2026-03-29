package orchestration

import (
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

	msg := Message{
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
		msg := Message{
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

	msg := Message{
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

func BenchmarkGetPendingMissions(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := tmpDir + "/benchmark.db"
	db, err := NewSIPDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to create SIPDB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Seed database with many missions
	for i := 0; i < 1000; i++ {
		task := Message{
			ID:      fmt.Sprintf("mission-%d", i),
			Content: "Benchmark task",
			Type:    EventTask,
		}
		db.DelegateMission(ctx, fmt.Sprintf("mission-%d", i), "SOFTWARE_ENGINEER", task)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetPendingMissions(ctx, "SOFTWARE_ENGINEER")
		if err != nil {
			b.Fatalf("Error in GetPendingMissions: %v", err)
		}
	}
}
