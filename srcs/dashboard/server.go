package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/integrations"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

// Server encapsulates the HTTP handlers and state for the One Human Corp dashboard.
//
// Constraints: Must be instantiated with a valid domain.Organization, orchestration.Hub, and billing.Tracker.
type Server struct {
	mu              sync.RWMutex
	org             domain.Organization
	hub             *orchestration.Hub
	tracker         *billing.Tracker
	approvals       []ApprovalRequest
	handoffs        []HandoffPackage
	skills          []SkillPack
	snapshots       []OrgSnapshot
	integReg        *integrations.Registry
	trustAgreements []TrustAgreement
	b2bGateway      *orchestration.B2BGateway
	incidents       []Incident
	computeProfiles []ComputeProfile
	budgetAlerts    []BudgetAlert
	pipelines       []Pipeline
	authStore       *auth.Store
	authHandlers    *auth.Handlers
	settings        Settings
}

type Settings struct {
	MinimaxAPIKey string `json:"minimaxApiKey"`
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
	{ID: "linear-mcp", Name: "Linear", Description: "Modern issue tracking: manage issues, cycles, and roadmaps for high-velocity teams.", Category: "project_management", Status: "available"},
	{ID: "figma-mcp", Name: "Figma", Description: "Design file access: read wireframes, export assets, inspect component specs.", Category: "design", Status: "available"},
	{ID: "aws-mcp", Name: "AWS", Description: "Cloud infrastructure: provision EC2 instances, manage S3, deploy Lambda functions.", Category: "infrastructure", Status: "available"},
	{ID: "gcp-mcp", Name: "Google Cloud Platform", Description: "Cloud infrastructure: manage GCE instances, Cloud Storage, Cloud Run, and GKE clusters.", Category: "infrastructure", Status: "available"},
	{ID: "azure-mcp", Name: "Microsoft Azure", Description: "Cloud infrastructure: provision VMs, manage Azure Blob Storage, deploy Azure Functions.", Category: "infrastructure", Status: "available"},
	{ID: "kubernetes-mcp", Name: "Kubernetes", Description: "Container orchestration: deploy workloads, scale pods, inspect cluster health.", Category: "infrastructure", Status: "available"},
	{ID: "slack-mcp", Name: "Slack / Mattermost", Description: "Human-in-the-loop approval: send HITL notifications, await human manager sign-off.", Category: "communication", Status: "available"},
	{ID: "telegram-mcp", Name: "Telegram", Description: "Agent messaging via Telegram bots: send notifications and collect HITL responses.", Category: "communication", Status: "available"},
	{ID: "teams-mcp", Name: "Microsoft Teams", Description: "Agent messaging via Teams webhooks: post updates and await approval from human managers.", Category: "communication", Status: "available"},
	{ID: "postgres-mcp", Name: "PostgreSQL", Description: "Database operations: run queries, manage schema, inspect table data.", Category: "database", Status: "available"},
	{ID: "mysql-mcp", Name: "MySQL", Description: "Database operations: run queries, manage schema, and inspect MySQL or MariaDB table data.", Category: "database", Status: "available"},
	{ID: "redis-mcp", Name: "Redis", Description: "In-memory data store: manage keys, queues, pub/sub channels, and caching layers.", Category: "database", Status: "available"},
	{ID: "opentelemetry-mcp", Name: "OpenTelemetry", Description: "Observability: push metrics and traces to Grafana / OpenObserve dashboards.", Category: "observability", Status: "available"},
	{ID: "datadog-mcp", Name: "Datadog", Description: "Monitoring and APM: query metrics, manage monitors, and inspect distributed traces.", Category: "observability", Status: "available"},
	{ID: "sentry-mcp", Name: "Sentry", Description: "Error tracking: capture exceptions, triage issues, and link errors to code changes.", Category: "observability", Status: "available"},
	{ID: "github-actions-mcp", Name: "GitHub Actions", Description: "CI/CD pipelines: trigger workflow runs, inspect job logs, and manage deployment environments.", Category: "cicd", Status: "available"},
	{ID: "notion-mcp", Name: "Notion", Description: "Knowledge base: read and write pages, manage databases, and retrieve structured documentation.", Category: "knowledge", Status: "available"},
	{ID: "spire-mcp", Name: "SPIFFE/SPIRE", Description: "Identity management: issue and rotate SVID certificates for agent workloads.", Category: "identity", Status: "available"},
}

