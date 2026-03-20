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
 * Intent: Defines the playbook, prompt, and capabilities for a specific role within the AI workforce.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

/**
 * Intent: Represents an individual contributor (human or AI agent) within the organisation.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
  managerId?: string;
  isHuman?: boolean;
};

/**
 * Intent: Aggregates the hierarchy, workforce details, and role playbooks for a domain.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Encapsulates a discrete event, command, or context update passed between agents or rooms.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Maintains a persistent, sequential transcript of inter-agent collaboration.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};

/**
 * Intent: Provides aggregated cost and token usage for an individual agent.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

/**
 * Intent: Aggregates total cost and token usage for a specific organisation.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  projectedMonthlyUSD?: number;
  agents: AgentCost[];
};

/**
 * Intent: Represents an aggregated count of agents in a specific operational phase.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * Intent: Represents the current runtime state of an active, instantiated worker within the AI organisation.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type AgentRuntime = {
  id: string;
  name: string;
  role: string;
  organizationId: string;
  status: string;
};

/**
 * Intent: A point-in-time snapshot of the entire organisation's operational state, including members, meetings, costs, and active agents.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Describes a supported organisational domain template.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};

/**
 * Intent: Represents a registered tool in the MCP gateway.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Represents the lifecycle state of a guardian-gate request.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * Intent: Created by the Guardian Agent when a high-risk action requires explicit human sign-off.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Carries the structured context an agent sends to a human manager when escalating a task.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Represents the SPIFFE SVID certificate issued to an agent workload.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Pairs a role name with its override base prompt.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * Intent: An importable module that extends or overrides agent capabilities.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: A point-in-time metadata record of an organization's state.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Describes a community-published asset.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Surfaces operational health metrics.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to IntegrationCategory.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type IntegrationCategory = "chat" | "git" | "issues";

/**
 * Intent: Handles operations related to IntegrationStatus.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * Intent: A configured external service connection.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Represents a message dispatched through a chat service.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to PullRequestStatus.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * Intent: Represents a PR/MR opened on a git hosting platform.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to IssueStatus.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";

/**
 * Intent: Handles operations related to IssuePriority.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Intent: Represents a ticket created in an external issue tracker.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to UserPublic.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to Role.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type Role = {
  id: string;
  name: string;
  permissions: string[];
};
/**
 * Intent: Handles operations related to LoginResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type LoginResponse = {
  token: string;
  user: UserPublic;
  expiresAt: string;
};
/**
 * Intent: Handles operations related to Settings.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export type Settings = {
  minimaxApiKey?: string;
};
