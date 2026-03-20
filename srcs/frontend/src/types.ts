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
 * Defines the playbook, prompt, and capabilities for a specific role within the AI workforce.

 * Summary: Provides RoleProfile functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

/**
 * Represents an individual contributor (human or AI agent) within the organisation.

 * Summary: Provides OrganizationMember functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
  managerId?: string;
  isHuman?: boolean;
};

/**
 * Aggregates the hierarchy, workforce details, and role playbooks for a domain.

 * Summary: Provides Organization functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Encapsulates a discrete event, command, or context update passed between agents or rooms.

 * Summary: Provides MeetingMessage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Maintains a persistent, sequential transcript of inter-agent collaboration.

 * Summary: Provides MeetingRoom functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};

/**
 * Provides aggregated cost and token usage for an individual agent.

 * Summary: Provides AgentCost functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

/**
 * Aggregates total cost and token usage for a specific organisation.

 * Summary: Provides CostSummary functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  projectedMonthlyUSD?: number;
  agents: AgentCost[];
};

/**
 * Represents an aggregated count of agents in a specific operational phase.

 * Summary: Provides StatusBucket functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * Represents the current runtime state of an active, instantiated worker within the AI organisation.

 * Summary: Provides AgentRuntime functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type AgentRuntime = {
  id: string;
  name: string;
  role: string;
  organizationId: string;
  status: string;
};

/**
 * A point-in-time snapshot of the entire organisation's operational state,
 * including members, meetings, costs, and active agents.

 * Summary: Provides DashboardSnapshot functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Describes a supported organisational domain template.

 * Summary: Provides DomainInfo functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};

/**
 * Represents a registered tool in the MCP gateway.

 * Summary: Provides MCPTool functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Represents the lifecycle state of a guardian-gate request.

 * Summary: Provides ApprovalStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * Created by the Guardian Agent when a high-risk action requires explicit human sign-off.

 * Summary: Provides ApprovalRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Carries the structured context an agent sends to a human manager when escalating a task.

 * Summary: Provides HandoffPackage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Represents the SPIFFE SVID certificate issued to an agent workload.

 * Summary: Provides AgentIdentity functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Pairs a role name with its override base prompt.

 * Summary: Provides SkillPackRole functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * An importable module that extends or overrides agent capabilities.

 * Summary: Provides SkillPack functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * A point-in-time metadata record of an organization's state.

 * Summary: Provides OrgSnapshot functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Describes a community-published asset.

 * Summary: Provides MarketplaceItem functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Surfaces operational health metrics.

 * Summary: Provides AnalyticsSummary functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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

/** Groups integrations by their function.
 * Summary: Provides IntegrationCategory functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type IntegrationCategory = "chat" | "git" | "issues";

/** Reflects whether an integration is reachable.
 * Summary: Provides IntegrationStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * A configured external service connection.

 * Summary: Provides Integration functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Represents a message dispatched through a chat service.

 * Summary: Provides ChatMessage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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

/** Tracks the lifecycle of a PR/MR.
 * Summary: Provides PullRequestStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * Represents a PR/MR opened on a git hosting platform.

 * Summary: Provides PullRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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

/** Tracks the lifecycle of an issue/ticket.
 * Summary: Provides IssueStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";

/** Indicates ticket urgency.
 * Summary: Provides IssuePriority functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Represents a ticket created in an external issue tracker.

 * Summary: Provides Issue functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Summary: Provides UserPublic functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
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
 * Summary: Provides Role functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type Role = {
  id: string;
  name: string;
  permissions: string[];
};

/**
 * Summary: Provides LoginResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type LoginResponse = {
  token: string;
  user: UserPublic;
  expiresAt: string;
};

/**
 * Summary: Provides Settings functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type Settings = {
  minimaxApiKey?: string;
};