var statusOrder = []orchestration.Status{
	orchestration.StatusActive,
	orchestration.StatusBlocked,
	orchestration.StatusIdle,
	orchestration.StatusInMeeting,
}

// NewServer initializes a new Dashboard HTTP handler that routes all API and frontend requests.
//
// Parameters:
//   - org: domain.Organization; The base organizational structure.
//   - hub: *orchestration.Hub; The agent communication and meeting room registry.
//   - tracker: *billing.Tracker; The cost and token tracking engine.
//
// Returns: An http.Handler that serves the dashboard REST APIs and static React frontend.
func NewServer(org domain.Organization, hub *orchestration.Hub, tracker *billing.Tracker, authStore ...*auth.Store) http.Handler {
	var store *auth.Store
	if len(authStore) > 0 && authStore[0] != nil {
		store = authStore[0]
	} else {
		store = auth.NewStore()
	}
	server := &Server{
		org:             org,
		hub:             hub,
		tracker:         tracker,
		approvals:       []ApprovalRequest{},
		handoffs:        []HandoffPackage{},
		skills:          defaultSkillPacks(),
		snapshots:       []OrgSnapshot{},
		integReg:        integrations.NewRegistry(),
		trustAgreements: []TrustAgreement{},
		b2bGateway:      orchestration.NewB2BGateway(hub),
		incidents:       []Incident{},
		computeProfiles: []ComputeProfile{},
		budgetAlerts:    []BudgetAlert{},
		pipelines:       []Pipeline{},
		authStore:       store,
		authHandlers:    auth.NewHandlers(store),
	}
	// Load Minimax API key from environment on startup.
	if key := os.Getenv("MINIMAX_API_KEY"); key != "" {
		hub.SetMinimaxAPIKey(key)
		server.settings.MinimaxAPIKey = key
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/api/dashboard", server.handleDashboard)
	mux.HandleFunc("/api/org", server.handleOrg)
	mux.HandleFunc("/api/meetings", server.handleMeetings)
	mux.HandleFunc("/api/costs", server.handleCosts)
	mux.HandleFunc("/api/messages", server.handleSendMessage)
	mux.HandleFunc("/api/agents/hire", server.handleHireAgent)
	mux.HandleFunc("/api/agents/fire", server.handleFireAgent)
	mux.HandleFunc("/api/domains", server.handleDomains)
	mux.HandleFunc("/api/mcp/tools", server.handleMCPTools)
	mux.HandleFunc("/api/mcp/tools/invoke", server.handleMCPInvoke)
	mux.HandleFunc("/api/dev/seed", server.handleDevSeed)
	mux.HandleFunc("/api/settings", server.handleSettings)
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
	// Phase 2 – External Integrations (chat, git, issues)
	mux.HandleFunc("/api/integrations", server.handleIntegrations)
	mux.HandleFunc("/api/integrations/connect", server.handleIntegrationConnect)
	mux.HandleFunc("/api/integrations/disconnect", server.handleIntegrationDisconnect)
	mux.HandleFunc("/api/integrations/chat/messages", server.handleChatMessages)
	mux.HandleFunc("/api/integrations/chat/send", server.handleChatSend)
	mux.HandleFunc("/api/integrations/chat/test", server.handleChatTest)
	mux.HandleFunc("/api/integrations/git/prs", server.handlePullRequests)
	mux.HandleFunc("/api/integrations/git/pr/create", server.handlePRCreate)
	mux.HandleFunc("/api/integrations/git/pr/merge", server.handlePRMerge)
	mux.HandleFunc("/api/integrations/git/pr/close", server.handlePRClose)
	mux.HandleFunc("/api/integrations/issues", server.handleIssues)
	mux.HandleFunc("/api/integrations/issues/create", server.handleIssueCreate)
	mux.HandleFunc("/api/integrations/issues/status", server.handleIssueUpdateStatus)
	mux.HandleFunc("/api/integrations/issues/assign", server.handleIssueAssign)
	// Phase 5 – B2B Cross-Org Collaboration
	mux.HandleFunc("/api/b2b/agreements", server.handleB2BAgreements)
	mux.HandleFunc("/api/b2b/handshake", server.handleB2BHandshake)
	mux.HandleFunc("/api/b2b/revoke", server.handleB2BRevoke)
	mux.HandleFunc("/api/b2b/tunnel", server.b2bGateway.HandleB2BEndpoint)
	// Phase 5 – Autonomous SRE / Incident Management
	mux.HandleFunc("/api/incidents", server.handleIncidents)
	mux.HandleFunc("/api/incidents/status", server.handleIncidentStatus)
	// Phase 5 – Compute Optimisation / Hardware-Aware Scheduling
	mux.HandleFunc("/api/compute/profiles", server.handleComputeProfiles)
	mux.HandleFunc("/api/clusters/", server.handleClusterStatus)
	// Phase 5 – Budget Alerts
	mux.HandleFunc("/api/billing/alerts", server.handleBudgetAlerts)
	// Phase 5 – Automated SDLC / Pipelines
	mux.HandleFunc("/api/pipelines", server.handlePipelines)
	mux.HandleFunc("/api/pipelines/promote", server.handlePipelinePromote)
	mux.HandleFunc("/api/pipelines/status", server.handlePipelineStatus)
	// Auth – login / logout / current user
	mux.HandleFunc("/api/auth/login", server.authHandlers.HandleLogin)
	mux.HandleFunc("/api/auth/logout", server.authHandlers.HandleLogout)
	mux.HandleFunc("/api/auth/me", server.authHandlers.HandleMe)
	// User management (admin only)
	mux.HandleFunc("/api/users", server.authHandlers.HandleUsers)
	mux.HandleFunc("/api/users/", server.authHandlers.HandleUser)
	// Role management
	mux.HandleFunc("/api/roles", server.authHandlers.HandleRoles)
	// Health / readiness probes
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", telemetry.MetricsHandler())
	return telemetry.Middleware(auth.Middleware(store)(mux))
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>One Human Corp Dashboard</title>
</head>
<body>
  <h1>One Human Corp Dashboard</h1>
  <p>Send Message to an agent or meeting room using the API.</p>
  <p>View Role Playbooks and agent skill sets in the Settings panel.</p>
  <div id="root"></div>
</body>
</html>`

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(indexHTML))
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
		telemetry.RecordHumanInteraction(r.Context(), "message")
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

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r.Method == http.MethodGet {
		writeJSON(w, s.settings)
		return
	}

	if r.Method == http.MethodPost {
		var req Settings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		s.settings = req
		s.hub.SetMinimaxAPIKey(req.MinimaxAPIKey)
		writeJSON(w, s.settings)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	hub.RegisterAgent(orchestration.Agent{ID: "CEO", Name: "Human CEO", Role: "CEO", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("launch-readiness", "Review launch blockers, sign-off on reliability checklist, assign post-launch owners.", []string{"pm-1", "swe-1", "ux-1", "CEO"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       orchestration.EventTask,
		Content:    "Ship the reliability checklist before launch.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-6 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-2",
		FromAgent:  "swe-1",
		ToAgent:    "pm-1",
		Type:       orchestration.EventStatus,
		Content:    "Checklist is 90% complete. Waiting on design assets for the final error states.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-4 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-3",
		FromAgent:  "ux-1",
		ToAgent:    "pm-1",
		Type:       orchestration.EventStatus,
		Content:    "Design QA pass completed with no blockers. Assets pushed to main.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-2 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-4",
		FromAgent:  "CEO",
		ToAgent:    "pm-1",
		Type:       orchestration.EventApprovalNeeded,
		Content:    "Looks good. Proceed with the final staging deployment, but keep a close eye on the latency metrics.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-1 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "pm-1",
		AgentRole:        "PRODUCT_MANAGER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o-mini",
		PromptTokens:     1200,
		CompletionTokens: 400,
		OccurredAt:       now.Add(-10 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		AgentRole:        "SOFTWARE_ENGINEER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     2600,
		CompletionTokens: 900,
		OccurredAt:       now.Add(-8 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "ux-1",
		AgentRole:        "DESIGNER",
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
		ID:         "seed-dma-1",
		FromAgent:  "growth-1",
		ToAgent:    "content-1",
		Type:       orchestration.EventTask,
		Content:    "Draft top-of-funnel blog series targeting enterprise SaaS buyers.",
		MeetingID:  "campaign-kickoff",
		OccurredAt: now.Add(-5 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "growth-1",
		AgentRole:        "GROWTH_AGENT",
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
		ID:         "seed-acc-1",
		FromAgent:  "cfo-1",
		ToAgent:    "bookkeeper-1",
		Type:       orchestration.EventTask,
		Content:    "Reconcile bank feeds and categorize uncategorized transactions before EOD.",
		MeetingID:  "month-close",
		OccurredAt: now.Add(-3 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "cfo-1",
		AgentRole:        "CFO",
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

// ── Chat test handler ─────────────────────────────────────────────────────────

type chatTestRequest struct {
	IntegrationID string `json:"integrationId"`
	BotToken      string `json:"botToken,omitempty"`
	ChatID        string `json:"chatId,omitempty"`
	WebhookURL    string `json:"webhookUrl,omitempty"`
	APIToken      string `json:"apiToken,omitempty"`
}

func (s *Server) handleChatTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

// ── MCP tool invocation ───────────────────────────────────────────────────────

type mcpInvokeRequest struct {
	ToolID string         `json:"toolId"`
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

func (s *Server) handleMCPInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req mcpInvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.ToolID == "" {
		http.Error(w, "toolId is required", http.StatusBadRequest)
		return
	}
	if req.Params == nil {
		req.Params = map[string]any{}
	}
	result, err := s.invokeMCPTool(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

func (s *Server) invokeMCPTool(req mcpInvokeRequest) (map[string]any, error) {
	getString := func(key string) string {
		if v, ok := req.Params[key]; ok {
			if str, ok := v.(string); ok {
				return str
			}
		}
		return ""
	}

	switch req.ToolID {
	// ── Communication tools ───────────────────────────────────────────────────
	case "telegram-mcp", "slack-mcp", "teams-mcp":
		integrationID := getString("integrationId")
		if integrationID == "" {
			switch req.ToolID {
			case "telegram-mcp":
				integrationID = "telegram"
			case "slack-mcp":
				integrationID = "slack"
			case "teams-mcp":
				integrationID = "teams"
			}
		}
		channel := getString("channel")
		fromAgent := getString("fromAgent")
		content := getString("content")
		threadID := getString("threadId")

		if content == "" {
			return nil, errors.New("content is required")
		}
		if fromAgent == "" {
			fromAgent = "system"
		}
		// Fall back to the configured chatspace if no channel given.
		if channel == "" {
			if integ, ok := s.integReg.Integration(integrationID); ok {
				channel = integ.Chatspace
			}
		}
		if channel == "" {
			return nil, errors.New("channel is required — configure the integration's chatspace first")
		}
		msg, err := s.integReg.SendChatMessage(integrationID, channel, fromAgent, content, threadID, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"message": msg, "delivered": true}, nil

	// ── Git tools ─────────────────────────────────────────────────────────────
	case "git-mcp":
		integrationID := getString("integrationId")
		if integrationID == "" {
			integrationID = "github"
		}
		repo := getString("repository")
		title := getString("title")
		body := getString("body")
		source := getString("sourceBranch")
		target := getString("targetBranch")
		createdBy := getString("createdBy")
		if target == "" {
			target = "main"
		}
		pr, err := s.integReg.CreatePullRequest(integrationID, repo, title, body, source, target, createdBy, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"pullRequest": pr}, nil

	// ── Issue tracker tools ───────────────────────────────────────────────────
	case "jira-mcp", "linear-mcp":
		integrationID := getString("integrationId")
		if integrationID == "" {
			if req.ToolID == "jira-mcp" {
				integrationID = "jira"
			} else {
				integrationID = "linear"
			}
		}
		project := getString("project")
		title := getString("title")
		description := getString("description")
		createdBy := getString("createdBy")
		priority := getString("priority")
		issue, err := s.integReg.CreateIssue(integrationID, project, title, description, createdBy,
			integrations.IssuePriority(priority), nil, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"issue": issue}, nil

	// ── Unimplemented tools — return a structured acknowledgement ─────────────
	default:
		return map[string]any{
			"toolId":  req.ToolID,
			"status":  "invoked",
			"message": "Tool invocation recorded. Connect the corresponding service integration to enable live execution.",
		}, nil
	}
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

// ── Integration request/response types ────────────────────────────────────────

type integrationConnectRequest struct {
	IntegrationID string `json:"integrationId"`
	BaseURL       string `json:"baseUrl,omitempty"`
	// Chat credentials — stored server-side, never returned to the client.
	BotToken   string `json:"botToken,omitempty"`
	ChatID     string `json:"chatId,omitempty"`
	WebhookURL string `json:"webhookUrl,omitempty"`
	APIToken   string `json:"apiToken,omitempty"`
}

type integrationDisconnectRequest struct {
	IntegrationID string `json:"integrationId"`
}

type chatSendRequest struct {
	IntegrationID string `json:"integrationId"`
	Channel       string `json:"channel"`
	FromAgent     string `json:"fromAgent"`
	Content       string `json:"content"`
	ThreadID      string `json:"threadId,omitempty"`
}

type prCreateRequest struct {
	IntegrationID string `json:"integrationId"`
	Repository    string `json:"repository"`
	Title         string `json:"title"`
	Body          string `json:"body,omitempty"`
	SourceBranch  string `json:"sourceBranch"`
	TargetBranch  string `json:"targetBranch"`
	CreatedBy     string `json:"createdBy,omitempty"`
}

type prActionRequest struct {
	PRID string `json:"prId"`
}

type issueCreateRequest struct {
	IntegrationID string   `json:"integrationId"`
	Project       string   `json:"project"`
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	CreatedBy     string   `json:"createdBy,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	Labels        []string `json:"labels,omitempty"`
}

type issueStatusRequest struct {
	IssueID string `json:"issueId"`
	Status  string `json:"status"`
}

type issueAssignRequest struct {
	IssueID  string `json:"issueId"`
	Assignee string `json:"assignee"`
}

// ── Integration handlers ──────────────────────────────────────────────────────

func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	category := r.URL.Query().Get("category")
	if category != "" {
		writeJSON(w, s.integReg.IntegrationsByCategory(integrations.Category(category)))
		return
	}
	writeJSON(w, s.integReg.Integrations())
}

