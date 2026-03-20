// Re-export proto-generated types for use throughout the application.
// These types are generated at build time from srcs/proto/*.proto and are NOT
// stored in the repo – they appear as src/proto_types.ts in the Bazel output.
export type {
  AgentAgentMessage,
  AgentAgent,
  ApiDashboardSnapshot,
  ApiMeetingRoom,
  ApiStatusCount,
  BillingCostSummary,
  BillingAgentCostSummary,
  CommonAgentStatus,
  CommonRole,
  OrchestrationAgent,
  OrchestrationMessage,
  OrchestrationMeetingRoom,
  OrganizationOrganization,
  OrganizationTeamMember,
  OrganizationRoleProfile,
} from "./proto_types";

/**
 * Summary: Defines the playbook, prompt, and capabilities for a specific role within the AI workforce.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

/**
 * Summary: Represents an individual contributor (human or AI agent) within the organisation.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
  managerId?: string;
  isHuman?: boolean;
};

/**
 * Summary: Aggregates the hierarchy, workforce details, and role playbooks for a domain.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Encapsulates a discrete event, command, or context update passed between agents or rooms.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Maintains a persistent, sequential transcript of inter-agent collaboration.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};

/**
 * Summary: Provides aggregated cost and token usage for an individual agent.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

/**
 * Summary: Aggregates total cost and token usage for a specific organisation.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  projectedMonthlyUSD?: number;
  agents: AgentCost[];
};

/**
 * Summary: Represents an aggregated count of agents in a specific operational phase.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * Summary: Represents the current runtime state of an active, instantiated worker within the AI organisation.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type AgentRuntime = {
  id: string;
  name: string;
  role: string;
  organizationId: string;
  status: string;
};

/**
 * Summary: A point-in-time snapshot of the entire organisation's operational state, including members, meetings, costs, and active agents.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Describes a supported organisational domain template.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};

/**
 * Summary: Represents a registered tool in the MCP gateway.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Represents the lifecycle state of a guardian-gate request.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * Summary: Created by the Guardian Agent when a high-risk action requires explicit human sign-off.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Carries the structured context an agent sends to a human manager when escalating a task.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Represents the SPIFFE SVID certificate issued to an agent workload.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Pairs a role name with its override base prompt.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * Summary: An importable module that extends or overrides agent capabilities.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: A point-in-time metadata record of an organization's state.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Describes a community-published asset.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Surfaces operational health metrics.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Groups integrations by their function.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IntegrationCategory = "chat" | "git" | "issues";

/**
 * Summary: Reflects whether an integration is reachable.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * Summary: A configured external service connection.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type Integration = {
  id: string;
  name: string;
  type: string;
  category: IntegrationCategory;
  baseUrl?: string;
  status: IntegrationStatus;
  description?: string;
  /** True when real credentials (bot token, webhook URL, etc.) are stored server-side. */
  hasCredentials?: boolean;
  /** Default delivery channel / chatspace for this integration (e.g. Telegram chat_id). */
  chatspace?: string;
  createdAt: string;
};

/**
 * Summary: Represents a message dispatched through a chat service.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Tracks the lifecycle of a PR/MR.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * Summary: Represents a PR/MR opened on a git hosting platform.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: Tracks the lifecycle of an issue/ticket.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";

/**
 * Summary: Indicates ticket urgency.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Summary: Represents a ticket created in an external issue tracker.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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

// ── Auth / User Management ────────────────────────────────────────────────────

/**
 * Summary: UserPublic is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type UserPublic = {
  id: string;
  username: string;
  email: string;
  roles: string[];
  active: boolean;
  createdAt: string;
};

/**
 * Summary: Role is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type Role = {
  id: string;
  name: string;
  permissions: string[];
};

/**
 * Summary: LoginResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type LoginResponse = {
  token: string;
  user: UserPublic;
  expiresAt: string;
};

/**
 * Summary: Settings is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type Settings = {
  minimaxApiKey?: string;
};
