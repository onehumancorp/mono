package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/onehumancorp/mono/srcs/integrations"
)

// MCPTool represents a registered tool in the MCP gateway.
type MCPTool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
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

func (s *Server) handleMCPTools(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, mcpTools)
}

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
