package orchestration

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Status indicates the current operational phase of an AI agent within the workforce.
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
type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
	Status         Status `json:"status"`
}

// Message encapsulates a discrete event, command, or context update passed between agents or rooms.
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
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

// Hub acts as the thread-safe central message broker and runtime state manager for the AI workforce.
//
// Constraints: Must be accessed via its exported methods to preserve data race safety.
type Hub struct {
	mu       sync.RWMutex
	agents   map[string]Agent
	inbox    map[string][]Message
	meetings map[string]MeetingRoom
}

// NewHub constructs a new instance of an orchestration Hub, pre-allocated with empty registries.
//
// Returns: An instantiated *Hub ready to register agents and route events.
func NewHub() *Hub {
	return &Hub{
		agents:   map[string]Agent{},
		inbox:    map[string][]Message{},
		meetings: map[string]MeetingRoom{},
	}
}

// RegisterAgent enrolls an agent into the Hub, allocating an inbox and initialising its Status.
//
// Parameters:
//   - agent: Agent; The worker object containing ID, Name, Role, and Organization context.
func (h *Hub) RegisterAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent.Status == "" {
		agent.Status = StatusIdle
	}

	h.agents[agent.ID] = agent
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

	return meeting
}

// FireAgent removes an agent from the hub and clears their inbox.
//
// Parameters:
//   - id: string; The unique identifier of the agent to terminate.
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
	if message.ToAgent != "" && message.ToAgent != "all" {
		if _, ok := h.agents[message.ToAgent]; !ok {
			return errors.New("recipient agent is not registered")
		}
		h.inbox[message.ToAgent] = append(h.inbox[message.ToAgent], message)
	}

	if message.MeetingID != "" {
		meeting, ok := h.meetings[message.MeetingID]
		if !ok {
			return errors.New("meeting room is not registered")
		}
		meeting.Transcript = append(meeting.Transcript, message)
		h.meetings[message.MeetingID] = meeting
	}
	sender := h.agents[message.FromAgent]
	if message.MeetingID != "" {
		sender.Status = StatusInMeeting
	} else {
		sender.Status = StatusActive
	}
	h.agents[message.FromAgent] = sender

	return nil
}

// Inbox retrieves all undelivered or direct messages routed exclusively to a single agent.
//
// Parameters:
//   - agentID: string; The unique identifier of the worker.
//
// Returns: A slice of direct Message objects.
func (h *Hub) Inbox(agentID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return append([]Message(nil), h.inbox[agentID]...)
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

	return meetings
}

// Agents retrieves a point-in-time snapshot of the entire registered workforce, ordered by ID.
//
// Returns: A slice of all active Agent objects in the orchestration Hub.
func (h *Hub) Agents() []Agent {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agents := make([]Agent, 0, len(h.agents))
	for _, agent := range h.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}
