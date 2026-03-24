package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/onehumancorp/mono/srcs/integrations"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

// Summary: Handles retrieving meetings.
// Intent: Handles retrieving meetings.
// Params: w, _
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleMeetings(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.hub.Meetings())
}

// Summary: Handles sending a message.
// Intent: Handles sending a message.
// Params: w, r
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form payload", http.StatusBadRequest)
		return
	}

	message := orchestration.Message{
		ID:         "web-" + time.Now().UTC().Format("20060102150405.000000000"),
		FromAgent:  r.FormValue("fromAgent"),
		ToAgent:    r.FormValue("toAgent"),
		Type:       r.FormValue("messageType"),
		Content:    r.FormValue("content"),
		MeetingID:  r.FormValue("meetingId"),
		OccurredAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for concurrent approval locks
	if message.Type == "SpecApproved" || message.Type == "direction" {
		for _, m := range s.hub.Meetings() {
			if m.ID == message.MeetingID {
				// Find the most recent ApprovalNeeded from the target agent to the sender
				var lastApprovalNeededIndex int = -1
				for i := len(m.Transcript) - 1; i >= 0; i-- {
					t := m.Transcript[i]
					if t.Type == "ApprovalNeeded" && t.FromAgent == message.ToAgent && t.ToAgent == message.FromAgent {
						lastApprovalNeededIndex = i
						break
					}
				}

				if lastApprovalNeededIndex != -1 {
					// Check if this specific ApprovalNeeded has already been resolved by a subsequent message
					for i := lastApprovalNeededIndex + 1; i < len(m.Transcript); i++ {
						t := m.Transcript[i]
						if t.FromAgent == message.FromAgent && t.ToAgent == message.ToAgent {
							if t.Type == "SpecApproved" || (t.Type == "direction" && strings.Contains(t.Content, "Rejected")) {
								http.Error(w, "State Changed: This action has already been approved or rejected by another user.", http.StatusConflict)
								return
							}
						}
					}
				}
				break
			}
		}
	}

	if err := s.hub.Publish(message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		telemetry.RecordHumanInteraction(r.Context(), "message")
		writeJSON(w, s.snapshotLocked())
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Summary: Handles testing chat connection.
// Intent: Handles testing chat connection.
// Params: w, r
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleChatTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatTestRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IntegrationID == "" {
		http.Error(w, "integrationId is required", http.StatusBadRequest)
		return
	}
	creds := integrations.IntegrationCredentials{
		BotToken:   req.BotToken,
		ChatID:     req.ChatID,
		WebhookURL: req.WebhookURL,
		APIToken:   req.APIToken,
	}
	if err := s.integReg.TestConnection(req.IntegrationID, creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]bool{"success": true})
}

// Summary: Handles retrieving chat messages.
// Intent: Handles retrieving chat messages.
// Params: w, r
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	integrationID := r.URL.Query().Get("integrationId")
	msgs := s.integReg.ChatMessages(integrationID)
	if msgs == nil {
		msgs = []integrations.ChatMessage{}
	}
	writeJSON(w, msgs)
}

// Summary: Handles sending a chat message.
// Intent: Handles sending a chat message.
// Params: w, r
// Returns: None
// Errors: None
// Side Effects: None
func (s *Server) handleChatSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatSendRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	msg, err := s.integReg.SendChatMessage(req.IntegrationID, req.Channel, req.FromAgent, req.Content, req.ThreadID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, msg)
}
