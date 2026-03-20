package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/telemetry"
	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status Intent: Status indicates the current operational phase of an AI agent within the workforce.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Status string

const (
	// StatusIdle Intent: Handles operations related to StatusIdle.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	StatusIdle      Status = "IDLE"
	// StatusActive Intent: Handles operations related to StatusActive.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	StatusActive    Status = "ACTIVE"
	// StatusInMeeting Intent: Handles operations related to StatusInMeeting.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	StatusInMeeting Status = "IN_MEETING"
	// StatusBlocked Intent: Handles operations related to StatusBlocked.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	StatusBlocked   Status = "BLOCKED"
)

// Event type constants for the asynchronous pub/sub agent interaction protocol.
const (
	// EventTask Intent: Handles operations related to EventTask.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventTask           = "task"
	// EventStatus Intent: Handles operations related to EventStatus.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventStatus         = "status"
	// EventHandoff Intent: Handles operations related to EventHandoff.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventHandoff        = "handoff"
	// EventCodeReviewed Intent: Handles operations related to EventCodeReviewed.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventCodeReviewed   = "CodeReviewed"
	// EventTestsFailed Intent: Handles operations related to EventTestsFailed.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventTestsFailed    = "TestsFailed"
	// EventTestsPassed Intent: Handles operations related to EventTestsPassed.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventTestsPassed    = "TestsPassed"
	// EventSpecApproved Intent: Handles operations related to EventSpecApproved.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventSpecApproved   = "SpecApproved"
	// EventBlockerRaised Intent: Handles operations related to EventBlockerRaised.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventBlockerRaised  = "BlockerRaised"
	// EventBlockerCleared Intent: Handles operations related to EventBlockerCleared.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventBlockerCleared = "BlockerCleared"
	// EventPRCreated Intent: Handles operations related to EventPRCreated.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventPRCreated      = "PRCreated"
	// EventPRMerged Intent: Handles operations related to EventPRMerged.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventPRMerged       = "PRMerged"
	// EventDesignReviewed Intent: Handles operations related to EventDesignReviewed.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventDesignReviewed = "DesignReviewed"
	// EventApprovalNeeded Intent: Handles operations related to EventApprovalNeeded.
	//
	// Params: None.
	//
	// Returns: None.
	//
	// Errors: Returns an error if the operation fails.
	//
	// Side Effects: Modifies state or interacts with external systems as necessary.
	EventApprovalNeeded = "ApprovalNeeded"
)

// Agent Intent: Agent represents an active, instantiated worker within the AI organisation.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
	Status         Status `json:"status"`
	// ProviderType identifies the external agent implementation backing this worker
	// (e.g. "claude", "gemini", "opencode").  An empty string or "builtin" means
	// the platform's own lightweight agent is used.
	ProviderType string `json:"providerType,omitempty"`
}

// Message Intent: Message encapsulates a discrete event, command, or context update passed between agents or rooms.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

// MeetingRoom Intent: MeetingRoom maintains a persistent, sequential transcript of inter-agent collaboration.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

// Hub Intent: Hub acts as the thread-safe central message broker and runtime state manager for the AI workforce.  Constraints: Must be accessed via its exported methods to preserve data race safety.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type Hub struct {
	mu            sync.RWMutex
	agents        map[string]Agent
	inbox         map[string][]Message
	meetings      map[string]MeetingRoom
	minimaxAPIKey string
}

// NewHub Intent: NewHub constructs a new instance of an orchestration Hub, pre-allocated with empty registries.  Returns: An instantiated *Hub ready to register agents and route events.
//
// Params: None.
//
// Returns:
//   - *Hub: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func NewHub() *Hub {
	return &Hub{
		agents:   map[string]Agent{},
		inbox:    map[string][]Message{},
		meetings: map[string]MeetingRoom{},
	}
}