func (s *Server) handleIntegrationConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req integrationConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	updated, err := s.integReg.Connect(req.IntegrationID, req.BaseURL, creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, updated)
}

func (s *Server) handleIntegrationDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req integrationDisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IntegrationID == "" {
		http.Error(w, "integrationId is required", http.StatusBadRequest)
		return
	}
	updated, err := s.integReg.Disconnect(req.IntegrationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, updated)
}

// ── Chat handlers ─────────────────────────────────────────────────────────────

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

func (s *Server) handleChatSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

// ── Git handlers ──────────────────────────────────────────────────────────────

func (s *Server) handlePullRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	integrationID := r.URL.Query().Get("integrationId")
	prs := s.integReg.PullRequests(integrationID)
	if prs == nil {
		prs = []integrations.PullRequest{}
	}
	writeJSON(w, prs)
}

func (s *Server) handlePRCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.CreatePullRequest(req.IntegrationID, req.Repository, req.Title, req.Body, req.SourceBranch, req.TargetBranch, req.CreatedBy, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

func (s *Server) handlePRMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PRID == "" {
		http.Error(w, "prId is required", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.MergePullRequest(req.PRID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

func (s *Server) handlePRClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PRID == "" {
		http.Error(w, "prId is required", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.ClosePullRequest(req.PRID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

// ── Issue tracker handlers ────────────────────────────────────────────────────

func (s *Server) handleIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	integrationID := r.URL.Query().Get("integrationId")
	issues := s.integReg.Issues(integrationID)
	if issues == nil {
		issues = []integrations.Issue{}
	}
	writeJSON(w, issues)
}

func (s *Server) handleIssueCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.CreateIssue(req.IntegrationID, req.Project, req.Title, req.Description, req.CreatedBy, integrations.IssuePriority(req.Priority), req.Labels, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, issue)
}

func (s *Server) handleIssueUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IssueID == "" || req.Status == "" {
		http.Error(w, "issueId and status are required", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.UpdateIssueStatus(req.IssueID, integrations.IssueStatus(req.Status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, issue)
}

func (s *Server) handleIssueAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IssueID == "" || req.Assignee == "" {
		http.Error(w, "issueId and assignee are required", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.AssignIssue(req.IssueID, req.Assignee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, issue)
}

// ── B2B Collaboration ─────────────────────────────────────────────────────────

// TrustAgreementStatus represents the lifecycle of a B2B trust agreement.
type TrustAgreementStatus string

const (
	TrustStatusPending TrustAgreementStatus = "PENDING"
	TrustStatusActive  TrustAgreementStatus = "ACTIVE"
	TrustStatusRevoked TrustAgreementStatus = "REVOKED"
)

// TrustAgreement is a federated trust relationship between two OHC organisations.
// It enables cross-org agent collaboration using SPIFFE-federated JWTs.
type TrustAgreement struct {
	ID           string               `json:"id"`
	PartnerOrg   string               `json:"partnerOrg"`
	PartnerJWKS  string               `json:"partnerJwksUrl"`
	AllowedRoles []string             `json:"allowedRoles"`
	Status       TrustAgreementStatus `json:"status"`
	CreatedAt    time.Time            `json:"createdAt"`
}

type b2bHandshakeRequest struct {
	PartnerOrg   string   `json:"partnerOrg"`
	PartnerJWKS  string   `json:"partnerJwksUrl"`
	AllowedRoles []string `json:"allowedRoles"`
}

func (s *Server) handleB2BAgreements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		agreements := append([]TrustAgreement(nil), s.trustAgreements...)
		s.mu.RUnlock()
		writeJSON(w, agreements)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleB2BHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req b2bHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PartnerOrg == "" || req.PartnerJWKS == "" {
		http.Error(w, "partnerOrg and partnerJwksUrl are required", http.StatusBadRequest)
		return
	}

	agreement := TrustAgreement{
		ID:           "ta-" + strings.ReplaceAll(req.PartnerOrg, ".", "-") + "-" + time.Now().Format("20060102150405"),
		PartnerOrg:   req.PartnerOrg,
		PartnerJWKS:  req.PartnerJWKS,
		AllowedRoles: req.AllowedRoles,
		Status:       TrustStatusActive,
		CreatedAt:    time.Now().UTC(),
	}

	s.mu.Lock()
	s.trustAgreements = append(s.trustAgreements, agreement)
	s.mu.Unlock()

	s.b2bGateway.AddAgreement(domain.TrustAgreement{
		ID: agreement.ID,
		PartnerOrg: agreement.PartnerOrg,
		PartnerJWKS: agreement.PartnerJWKS,
		AllowedRoles: agreement.AllowedRoles,
		Status: string(agreement.Status),
	})
	writeJSON(w, agreement)
}

func (s *Server) handleB2BRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		AgreementID string `json:"agreementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.AgreementID == "" {
		http.Error(w, "agreementId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, ag := range s.trustAgreements {
		if ag.ID == req.AgreementID {
			s.trustAgreements[i].Status = TrustStatusRevoked
			s.b2bGateway.RemoveAgreement(ag.PartnerOrg)
			writeJSON(w, s.trustAgreements[i])
			return
		}
	}
	http.Error(w, "agreement not found", http.StatusNotFound)
}

// ── Autonomous SRE / Incident Management ─────────────────────────────────────

// IncidentSeverity classifies the urgency of an operational incident.
type IncidentSeverity string

const (
	SeverityP0 IncidentSeverity = "P0"
	SeverityP1 IncidentSeverity = "P1"
	SeverityP2 IncidentSeverity = "P2"
)

// IncidentStatus reflects the investigation lifecycle state.
type IncidentStatus string

const (
	IncidentStatusInvestigating IncidentStatus = "INVESTIGATING"
	IncidentStatusProposed      IncidentStatus = "PROPOSED"
	IncidentStatusResolved      IncidentStatus = "RESOLVED"
)

// Incident represents an operational event requiring SRE attention.
type Incident struct {
	ID               string           `json:"id"`
	Severity         IncidentSeverity `json:"severity"`
	Summary          string           `json:"summary"`
	RCA              string           `json:"rootCauseAnalysis"`
	ResolutionPlanID string           `json:"resolutionPlanId,omitempty"`
	Status           IncidentStatus   `json:"status"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

type incidentCreateRequest struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	RCA      string `json:"rootCauseAnalysis,omitempty"`
}

type incidentStatusRequest struct {
	IncidentID       string `json:"incidentId"`
	Status           string `json:"status"`
	ResolutionPlanID string `json:"resolutionPlanId,omitempty"`
	RCA              string `json:"rootCauseAnalysis,omitempty"`
}

func (s *Server) handleIncidents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		incidents := append([]Incident(nil), s.incidents...)
		s.mu.RUnlock()
		writeJSON(w, incidents)
	case http.MethodPost:
		var req incidentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Severity == "" || req.Summary == "" {
			http.Error(w, "severity and summary are required", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		incident := Incident{
			ID:        "inc-" + now.Format("20060102150405"),
			Severity:  IncidentSeverity(req.Severity),
			Summary:   req.Summary,
			RCA:       req.RCA,
			Status:    IncidentStatusInvestigating,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.mu.Lock()
		s.incidents = append(s.incidents, incident)
		s.mu.Unlock()
		writeJSON(w, incident)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleIncidentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req incidentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IncidentID == "" || req.Status == "" {
		http.Error(w, "incidentId and status are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, inc := range s.incidents {
		if inc.ID == req.IncidentID {
			s.incidents[i].Status = IncidentStatus(req.Status)
			s.incidents[i].UpdatedAt = time.Now().UTC()
			if req.ResolutionPlanID != "" {
				s.incidents[i].ResolutionPlanID = req.ResolutionPlanID
			}
			if req.RCA != "" {
				s.incidents[i].RCA = req.RCA
			}
			writeJSON(w, s.incidents[i])
			return
		}
	}
	http.Error(w, "incident not found", http.StatusNotFound)
}

// ── Compute Optimization / Hardware-Aware Scheduling ─────────────────────────

// ComputeProfile defines the hardware requirements for a given agent role.
type ComputeProfile struct {
	RoleID             string    `json:"roleId"`
	MinVRAMGB          int       `json:"minVramGb"`
	PreferredGPUType   string    `json:"preferredGpuType"` // "h100", "a10g", "cpu"
	SchedulingPriority int       `json:"schedulingPriority"`
	CreatedAt          time.Time `json:"createdAt"`
}

type computeProfileRequest struct {
	RoleID             string `json:"roleId"`
	MinVRAMGB          int    `json:"minVramGb"`
	PreferredGPUType   string `json:"preferredGpuType"`
	SchedulingPriority int    `json:"schedulingPriority"`
}

// ClusterStatus reflects the health of a remote Kubernetes cluster region.
type ClusterStatus struct {
	Region         string    `json:"region"`
	Status         string    `json:"status"` // healthy, degraded, offline
	LatencyMS      int       `json:"latencyMs"`
	AvailableNodes int       `json:"availableNodes"`
	CheckedAt      time.Time `json:"checkedAt"`
}

func (s *Server) handleComputeProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		profiles := append([]ComputeProfile(nil), s.computeProfiles...)
		s.mu.RUnlock()
		writeJSON(w, profiles)
	case http.MethodPost:
		var req computeProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.RoleID == "" {
			http.Error(w, "roleId is required", http.StatusBadRequest)
			return
		}
		profile := ComputeProfile{
			RoleID:             req.RoleID,
			MinVRAMGB:          req.MinVRAMGB,
			PreferredGPUType:   req.PreferredGPUType,
			SchedulingPriority: req.SchedulingPriority,
			CreatedAt:          time.Now().UTC(),
		}
		s.mu.Lock()
		s.computeProfiles = append(s.computeProfiles, profile)
		s.mu.Unlock()
		writeJSON(w, profile)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract region from URL path: /api/clusters/{region}/status
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	region := ""
	for i, p := range parts {
		if p == "clusters" && i+1 < len(parts) {
			region = parts[i+1]
			break
		}
	}
	if region == "" {
		http.Error(w, "region is required in path", http.StatusBadRequest)
		return
	}
	// Simulated cluster health response (would call k8s API in production)
	status := ClusterStatus{
		Region:         region,
		Status:         "healthy",
		LatencyMS:      3,
		AvailableNodes: 5,
		CheckedAt:      time.Now().UTC(),
	}
	writeJSON(w, status)
}

// ── Budget Alerts ─────────────────────────────────────────────────────────────

// defaultBudgetAlertNotifyPct is the default notification threshold (80 %).
const defaultBudgetAlertNotifyPct = 0.8

// BudgetAlert defines a spending threshold with notification behaviour.
type BudgetAlert struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ThresholdUSD   float64   `json:"thresholdUsd"`
	NotifyAtPct    float64   `json:"notifyAtPct"` // e.g. 0.8 → notify at 80 %
	Triggered      bool      `json:"triggered"`
	CreatedAt      time.Time `json:"createdAt"`
}

type budgetAlertRequest struct {
	OrganizationID string  `json:"organizationId"`
	ThresholdUSD   float64 `json:"thresholdUsd"`
	NotifyAtPct    float64 `json:"notifyAtPct"`
}

func (s *Server) handleBudgetAlerts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		alerts := append([]BudgetAlert(nil), s.budgetAlerts...)
		s.mu.RUnlock()
		// Evaluate triggered state against current spend.
		summary := s.tracker.Summary(s.org.ID)
		for i, a := range alerts {
			alerts[i].Triggered = summary.TotalCostUSD >= a.ThresholdUSD*a.NotifyAtPct
		}
		writeJSON(w, alerts)
	case http.MethodPost:
		var req budgetAlertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.ThresholdUSD <= 0 {
			http.Error(w, "thresholdUsd must be greater than zero", http.StatusBadRequest)
			return
		}
		if req.NotifyAtPct <= 0 || req.NotifyAtPct > 1 {
			req.NotifyAtPct = defaultBudgetAlertNotifyPct // default 80 %
		}
		orgID := req.OrganizationID
		if orgID == "" {
			s.mu.RLock()
			orgID = s.org.ID
			s.mu.RUnlock()
		}
		alert := BudgetAlert{
			ID:             "alert-" + time.Now().Format("20060102150405"),
			OrganizationID: orgID,
			ThresholdUSD:   req.ThresholdUSD,
			NotifyAtPct:    req.NotifyAtPct,
			Triggered:      false,
			CreatedAt:      time.Now().UTC(),
		}
		s.mu.Lock()
		s.budgetAlerts = append(s.budgetAlerts, alert)
		s.mu.Unlock()
		writeJSON(w, alert)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ── Automated SDLC / Pipelines ────────────────────────────────────────────────

// PipelineStatus reflects the lifecycle of an autonomous CI/CD pipeline.
type PipelineStatus string

const (
	PipelineStatusPending      PipelineStatus = "PENDING"
	PipelineStatusImplementing PipelineStatus = "IMPLEMENTING"
	PipelineStatusTesting      PipelineStatus = "TESTING"
	PipelineStatusStaging      PipelineStatus = "STAGING"
	PipelineStatusPromoted     PipelineStatus = "PROMOTED"
	PipelineStatusFailed       PipelineStatus = "FAILED"
)

// Pipeline represents an autonomous implementation pipeline from spec to production.
type Pipeline struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Status      PipelineStatus `json:"status"`
	Branch      string         `json:"branch"`
	StagingURL  string         `json:"stagingUrl,omitempty"`
	InitiatedBy string         `json:"initiatedBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type pipelineCreateRequest struct {
	Name        string `json:"name"`
	Branch      string `json:"branch"`
	InitiatedBy string `json:"initiatedBy"`
}

type pipelinePromoteRequest struct {
	PipelineID string `json:"pipelineId"`
	ApprovedBy string `json:"approvedBy"`
}

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		pipelines := append([]Pipeline(nil), s.pipelines...)
		s.mu.RUnlock()
		writeJSON(w, pipelines)
	case http.MethodPost:
		var req pipelineCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		pipeline := Pipeline{
			ID:          "pipeline-" + now.Format("20060102150405"),
			Name:        req.Name,
			Status:      PipelineStatusPending,
			Branch:      req.Branch,
			InitiatedBy: req.InitiatedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		s.mu.Lock()
		s.pipelines = append(s.pipelines, pipeline)
		s.mu.Unlock()
		writeJSON(w, pipeline)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePipelinePromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pipelinePromoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PipelineID == "" {
		http.Error(w, "pipelineId is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pipelines {
		if p.ID == req.PipelineID {
			if s.pipelines[i].Status != PipelineStatusStaging {
				http.Error(w, "pipeline must be in STAGING status to promote", http.StatusBadRequest)
				return
			}
			s.pipelines[i].Status = PipelineStatusPromoted
			s.pipelines[i].UpdatedAt = time.Now().UTC()
			writeJSON(w, s.pipelines[i])
			return
		}
	}
	http.Error(w, "pipeline not found", http.StatusNotFound)
}

func (s *Server) handlePipelineStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PipelineID string `json:"pipelineId"`
		Status     string `json:"status"`
		StagingURL string `json:"stagingUrl,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PipelineID == "" || req.Status == "" {
		http.Error(w, "pipelineId and status are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pipelines {
		if p.ID == req.PipelineID {
			s.pipelines[i].Status = PipelineStatus(req.Status)
			s.pipelines[i].UpdatedAt = time.Now().UTC()
			if req.StagingURL != "" {
				s.pipelines[i].StagingURL = req.StagingURL
			}
			writeJSON(w, s.pipelines[i])
			return
		}
	}
	http.Error(w, "pipeline not found", http.StatusNotFound)
}
