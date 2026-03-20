package orchestration

import (
	"context"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
)

func BenchmarkStreamMessagesLatency(b *testing.B) {
	hub := NewHub()
	senderID := "sender"
	receiverID := "receiver"
	hub.RegisterAgent(Agent{ID: senderID, Name: "Sender", Role: "Test", OrganizationID: "Org"})
	hub.RegisterAgent(Agent{ID: receiverID, Name: "Receiver", Role: "Test", OrganizationID: "Org"})

	server := NewHubServiceServer(hub)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &mockStreamMessagesServer{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{AgentId: receiverID}.Build()

	go server.StreamMessages(req, mockStream)

	// Wait for the stream handler to initialize
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Publish(Message{
			ID:         "msg",
			FromAgent:  senderID,
			ToAgent:    receiverID,
			Type:       "test",
			Content:    "hello",
			OccurredAt: time.Now(),
		})

		// Instead of waiting with time.Sleep, we'll wait for the mock stream to receive the message
		for {
			mockStream.mu.Lock()
			count := len(mockStream.messages)
			mockStream.mu.Unlock()
			if count > i {
				break
			}
			time.Sleep(10 * time.Microsecond)
		}
	}
}
