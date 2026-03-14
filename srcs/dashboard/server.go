package dashboard

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

type Server struct {
	mu      sync.RWMutex
	org     domain.Organization
	hub     *orchestration.Hub
	tracker *billing.Tracker
}

type statusCount struct {
	Status orchestration.Status `json:"status"`
	Count  int                  `json:"count"`
}

type dashboardSnapshot struct {
	Organization domain.Organization         `json:"organization"`
	Meetings     []orchestration.MeetingRoom `json:"meetings"`
	Costs        billing.Summary             `json:"costs"`
	Agents       []orchestration.Agent       `json:"agents"`
	Statuses     []statusCount               `json:"statuses"`
	UpdatedAt    time.Time                   `json:"updatedAt"`
}

type seedRequest struct {
	Scenario string `json:"scenario"`
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
	mux.HandleFunc("/app", server.handleApp)
	if dist := frontendDistPath(); dist != "" {
		mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(dist))))
	} else {
		mux.HandleFunc("/app/", server.handleApp)
	}
	mux.HandleFunc("/api/dashboard", server.handleDashboard)
	mux.HandleFunc("/api/org", server.handleOrg)
	mux.HandleFunc("/api/meetings", server.handleMeetings)
	mux.HandleFunc("/api/costs", server.handleCosts)
	mux.HandleFunc("/api/messages", server.handleSendMessage)
	mux.HandleFunc("/api/dev/seed", server.handleDevSeed)
	return mux
}

func frontendDistPath() string {
	if fromEnv := os.Getenv("MONO_FRONTEND_DIST"); fromEnv != "" {
		if hasFrontendIndex(fromEnv) {
			return fromEnv
		}
	}

	candidates := []string{
		"srcs/frontend/dist",
		"../srcs/frontend/dist",
		"../../srcs/frontend/dist",
	}

	for _, candidate := range candidates {
		if hasFrontendIndex(candidate) {
			return candidate
		}
	}

	return ""
}

func hasFrontendIndex(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "index.html"))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (s *Server) handleApp(w http.ResponseWriter, r *http.Request) {
	if dist := frontendDistPath(); dist != "" {
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>One Human Corp Frontend</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; background: #0f172a; color: #e2e8f0; }
    .card { background: #1e293b; padding: 1rem 1.25rem; border-radius: 12px; }
    code { background: #334155; padding: 0.1rem 0.3rem; border-radius: 6px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>React Frontend Route</h1>
    <p>No production build found at <code>srcs/frontend/dist</code>.</p>
    <p>Run <code>cd srcs/frontend && npm install && npm run build</code> and refresh this page.</p>
    <p>For local development, run <code>npm run dev</code> in <code>srcs/frontend</code>.</p>
  </div>
</body>
</html>`))
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.snapshot()
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
		"Org":      snapshot.Organization,
		"Agents":   snapshot.Agents,
		"Statuses": snapshot.Statuses,
		"Meetings": snapshot.Meetings,
		"Summary":  snapshot.Costs,
	})
}

func (s *Server) handleOrg(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.org)
}

func (s *Server) handleMeetings(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.hub.Meetings())
}

func (s *Server) handleCosts(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.tracker.Summary(s.org.ID))
}

func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.snapshot())
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

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.hub.Publish(message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, s.snapshotLocked())
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleDevSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload seedRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	org, hub, tracker, err := seededScenario(payload.Scenario, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.org = org
	s.hub = hub
	s.tracker = tracker
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

func (s *Server) snapshot() dashboardSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

func (s *Server) snapshotLocked() dashboardSnapshot {
	agents := s.hub.Agents()
	return dashboardSnapshot{
		Organization: s.org,
		Meetings:     s.hub.Meetings(),
		Costs:        s.tracker.Summary(s.org.ID),
		Agents:       agents,
		Statuses:     summarizeStatuses(agents),
		UpdatedAt:    time.Now().UTC(),
	}
}

func seededScenario(name string, now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	scenario := name
	if scenario == "" {
		scenario = "launch-readiness"
	}

	if scenario != "launch-readiness" {
		return domain.Organization{}, nil, nil, errors.New("unsupported seed scenario")
	}

	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "ux-1", Name: "Design Lead", Role: "DESIGNER", OrganizationID: org.ID})
	hub.OpenMeeting("launch-readiness", []string{"pm-1", "swe-1", "ux-1"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       "task",
		Content:    "Ship the reliability checklist before launch.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-4 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-2",
		FromAgent:  "ux-1",
		ToAgent:    "pm-1",
		Type:       "status",
		Content:    "Design QA pass completed with no blockers.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-2 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "pm-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o-mini",
		PromptTokens:     1200,
		CompletionTokens: 400,
		OccurredAt:       now.Add(-10 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     2600,
		CompletionTokens: 900,
		OccurredAt:       now.Add(-8 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "ux-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o-mini",
		PromptTokens:     900,
		CompletionTokens: 300,
		OccurredAt:       now.Add(-6 * time.Minute),
	})

	return org, hub, tracker, nil
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
