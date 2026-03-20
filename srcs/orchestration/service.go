package orchestration

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Status defines the current operational phase of an AI agent.
//
// Parameters: none
// Returns: a Status enum string indicating agent state.
// Errors: none.
// Side Effects: none.
type Status string

const (
	// StatusIdle indicates the agent is waiting for work.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusIdle      Status = "IDLE"
	// StatusActive indicates the agent is executing a task.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusActive    Status = "ACTIVE"
	// StatusInMeeting indicates the agent is collaborating synchronously.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusInMeeting Status = "IN_MEETING"
	// StatusBlocked indicates the agent requires human intervention.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusBlocked   Status = "BLOCKED"
)

// Event type constants for the asynchronous pub/sub agent interaction protocol.
const (
	// EventTask signals a generic work request via Pub/Sub.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	EventTask           = "task"
	// EventStatus broadcasts an operational update via Pub/Sub.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	EventStatus         = "status"
	// EventHandoff initiates an escalation protocol to a human manager.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	EventHandoff        = "handoff"
	// EventCodeReviewed signals the completion of a PR check.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	EventCodeReviewed   = "CodeReviewed"
	// EventTestsFailed signals a regression during validation.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
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

// Agent represents an autonomous actor operating within the swarm.
//
// Parameters: none
// Returns: an Agent struct holding operational state and identification.
// Errors: none.
// Side Effects: none.
type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
	Status         Status `json:"status"`
}

// Message represents a serialized event or conversation snippet within the Pub/Sub system.
//
// Parameters: none
// Returns: a Message struct tracking interaction payloads.
// Errors: none.
// Side Effects: none.
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

// MeetingRoom defines a synchronous collaboration space for multiple agents.
//
// Parameters: none
// Returns: a MeetingRoom struct tracking discussion transcripts.
// Errors: none.
// Side Effects: none.
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

// Hub orchestrates the communication and state lifecycle of the entire agent network.
//
// Parameters: none
// Returns: a Hub instance for managing Pub/Sub and virtual meetings.
// Errors: none.
// Side Effects: none.
type Hub struct {
	mu       sync.RWMutex
	agents   map[string]Agent
	inbox    map[string][]Message
	meetings map[string]MeetingRoom
}

// NewHub initializes an event broker for asynchronous agent communication.
//
// Parameters: none
// Returns: A pointer to a newly instantiated Hub.
// Errors: none.
// Side Effects: creates underlying state management structures.
func NewHub() *Hub {
	return &Hub{
		agents:   map[string]Agent{},
		inbox:    map[string][]Message{},
		meetings: map[string]MeetingRoom{},
	}
}

func (h *Hub) RegisterAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if agent.Status == "" {
		agent.Status = StatusIdle
	}

	h.agents[agent.ID] = agent
}

func (h *Hub) Agent(id string) (Agent, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agent, ok := h.agents[id]
	return agent, ok
}

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
func (h *Hub) FireAgent(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.agents, id)
	delete(h.inbox, id)
}

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

	return nil
}

func (h *Hub) Inbox(agentID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return append([]Message(nil), h.inbox[agentID]...)
}

func (h *Hub) Meeting(id string) (MeetingRoom, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	meeting, ok := h.meetings[id]
	return meeting, ok
}

func (h *Hub) Meetings() []MeetingRoom {
	h.mu.RLock()
	defer h.mu.RUnlock()

	meetings := make([]MeetingRoom, 0, len(h.meetings))
	for _, meeting := range h.meetings {
		meetings = append(meetings, meeting)
	}

	return meetings
}

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
