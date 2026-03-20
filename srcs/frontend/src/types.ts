/**
 * RoleProfile defines the operational constraints, capabilities, and system prompts for a specific role.
 *
 * @summary Defines the structure and fields for RoleProfile within the application.
 * @param none
 * @returns {RoleProfile} a RoleProfile configuration.
 * @throws none
 * @sideeffects none
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

/**
 * OrganizationMember represents an individual entity, either human or AI, within an organization.
 *
 * @summary Defines the structure and fields for OrganizationMember within the application.
 * @param none
 * @returns {OrganizationMember} a member entity.
 * @throws none
 * @sideeffects none
 */
export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
  managerId?: string;
  isHuman?: boolean;
};

/**
 * Organization represents the entire hierarchical structure and metadata of a corporate entity.
 *
 * @summary Defines the structure and fields for Organization within the application.
 * @param none
 * @returns {Organization} the organization schema.
 * @throws none
 * @sideeffects none
 */
export type Organization = {
  id: string;
  name: string;
  domain: string;
  ceoId?: string;
  members: OrganizationMember[];
  roleProfiles: RoleProfile[];
};

/**
 * MeetingMessage represents a serialized event or conversation snippet within the Pub/Sub system.
 *
 * @summary Defines the structure and fields for MeetingMessage within the application.
 * @param none
 * @returns {MeetingMessage} a single transcript message.
 * @throws none
 * @sideeffects none
 */
export type MeetingMessage = {
  id: string;
  fromAgent: string;
  toAgent: string;
  type: string;
  content: string;
  meetingId: string;
  occurredAt: string;
};

/**
 * MeetingRoom defines a synchronous collaboration space for multiple agents.
 *
 * @summary Defines the structure and fields for MeetingRoom within the application.
 * @param none
 * @returns {MeetingRoom} a meeting state object.
 * @throws none
 * @sideeffects none
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};

/**
 * AgentCost tracks the specific token utilization and financial burn of a single AI actor.
 *
 * @summary Defines the structure and fields for AgentCost within the application.
 * @param none
 * @returns {AgentCost} the cost details for an agent.
 * @throws none
 * @sideeffects none
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

/**
 * CostSummary represents the total aggregated API usage and cost for an entire organization.
 *
 * @summary Defines the structure and fields for CostSummary within the application.
 * @param none
 * @returns {CostSummary} aggregate financial and token usage statistics.
 * @throws none
 * @sideeffects none
 */
export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  projectedMonthlyUSD?: number;
  agents: AgentCost[];
};

/**
 * StatusBucket aggregates the number of system operations grouped by their current phase.
 *
 * @summary Defines the structure and fields for StatusBucket within the application.
 * @param none
 * @returns {StatusBucket} grouped status counts.
 * @throws none
 * @sideeffects none
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * AgentRuntime maps the active execution state and organizational assignment of an AI worker.
 *
 * @summary Defines the structure and fields for AgentRuntime within the application.
 * @param none
 * @returns {AgentRuntime} runtime execution metadata.
 * @throws none
 * @sideeffects none
 */
export type AgentRuntime = {
  id: string;
  name: string;
  role: string;
  organizationId: string;
  status: string;
};

/**
 * DashboardSnapshot provides a complete, unified view of the organization's current state for the frontend.
 *
 * @summary Defines the structure and fields for DashboardSnapshot within the application.
 * @param none
 * @returns {DashboardSnapshot} the unified dashboard view.
 * @throws none
 * @sideeffects none
 */
export type DashboardSnapshot = {
  organization: Organization;
  meetings: MeetingRoom[];
  costs: CostSummary;
  agents: AgentRuntime[];
  statuses: StatusBucket[];
  updatedAt: string;
};

/**
 * DomainInfo holds metadata for available organizational templates.
 *
 * @summary Defines the structure and fields for DomainInfo within the application.
 * @param none
 * @returns {DomainInfo} domain metadata.
 * @throws none
 * @sideeffects none
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};

/**
 * MCPTool defines a Model Context Protocol tool available to the agent swarm.
 *
 * @summary Defines the structure and fields for MCPTool within the application.
 * @param none
 * @returns {MCPTool} the definition of an MCP capability.
 * @throws none
 * @sideeffects none
 */
