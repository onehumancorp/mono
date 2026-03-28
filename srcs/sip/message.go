package sip

import "time"

const (
	EventTask            = "task"
	EventStatus          = "status"
	EventApprovalNeeded  = "approval_needed"
	EventApprovalGranted = "approval_granted"
	EventError           = "error"
)

type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}
