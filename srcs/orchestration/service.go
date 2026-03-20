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

// Status indicates the current operational phase of an AI agent within the workforce.
// Summary: Status functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type Status string

const (
	StatusIdle      Status = "IDLE"
	StatusActive    Status = "ACTIVE"
	StatusInMeeting Status = "IN_MEETING"
	StatusBlocked   Status = "BLOCKED"
)

// Event type constants for the asynchronous pub/sub agent interaction protocol.
const (
	EventTask           = "task"
	EventStatus         = "status"
	EventHandoff        = "handoff"
	EventCodeReviewed   = "CodeReviewed"
	EventTestsFailed    = "TestsFailed"
	EventTestsPassed    = "TestsPassed"
	EventSpecApproved   = "SpecApproved"
	EventBlockerRaised  = "BlockerRaised"
	EventBlockerCleared = "BlockerCleared"
	EventPRCreated      = "PRCreated"
	EventPRMerged       = "PRMerged"
	EventDesignReviewed = "DesignReviewed"
	EventApprovalNeeded = "ApprovalNeeded"
)

// Agent represents an active, instantiated worker within the AI organisation.
// Summary: Agent functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// Message encapsulates a discrete event, command, or context update passed between agents or rooms.
// Summary: Message functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

// MeetingRoom maintains a persistent, sequential transcript of inter-agent collaboration.
// Summary: MeetingRoom functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

// Hub acts as the thread-safe central message broker and runtime state manager for the AI workforce.
//
// Constraints: Must be accessed via its exported methods to preserve data race safety.
// Summary: Hub functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: NewHub functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func NewHub() *Hub {
	return &Hub{
		agents:   map[string]Agent{},
		inbox:    map[string][]Message{},
		meetings: map[string]MeetingRoom{},
		subs:     map[string][]chan struct{}{},
	}
}

// RegisterAgent enrolls an agent into the Hub, allocating an inbox and initialising its Status.
//
// Parameters:
//   - agent: Agent; The worker object containing ID, Name, Role, and Organization context.
// Summary: RegisterAgent functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (h *Hub) RegisterAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent.Status == "" {
		agent.Status = StatusIdle
	}

	h.agents[agent.ID] = agent
}

// SetMinimaxAPIKey provides functionality for SetMinimaxAPIKey.
// Summary: SetMinimaxAPIKey functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (h *Hub) SetMinimaxAPIKey(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minimaxAPIKey = key
}

// MinimaxAPIKey provides functionality for MinimaxAPIKey.
// Summary: MinimaxAPIKey functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: Agent functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: OpenMeeting functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: OpenMeetingWithAgenda functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// FireAgent removes an agent from the hub and clears their inbox.
//
// Parameters:
//   - id: string; The unique identifier of the agent to terminate.
// Summary: FireAgent functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: Publish functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

	telemetry.RecordAgentApiCall(context.Background(), sender.ID, sender.Role, "publish")

	return nil
}

// Subscribe returns a channel that receives real-time messages for the given agent.
// Summary: Subscribe functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: Inbox functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (h *Hub) Inbox(agentID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	inbox := h.inbox[agentID]
	if len(inbox) == 0 {
		return nil
	}
	res := make([]Message, len(inbox))
	copy(res, inbox)
	return res
}

// Meeting retrieves the current state and transcript of a specified virtual meeting room.
//
// Parameters:
//   - id: string; The unique identifier of the room.
//
// Returns: The matching MeetingRoom object and a boolean indicating if it exists.
// Summary: Meeting functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: Meetings functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
// Summary: Agents functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// HubServiceServer implements the gRPC HubService defined in hub.proto.
// Summary: RegisterHubService functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func RegisterHubService(s *grpc.Server, hub *Hub) {
	pb.RegisterHubServiceServer(s, &HubServiceServer{hub: hub})
}

// HubServiceServer provides functionality for HubServiceServer.
// Summary: HubServiceServer functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type HubServiceServer struct {
	pb.UnimplementedHubServiceServer
	hub *Hub
}

// NewHubServiceServer provides functionality for NewHubServiceServer.
// Summary: NewHubServiceServer functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func NewHubServiceServer(hub *Hub) *HubServiceServer {
	return &HubServiceServer{hub: hub}
}

// RegisterAgent provides functionality for RegisterAgent.
// Summary: RegisterAgent functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// OpenMeeting provides functionality for OpenMeeting.
// Summary: OpenMeeting functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (s *HubServiceServer) OpenMeeting(ctx context.Context, req *pb.OpenMeetingRequest) (*pb.MeetingRoom, error) {
	meeting := s.hub.OpenMeetingWithAgenda(req.GetMeetingId(), req.GetAgenda(), req.GetParticipants())
	return pb.MeetingRoom_builder{
		Id:           meeting.ID,
		Agenda:       meeting.Agenda,
		Participants: meeting.Participants,
	}.Build(), nil
}

// Publish provides functionality for Publish.
// Summary: Publish functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// StreamMessages provides functionality for StreamMessages.
// Summary: StreamMessages functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (s *HubServiceServer) StreamMessages(req *pb.StreamMessagesRequest, stream pb.HubService_StreamMessagesServer) error {
	agentID := req.GetAgentId()

	// 1. Subscribe for new real-time messages to eliminate polling latency.
	// We must subscribe first to avoid missing any messages published between
	// draining the inbox and subscribing.
	ch, unsubscribe := s.hub.Subscribe(agentID)
	defer unsubscribe()

	var lastCount int

	sendNewMessages := func() error {
		msgs := s.hub.Inbox(agentID)
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

// Reason provides functionality for Reason.
// Summary: Reason functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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

// MinimaxClient handles interaction with the Minimax Model 2.7.
// Summary: MinimaxClient functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type MinimaxClient struct {
	APIKey string
}

// NewMinimaxClient provides functionality for NewMinimaxClient.
// Summary: NewMinimaxClient functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func NewMinimaxClient(apiKey string) *MinimaxClient {
	return &MinimaxClient{APIKey: apiKey}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Reason provides functionality for Reason.
// Summary: Reason functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
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
