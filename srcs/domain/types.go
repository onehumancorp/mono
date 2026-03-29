package domain

import "time"

// Status indicates the current operational phase of an AI agent within the workforce.
type Status string

const (
	StatusIdle            Status = "IDLE"
	StatusActive          Status = "ACTIVE"
	StatusInMeeting       Status = "IN_MEETING"
	StatusBlocked         Status = "BLOCKED"
	StatusWaitingForTools Status = "WAITING_FOR_TOOLS"
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

// Agent represents an autonomous AI actor registered in the orchestration Hub.
type Agent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
	Status         Status `json:"status"`
	ProviderType   string `json:"providerType,omitempty"`
	Region         string `json:"region,omitempty"`
}

// Message represents a discrete packet of communication between agents within a meeting room.
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

// MeetingRoom provides a thread-safe, isolated collaborative space.
type MeetingRoom struct {
	ID           string    `json:"id"`
	Agenda       string    `json:"agenda,omitempty"`
	Participants []string  `json:"participants"`
	Transcript   []Message `json:"transcript"`
}
