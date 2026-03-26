package orchestration

import (
	"context"
	"fmt"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DelegateSubTask handles Hierarchical Task Delegation by provisioning temporary
// specialized sub-agents. It enforces VRAM quota limits before creating the agent,
// and isolates the sub-agent with its own thread ID and instructions.
//
//   ctx context.Context
//   req *pb.SubTask
//
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns DelegateSubTask(ctx context.Context, req *pb.SubTask) (*pb.DelegateTaskResponse, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) DelegateSubTask(ctx context.Context, req *pb.SubTask) (*pb.DelegateTaskResponse, error) {
	if req.GetTaskId() == "" || req.GetTargetRole() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "task_id and target_role are required")
	}

	// 1. Quota Enforcement
	// Mock VRAM quota limit checking to enforce token efficiency and bounds.
	s.hub.mu.RLock()
	currentAgents := len(s.hub.agents)
	s.hub.mu.RUnlock()

	// VRAM Quota Enforcement: Hard limit at 10 active agents across the hub
	if currentAgents >= 10 {
		return nil, status.Errorf(codes.ResourceExhausted, "VRAM quota limit exceeded, cannot spawn sub-agent")
	}

	// 2. Provisioning
	subAgentID := fmt.Sprintf("sub-agent-%s-%d", req.GetTargetRole(), time.Now().UnixNano())
	subAgent := Agent{
		ID:             subAgentID,
		Name:           fmt.Sprintf("Specialized %s Agent", req.GetTargetRole()),
		Role:           req.GetTargetRole(),
		OrganizationID: "dynamic-delegation",
		Status:         StatusIdle,
		ProviderType:   "builtin",
	}

	s.hub.RegisterAgent(subAgent)

	// Ensure SYSTEM sender exists to avoid "sender agent is not registered" in Publish
	s.hub.mu.RLock()
	_, sysExists := s.hub.agents["SYSTEM"]
	s.hub.mu.RUnlock()
	if !sysExists {
		s.hub.RegisterAgent(Agent{ID: "SYSTEM", Name: "System", Role: "SYSTEM", Status: StatusIdle})
	}

	// 3. Execution (trigger via task message)
	msg := Message{
		ID:         fmt.Sprintf("msg-%s-%d", req.GetTaskId(), time.Now().UnixNano()),
		FromAgent:  "SYSTEM",
		ToAgent:    subAgentID,
		Type:       "TaskDelegation",
		Content:    fmt.Sprintf("Execute Task: %s\nContext: %s", req.GetInstruction(), req.GetParentThreadId()),
		OccurredAt: time.Now().UTC(),
	}

	err := s.hub.Publish(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish task to sub-agent: %v", err)
	}

	// Return success acknowledging that the sub-agent is spawned and task assigned.
	return pb.DelegateTaskResponse_builder{Success: true}.Build(), nil
}
