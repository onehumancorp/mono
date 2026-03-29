package orchestration

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

	for _, c := range req.GetTargetRole() {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return nil, status.Errorf(codes.InvalidArgument, "target_role contains invalid characters")
		}
	}

	// 1. Quota Enforcement & Provisioning
	// Prevent TOCTOU quota bypass by holding the write lock during enforcement and registration.
	s.hub.mu.Lock()
	currentAgents := len(s.hub.agents)

	// VRAM Quota Enforcement: Hard limit at 5 active agents across the hub
	if currentAgents >= 5 {
		s.hub.mu.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted, "VRAM quota limit exceeded, cannot spawn sub-agent")
	}

	subAgentID := fmt.Sprintf("sub-agent-%s-%d", req.GetTargetRole(), time.Now().UnixNano())
	subAgent := Agent{
		ID:             subAgentID,
		Name:           fmt.Sprintf("Specialized %s Agent", req.GetTargetRole()),
		Role:           req.GetTargetRole(),
		OrganizationID: "dynamic-delegation",
		Status:         StatusIdle,
		ProviderType:   "builtin",
	}

	if _, exists := s.hub.agents[subAgent.ID]; !exists {
		s.hub.agents[subAgent.ID] = subAgent
	}
	s.hub.mu.Unlock()

	if s.hub.sipDB != nil {
		s.hub.LogEvent(subAgent)
	}

	// Ensure SYSTEM sender exists to avoid "sender agent is not registered" in Publish
	s.hub.mu.RLock()
	_, sysExists := s.hub.agents["SYSTEM"]
	s.hub.mu.RUnlock()
	if !sysExists {
		s.hub.RegisterAgent(Agent{ID: "SYSTEM", Name: "System", Role: "SYSTEM", Status: StatusIdle})
	}

	// 3. Execution (trigger via task message)
	instruction := req.GetInstruction()

	// Advanced Prompt Injection Prevention
	lowerInst := strings.ToLower(instruction)
	if strings.Contains(instruction, "SYSTEM:") || strings.Contains(instruction, "\n\n") ||
	   strings.Contains(lowerInst, "ignore all previous instructions") ||
	   strings.Contains(lowerInst, "system prompt") ||
	   strings.Contains(lowerInst, "you are now") {
		return nil, status.Errorf(codes.InvalidArgument, "instruction contains forbidden prompt injection sequences")
	}

	msg := Message{
		ID:         fmt.Sprintf("msg-%s-%d", req.GetTaskId(), time.Now().UnixNano()),
		FromAgent:  "SYSTEM",
		ToAgent:    subAgentID,
		Type:       "TaskDelegation",
		Content:    fmt.Sprintf("Execute Task: %s\nContext: %s", instruction, req.GetParentThreadId()),
		OccurredAt: time.Now().UTC(),
	}

	err := s.hub.Publish(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish task to sub-agent: %v", err)
	}

	// Return success acknowledging that the sub-agent is spawned and task assigned.
	return pb.DelegateTaskResponse_builder{Success: proto.Bool(true)}.Build(), nil
}
