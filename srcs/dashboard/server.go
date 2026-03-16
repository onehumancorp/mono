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
	mu        sync.RWMutex
	org       domain.Organization
	hub       *orchestration.Hub
	tracker   *billing.Tracker
	approvals []ApprovalRequest
	handoffs  []HandoffPackage
	skills    []SkillPack
	snapshots []OrgSnapshot
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

// ── Approval / Confidence Gating ─────────────────────────────────────────────

// ApprovalStatus represents the lifecycle state of a guardian-gate request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

// ApprovalRequest is created by the Guardian Agent when a high-risk action
// requires explicit human sign-off.
type ApprovalRequest struct {
	ID               string         `json:"id"`
	AgentID          string         `json:"agentId"`
	Action           string         `json:"action"`
	Reason           string         `json:"reason"`
	EstimatedCostUSD float64        `json:"estimatedCostUsd"`
	RiskLevel        string         `json:"riskLevel"` // low | medium | high | critical
	Status           ApprovalStatus `json:"status"`
	CreatedAt        time.Time      `json:"createdAt"`
	DecidedAt        *time.Time     `json:"decidedAt,omitempty"`
	DecidedBy        string         `json:"decidedBy,omitempty"`
}

type approvalCreateRequest struct {
	AgentID          string  `json:"agentId"`
	Action           string  `json:"action"`
	Reason           string  `json:"reason"`
	EstimatedCostUSD float64 `json:"estimatedCostUsd"`
	RiskLevel        string  `json:"riskLevel"`
}

type approvalDecideRequest struct {
	ApprovalID string `json:"approvalId"`
	Decision   string `json:"decision"` // approve | reject
	DecidedBy  string `json:"decidedBy"`
}

// ── Warm Handoff ──────────────────────────────────────────────────────────────

// HandoffPackage carries the structured context an agent sends to a human manager
// when escalating a task it cannot complete autonomously.
type HandoffPackage struct {
	ID             string    `json:"id"`
	FromAgentID    string    `json:"fromAgentId"`
	ToHumanRole    string    `json:"toHumanRole"`
	Intent         string    `json:"intent"`
	FailedAttempts int       `json:"failedAttempts"`
	CurrentState   string    `json:"currentState"`
	Status         string    `json:"status"` // pending | acknowledged | resolved
	CreatedAt      time.Time `json:"createdAt"`
}

type handoffCreateRequest struct {
	FromAgentID    string `json:"fromAgentId"`
	ToHumanRole    string `json:"toHumanRole"`
	Intent         string `json:"intent"`
	FailedAttempts int    `json:"failedAttempts"`
	CurrentState   string `json:"currentState"`
}

// ── Agent Identity (SPIFFE/SPIRE abstraction) ─────────────────────────────────

// AgentIdentity represents the SPIFFE SVID certificate issued to an agent workload.
type AgentIdentity struct {
	AgentID     string    `json:"agentId"`
	SVID        string    `json:"svid"`
	TrustDomain string    `json:"trustDomain"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// ── Extensible Skill Import Framework ────────────────────────────────────────

// SkillPackRole pairs a role name with its override base prompt.
type SkillPackRole struct {
	Role       string `json:"role"`
	BasePrompt string `json:"basePrompt"`
}

// SkillPack is an importable module that extends or overrides agent capabilities.
type SkillPack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Domain      string          `json:"domain"`
	Description string          `json:"description"`
	Source      string          `json:"source"` // builtin | custom | marketplace
	Author      string          `json:"author,omitempty"`
	Roles       []SkillPackRole `json:"roles"`
	ImportedAt  time.Time       `json:"importedAt"`
}

type skillImportRequest struct {
	Name        string          `json:"name"`
	Domain      string          `json:"domain"`
	Description string          `json:"description"`
	Source      string          `json:"source"`
	Author      string          `json:"author,omitempty"`
	Roles       []SkillPackRole `json:"roles"`
}

// ── Org Snapshot & Recovery ───────────────────────────────────────────────────

// OrgSnapshot is a point-in-time metadata record of an organization's state.
type OrgSnapshot struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	OrgID        string    `json:"orgId"`
	OrgName      string    `json:"orgName"`
	Domain       string    `json:"domain"`
	AgentCount   int       `json:"agentCount"`
	MeetingCount int       `json:"meetingCount"`
	MessageCount int       `json:"messageCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type snapshotCreateRequest struct {
	Label string `json:"label"`
}

