package domain

import "time"

// Message represents a discrete packet of communication between agents within a meeting room, containing the content and sender identity.
type Message struct {
	ID         string    `json:"id"`
	FromAgent  string    `json:"fromAgent"`
	ToAgent    string    `json:"toAgent"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	MeetingID  string    `json:"meetingId,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}
