package orchestration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDelegateSubTask_Success(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender-1", Name: "Sender", Role: "PM", Status: StatusIdle})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	req := &pb.SubTask{
		TaskId:         "task-1",
		TargetRole:     "SWE",
		Instruction:    "Implement login component",
		ParentThreadId: "thread-1",
		FromAgentId:    "sender-1",
	}

	resp, err := server.DelegateSubTask(ctx, req)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatalf("expected success to be true")
	}

	// Verify that the sub-agent was registered and received the message.
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	var subAgentID string
	for id := range hub.agents {
		if id != "SYSTEM" {
			subAgentID = id
		}
	}

	if !strings.HasPrefix(subAgentID, "sub-agent-SWE-") {
		t.Fatalf("unexpected agent ID: %s", subAgentID)
	}

	msgs := hub.inbox[subAgentID]
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in inbox, got %d", len(msgs))
	}

	if msgs[0].Type != "TaskDelegation" {
		t.Fatalf("expected message type TaskDelegation, got %s", msgs[0].Type)
	}
	if !strings.Contains(msgs[0].Content, "Implement login component") {
		t.Fatalf("expected instruction in message content, got %s", msgs[0].Content)
	}
}

func TestDelegateSubTask_QuotaExhaustion(t *testing.T) {
	hub := NewHub()
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	// Fill the hub to reach the quota limit (10)
	for i := 0; i < 10; i++ {
		hub.RegisterAgent(Agent{
			ID:             fmt.Sprintf("filler-%d", i),
			Name:           "Filler Agent",
			Role:           "FILLER",
			OrganizationID: "org-1",
			Status:         StatusIdle,
		})
	}

	req := &pb.SubTask{
		TaskId:         "task-2",
		TargetRole:     "QA",
		Instruction:    "Test login component",
		ParentThreadId: "thread-1",
	}

	_, err := server.DelegateSubTask(ctx, req)
	if err == nil {
		t.Fatalf("expected quota exhaustion error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted code, got %v", st.Code())
	}
}

func TestDelegateSubTask_MissingFields(t *testing.T) {
	hub := NewHub()
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *pb.SubTask
	}{
		{
			name: "missing task_id",
			req: &pb.SubTask{
				TargetRole:  "SWE",
				Instruction: "Impl",
			},
		},
		{
			name: "missing target_role",
			req: &pb.SubTask{
				TaskId:      "task-3",
				Instruction: "Impl",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.DelegateSubTask(ctx, tt.req)
			if err == nil {
				t.Fatalf("expected error for missing fields, got nil")
			}
			st, _ := status.FromError(err)
			if st.Code() != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument code, got %v", st.Code())
			}
		})
	}
}

// TestDelegateSubTask_Integration checks the real data law by seeing if the message gets processed properly
func TestDelegateSubTask_Integration(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender-1", Name: "Sender", Role: "PM", Status: StatusIdle})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	req := &pb.SubTask{
		TaskId:         "task-int-1",
		TargetRole:     "QA",
		Instruction:    "Verify real data integration",
		ParentThreadId: "thread-int-1",
		FromAgentId:    "sender-1",
	}

	_, err := server.DelegateSubTask(ctx, req)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify agent properties
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	var subAgentID string
	for id := range hub.agents {
		if id != "SYSTEM" {
			subAgentID = id
		}
	}

	agent, exists := hub.agents[subAgentID]
	if !exists {
		t.Fatalf("agent does not exist")
	}

	if agent.ProviderType != "builtin" {
		t.Fatalf("expected ProviderType builtin, got %s", agent.ProviderType)
	}
	if agent.Status != StatusIdle {
		t.Fatalf("expected StatusIdle, got %s", agent.Status)
	}
}