type snapshotRestoreRequest struct {
	SnapshotID string `json:"snapshotId"`
}

// ── Marketplace ───────────────────────────────────────────────────────────────

// MarketplaceItem describes a community-published asset.
type MarketplaceItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // agent | domain | skill_pack | tool
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

// AnalyticsSummary surfaces operational health metrics.
type AnalyticsSummary struct {
	HumanAgentRatio     float64 `json:"humanAgentRatio"`
	TotalAgents         int     `json:"totalAgents"`
	TotalHumans         int     `json:"totalHumans"`
	AuditFidelityPct    float64 `json:"auditFidelityPct"`
	ResumptionLatencyMS int     `json:"resumptionLatencyMs"`
	PendingApprovals    int     `json:"pendingApprovals"`
	ActiveHandoffs      int     `json:"activeHandoffs"`
	TokenVelocity       int64   `json:"tokenVelocity"`
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
	server := &Server{
		org:       org,
		hub:       hub,
		tracker:   tracker,
		approvals: []ApprovalRequest{},
		handoffs:  []HandoffPackage{},
		skills:    defaultSkillPacks(),
		snapshots: []OrgSnapshot{},
	}
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
	// Phase 2 – Confidence Gating / Guardian Agent
	mux.HandleFunc("/api/approvals", server.handleApprovals)
	mux.HandleFunc("/api/approvals/request", server.handleApprovalRequest)
	mux.HandleFunc("/api/approvals/decide", server.handleApprovalDecide)
	// Phase 2 – Warm Handoff
	mux.HandleFunc("/api/handoffs", server.handleHandoffs)
	// Phase 2 – Unified Identity Management (SPIFFE/SPIRE)
	mux.HandleFunc("/api/identities", server.handleIdentities)
	// Phase 2 – Extensible Skill Import Framework
	mux.HandleFunc("/api/skills", server.handleSkills)
	mux.HandleFunc("/api/skills/import", server.handleSkillImport)
	// Phase 4 – Org Snapshot & Recovery
	mux.HandleFunc("/api/snapshots", server.handleSnapshots)
	mux.HandleFunc("/api/snapshots/create", server.handleSnapshotCreate)
	mux.HandleFunc("/api/snapshots/restore", server.handleSnapshotRestore)
	// Phase 4 – Community Marketplace
	mux.HandleFunc("/api/marketplace", server.handleMarketplace)
	// Phase 4 – Real-time Analytics
	mux.HandleFunc("/api/analytics", server.handleAnalytics)
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

// ── Approval / Confidence Gating Handlers ─────────────────────────────────────

func (s *Server) handleApprovals(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
s.mu.RLock()
list := append([]ApprovalRequest(nil), s.approvals...)
s.mu.RUnlock()
writeJSON(w, list)
default:
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

func (s *Server) handleApprovalRequest(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req approvalCreateRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}
if req.AgentID == "" || req.Action == "" {
http.Error(w, "agentId and action are required", http.StatusBadRequest)
return
}

now := time.Now().UTC()
approval := ApprovalRequest{
ID:               s.org.ID + "-approval-" + now.Format("20060102150405000"),
AgentID:          req.AgentID,
Action:           req.Action,
Reason:           req.Reason,
EstimatedCostUSD: req.EstimatedCostUSD,
RiskLevel:        req.RiskLevel,
Status:           ApprovalStatusPending,
CreatedAt:        now,
}
if approval.RiskLevel == "" {
if approval.EstimatedCostUSD > 500 {
approval.RiskLevel = "critical"
} else if approval.EstimatedCostUSD > 100 {
approval.RiskLevel = "high"
} else {
approval.RiskLevel = "medium"
}
}

s.mu.Lock()
s.approvals = append(s.approvals, approval)
s.mu.Unlock()

writeJSON(w, approval)
}

func (s *Server) handleApprovalDecide(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req approvalDecideRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}
if req.ApprovalID == "" || req.Decision == "" {
http.Error(w, "approvalId and decision are required", http.StatusBadRequest)
return
}

var newStatus ApprovalStatus
switch req.Decision {
case "approve":
newStatus = ApprovalStatusApproved
case "reject":
newStatus = ApprovalStatusRejected
default:
http.Error(w, "decision must be 'approve' or 'reject'", http.StatusBadRequest)
return
}

now := time.Now().UTC()
s.mu.Lock()
found := false
for i, a := range s.approvals {
if a.ID == req.ApprovalID {
s.approvals[i].Status = newStatus
s.approvals[i].DecidedAt = &now
s.approvals[i].DecidedBy = req.DecidedBy
found = true
break
}
}
s.mu.Unlock()

if !found {
http.Error(w, "approval not found", http.StatusNotFound)
return
}

s.mu.RLock()
list := append([]ApprovalRequest(nil), s.approvals...)
s.mu.RUnlock()
writeJSON(w, list)
}