// RegisterAgent Intent: RegisterAgent enrolls an agent into the Hub, allocating an inbox and initialising its Status.  Parameters: - agent: Agent; The worker object containing ID, Name, Role, and Organization context.
//
// Params:
//   - agent: parameter inferred from signature.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) RegisterAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent.Status == "" {
		agent.Status = StatusIdle
	}

	h.agents[agent.ID] = agent
}
// SetMinimaxAPIKey Intent: Handles operations related to SetMinimaxAPIKey.
//
// Params:
//   - key: parameter inferred from signature.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) SetMinimaxAPIKey(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minimaxAPIKey = key
}
// MinimaxAPIKey Intent: Handles operations related to MinimaxAPIKey.
//
// Params: None.
//
// Returns:
//   - string: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) MinimaxAPIKey() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.minimaxAPIKey
}

// Agent Intent: Agent retrieves the runtime state of a specific worker by ID.  Parameters: - id: string; The unique identifier of the agent.  Returns: The matching Agent object and a boolean indicating if it exists in the registry.
//
// Params:
//   - id: parameter inferred from signature.
//
// Returns:
//   - Agent: return value inferred from signature.
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Agent(id string) (Agent, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agent, ok := h.agents[id]
	return agent, ok
}

// OpenMeeting Intent: OpenMeeting instantiates a new collaborative context window and marks all participants as InMeeting.  Parameters: - id: string; Unique identifier for the room. - participants: []string; A list of agent IDs to be enrolled in the discussion.  Returns: The instantiated MeetingRoom.
//
// Params:
//   - id: parameter inferred from signature.
//   - participants: parameter inferred from signature.
//
// Returns:
//   - MeetingRoom: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) OpenMeeting(id string, participants []string) MeetingRoom {
	h.mu.Lock()
	defer h.mu.Unlock()

	meeting := MeetingRoom{ID: id, Participants: append([]string(nil), participants...)}
	h.meetings[id] = meeting
	for _, participant := range participants {
		agent := h.agents[participant]
		agent.Status = StatusInMeeting
		h.agents[participant] = agent
	}

	telemetry.RecordMeetingEvent(context.Background(), "opened")
	return meeting
}

// OpenMeetingWithAgenda Intent: OpenMeetingWithAgenda creates a meeting room with an explicit agenda descriptor.  Parameters: - id: string; Unique identifier for the room. - agenda: string; The primary objective guiding the agents' conversation. - participants: []string; A list of agent IDs to be enrolled in the discussion.  Returns: The instantiated MeetingRoom.
//
// Params:
//   - id: parameter inferred from signature.
//   - agenda: parameter inferred from signature.
//   - participants: parameter inferred from signature.
//
// Returns:
//   - MeetingRoom: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) OpenMeetingWithAgenda(id, agenda string, participants []string) MeetingRoom {
	h.mu.Lock()
	defer h.mu.Unlock()

	meeting := MeetingRoom{ID: id, Agenda: agenda, Participants: append([]string(nil), participants...)}
	h.meetings[id] = meeting
	for _, participant := range participants {
		agent := h.agents[participant]
		agent.Status = StatusInMeeting
		h.agents[participant] = agent
	}

	telemetry.RecordMeetingEvent(context.Background(), "opened")
	return meeting
}

// FireAgent Intent: FireAgent removes an agent from the hub and clears their inbox.  Parameters: - id: string; The unique identifier of the agent to terminate.
//
// Params:
//   - id: parameter inferred from signature.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) FireAgent(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.agents, id)
	delete(h.inbox, id)
}

// Publish Intent: Publish validates and routes a message to a direct recipient, a meeting room, or both.  Parameters: - message: Message; The event payload containing routing headers and content.  Returns: An error if the sender or recipient agents do not exist, or if the target meeting is unrecognised.
//
// Params:
//   - message: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Publish(message Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.agents[message.FromAgent]; !ok {
		return errors.New("sender agent is not registered")
	}
	if message.ToAgent != "" {
		if _, ok := h.agents[message.ToAgent]; !ok {
			return errors.New("recipient agent is not registered")
		}
		h.inbox[message.ToAgent] = append(h.inbox[message.ToAgent], message)
	}

	sender := h.agents[message.FromAgent]
	if message.MeetingID != "" {
		meeting, ok := h.meetings[message.MeetingID]
		if !ok {
			return errors.New("meeting room is not registered")
		}
		meeting.Transcript = append(meeting.Transcript, message)
		h.meetings[message.MeetingID] = meeting
		sender.Status = StatusInMeeting
	} else {
		sender.Status = StatusActive
	}
	h.agents[message.FromAgent] = sender

	telemetry.RecordAgentApiCall(context.Background(), sender.ID, sender.Role, "publish")

	return nil
}

