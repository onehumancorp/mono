package orchestration

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc/metadata"
)

// mockStreamServer implements pb.HubService_StreamMessagesServer for benchmarks.
type mockStreamServer struct {
	ctx      context.Context
	messages []*pb.Message
}

func (m *mockStreamServer) Context() context.Context { return m.ctx }
func (m *mockStreamServer) Send(msg *pb.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}
func (m *mockStreamServer) SendHeader(metadata.MD) error { return nil }
func (m *mockStreamServer) SetTrailer(metadata.MD)       {}
func (m *mockStreamServer) SetHeader(metadata.MD) error  { return nil }
func (m *mockStreamServer) SendMsg(m_ interface{}) error                  { return nil }
func (m *mockStreamServer) RecvMsg(m_ interface{}) error                  { return nil }

func BenchmarkStreamMessagesLatency(b *testing.B) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender"})
	hub.RegisterAgent(Agent{ID: "receiver"})
	srv := NewHubServiceServer(hub)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		stream := &mockStreamServer{ctx: ctx, messages: make([]*pb.Message, 0)}

		done := make(chan struct{})
		go func() {
			srv.StreamMessages(pb.StreamMessagesRequest_builder{AgentId: "receiver"}.Build(), stream)
			close(done)
		}()

		// Simulate message published after stream starts
		time.Sleep(10 * time.Millisecond)
		hub.Publish(Message{
			ID:        fmt.Sprintf("msg-%d", i),
			FromAgent: "sender",
			ToAgent:   "receiver",
			Content:   "bench",
		})

		// Wait for message to be received by stream (mocked by sleep in our benchmark case)
		// For a real latency benchmark, we measure time between publish and receive.
		start := time.Now()
		for len(stream.messages) == 0 {
			time.Sleep(1 * time.Millisecond)
		}
		latency := time.Since(start)
		b.ReportMetric(float64(latency.Milliseconds()), "ms/msg")

		cancel()
		<-done
	}
}

func BenchmarkAgentsSlice(b *testing.B) {
	hub := NewHub()
	for i := 0; i < 1000; i++ {
		hub.RegisterAgent(Agent{ID: fmt.Sprintf("agent-%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Agents()
	}
}

func BenchmarkReason(b *testing.B) {
	b.Skip("Skipping network benchmark unless needed")
}
