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

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"github.com/onehumancorp/mono/srcs/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Summary: Status indicates the current operational phase of an AI agent within the workforce.
// Intent: Status indicates the current operational phase of an AI agent within the workforce.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Status string

const (
	// Summary: Defines the StatusIdle type.
	// Intent: Defines the StatusIdle type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusIdle      Status = "IDLE"
	// Summary: Defines the StatusActive type.
	// Intent: Defines the StatusActive type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusActive    Status = "ACTIVE"
	// Summary: Defines the StatusInMeeting type.
	// Intent: Defines the StatusInMeeting type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusInMeeting Status = "IN_MEETING"
	// Summary: Defines the StatusBlocked type.
	// Intent: Defines the StatusBlocked type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusBlocked   Status = "BLOCKED"
)

// Event type constants for the asynchronous pub/sub agent interaction protocol.
const (
	// Summary: Defines the EventTask type.
	// Intent: Defines the EventTask type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventTask           = "task"
	// Summary: Defines the EventStatus type.
	// Intent: Defines the EventStatus type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventStatus         = "status"
	// Summary: Defines the EventHandoff type.
	// Intent: Defines the EventHandoff type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventHandoff        = "handoff"
	// Summary: Defines the EventCodeReviewed type.
	// Intent: Defines the EventCodeReviewed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventCodeReviewed   = "CodeReviewed"
	// Summary: Defines the EventTestsFailed type.
	// Intent: Defines the EventTestsFailed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventTestsFailed    = "TestsFailed"
	// Summary: Defines the EventTestsPassed type.
	// Intent: Defines the EventTestsPassed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventTestsPassed    = "TestsPassed"
	// Summary: Defines the EventSpecApproved type.
	// Intent: Defines the EventSpecApproved type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventSpecApproved   = "SpecApproved"
	// Summary: Defines the EventBlockerRaised type.
	// Intent: Defines the EventBlockerRaised type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventBlockerRaised  = "BlockerRaised"
	// Summary: Defines the EventBlockerCleared type.
	// Intent: Defines the EventBlockerCleared type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventBlockerCleared = "BlockerCleared"
	// Summary: Defines the EventPRCreated type.
	// Intent: Defines the EventPRCreated type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventPRCreated      = "PRCreated"
	// Summary: Defines the EventPRMerged type.
	// Intent: Defines the EventPRMerged type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventPRMerged       = "PRMerged"
	// Summary: Defines the EventDesignReviewed type.
	// Intent: Defines the EventDesignReviewed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventDesignReviewed = "DesignReviewed"
	// Summary: Defines the EventApprovalNeeded type.
	// Intent: Defines the EventApprovalNeeded type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	EventApprovalNeeded = "ApprovalNeeded"
)

// Summary: Agent represents an active, instantiated worker within the AI organisation.
// Intent: Agent represents an active, instantiated worker within the AI organisation.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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

// Summary: Message encapsulates a discrete event, command, or context update passed between agents or rooms.
// Intent: Message encapsulates a discrete event, command, or context update passed between agents or rooms.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

// DelegateTask allows an agent in Delegate Mode to act as a routing proxy.
// It inspects an incoming task, updates the sender and recipient fields,
// and forwards the task to the best-fit specialist agent from the registry.
//
// Parameters:
//   - fromAgentID: string; The unique identifier of the delegating agent.
//   - toAgentID: string; The unique identifier of the specialist agent.
//   - task: Message; The task payload to be delegated.
//
// Returns: An error if either the delegating agent or the specialist agent does not exist.
func (h *Hub) DelegateTask(fromAgentID, toAgentID string, task Message) error {
	task.FromAgent = fromAgentID
	task.ToAgent = toAgentID
	return h.Publish(task)
}

// Summary: MeetingRoom maintains a persistent, sequential transcript of inter-agent collaboration.
// Intent: MeetingRoom maintains a persistent, sequential transcript of inter-agent collaboration.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

// Summary: Hub acts as the thread-safe central message broker and runtime state manager for the AI workforce.  Constraints: Must be accessed via its exported methods to preserve data race safety.
// Intent: Hub acts as the thread-safe central message broker and runtime state manager for the AI workforce.  Constraints: Must be accessed via its exported methods to preserve data race safety.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Hub struct {
	mu            sync.RWMutex
	agents        map[string]Agent
	inbox         map[string][]Message
	meetings      map[string]MeetingRoom
	minimaxAPIKey string
	subs          map[string][]chan struct{}
}

// NewHub constructs a new instance of an orchestration Hub, pre-allocated with empty registries.
//
// Returns: An instantiated *Hub ready to register agents and route events.
func NewHub() *Hub {
	return &Hub{
		agents:   map[string]Agent{},
		inbox:    map[string][]Message{},
		meetings: map[string]MeetingRoom{},
		subs:     map[string][]chan struct{}{},
	}
}

