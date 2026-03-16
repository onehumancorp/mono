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

// hireRequest carries agent creation parameters.
type hireRequest struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Model string `json:"model,omitempty"`
}

// fireRequest carries the ID of the agent to remove.
type fireRequest struct {
	AgentID string `json:"agentId"`
}

// MCPTool represents a registered tool in the MCP gateway.
type MCPTool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

// DomainInfo describes a supported organizational domain template.
type DomainInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var availableDomains = []DomainInfo{
	{ID: "software_company", Name: "Software Company", Description: "Full-stack engineering org: CEO, Director, PM, SWEs, QA, Security, Designer, Marketing."},
	{ID: "digital_marketing_agency", Name: "Digital Marketing Agency", Description: "Full-service agency: CEO, Marketing Director, Growth, Content, SEO, Paid Media, Analytics, Designer."},
	{ID: "accounting_firm", Name: "Accounting Firm", Description: "Financial services firm: CEO, CFO, Bookkeepers, Tax, Audit, Payroll."},
}

var mcpTools = []MCPTool{
	{ID: "git-mcp", Name: "Git", Description: "Source control operations: clone, commit, pull-request, review via GitHub or Gitea.", Category: "code", Status: "available"},
	{ID: "jira-mcp", Name: "Jira / Plane", Description: "Task and issue tracking: create tickets, update status, list sprint items.", Category: "project_management", Status: "available"},
	{ID: "figma-mcp", Name: "Figma", Description: "Design file access: read wireframes, export assets, inspect component specs.", Category: "design", Status: "available"},
	{ID: "aws-mcp", Name: "AWS", Description: "Cloud infrastructure: provision EC2 instances, manage S3, deploy Lambda functions.", Category: "infrastructure", Status: "available"},
	{ID: "slack-mcp", Name: "Slack / Mattermost", Description: "Human-in-the-loop approval: send HITL notifications, await human manager sign-off.", Category: "communication", Status: "available"},
	{ID: "postgres-mcp", Name: "PostgreSQL", Description: "Database operations: run queries, manage schema, inspect table data.", Category: "database", Status: "available"},
	{ID: "opentelemetry-mcp", Name: "OpenTelemetry", Description: "Observability: push metrics and traces to Grafana / OpenObserve dashboards.", Category: "observability", Status: "available"},
	{ID: "spire-mcp", Name: "SPIFFE/SPIRE", Description: "Identity management: issue and rotate SVID certificates for agent workloads.", Category: "identity", Status: "available"},
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
	mux.HandleFunc("/api/agents/hire", server.handleHireAgent)
	mux.HandleFunc("/api/agents/fire", server.handleFireAgent)
	mux.HandleFunc("/api/domains", server.handleDomains)
	mux.HandleFunc("/api/mcp/tools", server.handleMCPTools)
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

func (s *Server) handleHireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req hireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Role == "" {
		http.Error(w, "name and role are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	id := s.org.ID + "-agent-" + time.Now().UTC().Format("20060102150405000")
	agent := orchestration.Agent{
		ID:             id,
		Name:           req.Name,
		Role:           req.Role,
		OrganizationID: s.org.ID,
		Status:         orchestration.StatusIdle,
	}
	s.hub.RegisterAgent(agent)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

func (s *Server) handleFireAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req fireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.AgentID == "" {
		http.Error(w, "agentId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.hub.FireAgent(req.AgentID)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}

func (s *Server) handleDomains(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, availableDomains)
}

func (s *Server) handleMCPTools(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, mcpTools)
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

	switch scenario {
	case "launch-readiness":
		return seededLaunchReadiness(now)
	case "digital-marketing":
		return seededDigitalMarketing(now)
	case "accounting":
		return seededAccounting(now)
	default:
		return domain.Organization{}, nil, nil, errors.New("unsupported seed scenario")
	}
}

func seededLaunchReadiness(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "ux-1", Name: "Design Lead", Role: "DESIGNER", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("launch-readiness", "Review launch blockers, sign-off on reliability checklist, assign post-launch owners.", []string{"pm-1", "swe-1", "ux-1"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       orchestration.EventTask,
		Content:    "Ship the reliability checklist before launch.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-4 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-2",
		FromAgent:  "ux-1",
		ToAgent:    "pm-1",
		Type:       orchestration.EventStatus,
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

func seededDigitalMarketing(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewDigitalMarketingAgency("dma-demo", "Demo Digital Agency", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "growth-1", Name: "Growth Agent", Role: "GROWTH_AGENT", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "content-1", Name: "Content Strategist", Role: "CONTENT_STRATEGIST", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "seo-1", Name: "SEO Specialist", Role: "SEO_SPECIALIST", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("campaign-kickoff", "Plan Q2 acquisition campaigns and assign channel ownership.", []string{"growth-1", "content-1", "seo-1"})

	_ = hub.Publish(orchestration.Message{
		ID:        "seed-dma-1",
		FromAgent: "growth-1",
		ToAgent:   "content-1",
		Type:      orchestration.EventTask,
		Content:   "Draft top-of-funnel blog series targeting enterprise SaaS buyers.",
		MeetingID: "campaign-kickoff",
		OccurredAt: now.Add(-5 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "growth-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1800,
		CompletionTokens: 600,
		OccurredAt:       now.Add(-5 * time.Minute),
	})

	return org, hub, tracker, nil
}

func seededAccounting(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewAccountingFirm("acc-demo", "Demo Accounting Firm", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "bookkeeper-1", Name: "Bookkeeper", Role: "BOOKKEEPER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "tax-1", Name: "Tax Specialist", Role: "TAX_SPECIALIST", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "cfo-1", Name: "CFO", Role: "CFO", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("month-close", "Reconcile April ledger, prepare estimated tax liability, and review payroll entries.", []string{"bookkeeper-1", "tax-1", "cfo-1"})

	_ = hub.Publish(orchestration.Message{
		ID:        "seed-acc-1",
		FromAgent: "cfo-1",
		ToAgent:   "bookkeeper-1",
		Type:      orchestration.EventTask,
		Content:   "Reconcile bank feeds and categorize uncategorized transactions before EOD.",
		MeetingID: "month-close",
		OccurredAt: now.Add(-3 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "cfo-1",
		OrganizationID:   org.ID,
		Model:            "claude-3.5-sonnet",
		PromptTokens:     1500,
		CompletionTokens: 500,
		OccurredAt:       now.Add(-3 * time.Minute),
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
