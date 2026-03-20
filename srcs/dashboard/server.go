package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/agents"
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
	mu                    sync.RWMutex
	org                   domain.Organization
	hub                   *orchestration.Hub
	tracker               *billing.Tracker
	approvals             []ApprovalRequest
	handoffs              []HandoffPackage
	skills                []SkillPack
	snapshots             []OrgSnapshot
	integReg              *integrations.Registry
	trustAgreements       []TrustAgreement
	incidents             []Incident
	computeProfiles       []ComputeProfile
	budgetAlerts          []BudgetAlert
	pipelines             []Pipeline
	authStore             *auth.Store
	authHandlers          *auth.Handlers
	settings              Settings
	agentProviderRegistry *agents.Registry
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
		org:                   org,
		hub:                   hub,
		tracker:               tracker,
		approvals:             []ApprovalRequest{},
		handoffs:              []HandoffPackage{},
		skills:                defaultSkillPacks(),
		snapshots:             []OrgSnapshot{},
		integReg:              integrations.NewRegistry(),
		trustAgreements:       []TrustAgreement{},
		incidents:             []Incident{},
		computeProfiles:       []ComputeProfile{},
		budgetAlerts:          []BudgetAlert{},
		pipelines:             []Pipeline{},
		authStore:             store,
		authHandlers:          auth.NewHandlers(store),
		agentProviderRegistry: agents.DefaultRegistry(),
	}
	// Load Minimax API key from environment on startup.
	if key := os.Getenv("MINIMAX_API_KEY"); key != "" {
		hub.SetMinimaxAPIKey(key)
		server.settings.MinimaxAPIKey = key
	}
	// Pre-authenticate providers from environment variables so the platform
	// forwards credentials to freshly hired agents without requiring manual auth.
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		_ = server.agentProviderRegistry.Authenticate(agents.ProviderTypeClaude, agents.Credentials{APIKey: key})
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		_ = server.agentProviderRegistry.Authenticate(agents.ProviderTypeGemini, agents.Credentials{APIKey: key})
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		_ = server.agentProviderRegistry.Authenticate(agents.ProviderTypeOpenCode, agents.Credentials{APIKey: key})
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
	// Agent provider management
	mux.HandleFunc("/api/agents/providers", server.handleAgentProviders)
	mux.HandleFunc("/api/agents/providers/auth", server.handleAgentProviderAuth)
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

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
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
	hub.RegisterAgent(orchestration.Agent{ID: "qa-1", Name: "QA Lead", Role: "QA_TESTER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "sec-1", Name: "Security Auditor", Role: "SECURITY_ENGINEER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "CEO", Name: "Human CEO", Role: "CEO", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("launch-readiness", "Review launch blockers, sign-off on reliability checklist, assign post-launch owners.", []string{"pm-1", "swe-1", "ux-1", "qa-1", "sec-1", "CEO"})

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
		FromAgent:  "pm-1",
		ToAgent:    "CEO",
		Type:       orchestration.EventApprovalNeeded,
		Content:    "All pre-launch checks passed. Requesting final CEO approval to deploy to production.",
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
