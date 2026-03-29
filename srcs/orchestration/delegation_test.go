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
	"google.golang.org/protobuf/proto"
)

func TestDelegateSubTask_Success(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender-1", Name: "Sender", Role: "PM", Status: StatusIdle})
	server := NewHubServiceServer(hub)
	ctx := context.Background()

	req := pb.SubTask_builder{
		TaskId:         proto.String("task-1"),
		TargetRole:     proto.String("SWE"),
		Instruction:    proto.String("Implement login component"),
		ParentThreadId: proto.String("thread-1"),
		FromAgentId:    proto.String("sender-1"),
	}.Build()

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
		if strings.HasPrefix(id, "sub-agent-SWE-") {
			subAgentID = id
			break
		}
	}

	if subAgentID == "" {
		t.Fatalf("expected sub-agent to be created")
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

	req := pb.SubTask_builder{
		TaskId:         proto.String("task-2"),
		TargetRole:     proto.String("QA"),
		Instruction:    proto.String("Test login component"),
		ParentThreadId: proto.String("thread-1"),
	}.Build()

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
			req: pb.SubTask_builder{
				TargetRole:  proto.String("SWE"),
				Instruction: proto.String("Impl"),
			}.Build(),
		},
		{
			name: "missing target_role",
			req: pb.SubTask_builder{
				TaskId:      proto.String("task-3"),
				Instruction: proto.String("Impl"),
			}.Build(),
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

	req := pb.SubTask_builder{
		TaskId:         proto.String("task-int-1"),
		TargetRole:     proto.String("QA"),
		Instruction:    proto.String("Verify real data integration"),
		ParentThreadId: proto.String("thread-int-1"),
		FromAgentId:    proto.String("sender-1"),
	}.Build()

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
		if strings.HasPrefix(id, "sub-agent-QA-") {
			subAgentID = id
			break
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