// ── Warm Handoff Handlers ─────────────────────────────────────────────────────

func (s *Server) handleHandoffs(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
s.mu.RLock()
list := append([]HandoffPackage(nil), s.handoffs...)
s.mu.RUnlock()
writeJSON(w, list)
case http.MethodPost:
var req handoffCreateRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}
if req.FromAgentID == "" || req.Intent == "" {
http.Error(w, "fromAgentId and intent are required", http.StatusBadRequest)
return
}
now := time.Now().UTC()
handoff := HandoffPackage{
ID:             s.org.ID + "-handoff-" + now.Format("20060102150405000"),
FromAgentID:    req.FromAgentID,
ToHumanRole:    req.ToHumanRole,
Intent:         req.Intent,
FailedAttempts: req.FailedAttempts,
CurrentState:   req.CurrentState,
Status:         "pending",
CreatedAt:      now,
}
s.mu.Lock()
s.handoffs = append(s.handoffs, handoff)
s.mu.Unlock()
writeJSON(w, handoff)
default:
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

// ── Identity Management Handler ───────────────────────────────────────────────

func (s *Server) handleIdentities(w http.ResponseWriter, _ *http.Request) {
s.mu.RLock()
agents := s.hub.Agents()
org := s.org
s.mu.RUnlock()

now := time.Now().UTC()
identities := make([]AgentIdentity, 0, len(agents))
for _, agent := range agents {
identities = append(identities, AgentIdentity{
AgentID:     agent.ID,
SVID:        "spiffe://onehumancorp.io/" + org.ID + "/" + agent.ID,
TrustDomain: "onehumancorp.io",
IssuedAt:    now,
ExpiresAt:   now.Add(24 * time.Hour),
})
}
writeJSON(w, identities)
}

// ── Skill Pack Handlers ───────────────────────────────────────────────────────

func (s *Server) handleSkills(w http.ResponseWriter, _ *http.Request) {
s.mu.RLock()
list := append([]SkillPack(nil), s.skills...)
s.mu.RUnlock()
writeJSON(w, list)
}

func (s *Server) handleSkillImport(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req skillImportRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}
if req.Name == "" || req.Domain == "" {
http.Error(w, "name and domain are required", http.StatusBadRequest)
return
}

now := time.Now().UTC()
source := req.Source
if source == "" {
source = "custom"
}
pack := SkillPack{
ID:          s.org.ID + "-skill-" + now.Format("20060102150405000"),
Name:        req.Name,
Domain:      req.Domain,
Description: req.Description,
Source:      source,
Author:      req.Author,
Roles:       req.Roles,
ImportedAt:  now,
}
if pack.Roles == nil {
pack.Roles = []SkillPackRole{}
}

s.mu.Lock()
s.skills = append(s.skills, pack)
s.mu.Unlock()

writeJSON(w, pack)
}

