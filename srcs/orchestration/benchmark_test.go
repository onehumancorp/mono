package orchestration

import (
	"fmt"
	"testing"
)

func BenchmarkHubAgents(b *testing.B) {
	hub := NewHub()
	for i := 0; i < 1000; i++ {
		hub.RegisterAgent(Agent{ID: fmt.Sprintf("agent-%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agents := hub.Agents()
		ReleaseAgents(agents)
	}
}

func BenchmarkHubPublish(b *testing.B) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "agent-1"})
	hub.RegisterAgent(Agent{ID: "agent-2"})

	msg := Message{
		ID:        "msg-1",
		FromAgent: "agent-1",
		ToAgent:   "agent-2",
		Content:   "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Publish(msg)
	}
}
