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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ── Integration types ─────────────────────────────────────────────────────────

// Category groups integrations by their function (e.g., chat, git, issues).
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Category string

const (
	// Summary: Defines the CategoryChat type.
	// Intent: Defines the CategoryChat type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	CategoryChat Category = "chat"
	// Summary: Defines the CategoryGit type.
	// Intent: Defines the CategoryGit type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	CategoryGit Category = "git"
	// Summary: Defines the CategoryIssues type.
	// Intent: Defines the CategoryIssues type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	CategoryIssues Category = "issues"
)

// IntegrationType identifies the specific external service platform (e.g., github, slack).
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type IntegrationType string

const (
	// Summary: Chat services.
	// Intent: Chat services.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeSlack IntegrationType = "slack"
	// Summary: Defines the IntegrationTypeDiscord type.
	// Intent: Defines the IntegrationTypeDiscord type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeDiscord IntegrationType = "discord"
	// Summary: Defines the IntegrationTypeGoogleChat type.
	// Intent: Defines the IntegrationTypeGoogleChat type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeGoogleChat IntegrationType = "google_chat"
	// Summary: Defines the IntegrationTypeTelegram type.
	// Intent: Defines the IntegrationTypeTelegram type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeTelegram IntegrationType = "telegram"
	// Summary: Defines the IntegrationTypeTeams type.
	// Intent: Defines the IntegrationTypeTeams type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeTeams IntegrationType = "teams"

	// Summary: Git platforms.
	// Intent: Git platforms.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeGitHub IntegrationType = "github"
	// Summary: Defines the IntegrationTypeGitLab type.
	// Intent: Defines the IntegrationTypeGitLab type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeGitLab IntegrationType = "gitlab"
	// Summary: Defines the IntegrationTypeGitea type.
	// Intent: Defines the IntegrationTypeGitea type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeGitea IntegrationType = "gitea"

	// Summary: Issue trackers.
	// Intent: Issue trackers.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeJIRA IntegrationType = "jira"
	// Summary: Defines the IntegrationTypePlane type.
	// Intent: Defines the IntegrationTypePlane type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypePlane IntegrationType = "plane"
	// Summary: Defines the IntegrationTypeGitHubIssues type.
	// Intent: Defines the IntegrationTypeGitHubIssues type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeGitHubIssues IntegrationType = "github_issues"
	// Summary: Defines the IntegrationTypeLinear type.
	// Intent: Defines the IntegrationTypeLinear type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IntegrationTypeLinear IntegrationType = "linear"
)

// ConnectionStatus reflects whether an integration is currently active and reachable.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type ConnectionStatus string

const (
	// Summary: Defines the StatusConnected type.
	// Intent: Defines the StatusConnected type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusConnected ConnectionStatus = "connected"
	// Summary: Defines the StatusDisconnected type.
	// Intent: Defines the StatusDisconnected type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusDisconnected ConnectionStatus = "disconnected"
	// Summary: Defines the StatusError type.
	// Intent: Defines the StatusError type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StatusError ConnectionStatus = "error"
)

// Integration represents a configured external service connection.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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

// IntegrationCredentials holds the secret configuration for an integration. These are stored server-side only and never serialised to the client.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type IntegrationCredentials struct {
	BotToken   string // Telegram Bot API token
	ChatID     string // Telegram chat / group ID
	WebhookURL string // Discord (or generic) inbound webhook URL
	APIToken   string // Generic API token / Bearer credential
}

// IsEmpty reports whether no fields are set.
// Params: None
// Returns: bool
// Errors: None
// Side Effects: None
func (c IntegrationCredentials) IsEmpty() bool {
	return c.BotToken == "" && c.ChatID == "" && c.WebhookURL == "" && c.APIToken == ""
}

// ── Chat types ────────────────────────────────────────────────────────────────

// ChatMessage represents a message dispatched through a chat service.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type PullRequestStatus string

