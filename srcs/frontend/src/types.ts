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
 * @summary Defines the playbook, prompt, and capabilities for a specific role within the AI workforce.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};
/**
 * @summary Represents an individual contributor (human or AI agent) within the organisation.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
  managerId?: string;
  isHuman?: boolean;
};
/**
 * @summary Aggregates the hierarchy, workforce details, and role playbooks for a domain.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Encapsulates a discrete event, command, or context update passed between agents or rooms.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Maintains a persistent, sequential transcript of inter-agent collaboration.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};
/**
 * @summary Provides aggregated cost and token usage for an individual agent.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};
/**
 * @summary Aggregates total cost and token usage for a specific organisation.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  projectedMonthlyUSD?: number;
  agents: AgentCost[];
};
/**
 * @summary Represents an aggregated count of agents in a specific operational phase.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type StatusBucket = {
  status: string;
  count: number;
};
/**
 * @summary Represents the current runtime state of an active, instantiated worker within the AI organisation.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type AgentRuntime = {
  id: string;
  name: string;
  role: string;
  organizationId: string;
  status: string;
};
/**
 * @summary A point-in-time snapshot of the entire organisation's operational state, including members, meetings, costs, and active agents.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Describes a supported organisational domain template.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};
/**
 * @summary Represents a registered tool in the MCP gateway.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type MCPTool = {
  id: string;
  name: string;
  description: string;
  category: string;
  status: string;
};

/**
 * ── Approval / Confidence Gating ─────────────────────────────────────────────
 * @summary Represents the lifecycle state of a guardian-gate request.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";
/**
 * @summary Created by the Guardian Agent when a high-risk action requires explicit human sign-off.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── Warm Handoff ──────────────────────────────────────────────────────────────
 * @summary Carries the structured context an agent sends to a human manager when escalating a task.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── Identity Management ───────────────────────────────────────────────────────
 * @summary Represents the SPIFFE SVID certificate issued to an agent workload.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type AgentIdentity = {
  agentId: string;
  svid: string;
  trustDomain: string;
  issuedAt: string;
  expiresAt: string;
};

/**
 * ── Skill Import Framework ────────────────────────────────────────────────────
 * @summary Pairs a role name with its override base prompt.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};
/**
 * @summary An importable module that extends or overrides agent capabilities.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── Org Snapshot & Recovery ───────────────────────────────────────────────────
 * @summary A point-in-time metadata record of an organization's state.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── Marketplace ───────────────────────────────────────────────────────────────
 * @summary Describes a community-published asset.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── Real-time Analytics ───────────────────────────────────────────────────────
 * @summary Surfaces operational health metrics.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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

/**
 * ── External Integrations ─────────────────────────────────────────────────────
 * @summary Defines IntegrationCategory.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type IntegrationCategory = "chat" | "git" | "issues";
/**
 * @summary Defines IntegrationStatus.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";
/**
 * @summary A configured external service connection.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Represents a message dispatched through a chat service.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Defines PullRequestStatus.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type PullRequestStatus = "open" | "merged" | "closed";
/**
 * @summary Represents a PR/MR opened on a git hosting platform.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Defines IssueStatus.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";
/**
 * @summary Defines IssuePriority.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";
/**
 * @summary Represents a ticket created in an external issue tracker.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
/**
 * @summary ── Auth / User Management ────────────────────────────────────────────────────
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
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
 * @summary Defines Role.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type Role = {
  id: string;
  name: string;
  permissions: string[];
};
/**
 * @summary Defines LoginResponse.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type LoginResponse = {
  token: string;
  user: UserPublic;
  expiresAt: string;
};
/**
 * @summary Defines Settings.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: None
 */
export type Settings = {
  minimaxApiKey?: string;
  theme?: string;
};