// Summary: RegisterAgent enrolls an agent into the Hub, allocating an inbox and initialising its Status.  Parameters:   - agent: Agent; The worker object containing ID, Name, Role, and Organization context.
// Intent: RegisterAgent enrolls an agent into the Hub, allocating an inbox and initialising its Status.  Parameters:   - agent: Agent; The worker object containing ID, Name, Role, and Organization context.
// Params: agent
// Returns: None
// Errors: None
// Side Effects: None
func (h *Hub) RegisterAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent.Status == "" {
		agent.Status = StatusIdle
	}

	h.agents[agent.ID] = agent
}

// Summary: SetMinimaxAPIKey functionality.
// Intent: SetMinimaxAPIKey functionality.
// Params: key
// Returns: None
// Errors: None
// Side Effects: None
func (h *Hub) SetMinimaxAPIKey(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minimaxAPIKey = key
}

// Summary: MinimaxAPIKey functionality.
// Intent: MinimaxAPIKey functionality.
// Params: None
// Returns: string
// Errors: None
// Side Effects: None
func (h *Hub) MinimaxAPIKey() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.minimaxAPIKey
}

// Agent retrieves the runtime state of a specific worker by ID.
//
// Parameters:
//   - id: string; The unique identifier of the agent.
//
// Returns: The matching Agent object and a boolean indicating if it exists in the registry.
func (h *Hub) Agent(id string) (Agent, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agent, ok := h.agents[id]
	return agent, ok
}

// OpenMeeting instantiates a new collaborative context window and marks all participants as InMeeting.
//
// Parameters:
//   - id: string; Unique identifier for the room.
//   - participants: []string; A list of agent IDs to be enrolled in the discussion.
//
// Returns: The instantiated MeetingRoom.
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

// OpenMeetingWithAgenda creates a meeting room with an explicit agenda descriptor.
//
// Parameters:
//   - id: string; Unique identifier for the room.
//   - agenda: string; The primary objective guiding the agents' conversation.
//   - participants: []string; A list of agent IDs to be enrolled in the discussion.
//
// Returns: The instantiated MeetingRoom.
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

// Summary: FireAgent removes an agent from the hub and clears their inbox.  Parameters:   - id: string; The unique identifier of the agent to terminate.
// Intent: FireAgent removes an agent from the hub and clears their inbox.  Parameters:   - id: string; The unique identifier of the agent to terminate.
// Params: id
// Returns: None
// Errors: None
// Side Effects: None
func (h *Hub) FireAgent(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.agents, id)
	delete(h.inbox, id)
}

// Publish validates and routes a message to a direct recipient, a meeting room, or both.
//
// Parameters:
//   - message: Message; The event payload containing routing headers and content.
//
// Returns: An error if the sender or recipient agents do not exist, or if the target meeting is unrecognised.
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

		// Optimization: Check cap to avoid small reallocations
		inbox := h.inbox[message.ToAgent]
		if cap(inbox) == 0 {
			inbox = make([]Message, 0, 16)
		}
		h.inbox[message.ToAgent] = append(inbox, message)

		subs := h.subs[message.ToAgent]
		for i := 0; i < len(subs); i++ {
			select {
			case subs[i] <- struct{}{}:
			default:
			}
		}
	}

	sender := h.agents[message.FromAgent]
	if message.MeetingID != "" {
		meeting, ok := h.meetings[message.MeetingID]
		if !ok {
			return errors.New("meeting room is not registered")
		}

		if cap(meeting.Transcript) == 0 {
			meeting.Transcript = make([]Message, 0, 16)
		}
		meeting.Transcript = append(meeting.Transcript, message)
		h.meetings[message.MeetingID] = meeting
		sender.Status = StatusInMeeting

		for _, participant := range meeting.Participants {
			subs := h.subs[participant]
			for i := 0; i < len(subs); i++ {
				select {
				case subs[i] <- struct{}{}:
				default:
				}
			}
		}
	} else {
		sender.Status = StatusActive
	}
	h.agents[message.FromAgent] = sender

	// ⚡ BOLT: [Asynchronous telemetry recording to reduce critical path latency] - Randomized Selection from Top 5
	go func() {
		telemetry.RecordAgentApiCall(context.Background(), sender.ID, sender.Role, "publish")

		// Structured logging for agent execution traces
		telemetry.LogAgentExecution(context.Background(), sender.ID, sender.Role, "publish", message.Type, message.Content)
	}()

	return nil
}