const (
	// Summary: Defines the PRStatusOpen type.
	// Intent: Defines the PRStatusOpen type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	PRStatusOpen PullRequestStatus = "open"
	// Summary: Defines the PRStatusMerged type.
	// Intent: Defines the PRStatusMerged type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	PRStatusMerged PullRequestStatus = "merged"
	// Summary: Defines the PRStatusClosed type.
	// Intent: Defines the PRStatusClosed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	PRStatusClosed PullRequestStatus = "closed"
)

// PullRequest records an issue or code change request opened on a git hosting platform.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type IssueStatus string

const (
	// Summary: Defines the IssueStatusOpen type.
	// Intent: Defines the IssueStatusOpen type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssueStatusOpen IssueStatus = "open"
	// Summary: Defines the IssueStatusInProgress type.
	// Intent: Defines the IssueStatusInProgress type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssueStatusInProgress IssueStatus = "in_progress"
	// Summary: Defines the IssueStatusDone type.
	// Intent: Defines the IssueStatusDone type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssueStatusDone IssueStatus = "done"
	// Summary: Defines the IssueStatusClosed type.
	// Intent: Defines the IssueStatusClosed type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssueStatusClosed IssueStatus = "closed"
)

// IssuePriority indicates the urgency of a ticket.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type IssuePriority string

const (
	// Summary: Defines the IssuePriorityLow type.
	// Intent: Defines the IssuePriorityLow type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssuePriorityLow IssuePriority = "low"
	// Summary: Defines the IssuePriorityMedium type.
	// Intent: Defines the IssuePriorityMedium type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssuePriorityMedium IssuePriority = "medium"
	// Summary: Defines the IssuePriorityHigh type.
	// Intent: Defines the IssuePriorityHigh type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssuePriorityHigh IssuePriority = "high"
	// Summary: Defines the IssuePriorityCritical type.
	// Intent: Defines the IssuePriorityCritical type.
	// Params: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	IssuePriorityCritical IssuePriority = "critical"
)

// Issue records a ticket created in an external issue tracker.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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

// Registry manages all configured external service integrations and records every action taken through them (messages sent, PRs opened, tickets created).  Constraints: Thread-safe via sync.RWMutex.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
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
// Returns: A newly instantiated Registry pointer.
//
// Parameters: None
// Errors: None
// Side Effects: Allocates memory and initializes internal state.
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

// Integrations retrieves a snapshot of all registered external service integrations.
// Returns: A slice of Integration objects representing the current connection states.
//
// Parameters: None
// Errors: None
// Side Effects: Executes within local scope; may involve context-dependent side effects based on implementation.
func (r *Registry) Integrations() []Integration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]Integration(nil), r.integrations...)
}

// IntegrationsByCategory returns integrations filtered by their service category.
// Parameters:
//   - cat: Category; The category to filter by (e.g., CategoryChat).
// Returns: A slice of Integration objects belonging to the specified category.
//
// Errors: None
// Side Effects: Executes within local scope; may involve context-dependent side effects based on implementation.
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
// Parameters:
//   - id: string; The identifier of the integration.
// Returns: The matching Integration and a boolean indicating if it was found.
//
// Errors: None
// Side Effects: Executes within local scope; may involve context-dependent side effects based on implementation.
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

// LookupIPFunc is a variable to allow mocking net.LookupIP in tests across packages.
var // Summary: LookupIPFunc is a variable to allow mocking net.LookupIP in tests across packages.
// Intent: LookupIPFunc is a variable to allow mocking net.LookupIP in tests across packages.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
LookupIPFunc = net.LookupIP

// AllowLocalIPsForTesting can be set to true in tests to bypass SSRF IP checks
var // Summary: AllowLocalIPsForTesting can be set to true in tests to bypass SSRF IP checks
// Intent: AllowLocalIPsForTesting can be set to true in tests to bypass SSRF IP checks
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
AllowLocalIPsForTesting = false

