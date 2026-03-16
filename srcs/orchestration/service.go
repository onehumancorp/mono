package orchestration

import (
	"errors"
	"sort"
	"sync"
	"time"
)

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

type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
	Status         Status `json:"status"`
}

type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}

type Hub struct {
	mu       sync.RWMutex
	agents   map[string]Agent
	inbox    map[string][]Message
	meetings map[string]MeetingRoom
}

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
