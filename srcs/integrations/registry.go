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
// the platform.  Chat integrations with stored credentials (Telegram, Discord)
// make real outbound HTTP API calls in addition to recording messages locally.
package integrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ── Integration types ─────────────────────────────────────────────────────────

// Category groups integrations by their function.
type Category string

const (
	CategoryChat   Category = "chat"
	CategoryGit    Category = "git"
	CategoryIssues Category = "issues"
)

// IntegrationType identifies the specific external service.
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

// ConnectionStatus reflects whether an integration is reachable.
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusError        ConnectionStatus = "error"
)

// Integration is a configured external service connection.
type Integration struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Type           IntegrationType  `json:"type"`
	Category       Category         `json:"category"`
	BaseURL        string           `json:"baseUrl,omitempty"`
	Status         ConnectionStatus `json:"status"`
	Description    string           `json:"description,omitempty"`
	HasCredentials bool             `json:"hasCredentials,omitempty"`
	Chatspace      string           `json:"chatspace,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
}

// IntegrationCredentials holds the secret configuration for an integration.
// These are stored server-side only and never serialised to the client.
type IntegrationCredentials struct {
	BotToken   string // Telegram Bot API token
	ChatID     string // Telegram chat / group ID
	WebhookURL string // Discord (or generic) inbound webhook URL
	APIToken   string // Generic API token / Bearer credential
}

// IsEmpty reports whether no fields are set.
func (c IntegrationCredentials) IsEmpty() bool {
	return c.BotToken == "" && c.ChatID == "" && c.WebhookURL == "" && c.APIToken == ""
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

// PullRequestStatus tracks the lifecycle of a PR/MR.
type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "open"
	PRStatusMerged PullRequestStatus = "merged"
	PRStatusClosed PullRequestStatus = "closed"
)

// PullRequest represents a PR/MR opened on a git hosting platform.
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

// IssueStatus tracks the lifecycle of an issue/ticket.
type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusClosed     IssueStatus = "closed"
)

// IssuePriority indicates ticket urgency.
type IssuePriority string

const (
	IssuePriorityLow      IssuePriority = "low"
	IssuePriorityMedium   IssuePriority = "medium"
	IssuePriorityHigh     IssuePriority = "high"
	IssuePriorityCritical IssuePriority = "critical"
)

// Issue represents a ticket created in an external issue tracker.
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
type Registry struct {
	mu           sync.RWMutex
	integrations []Integration
	credentials  map[string]IntegrationCredentials // keyed by integration ID; never exposed to clients
	chatMessages []ChatMessage
	pullRequests []PullRequest
	issues       []Issue
}

// NewRegistry returns an initialised Registry pre-populated with the default
// set of supported integrations (all marked as disconnected until configured).
func NewRegistry() *Registry {
	return &Registry{
		integrations: defaultIntegrations(),
		credentials:  map[string]IntegrationCredentials{},
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
// An optional IntegrationCredentials value stores secrets (e.g. bot tokens)
// for integrations that make real outbound API calls.
func (r *Registry) Connect(id, baseURL string, creds ...IntegrationCredentials) (Integration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, i := range r.integrations {
		if i.ID == id {
			r.integrations[idx].Status = StatusConnected
			if baseURL != "" {
				r.integrations[idx].BaseURL = baseURL
			}
			if len(creds) > 0 && !creds[0].IsEmpty() {
				r.credentials[id] = creds[0]
				r.integrations[idx].HasCredentials = true
				// Populate the default chatspace from the ChatID credential so the
				// UI can display which channel messages will be delivered to.
				if creds[0].ChatID != "" {
					r.integrations[idx].Chatspace = creds[0].ChatID
				}
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
// records it in the registry log.  When real credentials are stored for the
// integration (Telegram bot token, Discord webhook URL) the message is also
// delivered via the external API.
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

	// Attempt real delivery when credentials are configured.
	if creds, hasCreds := r.credentials[integrationID]; hasCreds {
		text := fmt.Sprintf("[%s] %s", fromAgent, content)
		switch integ.Type {
		case IntegrationTypeTelegram:
			if creds.BotToken != "" {
				// Use provided channel (chat_id) or fall back to the stored ChatID.
				chatID := channel
				if creds.ChatID != "" {
					chatID = creds.ChatID
				}
				// Best-effort: log but do not fail the in-memory record.
				_ = sendTelegramMessage(creds.BotToken, chatID, text)
			}
		case IntegrationTypeDiscord:
			if creds.WebhookURL != "" {
				_ = sendDiscordWebhook(creds.WebhookURL, fromAgent, content)
			}
		}
	}

	return msg, nil
}

// TestConnection validates that the provided credentials can reach the external
// service by sending a short test message.  Use this during setup wizards
// before persisting credentials.
func (r *Registry) TestConnection(id string, creds IntegrationCredentials) error {
	r.mu.RLock()
	integ, ok := r.findIntegration(id)
	// If no credentials supplied, fall back to stored ones.
	stored := r.credentials[id]
	r.mu.RUnlock()

	if !ok {
		return errors.New("integration not found")
	}

	active := creds
	if active.IsEmpty() {
		active = stored
	}

	switch integ.Type {
	case IntegrationTypeTelegram:
		if active.BotToken == "" {
			return errors.New("bot token is required")
		}
		if active.ChatID == "" {
			return errors.New("chat ID is required")
		}
		return sendTelegramMessage(active.BotToken, active.ChatID,
			"✅ Test message from One Human Corp — Telegram integration confirmed!")
	case IntegrationTypeDiscord:
		if active.WebhookURL == "" {
			return errors.New("webhook URL is required")
		}
		return sendDiscordWebhook(active.WebhookURL, "One Human Corp",
			"✅ Test message — Discord integration confirmed!")
	default:
		// No live endpoint to test; accept unconditionally.
		return nil
	}
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

// ── Real outbound HTTP helpers ────────────────────────────────────────────────

// TelegramAPIBase is the base URL for the Telegram Bot API.
// Override in tests to point to a mock server.
var TelegramAPIBase = "https://api.telegram.org"

// sendTelegramMessage posts a text message to a Telegram chat via the Bot API.
func sendTelegramMessage(botToken, chatID, text string) error {
	apiURL := TelegramAPIBase + "/bot" + botToken + "/sendMessage"
	payload, err := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(payload)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("telegram API: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram decode: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram error: %s", result.Description)
	}
	return nil
}

// sendDiscordWebhook posts a message to a Discord channel via an inbound webhook.
func sendDiscordWebhook(webhookURL, username, content string) error {
	payload, err := json.Marshal(map[string]string{
		"username": username,
		"content":  content,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(payload)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("discord API: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()
	// Discord webhooks return 204 No Content on success.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord API returned status %d", resp.StatusCode)
	}
	return nil
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
