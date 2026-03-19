// Package integrations provides a registry for external service adapters.
//
// It implements three categories of integration that allow AI agents to
// interact with the same tools that human team members use:
//
//   - Chat services: Slack, Discord, Google Chat, Telegram, Microsoft Teams — for human–agent messaging
//   - Git platforms: GitHub, GitLab, Gitea    — for PR/MR creation
//   - Issue trackers: JIRA, Plane, GitHub Issues, Linear — for ticket management
//
// All state is held in-memory following the same pattern used by the rest of
// the platform.  Real-world deployments would replace the stub send/create
// calls with actual HTTP client requests.
package integrations

import (
	"errors"
	"sync"
	"time"
)

// ── Integration types ─────────────────────────────────────────────────────────

// Category groups integrations by their function (e.g., chat, git, issues).
type Category string

const (
	CategoryChat   Category = "chat"
	CategoryGit    Category = "git"
	CategoryIssues Category = "issues"
)

// IntegrationType identifies the specific external service platform (e.g., github, slack).
type IntegrationType string

const (
	// Chat services.
	IntegrationTypeSlack      IntegrationType = "slack"
	IntegrationTypeDiscord    IntegrationType = "discord"
	IntegrationTypeGoogleChat IntegrationType = "google_chat"
	IntegrationTypeTelegram   IntegrationType = "telegram"
	IntegrationTypeTeams      IntegrationType = "teams"

	// Git platforms.
	IntegrationTypeGitHub IntegrationType = "github"
	IntegrationTypeGitLab IntegrationType = "gitlab"
	IntegrationTypeGitea  IntegrationType = "gitea"

	// Issue trackers.
	IntegrationTypeJIRA         IntegrationType = "jira"
	IntegrationTypePlane        IntegrationType = "plane"
	IntegrationTypeGitHubIssues IntegrationType = "github_issues"
	IntegrationTypeLinear       IntegrationType = "linear"
)

// ConnectionStatus reflects whether an integration is currently active and reachable.
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusError        ConnectionStatus = "error"
)

// Integration represents a configured external service connection.
type Integration struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        IntegrationType  `json:"type"`
	Category    Category         `json:"category"`
	BaseURL     string           `json:"baseUrl,omitempty"`
	Status      ConnectionStatus `json:"status"`
	Description string           `json:"description,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// ── Chat types ────────────────────────────────────────────────────────────────

// ChatMessage represents a message dispatched through a chat service.
type ChatMessage struct {
	ID            string    `json:"id"`
	IntegrationID string    `json:"integrationId"`
	Channel       string    `json:"channel"`
	FromAgent     string    `json:"fromAgent"`
	Content       string    `json:"content"`
	ThreadID      string    `json:"threadId,omitempty"`
	SentAt        time.Time `json:"sentAt"`
}

// ── Git types ─────────────────────────────────────────────────────────────────

// PullRequestStatus tracks the lifecycle status of a PR/MR on a git platform.
type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "open"
	PRStatusMerged PullRequestStatus = "merged"
	PRStatusClosed PullRequestStatus = "closed"
)

// PullRequest records an issue or code change request opened on a git hosting platform.
type PullRequest struct {
	ID             string            `json:"id"`
	IntegrationID  string            `json:"integrationId"`
	Repository     string            `json:"repository"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	SourceBranch   string            `json:"sourceBranch"`
	TargetBranch   string            `json:"targetBranch"`
	URL            string            `json:"url"`
	CreatedByAgent string            `json:"createdByAgent"`
	Status         PullRequestStatus `json:"status"`
	CreatedAt      time.Time         `json:"createdAt"`
}

// ── Issue types ───────────────────────────────────────────────────────────────

// IssueStatus tracks the lifecycle phase of an issue or ticket.
type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusClosed     IssueStatus = "closed"
)

// IssuePriority indicates the urgency of a ticket.
type IssuePriority string

