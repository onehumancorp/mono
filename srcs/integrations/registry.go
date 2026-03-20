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

// Category represents the functional domain of an external integration.
//
// Parameters: none
// Returns: a Category enum string.
// Errors: none.
// Side Effects: none.
type Category string

const (
	// CategoryChat maps to conversational platforms like Slack or Mattermost.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	CategoryChat   Category = "chat"
	// CategoryGit maps to source control systems like GitHub or Gitea.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	CategoryGit    Category = "git"
	// CategoryIssues maps to ticketing software like Jira or Linear.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	CategoryIssues Category = "issues"
)

// IntegrationType identifies the specific external service being connected.
//
// Parameters: none
// Returns: an IntegrationType enum string.
// Errors: none.
// Side Effects: none.
type IntegrationType string

const (
	// Chat services.
	// IntegrationTypeSlack defines a cloud-hosted Slack chat connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeSlack      IntegrationType = "slack"
	// IntegrationTypeDiscord defines a Discord server connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeDiscord    IntegrationType = "discord"
	// IntegrationTypeGoogleChat defines a Google Chat connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeGoogleChat IntegrationType = "google_chat"
	// IntegrationTypeTelegram defines a Telegram connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeTelegram   IntegrationType = "telegram"
	// IntegrationTypeTeams defines a Microsoft Teams connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeTeams      IntegrationType = "teams"

	// Git platforms.
	// IntegrationTypeGitHub defines an external GitHub repository connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeGitHub IntegrationType = "github"
	// IntegrationTypeGitLab defines a GitLab repository connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeGitLab IntegrationType = "gitlab"
	// IntegrationTypeGitea defines a self-hosted Gitea repository connection.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeGitea  IntegrationType = "gitea"

	// Issue trackers.
	IntegrationTypeJIRA         IntegrationType = "jira"
	IntegrationTypePlane        IntegrationType = "plane"
	// IntegrationTypeGitHubIssues defines a connection to GitHub Issues.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IntegrationTypeGitHubIssues IntegrationType = "github_issues"
	IntegrationTypeLinear       IntegrationType = "linear"
)

// ConnectionStatus tracks the health and connectivity of an external integration.
//
// Parameters: none
// Returns: a ConnectionStatus enum string.
// Errors: none.
// Side Effects: none.
type ConnectionStatus string

const (
	// StatusConnected indicates an active, healthy external API link.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusConnected    ConnectionStatus = "connected"
	// StatusDisconnected indicates an intentional connection tear-down.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusDisconnected ConnectionStatus = "disconnected"
	// StatusError indicates an unreachable API or bad authentication state.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	StatusError        ConnectionStatus = "error"
)

// Integration describes a configured connection to a third-party service.
//
// Parameters: none
// Returns: an Integration struct.
// Errors: none.
// Side Effects: none.
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

// ChatMessage represents a single communication event synced from an external chat tool.
//
// Parameters: none
// Returns: a ChatMessage struct.
// Errors: none.
// Side Effects: none.
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

// PullRequestStatus tracks the lifecycle of code changes in a Git provider.
//
// Parameters: none
// Returns: a PullRequestStatus enum string.
// Errors: none.
// Side Effects: none.
type PullRequestStatus string

const (
	// PRStatusOpen indicates an ongoing code review.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	PRStatusOpen   PullRequestStatus = "open"
	// PRStatusMerged indicates the proposal was applied.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	PRStatusMerged PullRequestStatus = "merged"
	// PRStatusClosed indicates the proposal was rejected.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	PRStatusClosed PullRequestStatus = "closed"
)

// PullRequest represents a proposed code modification mapped from an external VCS.
//
// Parameters: none
// Returns: a PullRequest struct tracking code review states.
// Errors: none.
// Side Effects: none.
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

// IssueStatus represents the progress state of a tracked ticket or bug.
//
// Parameters: none
// Returns: an IssueStatus enum string.
// Errors: none.
// Side Effects: none.
type IssueStatus string

const (
	// IssueStatusOpen signifies work that has not started.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssueStatusOpen       IssueStatus = "open"
	// IssueStatusInProgress signifies active execution by an agent or human.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssueStatusInProgress IssueStatus = "in_progress"
	// IssueStatusDone signifies the task is successfully fulfilled.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssueStatusDone       IssueStatus = "done"
	// IssueStatusClosed signifies the task was dismissed or aborted.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssueStatusClosed     IssueStatus = "closed"
)

// IssuePriority indicates the urgency of resolving an external ticket.
//
// Parameters: none
// Returns: an IssuePriority enum string.
// Errors: none.
// Side Effects: none.
type IssuePriority string

const (
	// IssuePriorityLow indicates work with no urgent deadlines.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssuePriorityLow      IssuePriority = "low"
	// IssuePriorityMedium indicates standard operational work.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssuePriorityMedium   IssuePriority = "medium"
	// IssuePriorityHigh indicates blocking tasks requiring immediate attention.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssuePriorityHigh     IssuePriority = "high"
	// IssuePriorityCritical indicates a severe organizational outage.
	// Parameters: none. Returns: none. Errors: none. Side Effects: none.
	IssuePriorityCritical IssuePriority = "critical"
)

// Issue holds the details of a work item synced from an external task tracker.
//
// Parameters: none
// Returns: an Issue struct tracking project management state.
// Errors: none.
// Side Effects: none.
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
// Registry acts as the central hub for managing all third-party integrations.
//
// Parameters: none
// Returns: a Registry instance handling integration state.
// Errors: none.
// Side Effects: none.
type Registry struct {
	mu           sync.RWMutex
	integrations []Integration
	chatMessages []ChatMessage
	pullRequests []PullRequest
	issues       []Issue
}

// NewRegistry returns an initialised Registry pre-populated with the default
// NewRegistry initializes an empty integration manager.
//
// Parameters: none
// Returns: A pointer to a newly instantiated Registry.
// Errors: none.
// Side Effects: allocates memory for internal maps and slices.
func NewRegistry() *Registry {
	return &Registry{
		integrations: defaultIntegrations(),
		chatMessages: []ChatMessage{},
		pullRequests: []PullRequest{},
		issues:       []Issue{},
	}
}

// ── Integration management ────────────────────────────────────────────────────

// Integrations returns a copy of all registered integrations.
func (r *Registry) Integrations() []Integration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]Integration(nil), r.integrations...)
}

// IntegrationsByCategory returns integrations filtered by category.
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

// Integration returns the integration with the given ID.
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

// Connect marks an integration as connected and sets its base URL.
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

// Disconnect marks an integration as disconnected.
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

// SendChatMessage dispatches a message through the named chat integration and
// records it in the registry log.
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

// ChatMessages returns all recorded chat messages, optionally filtered by
// integration ID (pass empty string to return all).
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

// CreatePullRequest opens a pull request on the specified git integration and
// records it in the registry log.
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

// MergePullRequest transitions a PR to merged status.
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

// ClosePullRequest transitions a PR to closed status.
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

// PullRequests returns all recorded pull requests, optionally filtered by
// integration ID.
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

// CreateIssue opens a new ticket in the specified issue tracker and records it.
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

// UpdateIssueStatus transitions an issue to the given status.
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

// AssignIssue sets the assignee on an issue.
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

// Issues returns all recorded issues, optionally filtered by integration ID.
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
