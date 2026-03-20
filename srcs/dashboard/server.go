package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
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

func (s *Server) handleCosts(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.tracker.Summary(s.org.ID))
}

func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.snapshot())
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

// handleAgentProviders lists all registered external agent providers and their
// authentication status.  Responds to GET only.

// handleAgentProviderAuth accepts POST requests to authenticate with an external
// agent provider.  Credentials are stored in memory and forwarded to any
// subsequently hired agent of that provider type.

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

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

// ── Chat test handler ─────────────────────────────────────────────────────────

// ── MCP tool invocation ───────────────────────────────────────────────────────

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

// ── Warm Handoff Handlers ─────────────────────────────────────────────────────

// ── Identity Management Handler ───────────────────────────────────────────────

// ── Skill Pack Handlers ───────────────────────────────────────────────────────

// ── Snapshot Handlers ─────────────────────────────────────────────────────────

// ── Marketplace Handler ───────────────────────────────────────────────────────

// ── Analytics Handler ─────────────────────────────────────────────────────────

// ── Default Data Factories ────────────────────────────────────────────────────

// ── Integration request/response types ────────────────────────────────────────

// ── Integration handlers ──────────────────────────────────────────────────────

// ── Chat handlers ─────────────────────────────────────────────────────────────

// ── Git handlers ──────────────────────────────────────────────────────────────

// ── Issue tracker handlers ────────────────────────────────────────────────────

// ── B2B Collaboration ─────────────────────────────────────────────────────────

// ── Autonomous SRE / Incident Management ─────────────────────────────────────

// ── Compute Optimization / Hardware-Aware Scheduling ─────────────────────────

// ── Budget Alerts ─────────────────────────────────────────────────────────────

// ── Automated SDLC / Pipelines ────────────────────────────────────────────────
