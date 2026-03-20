package dashboard

import (
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

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

type hireRequest struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Model        string `json:"model,omitempty"`
	ProviderType string `json:"providerType,omitempty"`
}

type fireRequest struct {
	AgentID string `json:"agentId"`
}

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

type ApprovalRequest struct {
	ID               string         `json:"id"`
	AgentID          string         `json:"agentId"`
	Action           string         `json:"action"`
	Reason           string         `json:"reason"`
	EstimatedCostUSD float64        `json:"estimatedCostUsd"`
	RiskLevel        string         `json:"riskLevel"`
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
	Decision   string `json:"decision"`
	DecidedBy  string `json:"decidedBy"`
}

type HandoffPackage struct {
	ID             string    `json:"id"`
	FromAgentID    string    `json:"fromAgentId"`
	ToHumanRole    string    `json:"toHumanRole"`
	Intent         string    `json:"intent"`
	FailedAttempts int       `json:"failedAttempts"`
	CurrentState   string    `json:"currentState"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type handoffCreateRequest struct {
	FromAgentID    string `json:"fromAgentId"`
	ToHumanRole    string `json:"toHumanRole"`
	Intent         string `json:"intent"`
	FailedAttempts int    `json:"failedAttempts"`
	CurrentState   string `json:"currentState"`
}

type AgentIdentity struct {
	AgentID     string    `json:"agentId"`
	SVID        string    `json:"svid"`
	TrustDomain string    `json:"trustDomain"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type SkillPackRole struct {
	Role       string `json:"role"`
	BasePrompt string `json:"basePrompt"`
}

type SkillPack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Domain      string          `json:"domain"`
	Description string          `json:"description"`
	Source      string          `json:"source"`
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

type MarketplaceItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
}

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

type MCPTool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

type DomainInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type providerAuthRequest struct {
	ProviderType string            `json:"providerType"`
	APIKey       string            `json:"apiKey,omitempty"`
	OAuthToken   string            `json:"oauthToken,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

type chatTestRequest struct {
	IntegrationID string `json:"integrationId"`
	BotToken      string `json:"botToken,omitempty"`
	ChatID        string `json:"chatId,omitempty"`
	WebhookURL    string `json:"webhookUrl,omitempty"`
	APIToken      string `json:"apiToken,omitempty"`
}

type mcpInvokeRequest struct {
	ToolID string         `json:"toolId"`
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

type integrationConnectRequest struct {
	IntegrationID string `json:"integrationId"`
	BaseURL       string `json:"baseUrl,omitempty"`
	BotToken      string `json:"botToken,omitempty"`
	ChatID        string `json:"chatId,omitempty"`
	WebhookURL    string `json:"webhookUrl,omitempty"`
	APIToken      string `json:"apiToken,omitempty"`
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

type TrustAgreementStatus string

const (
	TrustStatusPending TrustAgreementStatus = "PENDING"
	TrustStatusActive  TrustAgreementStatus = "ACTIVE"
	TrustStatusRevoked TrustAgreementStatus = "REVOKED"
)

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

type IncidentSeverity string

const (
	SeverityP0 IncidentSeverity = "P0"
	SeverityP1 IncidentSeverity = "P1"
	SeverityP2 IncidentSeverity = "P2"
)

type IncidentStatus string

const (
	IncidentStatusInvestigating IncidentStatus = "INVESTIGATING"
	IncidentStatusProposed      IncidentStatus = "PROPOSED"
	IncidentStatusResolved      IncidentStatus = "RESOLVED"
)

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

type ComputeProfile struct {
	RoleID             string    `json:"roleId"`
	MinVRAMGB          int       `json:"minVramGb"`
	PreferredGPUType   string    `json:"preferredGpuType"`
	SchedulingPriority int       `json:"schedulingPriority"`
	CreatedAt          time.Time `json:"createdAt"`
}

type computeProfileRequest struct {
	RoleID             string `json:"roleId"`
	MinVRAMGB          int    `json:"minVramGb"`
	PreferredGPUType   string `json:"preferredGpuType"`
	SchedulingPriority int    `json:"schedulingPriority"`
}

type ClusterStatus struct {
	Region         string    `json:"region"`
	Status         string    `json:"status"`
	LatencyMS      int       `json:"latencyMs"`
	AvailableNodes int       `json:"availableNodes"`
	CheckedAt      time.Time `json:"checkedAt"`
}

const defaultBudgetAlertNotifyPct = 0.8

type BudgetAlert struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ThresholdUSD   float64   `json:"thresholdUsd"`
	NotifyAtPct    float64   `json:"notifyAtPct"`
	Triggered      bool      `json:"triggered"`
	CreatedAt      time.Time `json:"createdAt"`
}

type budgetAlertRequest struct {
	OrganizationID string  `json:"organizationId"`
	ThresholdUSD   float64 `json:"thresholdUsd"`
	NotifyAtPct    float64 `json:"notifyAtPct"`
}

type PipelineStatus string

const (
	PipelineStatusPending      PipelineStatus = "PENDING"
	PipelineStatusImplementing PipelineStatus = "IMPLEMENTING"
	PipelineStatusTesting      PipelineStatus = "TESTING"
	PipelineStatusStaging      PipelineStatus = "STAGING"
	PipelineStatusPromoted     PipelineStatus = "PROMOTED"
	PipelineStatusFailed       PipelineStatus = "FAILED"
)

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