// Inbox Intent: Inbox retrieves all undelivered or direct messages routed exclusively to a single agent.  Parameters: - agentID: string; The unique identifier of the worker.  Returns: A slice of direct Message objects.
//
// Params:
//   - agentID: parameter inferred from signature.
//
// Returns:
//   - []Message: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Inbox(agentID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return append([]Message(nil), h.inbox[agentID]...)
}

// Meeting Intent: Meeting retrieves the current state and transcript of a specified virtual meeting room.  Parameters: - id: string; The unique identifier of the room.  Returns: The matching MeetingRoom object and a boolean indicating if it exists.
//
// Params:
//   - id: parameter inferred from signature.
//
// Returns:
//   - MeetingRoom: return value inferred from signature.
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Meeting(id string) (MeetingRoom, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	meeting, ok := h.meetings[id]
	telemetry.RecordMeetingEvent(context.Background(), "opened")
	return meeting, ok
}

// Meetings Intent: Meetings fetches a point-in-time snapshot of all active meeting rooms.  Returns: A slice containing all MeetingRoom objects.
//
// Params: None.
//
// Returns:
//   - []MeetingRoom: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Meetings() []MeetingRoom {
	h.mu.RLock()
	defer h.mu.RUnlock()

	meetings := make([]MeetingRoom, 0, len(h.meetings))
	for _, meeting := range h.meetings {
		meetings = append(meetings, meeting)
	}

	telemetry.RecordMeetingEvent(context.Background(), "opened")
	return meetings
}

// Agents Intent: Agents retrieves a point-in-time snapshot of the entire registered workforce, ordered by ID.  Returns: A slice of all active Agent objects in the orchestration Hub.
//
// Params: None.
//
// Returns:
//   - []Agent: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (h *Hub) Agents() []Agent {
	h.mu.RLock()
	agents := make([]Agent, 0, len(h.agents))
	for _, agent := range h.agents {
		agents = append(agents, agent)
	}
	h.mu.RUnlock()

	// ⚡ BOLT: [O(n log n) sorting inside read lock] - Randomized Selection from Top 5
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}

