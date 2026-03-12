package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

type Server struct {
	org     domain.Organization
	hub     *orchestration.Hub
	tracker *billing.Tracker
}

type statusCount struct {
	Status orchestration.Status
	Count  int
}

var statusOrder = []orchestration.Status{
	orchestration.StatusActive,
	orchestration.StatusBlocked,
	orchestration.StatusIdle,
	orchestration.StatusInMeeting,
}

func NewServer(org domain.Organization, hub *orchestration.Hub, tracker *billing.Tracker) http.Handler {
	server := &Server{org: org, hub: hub, tracker: tracker}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/api/org", server.handleOrg)
	mux.HandleFunc("/api/meetings", server.handleMeetings)
	mux.HandleFunc("/api/costs", server.handleCosts)
	mux.HandleFunc("/api/messages", server.handleSendMessage)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	agents := s.hub.Agents()
	page := template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>One Human Corp Dashboard</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; background: #0f172a; color: #e2e8f0; }
    .card { background: #1e293b; padding: 1rem 1.25rem; border-radius: 12px; margin-bottom: 1rem; }
    h1, h2 { margin-top: 0; }
    ul { padding-left: 1.25rem; }
  </style>
</head>
<body>
  <h1>One Human Corp Dashboard</h1>
  <div class="card">
    <h2>{{.Org.Name}}</h2>
    <p>Domain: {{.Org.Domain}}</p>
    <p>Members: {{len .Org.Members}}</p>
  </div>
  <div class="card">
    <h2>Org Chart</h2>
    <ul>
    {{range .Org.Members}}
      <li>{{.Name}} — {{.Role}}</li>
    {{end}}
    </ul>
  </div>
  <div class="card">
    <h2>Role Playbooks</h2>
    {{range .Org.RoleProfiles}}
    <h3>{{.Role}}</h3>
    <p>{{.BasePrompt}}</p>
    <p><strong>Capabilities:</strong> {{range $index, $capability := .Capabilities}}{{if $index}}, {{end}}{{$capability}}{{end}}</p>
    <p><strong>Context Inputs:</strong> {{range $index, $input := .ContextInputs}}{{if $index}}, {{end}}{{$input}}{{end}}</p>
    {{end}}
  </div>
  <div class="card">
    <h2>Project Status</h2>
    <p>Registered agents: {{len .Agents}}</p>
    <ul>
    {{range .Statuses}}
      <li>{{.Status}} — {{.Count}}</li>
    {{end}}
    </ul>
    <ul>
    {{range .Agents}}
      <li>{{.Name}} — {{.Status}}</li>
    {{end}}
    </ul>
  </div>
  <div class="card">
    <h2>Active Meetings</h2>
    <p>{{len .Meetings}} meeting(s)</p>
    {{range .Meetings}}
    <h3>{{.ID}}</h3>
    <ul>
      {{range .Transcript}}
      <li>{{.FromAgent}} → {{.ToAgent}}: {{.Content}}</li>
      {{else}}
      <li>No messages yet.</li>
      {{end}}
    </ul>
    {{end}}
  </div>
  <div class="card">
    <h2>Cost Summary</h2>
    <p>Total cost: ${{printf "%.6f" .Summary.TotalCostUSD}}</p>
    <p>Total tokens: {{.Summary.TotalTokens}}</p>
    <ul>
    {{range .Summary.Agents}}
      <li>{{.AgentID}} — ${{printf "%.6f" .CostUSD}} ({{.TokenUsed}} tokens)</li>
    {{end}}
    </ul>
  </div>
  <div class="card">
    <h2>Send Message</h2>
    <form method="post" action="/api/messages">
      <label>From Agent <input name="fromAgent" value="pm-1"></label><br>
      <label>To Agent <input name="toAgent" value="swe-1"></label><br>
      <label>Meeting ID <input name="meetingId" value="kickoff"></label><br>
      <label>Message Type <input name="messageType" value="task"></label><br>
      <label>Content <input name="content" value="Review the roadmap"></label><br>
      <button type="submit">Send Message</button>
    </form>
  </div>
</body>
</html>`))

	_ = page.Execute(w, map[string]any{
		"Org":      s.org,
		"Agents":   agents,
		"Statuses": summarizeStatuses(agents),
		"Meetings": s.hub.Meetings(),
		"Summary":  s.tracker.Summary(s.org.ID),
	})
}

func (s *Server) handleOrg(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.org)
}

func (s *Server) handleMeetings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.hub.Meetings())
}

func (s *Server) handleCosts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.tracker.Summary(s.org.ID))
}

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

	if err := s.hub.Publish(message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func summarizeStatuses(agents []orchestration.Agent) []statusCount {
	counts := map[orchestration.Status]int{
		orchestration.StatusIdle:      0,
		orchestration.StatusActive:    0,
		orchestration.StatusInMeeting: 0,
		orchestration.StatusBlocked:   0,
	}
	for _, agent := range agents {
		counts[agent.Status]++
	}

	statuses := make([]statusCount, 0, len(counts))
	for _, status := range statusOrder {
		statuses = append(statuses, statusCount{
			Status: status,
			Count:  counts[status],
		})
	}

	return statuses
}