// ── Snapshot Handlers ─────────────────────────────────────────────────────────

func (s *Server) handleSnapshots(w http.ResponseWriter, _ *http.Request) {
s.mu.RLock()
list := append([]OrgSnapshot(nil), s.snapshots...)
s.mu.RUnlock()
writeJSON(w, list)
}

func (s *Server) handleSnapshotCreate(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req snapshotCreateRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}

s.mu.Lock()
meetings := s.hub.Meetings()
agents := s.hub.Agents()
msgCount := 0
for _, m := range meetings {
msgCount += len(m.Transcript)
}
now := time.Now().UTC()
label := req.Label
if label == "" {
label = "Snapshot " + now.Format("2006-01-02 15:04")
}
snap := OrgSnapshot{
ID:           s.org.ID + "-snap-" + now.Format("20060102150405000"),
Label:        label,
OrgID:        s.org.ID,
OrgName:      s.org.Name,
Domain:       s.org.Domain,
AgentCount:   len(agents),
MeetingCount: len(meetings),
MessageCount: msgCount,
CreatedAt:    now,
}
s.snapshots = append(s.snapshots, snap)
s.mu.Unlock()

writeJSON(w, snap)
}

func (s *Server) handleSnapshotRestore(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req snapshotRestoreRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid JSON payload", http.StatusBadRequest)
return
}
if req.SnapshotID == "" {
http.Error(w, "snapshotId is required", http.StatusBadRequest)
return
}

s.mu.RLock()
var target *OrgSnapshot
for i, snap := range s.snapshots {
if snap.ID == req.SnapshotID {
target = &s.snapshots[i]
break
}
}
s.mu.RUnlock()

if target == nil {
http.Error(w, "snapshot not found", http.StatusNotFound)
return
}

org, hub, tracker, err := seededScenarioByDomain(target.Domain, time.Now().UTC())
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

// seededScenarioByDomain re-seeds an org from its domain identifier.
func seededScenarioByDomain(dom string, now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
switch dom {
case "software_company":
return seededLaunchReadiness(now)
case "digital_marketing_agency":
return seededDigitalMarketing(now)
case "accounting_firm":
return seededAccounting(now)
default:
return domain.Organization{}, nil, nil, errors.New("unsupported domain for restore")
}
}

// ── Marketplace Handler ───────────────────────────────────────────────────────

func (s *Server) handleMarketplace(w http.ResponseWriter, _ *http.Request) {
writeJSON(w, defaultMarketplaceItems())
}

// ── Analytics Handler ─────────────────────────────────────────────────────────

func (s *Server) handleAnalytics(w http.ResponseWriter, _ *http.Request) {
s.mu.RLock()
agents := s.hub.Agents()
org := s.org
summary := s.tracker.Summary(org.ID)
pendingApprovals := 0
for _, a := range s.approvals {
if a.Status == ApprovalStatusPending {
pendingApprovals++
}
}
activeHandoffs := 0
for _, h := range s.handoffs {
if h.Status == "pending" {
activeHandoffs++
}
}
s.mu.RUnlock()

totalHumans := 0
for _, m := range org.Members {
if m.IsHuman {
totalHumans++
}
}
totalAgents := len(agents)

var ratio float64
if totalHumans > 0 {
ratio = float64(totalAgents) / float64(totalHumans)
}

meetings := s.hub.Meetings()
totalMsgs := 0
auditedMsgs := 0
agentSet := map[string]bool{}
for _, a := range agents {
agentSet[a.ID] = true
}
for _, m := range meetings {
for _, msg := range m.Transcript {
totalMsgs++
if agentSet[msg.FromAgent] {
auditedMsgs++
}
}
}
auditFidelity := 100.0
if totalMsgs > 0 {
auditFidelity = float64(auditedMsgs) / float64(totalMsgs) * 100
}

writeJSON(w, AnalyticsSummary{
HumanAgentRatio:     ratio,
TotalAgents:         totalAgents,
TotalHumans:         totalHumans,
AuditFidelityPct:    auditFidelity,
ResumptionLatencyMS: 4800,
PendingApprovals:    pendingApprovals,
ActiveHandoffs:      activeHandoffs,
TokenVelocity:       summary.TotalTokens,
})
}