export type MCPTool = {
  id: string;
  name: string;
  description: string;
  category: string;
  status: string;
};

// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * ApprovalStatus defines the current state of a human-in-the-loop approval request.
 *
 * @summary Defines the structure and fields for ApprovalStatus within the application.
 * @param none
 * @returns {ApprovalStatus} an enum reflecting approval state.
 * @throws none
 * @sideeffects none
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * ApprovalRequest represents a pending action requiring human verification.
 *
 * @summary Defines the structure and fields for ApprovalRequest within the application.
 * @param none
 * @returns {ApprovalRequest} an approval struct with risk and cost details.
 * @throws none
 * @sideeffects none
 */
export type ApprovalRequest = {
  id: string;
  agentId: string;
  action: string;
  reason: string;
  estimatedCostUsd: number;
  riskLevel: "low" | "medium" | "high" | "critical";
  status: ApprovalStatus;
  createdAt: string;
  decidedAt?: string;
  decidedBy?: string;
};

// ── Warm Handoff ──────────────────────────────────────────────────────────────

/**
 * HandoffPackage contains contextual state transferred from an agent to a human.
 *
 * @summary Defines the structure and fields for HandoffPackage within the application.
 * @param none
 * @returns {HandoffPackage} failure state and agent intent context.
 * @throws none
 * @sideeffects none
 */
export type HandoffPackage = {
  id: string;
  fromAgentId: string;
  toHumanRole: string;
  intent: string;
  failedAttempts: number;
  currentState: string;
  status: "pending" | "acknowledged" | "resolved";
  createdAt: string;
};

// ── Identity Management ───────────────────────────────────────────────────────

/**
 * AgentIdentity holds the SPIFFE identity metadata for a deployed AI agent.
 *
 * @summary Defines the structure and fields for AgentIdentity within the application.
 * @param none
 * @returns {AgentIdentity} SPIRE identity records.
 * @throws none
 * @sideeffects none
 */
export type AgentIdentity = {
  agentId: string;
  svid: string;
  trustDomain: string;
  issuedAt: string;
  expiresAt: string;
};

// ── Skill Import Framework ────────────────────────────────────────────────────

/**
 * SkillPackRole defines an individual system prompt template within a SkillPack.
 *
 * @summary Defines the structure and fields for SkillPackRole within the application.
 * @param none
 * @returns {SkillPackRole} an individual role prompt.
 * @throws none
 * @sideeffects none
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * SkillPack groups related operational capabilities for importing into an organization.
 *
 * @summary Defines the structure and fields for SkillPack within the application.
 * @param none
 * @returns {SkillPack} imported tools and capabilities mapping.
 * @throws none
 * @sideeffects none
 */
export type SkillPack = {
  id: string;
  name: string;
  domain: string;
  description: string;
  source: "builtin" | "custom" | "marketplace";
  author?: string;
  roles: SkillPackRole[];
  importedAt: string;
};

// ── Org Snapshot & Recovery ───────────────────────────────────────────────────

/**
 * OrgSnapshot captures a point-in-time backup of the entire organization state.
 *
 * @summary Defines the structure and fields for OrgSnapshot within the application.
 * @param none
 * @returns {OrgSnapshot} a point-in-time data snapshot.
 * @throws none
 * @sideeffects none
 */
export type OrgSnapshot = {
  id: string;
  label: string;
  orgId: string;
  orgName: string;
  domain: string;
  agentCount: number;
  meetingCount: number;
  messageCount: number;
  createdAt: string;
};

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * MarketplaceItem represents a downloadable capability available to organizations.
 *
 * @summary Defines the structure and fields for MarketplaceItem within the application.
 * @param none
 * @returns {MarketplaceItem} details for a downloadable feature.
 * @throws none
 * @sideeffects none
 */