func isBlockedIP(ip net.IP) bool {
	if AllowLocalIPsForTesting {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// validateURL checks if a given URL string is safe from SSRF attacks.
// It explicitly blocks loopback, private, unspecified, and link-local IP addresses.
// It fails closed on DNS resolution errors.
func validateURL(u string) error {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("invalid URL scheme")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return errors.New("URL must contain a host")
	}

	ips, err := LookupIPFunc(host)
	if err != nil {
		// Fail closed on DNS resolution error
		return errors.New("DNS resolution failed")
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return errors.New("URL resolves to a blocked IP address")
		}
	}

	return nil
}

// initSafeHTTPClient returns an http.Client with a custom DialContext that prevents
// DNS rebinding (TOCTOU) attacks by pinning the connection to the validated IP.
func initSafeHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := LookupIPFunc(host)
			if err != nil {
				return nil, fmt.Errorf("DNS resolution failed: %w", err)
			}
			if len(ips) == 0 {
				return nil, errors.New("no IP addresses found for host")
			}

			// Validate all resolved IPs
			for _, ip := range ips {
				if isBlockedIP(ip) {
					return nil, errors.New("URL resolves to a blocked IP address")
				}
			}

			// Connect directly to the first validated IP
			safeAddr := net.JoinHostPort(ips[0].String(), port)
			return dialer.DialContext(ctx, network, safeAddr)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
}

var safeClient = initSafeHTTPClient()

// Connect marks an integration as connected and sets its base URL.
// An optional IntegrationCredentials value stores secrets (e.g. bot tokens)
// for integrations that make real outbound API calls.
// Parameters:
//   - id: string; The identifier of the integration to connect.
//   - baseURL: string; The API base URL to use for requests.
//   - creds: IntegrationCredentials; Optional credentials for outbound API calls.
// Returns: The updated Integration, or an error if it was not found.
//
// Errors: Returns a standard Go error if preconditions fail or validation errors occur
// Side Effects: Mutates internal state or performs external system modifications (write operation).
func (r *Registry) Connect(id, baseURL string, creds ...IntegrationCredentials) (Integration, error) {
	if baseURL != "" {
		if err := validateURL(baseURL); err != nil {
			return Integration{}, err
		}
	}
	if len(creds) > 0 && creds[0].WebhookURL != "" {
		if err := validateURL(creds[0].WebhookURL); err != nil {
			return Integration{}, err
		}
	}

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

// Disconnect marks a previously connected integration as disconnected.
// Parameters:
//   - id: string; The identifier of the integration to disconnect.
// Returns: The updated Integration, or an error if it was not found.
//
// Errors: Returns a standard Go error if preconditions fail or validation errors occur
// Side Effects: Mutates internal state or performs external system modifications (write operation).
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
// Parameters:
//   - integrationID: string; The ID of the chat integration (e.g., "slack").
//   - channel: string; The target channel or space.
//   - fromAgent: string; The ID of the agent sending the message.
//   - content: string; The message payload.
//   - threadID: string; The thread context, if applicable.
//   - now: time.Time; The current timestamp.
// Returns: A ChatMessage record of the action, or an error if the integration is invalid.
//
// Errors: Returns a standard Go error if preconditions fail or validation errors occur
// Side Effects: Executes within local scope; may involve context-dependent side effects based on implementation.
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
				_ = sendTelegramMessage(context.Background(), creds.BotToken, chatID, text)
			}
		case IntegrationTypeDiscord:
			if creds.WebhookURL != "" {
				_ = sendDiscordWebhook(context.Background(), creds.WebhookURL, fromAgent, content)
			}
		}
	}

	return msg, nil
}

// TestConnection validates that the provided credentials can reach the external
// service by sending a short test message.  Use this during setup wizards
// before persisting credentials.
// Parameters:
//   - id: string; The identifier of the integration to test.
//   - creds: IntegrationCredentials; The credentials to validate.
// Returns: An error if the connection test fails.
// Errors: Fails if the integration is missing or if the external API call fails.
// Side Effects: Triggers real outbound HTTP API calls to Telegram or Discord.
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
		return sendTelegramMessage(context.Background(), active.BotToken, active.ChatID,
			"✅ Test message from One Human Corp — Telegram integration confirmed!")
	case IntegrationTypeDiscord:
		if active.WebhookURL == "" {
			return errors.New("webhook URL is required")
		}
		if err := validateURL(active.WebhookURL); err != nil {
			return err
		}
		return sendDiscordWebhook(context.Background(), active.WebhookURL, "One Human Corp",
			"✅ Test message — Discord integration confirmed!")
	default:
		// No live endpoint to test; accept unconditionally.
		return nil
	}
}

