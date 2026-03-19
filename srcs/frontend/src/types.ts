/**
 * Defines the playbook, prompt, and capabilities for a specific role within the AI workforce.
 */
export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

/**
 * Represents an individual contributor (human or AI agent) within the organisation.
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
 */
export type MeetingRoom = {
  id: string;
  agenda?: string;
  participants: string[];
  transcript: MeetingMessage[];
};

/**
 * Provides aggregated cost and token usage for an individual agent.
 */
export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

/**
 * Aggregates total cost and token usage for a specific organisation.
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
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * Represents the current runtime state of an active, instantiated worker within the AI organisation.
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
 */
export type DomainInfo = {
  id: string;
  name: string;
  description: string;
};

/**
 * Represents a registered tool in the MCP gateway.
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
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * Created by the Guardian Agent when a high-risk action requires explicit human sign-off.
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
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * An importable module that extends or overrides agent capabilities.
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

/** Groups integrations by their function. */
export type IntegrationCategory = "chat" | "git" | "issues";

/** Reflects whether an integration is reachable. */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * A configured external service connection.
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
 * Represents a message dispatched through a chat service.
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

/** Tracks the lifecycle of a PR/MR. */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * Represents a PR/MR opened on a git hosting platform.
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

/** Tracks the lifecycle of an issue/ticket. */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";

/** Indicates ticket urgency. */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Represents a ticket created in an external issue tracker.
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