// Subscribe returns a channel that receives real-time messages for the given agent.
func (h *Hub) Subscribe(agentID string) (<-chan struct{}, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan struct{}, 1)
	h.subs[agentID] = append(h.subs[agentID], ch)

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subs[agentID]
		for i, sub := range subs {
			if sub == ch {
				// Prevent memory leak from lingering reference in underlying array
				subs[i] = nil
				h.subs[agentID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	return ch, unsubscribe
}

// Inbox retrieves all undelivered or direct messages routed exclusively to a single agent.
//
// Parameters:
//   - agentID: string; The unique identifier of the worker.
//
// Returns: A slice of direct Message objects.
func (h *Hub) Inbox(agentID string) []Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	inbox := h.inbox[agentID]
	if len(inbox) == 0 {
		return nil
	}
	// ⚡ BOLT: [O(1) Inbox draining instead of O(N) slice copy] - Randomized Selection from Top 5
	h.inbox[agentID] = nil
	return inbox
}

// Meeting retrieves the current state and transcript of a specified virtual meeting room.
//
// Parameters:
//   - id: string; The unique identifier of the room.
//
// Returns: The matching MeetingRoom object and a boolean indicating if it exists.
func (h *Hub) Meeting(id string) (MeetingRoom, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	meeting, ok := h.meetings[id]
	telemetry.RecordMeetingEvent(context.Background(), "opened")
	return meeting, ok
}

// Meetings fetches a point-in-time snapshot of all active meeting rooms.
//
// Returns: A slice containing all MeetingRoom objects.
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

// Agents retrieves a point-in-time snapshot of the entire registered workforce, ordered by ID.
//
// Returns: A slice of all active Agent objects in the orchestration Hub.
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

// Summary: HubServiceServer implements the gRPC HubService defined in hub.proto.
// Intent: HubServiceServer implements the gRPC HubService defined in hub.proto.
// Params: s, hub
// Returns: None
// Errors: None
// Side Effects: None
func RegisterHubService(s *grpc.Server, hub *Hub) {
	pb.RegisterHubServiceServer(s, &HubServiceServer{hub: hub})
}

// Summary: Defines the HubServiceServer type.
// Intent: Defines the HubServiceServer type.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type HubServiceServer struct {
	pb.UnimplementedHubServiceServer
	hub *Hub
}

// Summary: NewHubServiceServer functionality.
// Intent: NewHubServiceServer functionality.
// Params: hub
// Returns: *HubServiceServer
// Errors: None
// Side Effects: None
func NewHubServiceServer(hub *Hub) *HubServiceServer {
	return &HubServiceServer{hub: hub}
}

// Summary: RegisterAgent functionality.
// Intent: RegisterAgent functionality.
// Params: ctx, req
// Returns: (*pb.RegisterAgentResponse, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: OpenMeeting functionality.
// Intent: OpenMeeting functionality.
// Params: ctx, req
// Returns: (*pb.MeetingRoom, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (s *HubServiceServer) OpenMeeting(ctx context.Context, req *pb.OpenMeetingRequest) (*pb.MeetingRoom, error) {
	meeting := s.hub.OpenMeetingWithAgenda(req.GetMeetingId(), req.GetAgenda(), req.GetParticipants())
	return pb.MeetingRoom_builder{
		Id:           meeting.ID,
		Agenda:       meeting.Agenda,
		Participants: meeting.Participants,
	}.Build(), nil
}

// Summary: Publish functionality.
// Intent: Publish functionality.
// Params: ctx, req
// Returns: (*pb.PublishMessageResponse, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: DelegateTask functionality.
// Intent: DelegateTask functionality.
// Params: ctx, req
// Returns: (*pb.DelegateTaskResponse, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

	return pb.DelegateTaskResponse_builder{Success: true}.Build(), nil
}

// Summary: StreamMessages functionality.
// Intent: StreamMessages functionality.
// Params: req, stream
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (s *HubServiceServer) StreamMessages(req *pb.StreamMessagesRequest, stream pb.HubService_StreamMessagesServer) error {
	agentID := req.GetAgentId()

	// 1. Subscribe for new real-time messages to eliminate polling latency.
	// We must subscribe first to avoid missing any messages published between
	// draining the inbox and subscribing.
	ch, unsubscribe := s.hub.Subscribe(agentID)
	defer unsubscribe()

	sendNewMessages := func() error {
		msgs := s.hub.Inbox(agentID)
		for _, m := range msgs {
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

// Summary: Reason functionality.
// Intent: Reason functionality.
// Params: ctx, req
// Returns: (*pb.ReasonResponse, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: MinimaxClient handles interaction with the Minimax Model 2.7.
// Intent: MinimaxClient handles interaction with the Minimax Model 2.7.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type MinimaxClient struct {
	APIKey string
}

// Summary: NewMinimaxClient functionality.
// Intent: NewMinimaxClient functionality.
// Params: apiKey
// Returns: *MinimaxClient
// Errors: None
// Side Effects: None
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

// Summary: Reason functionality.
// Intent: Reason functionality.
// Params: ctx, prompt
// Returns: (string, error)
// Errors: Returns an error if applicable
// Side Effects: None
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
	if err := enc.Encode(prompt); err != nil {
		return "", err
	}
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
