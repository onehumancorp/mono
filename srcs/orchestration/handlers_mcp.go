package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// RegisterHubService HubServiceServer implements the gRPC HubService defined in hub.proto.
// Accepts parameters: s *grpc.Server (No Constraints), hub *Hub (No Constraints).
// Returns nothing.
// Produces no errors.
// Has no side effects.
func RegisterHubService(s *grpc.Server, hub *Hub) {
	pb.RegisterHubServiceServer(s, &HubServiceServer{hub: hub})
}

// HubServiceServer implements the gRPC interface for the orchestration Hub, facilitating remote agent registration and message streaming.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type HubServiceServer struct {
	pb.UnimplementedHubServiceServer
	hub *Hub
}

// NewHubServiceServer functionality.
// Accepts parameters: hub *Hub (No Constraints).
// Returns *HubServiceServer.
// Produces no errors.
// Has no side effects.
func NewHubServiceServer(hub *Hub) *HubServiceServer {
	return &HubServiceServer{hub: hub}
}

// RegisterAgent functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns (*pb.RegisterAgentResponse, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	agentReq := req.GetAgent()
	agent := Agent{
		ID:             agentReq.GetId(),
		Name:           agentReq.GetName(),
		Role:           agentReq.GetRole(),
		OrganizationID: agentReq.GetOrganizationId(),
		Status:         Status(agentReq.GetStatus()),
		ProviderType:   agentReq.GetProviderType(),
	}
	s.hub.RegisterAgent(agent)
	return pb.RegisterAgentResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

// OpenMeeting functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns (*pb.MeetingRoom, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) OpenMeeting(ctx context.Context, req *pb.OpenMeetingRequest) (*pb.MeetingRoom, error) {
	meeting := s.hub.OpenMeetingWithAgenda(req.GetMeetingId(), req.GetAgenda(), req.GetParticipants())
	return pb.MeetingRoom_builder{
		Id:           proto.String(meeting.ID),
		Agenda:       proto.String(meeting.Agenda),
		Participants: meeting.Participants,
	}.Build(), nil
}

// Publish functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns (*pb.PublishMessageResponse, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) Publish(ctx context.Context, req *pb.PublishMessageRequest) (*pb.PublishMessageResponse, error) {
	msgReq := req.GetMessage()
	msg := Message{
		ID:         msgReq.GetId(),
		FromAgent:  msgReq.GetFromAgent(),
		ToAgent:    msgReq.GetToAgent(),
		Type:       msgReq.GetType(),
		Content:    msgReq.GetContent(),
		MeetingID:  msgReq.GetMeetingId(),
		OccurredAt: time.Unix(msgReq.GetOccurredAtUnix(), 0),
	}
	if err := s.hub.Publish(msg); err != nil {
		return nil, status.Errorf(codes.Internal, "publish failed: %v", err)
	}
	return pb.PublishMessageResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

// DelegateTask functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns (*pb.DelegateTaskResponse, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) DelegateTask(ctx context.Context, req *pb.DelegateTaskRequest) (*pb.DelegateTaskResponse, error) {
	msgReq := req.GetTask()
	msg := Message{
		ID:         msgReq.GetId(),
		FromAgent:  msgReq.GetFromAgent(),
		ToAgent:    msgReq.GetToAgent(),
		Type:       msgReq.GetType(),
		Content:    msgReq.GetContent(),
		MeetingID:  msgReq.GetMeetingId(),
		OccurredAt: time.Unix(msgReq.GetOccurredAtUnix(), 0),
	}

	err := s.hub.DelegateTask(req.GetFromAgentId(), req.GetToAgentId(), msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delegate task failed: %v", err)
	}

	return pb.DelegateTaskResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

// StreamMessages functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) StreamMessages(req *pb.StreamMessagesRequest, stream pb.HubService_StreamMessagesServer) error {
	agentID := req.GetAgentId()

	// 1. Subscribe for new real-time messages to eliminate polling latency.
	// We must subscribe first to avoid missing any messages published between
	// draining the inbox and subscribing.
	ch, unsubscribe := s.hub.Subscribe(agentID)
	defer unsubscribe()

	sendNewMessages := func() error {
		msgs := s.hub.Inbox(agentID)
		if len(msgs) > 0 {
			defer putMessageSlice(msgs)
		}
		for _, m := range msgs {
			if err := stream.Send(pb.Message_builder{
				Id:             proto.String(m.ID),
				FromAgent:      proto.String(m.FromAgent),
				ToAgent:        proto.String(m.ToAgent),
				Type:           proto.String(m.Type),
				Content:        proto.String(m.Content),
				MeetingId:      proto.String(m.MeetingID),
				OccurredAtUnix: proto.Int64(m.OccurredAt.Unix()),
			}.Build()); err != nil {
				return err
			}
		}
		return nil
	}

	// 2. Drain any pre-existing messages in the inbox.
	if err := sendNewMessages(); err != nil {
		return err
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ch:
			if err := sendNewMessages(); err != nil {
				return err
			}
		}
	}
}

// Reason functionality.
// Accepts parameters: s *HubServiceServer (No Constraints).
// Returns (*pb.ReasonResponse, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (s *HubServiceServer) Reason(ctx context.Context, req *pb.ReasonRequest) (*pb.ReasonResponse, error) {
	client := NewMinimaxClient(s.hub.MinimaxAPIKey())
	content, err := client.Reason(ctx, req.GetPrompt())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "minimax reasoning failed: %v", err)
	}
	return pb.ReasonResponse_builder{Content: proto.String(content)}.Build(), nil
}

// minimaxAPIURL is the endpoint for Minimax reasoning.
// ⚡ BOLT: [Configurable endpoint] - Randomized Selection from Top 5
var minimaxAPIURL = "https://api.minimax.io/v1/chat/completions"

// MinimaxClient handles interaction with the Minimax Model 2.7.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type MinimaxClient struct {
	APIKey string
}

// NewMinimaxClient functionality.
// Accepts parameters: apiKey string (No Constraints).
// Returns *MinimaxClient.
// Produces no errors.
// Has no side effects.
func NewMinimaxClient(apiKey string) *MinimaxClient {
	return &MinimaxClient{APIKey: apiKey}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// Reason functionality.
// Accepts parameters: c *MinimaxClient (No Constraints).
// Returns (string, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (c *MinimaxClient) Reason(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("minimax API key is not configured")
	}

	url := minimaxAPIURL
	// Optimization: construct the JSON payload manually to avoid
	// maps and slices allocations.
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(`{"model":"MiniMax-M2.7","messages":[{"role":"user","content":`)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(prompt)
	// Encode adds a newline, so we slice it off and add the closing brackets
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(`}]}`)

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// ⚡ BOLT: [Reused HTTP Client] - Randomized Selection from Top 5
	// Prevents severe connection and resource leaks by reusing connection pools on every request.
	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("minimax API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("empty response from minimax")
	}

	return result.Choices[0].Message.Content, nil
}

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

	// VRAM Quota Enforcement: Hard limit at 10 active agents across the hub
	if currentAgents >= 10 {
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
	if strings.Contains(instruction, "SYSTEM:") || strings.Contains(instruction, "\n\n") {
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