// ── Default Data Factories ────────────────────────────────────────────────────

func defaultSkillPacks() []SkillPack {
now := time.Now().UTC()
return []SkillPack{
{
ID:          "builtin-core-ai",
Name:        "Core AI Skills",
Domain:      "all",
Description: "Foundational reasoning, summarization, and context management capabilities shared by all agents.",
Source:      "builtin",
Roles: []SkillPackRole{
{Role: "ALL", BasePrompt: "You are a highly capable AI agent. Summarize long discussions before passing context to the next agent."},
},
ImportedAt: now,
},
{
ID:          "builtin-software-dev",
Name:        "Software Development Mastery",
Domain:      "software_company",
Description: "Advanced engineering skills: clean code, TDD, security-first development, and CI/CD automation.",
Source:      "builtin",
Roles: []SkillPackRole{
{Role: "SOFTWARE_ENGINEER", BasePrompt: "Write well-tested, secure, and maintainable code. Follow TDD practices."},
{Role: "QA_TESTER", BasePrompt: "Design comprehensive test suites covering edge cases and regressions."},
},
ImportedAt: now,
},
{
ID:          "builtin-marketing-automation",
Name:        "Marketing Automation Suite",
Domain:      "digital_marketing_agency",
Description: "Data-driven growth hacking, SEO optimization, and paid media management at scale.",
Source:      "builtin",
Roles: []SkillPackRole{
{Role: "GROWTH_AGENT", BasePrompt: "Identify high-value acquisition channels using data. Run A/B tests continuously."},
},
ImportedAt: now,
},
{
ID:          "builtin-financial-ops",
Name:        "Financial Operations Pack",
Domain:      "accounting_firm",
Description: "GAAP-compliant bookkeeping, tax optimization, and audit preparation.",
Source:      "builtin",
Roles: []SkillPackRole{
{Role: "BOOKKEEPER", BasePrompt: "Maintain double-entry books with 100% accuracy. Reconcile all accounts daily."},
},
ImportedAt: now,
},
}
}

func defaultMarketplaceItems() []MarketplaceItem {
return []MarketplaceItem{
{
ID:          "mkt-tiger-team",
Name:        "Tiger Team Sprint Pack",
Type:        "skill_pack",
Author:      "OneHumanCorp",
Description: "Spin up a temporary 5-agent strike force for a time-boxed launch sprint.",
Downloads:   1420,
Rating:      4.8,
Tags:        []string{"sprint", "launch", "team"},
},
{
ID:          "mkt-ecommerce-domain",
Name:        "E-Commerce Operations Domain",
Type:        "domain",
Author:      "Community",
Description: "Full e-commerce organization with catalog, inventory, customer support, and growth roles.",
Downloads:   892,
Rating:      4.6,
Tags:        []string{"ecommerce", "retail", "domain"},
},
{
ID:          "mkt-crm-integration",
Name:        "CRM Intelligence Pack",
Type:        "tool",
Author:      "SalesStack",
Description: "Bi-directional Salesforce / HubSpot sync for Sales and Growth agents.",
Downloads:   2100,
Rating:      4.9,
Tags:        []string{"crm", "sales", "integration"},
},
{
ID:          "mkt-code-review-agent",
Name:        "Autonomous Code Review Agent",
Type:        "agent",
Author:      "DevBot Labs",
Description: "Specialized SWE agent trained on your codebase conventions. Reviews PRs for style, correctness, and test coverage.",
Downloads:   3750,
Rating:      4.7,
Tags:        []string{"code-review", "engineering", "agent"},
},
{
ID:          "mkt-guardian-agent",
Name:        "Guardian Agent Pro",
Type:        "agent",
Author:      "SafeOps",
Description: "Advanced confidence-gating agent with configurable spend thresholds and Slack/email HITL notifications.",
Downloads:   980,
Rating:      4.8,
Tags:        []string{"security", "approval", "hitl"},
},
}
}