const (
	IssuePriorityLow      IssuePriority = "low"
	IssuePriorityMedium   IssuePriority = "medium"
	IssuePriorityHigh     IssuePriority = "high"
	IssuePriorityCritical IssuePriority = "critical"
)

// Issue records a ticket created in an external issue tracker.
type Issue struct {
	ID             string        `json:"id"`
	IntegrationID  string        `json:"integrationId"`
	Project        string        `json:"project"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Priority       IssuePriority `json:"priority"`
	Status         IssueStatus   `json:"status"`
	AssignedTo     string        `json:"assignedTo,omitempty"`
	Labels         []string      `json:"labels,omitempty"`
	CreatedByAgent string        `json:"createdByAgent"`
	URL            string        `json:"url"`
	CreatedAt      time.Time     `json:"createdAt"`
}

// ── Registry ─────────────────────────────────────────────────────────────────

// Registry manages all configured external service integrations and records
// every action taken through them (messages sent, PRs opened, tickets created).
//
// Constraints: Thread-safe via sync.RWMutex.
type Registry struct {
	mu           sync.RWMutex
	integrations []Integration
	chatMessages []ChatMessage
	pullRequests []PullRequest
	issues       []Issue
}

// NewRegistry returns an initialised Registry pre-populated with the default
// set of supported integrations (all marked as disconnected until configured).
//
// Returns: A newly instantiated Registry pointer.
func NewRegistry() *Registry {
	return &Registry{
		integrations: defaultIntegrations(),
		chatMessages: []ChatMessage{},
		pullRequests: []PullRequest{},
		issues:       []Issue{},
	}
}

// ── Integration management ────────────────────────────────────────────────────

// Integrations retrieves a snapshot of all registered external service integrations.
//
// Returns: A slice of Integration objects representing the current connection states.
func (r *Registry) Integrations() []Integration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]Integration(nil), r.integrations...)
}

// IntegrationsByCategory returns integrations filtered by their service category.
//
// Parameters:
//   - cat: Category; The category to filter by (e.g., CategoryChat).
//
// Returns: A slice of Integration objects belonging to the specified category.
func (r *Registry) IntegrationsByCategory(cat Category) []Integration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Integration
	for _, i := range r.integrations {
		if i.Category == cat {
			result = append(result, i)
		}
	}
	return result
}

// Integration looks up a specific integration by its unique ID.
//
// Parameters:
//   - id: string; The identifier of the integration.
//
// Returns: The matching Integration and a boolean indicating if it was found.
func (r *Registry) Integration(id string) (Integration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, i := range r.integrations {
		if i.ID == id {
			return i, true
		}
	}
	return Integration{}, false
}

// Connect enables an integration by marking it connected and setting its API base URL.
//
// Parameters:
//   - id: string; The identifier of the integration to connect.
//   - baseURL: string; The API base URL to use for requests.
//
// Returns: The updated Integration, or an error if it was not found.
func (r *Registry) Connect(id, baseURL string) (Integration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, i := range r.integrations {
		if i.ID == id {
			r.integrations[idx].Status = StatusConnected
			if baseURL != "" {
				r.integrations[idx].BaseURL = baseURL
			}
			return r.integrations[idx], nil
		}
	}
	return Integration{}, errors.New("integration not found")
}

// Disconnect marks a previously connected integration as disconnected.
//
// Parameters:
//   - id: string; The identifier of the integration to disconnect.
//
// Returns: The updated Integration, or an error if it was not found.
func (r *Registry) Disconnect(id string) (Integration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, i := range r.integrations {
		if i.ID == id {
			r.integrations[idx].Status = StatusDisconnected
			return r.integrations[idx], nil
		}
	}
	return Integration{}, errors.New("integration not found")
}

// ── Chat operations ───────────────────────────────────────────────────────────

// SendChatMessage records the dispatch of a message through the specified chat integration.
//
// Parameters:
//   - integrationID: string; The ID of the chat integration (e.g., "slack").
//   - channel: string; The target channel or space.
//   - fromAgent: string; The ID of the agent sending the message.
//   - content: string; The message payload.
//   - threadID: string; The thread context, if applicable.
//   - now: time.Time; The current timestamp.
//
// Returns: A ChatMessage record of the action, or an error if the integration is invalid.
func (r *Registry) SendChatMessage(integrationID, channel, fromAgent, content, threadID string, now time.Time) (ChatMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	integ, ok := r.findIntegration(integrationID)
	if !ok {
		return ChatMessage{}, errors.New("integration not found")
	}
	if integ.Category != CategoryChat {
		return ChatMessage{}, errors.New("integration is not a chat service")
	}
	if channel == "" {
		return ChatMessage{}, errors.New("channel is required")
	}
	if fromAgent == "" {
		return ChatMessage{}, errors.New("fromAgent is required")
	}
	if content == "" {
		return ChatMessage{}, errors.New("content is required")
	}

	msg := ChatMessage{
		ID:            generateID(integrationID+"-msg", now),
		IntegrationID: integrationID,
		Channel:       channel,
		FromAgent:     fromAgent,
		Content:       content,
		ThreadID:      threadID,
		SentAt:        now.UTC(),
	}
	r.chatMessages = append(r.chatMessages, msg)
	return msg, nil
}

// ChatMessages retrieves all recorded chat messages, with an optional integration ID filter.
//
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string for all messages.
//
// Returns: A slice of ChatMessage records.
func (r *Registry) ChatMessages(integrationID string) []ChatMessage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ChatMessage
	for _, m := range r.chatMessages {
		if integrationID == "" || m.IntegrationID == integrationID {
			result = append(result, m)
		}
	}
	return result
}

// ── Git operations ────────────────────────────────────────────────────────────

// CreatePullRequest registers a new PR/MR action on the specified git integration.
//
// Parameters:
//   - integrationID: string; The ID of the git integration (e.g., "github").
//   - repo: string; Target repository name.
//   - title: string; PR title.
//   - body: string; PR description.
//   - source: string; Branch name containing the changes.
//   - target: string; Base branch to merge into.
//   - createdBy: string; Agent ID opening the PR.
//   - now: time.Time; Timestamp.
//
// Returns: A PullRequest record of the action, or an error if parameters are invalid.
func (r *Registry) CreatePullRequest(integrationID, repo, title, body, source, target, createdBy string, now time.Time) (PullRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	integ, ok := r.findIntegration(integrationID)
	if !ok {
		return PullRequest{}, errors.New("integration not found")
	}
	if integ.Category != CategoryGit {
		return PullRequest{}, errors.New("integration is not a git platform")
	}
	if repo == "" {
		return PullRequest{}, errors.New("repository is required")
	}
	if title == "" {
		return PullRequest{}, errors.New("title is required")
	}
	if source == "" || target == "" {
		return PullRequest{}, errors.New("sourceBranch and targetBranch are required")
	}

	prID := generateID(integrationID+"-pr", now)
	pr := PullRequest{
		ID:             prID,
		IntegrationID:  integrationID,
		Repository:     repo,
		Title:          title,
		Body:           body,
		SourceBranch:   source,
		TargetBranch:   target,
		URL:            integ.BaseURL + "/" + repo + "/pull/" + prID,
		CreatedByAgent: createdBy,
		Status:         PRStatusOpen,
		CreatedAt:      now.UTC(),
	}
	r.pullRequests = append(r.pullRequests, pr)
	return pr, nil
}

// MergePullRequest transitions an open Pull Request to merged status.
//
// Parameters:
//   - prID: string; The unique registry ID of the pull request.
//
// Returns: The updated PullRequest record.
func (r *Registry) MergePullRequest(prID string) (PullRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, pr := range r.pullRequests {
		if pr.ID == prID {
			if pr.Status != PRStatusOpen {
				return PullRequest{}, errors.New("pull request is not open")
			}
			r.pullRequests[idx].Status = PRStatusMerged
			return r.pullRequests[idx], nil
		}
	}
	return PullRequest{}, errors.New("pull request not found")
}

// ClosePullRequest transitions an open Pull Request to closed status without merging.
//
// Parameters:
//   - prID: string; The unique registry ID of the pull request.
//
// Returns: The updated PullRequest record.
func (r *Registry) ClosePullRequest(prID string) (PullRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, pr := range r.pullRequests {
		if pr.ID == prID {
			if pr.Status != PRStatusOpen {
				return PullRequest{}, errors.New("pull request is not open")
			}
			r.pullRequests[idx].Status = PRStatusClosed
			return r.pullRequests[idx], nil
		}
	}
	return PullRequest{}, errors.New("pull request not found")
}

// PullRequests retrieves all recorded pull requests, with an optional integration ID filter.
//
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string to return all.
//
// Returns: A slice of PullRequest records.
func (r *Registry) PullRequests(integrationID string) []PullRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []PullRequest
	for _, pr := range r.pullRequests {
		if integrationID == "" || pr.IntegrationID == integrationID {
			result = append(result, pr)
		}
	}
	return result
}

// ── Issue operations ──────────────────────────────────────────────────────────

// CreateIssue registers a new ticket action in the specified issue tracker integration.
//
// Parameters:
//   - integrationID: string; The ID of the issue integration (e.g., "jira").
//   - project: string; The target project or board.
//   - title: string; The issue summary.
//   - description: string; The detailed description of the issue.
//   - createdBy: string; The ID of the agent reporting the issue.
//   - priority: IssuePriority; The urgency of the ticket.
//   - labels: []string; Categorisation tags.
//   - now: time.Time; Current timestamp.
//
// Returns: An Issue record of the action, or an error if parameters are invalid.
func (r *Registry) CreateIssue(integrationID, project, title, description, createdBy string, priority IssuePriority, labels []string, now time.Time) (Issue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	integ, ok := r.findIntegration(integrationID)
	if !ok {
		return Issue{}, errors.New("integration not found")
	}
	if integ.Category != CategoryIssues {
		return Issue{}, errors.New("integration is not an issue tracker")
	}
	if project == "" {
		return Issue{}, errors.New("project is required")
	}
	if title == "" {
		return Issue{}, errors.New("title is required")
	}

	if priority == "" {
		priority = IssuePriorityMedium
	}

	issueID := generateID(integrationID+"-issue", now)
	labelsCopy := make([]string, len(labels))
	copy(labelsCopy, labels)
	issue := Issue{
		ID:             issueID,
		IntegrationID:  integrationID,
		Project:        project,
		Title:          title,
		Description:    description,
		Priority:       priority,
		Status:         IssueStatusOpen,
		Labels:         labelsCopy,
		CreatedByAgent: createdBy,
		URL:            integ.BaseURL + "/issues/" + issueID,
		CreatedAt:      now.UTC(),
	}
	r.issues = append(r.issues, issue)
	return issue, nil
}

// UpdateIssueStatus transitions an existing issue to the specified lifecycle phase.
//
// Parameters:
//   - issueID: string; The unique registry ID of the issue.
//   - status: IssueStatus; The new status phase (e.g., IssueStatusDone).
//
// Returns: The updated Issue record.
func (r *Registry) UpdateIssueStatus(issueID string, status IssueStatus) (Issue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, issue := range r.issues {
		if issue.ID == issueID {
			r.issues[idx].Status = status
			return r.issues[idx], nil
		}
	}
	return Issue{}, errors.New("issue not found")
}

// AssignIssue sets or transfers ownership of an issue to a specific agent or human.
//
// Parameters:
//   - issueID: string; The unique registry ID of the issue.
//   - assignee: string; The identifier of the assigned worker.
//
// Returns: The updated Issue record.
func (r *Registry) AssignIssue(issueID, assignee string) (Issue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, issue := range r.issues {
		if issue.ID == issueID {
			r.issues[idx].AssignedTo = assignee
			return r.issues[idx], nil
		}
	}
	return Issue{}, errors.New("issue not found")
}

// Issues retrieves all recorded tickets, with an optional integration ID filter.
//
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string for all tickets.
//
// Returns: A slice of Issue records.
func (r *Registry) Issues(integrationID string) []Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Issue
	for _, issue := range r.issues {
		if integrationID == "" || issue.IntegrationID == integrationID {
			result = append(result, issue)
		}
	}
	return result
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// findIntegration looks up an integration by ID; caller must hold mu.
func (r *Registry) findIntegration(id string) (Integration, bool) {
	for _, i := range r.integrations {
		if i.ID == id {
			return i, true
		}
	}
	return Integration{}, false
}

// generateID produces a namespaced, time-stamped identifier for an activity record.
func generateID(prefix string, now time.Time) string {
	return prefix + "-" + now.UTC().Format("20060102150405.000000000")
}

// defaultIntegrations returns the built-in set of supported external services,
// all initially disconnected.
func defaultIntegrations() []Integration {
	now := time.Now().UTC()
	return []Integration{
		// Chat services
		{
			ID:          "slack",
			Name:        "Slack",
			Type:        IntegrationTypeSlack,
			Category:    CategoryChat,
			Description: "Send agent-to-human notifications and HITL approval requests via Slack channels.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "discord",
			Name:        "Discord",
			Type:        IntegrationTypeDiscord,
			Category:    CategoryChat,
			Description: "Post agent messages and meeting summaries to Discord channels or threads.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "google-chat",
			Name:        "Google Chat",
			Type:        IntegrationTypeGoogleChat,
			Category:    CategoryChat,
			Description: "Deliver agent updates and approval requests via Google Chat spaces.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "telegram",
			Name:        "Telegram",
			Type:        IntegrationTypeTelegram,
			Category:    CategoryChat,
			Description: "Send agent notifications and HITL approval requests via Telegram bots and channels.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "teams",
			Name:        "Microsoft Teams",
			Type:        IntegrationTypeTeams,
			Category:    CategoryChat,
			Description: "Deliver agent updates and approval requests to Microsoft Teams channels via webhooks.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		// Git platforms
		{
			ID:          "github",
			Name:        "GitHub",
			Type:        IntegrationTypeGitHub,
			Category:    CategoryGit,
			BaseURL:     "https://github.com",
			Description: "Open pull requests, review code, and manage branches on GitHub.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "gitlab",
			Name:        "GitLab",
			Type:        IntegrationTypeGitLab,
			Category:    CategoryGit,
			BaseURL:     "https://gitlab.com",
			Description: "Create merge requests and manage repositories on GitLab or self-hosted instances.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "gitea",
			Name:        "Gitea",
			Type:        IntegrationTypeGitea,
			Category:    CategoryGit,
			Description: "Open PRs on a self-hosted Gitea instance — the zero-lock-in OSS git option.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		// Issue trackers
		{
			ID:          "jira",
			Name:        "Jira",
			Type:        IntegrationTypeJIRA,
			Category:    CategoryIssues,
			Description: "Create and manage issues, epics, and sprints in Atlassian Jira.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "plane",
			Name:        "Plane",
			Type:        IntegrationTypePlane,
			Category:    CategoryIssues,
			Description: "Manage issues and cycles with Plane — the open-source Jira alternative.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "github-issues",
			Name:        "GitHub Issues",
			Type:        IntegrationTypeGitHubIssues,
			Category:    CategoryIssues,
			Description: "Track tasks and bugs directly in GitHub Issues alongside your repositories.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
		{
			ID:          "linear",
			Name:        "Linear",
			Type:        IntegrationTypeLinear,
			Category:    CategoryIssues,
			Description: "Manage issues, cycles, and roadmaps with Linear — the modern issue tracker for high-velocity teams.",
			Status:      StatusDisconnected,
			CreatedAt:   now,
		},
	}
}