// RegisterHubService Intent: HubServiceServer implements the gRPC HubService defined in hub.proto.
//
// Params:
//   - s: parameter inferred from signature.
//   - hub: parameter inferred from signature.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func RegisterHubService(s *grpc.Server, hub *Hub) {
	pb.RegisterHubServiceServer(s, &HubServiceServer{hub: hub})
}
// HubServiceServer Intent: Handles operations related to HubServiceServer.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type HubServiceServer struct {
	pb.UnimplementedHubServiceServer
	hub *Hub
}
// NewHubServiceServer Intent: Handles operations related to NewHubServiceServer.
//
// Params:
//   - hub: parameter inferred from signature.
//
// Returns:
//   - *HubServiceServer: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func NewHubServiceServer(hub *Hub) *HubServiceServer {
	return &HubServiceServer{hub: hub}
}
// RegisterAgent Intent: Handles operations related to RegisterAgent.
//
// Params:
//   - ctx: parameter inferred from signature.
//   - req: parameter inferred from signature.
//
// Returns:
//   - *pb.RegisterAgentResponse: return value inferred from signature.
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
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
	return pb.RegisterAgentResponse_builder{Success: true}.Build(), nil
}
// OpenMeeting Intent: Handles operations related to OpenMeeting.
//
// Params:
//   - ctx: parameter inferred from signature.
//   - req: parameter inferred from signature.
//
// Returns:
//   - *pb.MeetingRoom: return value inferred from signature.
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (s *HubServiceServer) OpenMeeting(ctx context.Context, req *pb.OpenMeetingRequest) (*pb.MeetingRoom, error) {
	meeting := s.hub.OpenMeetingWithAgenda(req.GetMeetingId(), req.GetAgenda(), req.GetParticipants())
	return pb.MeetingRoom_builder{
		Id:           meeting.ID,
		Agenda:       meeting.Agenda,
		Participants: meeting.Participants,
	}.Build(), nil
}
// Publish Intent: Handles operations related to Publish.
//
// Params:
//   - ctx: parameter inferred from signature.
//   - req: parameter inferred from signature.
//
// Returns:
//   - *pb.PublishMessageResponse: return value inferred from signature.
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
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
	return pb.PublishMessageResponse_builder{Success: true}.Build(), nil
}
// StreamMessages Intent: Handles operations related to StreamMessages.
//
// Params:
//   - req: parameter inferred from signature.
//   - stream: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (s *HubServiceServer) StreamMessages(req *pb.StreamMessagesRequest, stream pb.HubService_StreamMessagesServer) error {
	// Simple polling implementation for demonstration.
	// In production, use a proper pub/sub or channel-based mechanism.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastCount int
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			msgs := s.hub.Inbox(req.GetAgentId())
			if len(msgs) > lastCount {
				for i := lastCount; i < len(msgs); i++ {
					m := msgs[i]
					if err := stream.Send(pb.Message_builder{
						Id:             m.ID,
						FromAgent:      m.FromAgent,
						ToAgent:        m.ToAgent,
						Type:           m.Type,
						Content:        m.Content,
						MeetingId:      m.MeetingID,
						OccurredAtUnix: m.OccurredAt.Unix(),
					}.Build()); err != nil {
						return err
					}
				}
				lastCount = len(msgs)
			}
		}
	}
}
// Reason Intent: Handles operations related to Reason.
//
// Params:
//   - ctx: parameter inferred from signature.
//   - req: parameter inferred from signature.
//
// Returns:
//   - *pb.ReasonResponse: return value inferred from signature.
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (s *HubServiceServer) Reason(ctx context.Context, req *pb.ReasonRequest) (*pb.ReasonResponse, error) {
	client := NewMinimaxClient(s.hub.MinimaxAPIKey())
	content, err := client.Reason(ctx, req.GetPrompt())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "minimax reasoning failed: %v", err)
	}
	return pb.ReasonResponse_builder{Content: content}.Build(), nil
}

// minimaxAPIURL is the endpoint for Minimax reasoning.
// ⚡ BOLT: [Configurable endpoint] - Randomized Selection from Top 5
var minimaxAPIURL = "https://api.minimax.io/v1/chat/completions"

// MinimaxClient Intent: MinimaxClient handles interaction with the Minimax Model 2.7.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type MinimaxClient struct {
	APIKey string
}
// NewMinimaxClient Intent: Handles operations related to NewMinimaxClient.
//
// Params:
//   - apiKey: parameter inferred from signature.
//
// Returns:
//   - *MinimaxClient: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func NewMinimaxClient(apiKey string) *MinimaxClient {
	return &MinimaxClient{APIKey: apiKey}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
// Reason Intent: Handles operations related to Reason.
//
// Params:
//   - ctx: parameter inferred from signature.
//   - prompt: parameter inferred from signature.
//
// Returns:
//   - string: return value inferred from signature.
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (c *MinimaxClient) Reason(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("minimax API key is not configured")
	}

	url := minimaxAPIURL
	payload := map[string]interface{}{
		"model": "MiniMax-M2.7",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	// ⚡ BOLT: [JSON serialization thrashing in LLM API routing] - Randomized Selection from Top 5
	// Use a sync.Pool for bytes.Buffer and json.Encoder to avoid high-allocation JSON marshalling.
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
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