export type MarketplaceItem = {
  id: string;
  name: string;
  type: "agent" | "domain" | "skill_pack" | "tool";
  author: string;
  description: string;
  downloads: number;
  rating: number;
  tags: string[];
};

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * AnalyticsSummary calculates high-level operational metrics across the platform.
 *
 * @summary Defines the structure and fields for AnalyticsSummary within the application.
 * @param none
 * @returns {AnalyticsSummary} high-level efficiency and performance metadata.
 * @throws none
 * @sideeffects none
 */
export type AnalyticsSummary = {
  humanAgentRatio: number;
  totalAgents: number;
  totalHumans: number;
  auditFidelityPct: number;
  resumptionLatencyMs: number;
  pendingApprovals: number;
  activeHandoffs: number;
  tokenVelocity: number;
};

// ── External Integrations ─────────────────────────────────────────────────────

/**
 * IntegrationCategory represents the functional domain of an external integration.
 *
 * @summary Defines the structure and fields for IntegrationCategory within the application.
 * @param none
 * @returns {IntegrationCategory} an integration domain enum.
 * @throws none
 * @sideeffects none
 */
export type IntegrationCategory = "chat" | "git" | "issues";
/**
 * IntegrationStatus tracks the health and connectivity of an external integration.
 *
 * @summary Defines the structure and fields for IntegrationStatus within the application.
 * @param none
 * @returns {IntegrationStatus} connectivity enum state.
 * @throws none
 * @sideeffects none
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * Integration describes a configured connection to a third-party service.
 *
 * @summary Defines the structure and fields for Integration within the application.
 * @param none
 * @returns {Integration} full integration configuration.
 * @throws none
 * @sideeffects none
 */
export type Integration = {
  id: string;
  name: string;
  type: string;
  category: IntegrationCategory;
  baseUrl?: string;
  status: IntegrationStatus;
  description?: string;
  createdAt: string;
};

/**
 * ChatMessage represents a single communication event synced from an external chat tool.
 *
 * @summary Defines the structure and fields for ChatMessage within the application.
 * @param none
 * @returns {ChatMessage} chat log structure.
 * @throws none
 * @sideeffects none
 */
export type ChatMessage = {
  id: string;
  integrationId: string;
  channel: string;
  fromAgent: string;
  content: string;
  threadId?: string;
  sentAt: string;
};

/**
 * PullRequestStatus tracks the lifecycle of code changes in a Git provider.
 *
 * @summary Defines the structure and fields for PullRequestStatus within the application.
 * @param none
 * @returns {PullRequestStatus} Git provider PR state enum.
 * @throws none
 * @sideeffects none
 */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * PullRequest represents a proposed code modification mapped from an external VCS.
 *
 * @summary Defines the structure and fields for PullRequest within the application.
 * @param none
 * @returns {PullRequest} Git PR details.
 * @throws none
 * @sideeffects none
 */
export type PullRequest = {
  id: string;
  integrationId: string;
  repository: string;
  title: string;
  body: string;
  sourceBranch: string;
  targetBranch: string;
  url: string;
  createdByAgent: string;
  status: PullRequestStatus;
  createdAt: string;
};

/**
 * IssueStatus represents the progress state of a tracked ticket or bug.
 *
 * @summary Defines the structure and fields for IssueStatus within the application.
 * @param none
 * @returns {IssueStatus} ticket completion phase enum.
 * @throws none
 * @sideeffects none
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";
/**
 * IssuePriority indicates the urgency of resolving an external ticket.
 *
 * @summary Defines the structure and fields for IssuePriority within the application.
 * @param none
 * @returns {IssuePriority} ticket severity enum.
 * @throws none
 * @sideeffects none
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Issue holds the details of a work item synced from an external task tracker.
 *
 * @summary Defines the structure and fields for Issue within the application.
 * @param none
 * @returns {Issue} structured ticketing data.
 * @throws none
 * @sideeffects none
 */
export type Issue = {
  id: string;
  integrationId: string;
  project: string;
  title: string;
  description: string;
  priority: IssuePriority;
  status: IssueStatus;
  assignedTo?: string;
  labels?: string[];
  createdByAgent: string;
  url: string;
  createdAt: string;
};
