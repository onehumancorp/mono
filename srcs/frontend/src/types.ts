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
 * Summary: Defines RoleProfile.
 * Parameters: None
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
 * Summary: Defines OrganizationMember.
 * Parameters: None
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
 * Summary: Defines Organization.
 * Parameters: None
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
 * Summary: Defines MeetingMessage.
 * Parameters: None
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
 * Summary: Defines MeetingRoom.
 * Parameters: None
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
 * Summary: Defines AgentCost.
 * Parameters: None
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
 * Summary: Defines CostSummary.
 * Parameters: None
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
 * Summary: Defines StatusBucket.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * Summary: Defines AgentRuntime.
 * Parameters: None
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
 * Summary: Defines DashboardSnapshot.
 * Parameters: None
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
 * Summary: Defines DomainInfo.
 * Parameters: None
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
 * Summary: Defines MCPTool.
 * Parameters: None
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
 * Summary: Defines ApprovalStatus.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type ApprovalStatus = "PENDING" | "APPROVED" | "REJECTED";

/**
 * Summary: Defines ApprovalRequest.
 * Parameters: None
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
 * Summary: Defines HandoffPackage.
 * Parameters: None
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
 * Summary: Defines AgentIdentity.
 * Parameters: None
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
 * Summary: Defines SkillPackRole.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type SkillPackRole = {
  role: string;
  basePrompt: string;
};

/**
 * Summary: Defines SkillPack.
 * Parameters: None
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
 * Summary: Defines OrgSnapshot.
 * Parameters: None
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
 * Summary: Defines MarketplaceItem.
 * Parameters: None
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
 * Summary: Defines AnalyticsSummary.
 * Parameters: None
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
 * Summary: Defines IntegrationCategory.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IntegrationCategory = "chat" | "git" | "issues";

/**
 * Summary: Defines IntegrationStatus.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IntegrationStatus = "connected" | "disconnected" | "error";

/**
 * Summary: Defines Integration.
 * Parameters: None
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
 * Summary: Defines ChatMessage.
 * Parameters: None
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
 * Summary: Defines PullRequestStatus.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type PullRequestStatus = "open" | "merged" | "closed";

/**
 * Summary: Defines PullRequest.
 * Parameters: None
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
 * Summary: Defines IssueStatus.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IssueStatus = "open" | "in_progress" | "done" | "closed";

/**
 * Summary: Defines IssuePriority.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type IssuePriority = "low" | "medium" | "high" | "critical";

/**
 * Summary: Defines Issue.
 * Parameters: None
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
/**
 * Summary: Defines UserPublic.
 * Parameters: None
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
 * Summary: Defines Role.
 * Parameters: None
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
 * Summary: Defines LoginResponse.
 * Parameters: None
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
 * Summary: Defines Settings.
 * Parameters: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export type Settings = {
  minimaxApiKey?: string;
};