// ChatMessages retrieves all recorded chat messages, with an optional integration ID filter.
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string for all messages.
// Returns: A slice of ChatMessage records.
// Errors: None.
// Side Effects: None. Executes a read-only lock.
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
// Parameters:
//   - integrationID: string; The ID of the git integration (e.g., "github").
//   - repo: string; Target repository name.
//   - title: string; PR title.
//   - body: string; PR description.
//   - source: string; Branch name containing the changes.
//   - target: string; Base branch to merge into.
//   - createdBy: string; Agent ID opening the PR.
//   - now: time.Time; Timestamp.
// Returns: A PullRequest record of the action, or an error if parameters are invalid.
// Errors: Fails if the integration is not a git platform or if required fields are missing.
// Side Effects: Appends a new PullRequest to the internal memory store.
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
// Parameters:
//   - prID: string; The unique registry ID of the pull request.
// Returns: The updated PullRequest record.
// Errors: Fails if the PR is not found or is not in the open state.
// Side Effects: Mutates the status of the PullRequest to PRStatusMerged.
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
// Parameters:
//   - prID: string; The unique registry ID of the pull request.
// Returns: The updated PullRequest record.
// Errors: Fails if the PR is not found or is not in the open state.
// Side Effects: Mutates the status of the PullRequest to PRStatusClosed.
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
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string to return all.
// Returns: A slice of PullRequest records.
// Errors: None.
// Side Effects: None. Executes a read-only lock.
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
// Parameters:
//   - integrationID: string; The ID of the issue integration (e.g., "jira").
//   - project: string; The target project or board.
//   - title: string; The issue summary.
//   - description: string; The detailed description of the issue.
//   - createdBy: string; The ID of the agent reporting the issue.
//   - priority: IssuePriority; The urgency of the ticket.
//   - labels: []string; Categorisation tags.
//   - now: time.Time; Current timestamp.
// Returns: An Issue record of the action, or an error if parameters are invalid.
// Errors: Fails if the integration is not an issue tracker or required fields are missing.
// Side Effects: Appends a new Issue to the internal memory store.
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
// Parameters:
//   - issueID: string; The unique registry ID of the issue.
//   - status: IssueStatus; The new status phase (e.g., IssueStatusDone).
// Returns: The updated Issue record.
// Errors: Fails if the issue cannot be found.
// Side Effects: Mutates the status of the specific Issue record.
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
// Parameters:
//   - issueID: string; The unique registry ID of the issue.
//   - assignee: string; The identifier of the assigned worker.
// Returns: The updated Issue record.
// Errors: Fails if the issue cannot be found.
// Side Effects: Mutates the AssignedTo field of the specific Issue record.
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
// Parameters:
//   - integrationID: string; Filter by integration. Pass an empty string for all tickets.
// Returns: A slice of Issue records.
// Errors: None.
// Side Effects: None. Executes a read-only lock.
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

// TelegramAPIBase is the base URL for the Telegram Bot API. Override in tests to point to a mock server.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
var TelegramAPIBase = "https://api.telegram.org"

// sendTelegramMessage posts a text message to a Telegram chat via the Bot API.
func sendTelegramMessage(ctx context.Context, botToken, chatID, text string) error {
	if err := validateURL(TelegramAPIBase); err != nil {
		return err
	}

	apiURL := TelegramAPIBase + "/bot" + botToken + "/sendMessage"
	payload, _ := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := safeClient.Do(req)
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
func sendDiscordWebhook(ctx context.Context, webhookURL, username, content string) error {
	if err := validateURL(webhookURL); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"username": username,
		"content":  content,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := safeClient.Do(req)
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
