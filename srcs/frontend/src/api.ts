import type {
  AnalyticsSummary,
  AgentIdentity,
  ApprovalRequest,
  CostSummary,
  DashboardSnapshot,
  DomainInfo,
  HandoffPackage,
  MarketplaceItem,
  MCPTool,
  MeetingRoom,
  OrgSnapshot,
  Organization,
  SkillPack,
} from "./types";

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

/**
 * fetchOrganization retrieves the full state and structure of the requested organization.
 *
 * @summary Executes the fetchOrganization API operation against the backend server.
 * @param none
 * @returns {Promise<Organization>} an asynchronous Promise resolving to the Organization object.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchOrganization(): Promise<Organization> {
  return getJSON<Organization>("/api/org");
}

/**
 * fetchMeetings retrieves all active virtual meeting spaces and transcripts within the organization.
 *
 * @summary Executes the fetchMeetings API operation against the backend server.
 * @param none
 * @returns {Promise<MeetingRoom[]>} an asynchronous Promise resolving to a list of MeetingRooms.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchMeetings(): Promise<MeetingRoom[]> {
  return getJSON<MeetingRoom[]>("/api/meetings").then(normalizeMeetings);
}

function normalizeCosts(response: Record<string, unknown>): CostSummary {
  const agents = Array.isArray(response.agents) ? response.agents : [];

  return {
    organizationID: String(response.organizationID ?? response.organizationId ?? ""),
    totalTokens: Number(response.totalTokens ?? 0),
    totalCostUSD: Number(response.totalCostUSD ?? response.totalCostUsd ?? 0),
    projectedMonthlyUSD: response.projectedMonthlyUSD !== undefined
      ? Number(response.projectedMonthlyUSD ?? response.projectedMonthlyUsd ?? 0)
      : undefined,
    agents: agents.map((agent) => {
      const value = agent as Record<string, unknown>;
      return {
        agentID: String(value.agentID ?? value.agentId ?? ""),
        model: String(value.model ?? ""),
        tokenUsed: Number(value.tokenUsed ?? 0),
        costUSD: Number(value.costUSD ?? value.costUsd ?? 0),
      };
    }),
  };
}

function normalizeMeetings(meetings: MeetingRoom[]): MeetingRoom[] {
  return meetings.map((meeting) => ({
    ...meeting,
    transcript: meeting.transcript ?? [],
  }));
}

/**
 * fetchCosts requests real-time, model-aware token pricing data for all active agents.
 *
 * @summary Executes the fetchCosts API operation against the backend server.
 * @param none
 * @returns {Promise<CostSummary>} an asynchronous Promise resolving to current CostSummary metrics.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export async function fetchCosts(): Promise<CostSummary> {
  const response = await getJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

/**
 * fetchDashboard downloads an all-inclusive snapshot of the organization's state for UI rendering.
 *
 * @summary Executes the fetchDashboard API operation against the backend server.
 * @param none
 * @returns {Promise<DashboardSnapshot>} an asynchronous Promise resolving to the DashboardSnapshot.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await getJSON<Record<string, unknown>>("/api/dashboard");
  const rawOrganization = (response.organization ?? {}) as Record<string, unknown>;
  const rawMeetings = Array.isArray(response.meetings)
    ? (response.meetings as MeetingRoom[])
    : [];
  const rawCosts = (response.costs ?? {}) as Record<string, unknown>;
  const rawAgents = Array.isArray(response.agents) ? response.agents : [];
  const rawStatuses = Array.isArray(response.statuses) ? response.statuses : [];

  return {
    organization: {
      id: String(rawOrganization.id ?? ""),
      name: String(rawOrganization.name ?? ""),
      domain: String(rawOrganization.domain ?? ""),
      ceoId: rawOrganization.ceoId !== undefined ? String(rawOrganization.ceoId) : undefined,
      members: Array.isArray(rawOrganization.members)
        ? (rawOrganization.members as Organization["members"])
        : [],
      roleProfiles: Array.isArray(rawOrganization.roleProfiles)
        ? (rawOrganization.roleProfiles as Organization["roleProfiles"])
        : [],
    },
    meetings: normalizeMeetings(rawMeetings),
    costs: normalizeCosts(rawCosts),
    agents: rawAgents.map((agent) => {
      const value = agent as Record<string, unknown>;
      return {
        id: String(value.id ?? ""),
        name: String(value.name ?? ""),
        role: String(value.role ?? ""),
        organizationId: String(value.organizationId ?? value.organizationID ?? ""),
        status: String(value.status ?? ""),
      };
    }),
    statuses: rawStatuses.map((status) => {
      const value = status as Record<string, unknown>;
      return {
        status: String(value.status ?? "UNKNOWN"),
        count: Number(value.count ?? 0),
      };
    }),
    updatedAt: String(response.updatedAt ?? new Date().toISOString()),
  };
}

/**
 * sendMessage posts a new message payload to a specific agent within a virtual meeting room.
 *
 * @summary Executes the sendMessage API operation against the backend server.
 * @param {Object} form - Form data encapsulating sender, receiver, meeting context, and content.
 * @returns {Promise<void>} an asynchronous Promise that resolves upon successful delivery.
 * @throws throws an Error if the server response is not ok.
 * @sideeffects mutates server-side conversation state.
 */
export async function sendMessage(form: {
  fromAgent: string;
  toAgent: string;
  meetingId: string;
  messageType: string;
  content: string;
}): Promise<void> {
  const params = new URLSearchParams(form);
  const response = await fetch("/api/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      Accept: "application/json",
    },
    body: params.toString(),
    redirect: "follow",
  });

  if (!response.ok) {
    throw new Error(`Failed to send message: ${response.status}`);
  }
}

/**
 * hireAgent dynamically instantiates a new AI agent within the active organization.
 *
 * @summary Executes the hireAgent API operation against the backend server.
 * @param {string} name - the designated name for the new agent.
 * @param {string} role - the designated organizational role.
 * @returns {Promise<DashboardSnapshot>} an asynchronous Promise resolving to the updated DashboardSnapshot.
 * @throws throws an error on network or parsing failure.
 * @sideeffects creates a new backend agent instance.
 */
export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

/**
 * fireAgent terminates an active AI agent's execution and removes it from the organization.
 *
 * @summary Executes the fireAgent API operation against the backend server.
 * @param {string} agentId - the unique identifier of the agent to terminate.
 * @returns {Promise<DashboardSnapshot>} an asynchronous Promise resolving to the updated DashboardSnapshot.
 * @throws throws an error on network or parsing failure.
 * @sideeffects destroys an existing backend agent instance.
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

/**
 * fetchDomains retrieves available corporate domain templates (e.g. Software, Accounting).
 *
 * @summary Executes the fetchDomains API operation against the backend server.
 * @param none
 * @returns {Promise<DomainInfo[]>} an asynchronous Promise resolving to the list of available domains.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

/**
 * fetchMCPTools queries the backend for registered Model Context Protocol tools.
 *
 * @summary Executes the fetchMCPTools API operation against the backend server.
 * @param none
 * @returns {Promise<MCPTool[]>} an asynchronous Promise resolving to an array of MCPTools.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

/**
 * seedScenario forces the backend to populate mock entities to simulate an operational state.
 *
 * @summary Executes the seedScenario API operation against the backend server.
 * @param {string} scenario - string identifier for the target scenario to seed.
 * @returns {Promise<DashboardSnapshot>} an asynchronous Promise resolving to the updated snapshot.
 * @throws throws an error on network or parsing failure.
 * @sideeffects severely alters backend database state by injecting mock data.
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * fetchApprovals retrieves pending operations awaiting human security clearance.
 *
 * @summary Executes the fetchApprovals API operation against the backend server.
 * @param none
 * @returns {Promise<ApprovalRequest[]>} an asynchronous Promise resolving to pending approval requests.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

/**
 * requestApproval submits a high-risk operation from an agent for explicit human authorization.
 *
 * @summary Executes the requestApproval API operation against the backend server.
 * @param {Object} body - approval request payload detailing agent, cost, and risk level.
 * @returns {Promise<ApprovalRequest>} an asynchronous Promise resolving to the generated request.
 * @throws throws an error on network or parsing failure.
 * @sideeffects writes new pending approval state.
 */
export function requestApproval(body: {
  agentId: string;
  action: string;
  reason?: string;
  estimatedCostUsd?: number;
  riskLevel?: string;
}): Promise<ApprovalRequest> {
  return postJSON<ApprovalRequest>("/api/approvals/request", body);
}

/**
 * decideApproval submits the human manager's decision regarding a pending agent operation.
 *
 * @summary Executes the decideApproval API operation against the backend server.
 * @param {string} approvalId - target approval request ID.
 * @param {"approve" | "reject"} decision - explicit decision string.
 * @param {string} [decidedBy] - identifier of the human making the decision.
 * @returns {Promise<ApprovalRequest[]>} an asynchronous Promise resolving to the updated requests list.
 * @throws throws an error on network or parsing failure.
 * @sideeffects mutates the backend approval queue.
 */
export function decideApproval(
  approvalId: string,
  decision: "approve" | "reject",
  decidedBy?: string,
): Promise<ApprovalRequest[]> {
  return postJSON<ApprovalRequest[]>("/api/approvals/decide", { approvalId, decision, decidedBy });
}

// ── Warm Handoff ──────────────────────────────────────────────────────────────

/**
 * fetchHandoffs lists all incomplete operations an agent has explicitly escalated to a human.
 *
 * @summary Executes the fetchHandoffs API operation against the backend server.
 * @param none
 * @returns {Promise<HandoffPackage[]>} an asynchronous Promise resolving to active handoffs.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

/**
 * createHandoff manually triggers an escalation process transferring operational context to a human.
 *
 * @summary Executes the createHandoff API operation against the backend server.
 * @param {Object} body - escalation details including failing state.
 * @returns {Promise<HandoffPackage>} an asynchronous Promise resolving to the created handoff struct.
 * @throws throws an error on network or parsing failure.
 * @sideeffects creates a new backend escalation ticket.
 */
export function createHandoff(body: {
  fromAgentId: string;
  toHumanRole?: string;
  intent: string;
  failedAttempts?: number;
  currentState?: string;
}): Promise<HandoffPackage> {
  return postJSON<HandoffPackage>("/api/handoffs", body);
}

// ── Identity Management ───────────────────────────────────────────────────────

/**
 * fetchIdentities returns the SPIFFE/SPIRE certificate data validating agent authorization.
 *
 * @summary Executes the fetchIdentities API operation against the backend server.
 * @param none
 * @returns {Promise<AgentIdentity[]>} an asynchronous Promise resolving to certificate metadata.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

/**
 * fetchSkillPacks loads all external capability modules installed on the host organization.
 *
 * @summary Executes the fetchSkillPacks API operation against the backend server.
 * @param none
 * @returns {Promise<SkillPack[]>} an asynchronous Promise resolving to a list of loaded skills.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

/**
 * importSkillPack dynamically downloads and instantiates external capabilities into the environment.
 *
 * @summary Executes the importSkillPack API operation against the backend server.
 * @param {Object} body - details defining the new capabilities and metadata.
 * @returns {Promise<SkillPack>} an asynchronous Promise resolving to the successful import status.
 * @throws throws an error on network or parsing failure.
 * @sideeffects mutates available agent prompts and organization capabilities.
 */
export function importSkillPack(body: {
  name: string;
  domain: string;
  description?: string;
  source?: string;
  author?: string;
}): Promise<SkillPack> {
  return postJSON<SkillPack>("/api/skills/import", body);
}

// ── Snapshots ─────────────────────────────────────────────────────────────────

/**
 * fetchSnapshots retrieves a history of point-in-time organization backups.
 *
 * @summary Executes the fetchSnapshots API operation against the backend server.
 * @param none
 * @returns {Promise<OrgSnapshot[]>} an asynchronous Promise resolving to the backup manifest.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

/**
 * createSnapshot triggers the underlying CSI system to serialize and freeze the organization.
 *
 * @summary Executes the createSnapshot API operation against the backend server.
 * @param {string} [label] - optional human-readable tag for the snapshot.
 * @returns {Promise<OrgSnapshot>} an asynchronous Promise resolving to the created snapshot details.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers heavy disk I/O to capture full state.
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

/**
 * restoreSnapshot rolls the entire organization context backwards to a known-good immutable state.
 *
 * @summary Executes the restoreSnapshot API operation against the backend server.
 * @param {string} snapshotId - explicit backup ID to restore from.
 * @returns {Promise<DashboardSnapshot>} an asynchronous Promise resolving to the newly rolled-back dashboard.
 * @throws throws an error on network or parsing failure.
 * @sideeffects instantly truncates and replaces current live memory and state.
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * fetchMarketplace loads the global directory of third-party agents, domains, and tools.
 *
 * @summary Executes the fetchMarketplace API operation against the backend server.
 * @param none
 * @returns {Promise<MarketplaceItem[]>} an asynchronous Promise resolving to marketplace listings.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * fetchAnalytics calculates token velocity, fidelity ratios, and high-level efficiency stats.
 *
 * @summary Executes the fetchAnalytics API operation against the backend server.
 * @param none
 * @returns {Promise<AnalyticsSummary>} an asynchronous Promise resolving to aggregated platform metrics.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchAnalytics(): Promise<AnalyticsSummary> {
  return getJSON<AnalyticsSummary>("/api/analytics");
}

// ── External Integrations ─────────────────────────────────────────────────────

import type {
  ChatMessage,
  Integration,
  Issue,
  PullRequest,
} from "./types";

/**
 * fetchIntegrations lists configured connections to external tools like Jira or GitHub.
 *
 * @summary Executes the fetchIntegrations API operation against the backend server.
 * @param {string} [category] - optional filter for integration type.
 * @returns {Promise<Integration[]>} an asynchronous Promise resolving to the connection configurations.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return getJSON<Integration[]>(`/api/integrations${q}`);
}

/**
 * connectIntegration negotiates a live link to an external tool via the registry.
 *
 * @summary Executes the connectIntegration API operation against the backend server.
 * @param {string} integrationId - target identifier.
 * @param {string} [baseUrl] - optional endpoint configuration.
 * @returns {Promise<Integration>} an asynchronous Promise resolving to updated connection status.
 * @throws throws an error on network or parsing failure.
 * @sideeffects attempts an outbound handshake connection.
 */
export function connectIntegration(integrationId: string, baseUrl?: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/connect", { integrationId, baseUrl });
}

/**
 * disconnectIntegration severs an active API link and destroys cached external tool data.
 *
 * @summary Executes the disconnectIntegration API operation against the backend server.
 * @param {string} integrationId - target identifier to sever.
 * @returns {Promise<Integration>} an asynchronous Promise resolving to offline connection status.
 * @throws throws an error on network or parsing failure.
 * @sideeffects destroys backend routing context for the tool.
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/disconnect", { integrationId });
}

/**
 * fetchChatMessages retrieves historical synchronized discussion logs from an external tool.
 *
 * @summary Executes the fetchChatMessages API operation against the backend server.
 * @param {string} [integrationId] - optional filter for a specific chat provider.
 * @returns {Promise<ChatMessage[]>} an asynchronous Promise resolving to ChatMessage objects.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}

/**
 * sendChatMessage pushes an internal payload outbound to an external messaging provider.
 *
 * @summary Executes the sendChatMessage API operation against the backend server.
 * @param {Object} body - outbound payload containing channel and message text.
 * @returns {Promise<ChatMessage>} an asynchronous Promise resolving to the success state of the push.
 * @throws throws an error on network or parsing failure.
 * @sideeffects fires HTTP POSTs to third party chat integrations.
 */
export function sendChatMessage(body: {
  integrationId: string;
  channel: string;
  fromAgent: string;
  content: string;
  threadId?: string;
}): Promise<ChatMessage> {
  return postJSON<ChatMessage>("/api/integrations/chat/send", body);
}

/**
 * fetchPullRequests synchronizes open code proposals from external Git providers.
 *
 * @summary Executes the fetchPullRequests API operation against the backend server.
 * @param {string} [integrationId] - optional specific Git provider filter.
 * @returns {Promise<PullRequest[]>} an asynchronous Promise resolving to PR states.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}

/**
 * createPullRequest pushes local generated code externally to initiate a version control PR.
 *
 * @summary Executes the createPullRequest API operation against the backend server.
 * @param {Object} body - payload defining the PR branches and summary.
 * @returns {Promise<PullRequest>} an asynchronous Promise resolving to the created PR.
 * @throws throws an error on network or parsing failure.
 * @sideeffects initiates code review requests on remote VCS hosts.
 */
export function createPullRequest(body: {
  integrationId: string;
  repository: string;
  title: string;
  body?: string;
  sourceBranch: string;
  targetBranch: string;
  createdBy?: string;
}): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/create", body);
}

/**
 * mergePullRequest finalizes a code proposal on the remote Git provider.
 *
 * @summary Executes the mergePullRequest API operation against the backend server.
 * @param {string} prId - unique identifier of the target PR.
 * @returns {Promise<PullRequest>} an asynchronous Promise resolving to the closed PR.
 * @throws throws an error on network or parsing failure.
 * @sideeffects causes an immediate code merge on a remote repository.
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}

/**
 * closePullRequest aborts and rejects an open code proposal on the remote Git provider.
 *
 * @summary Executes the closePullRequest API operation against the backend server.
 * @param {string} prId - unique identifier of the target PR.
 * @returns {Promise<PullRequest>} an asynchronous Promise resolving to the closed PR.
 * @throws throws an error on network or parsing failure.
 * @sideeffects alters the status of a remote repository.
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}

/**
 * fetchIssues synchronizes active task progress from an external ticketing platform.
 *
 * @summary Executes the fetchIssues API operation against the backend server.
 * @param {string} [integrationId] - optional tracking provider filter.
 * @returns {Promise<Issue[]>} an asynchronous Promise resolving to the active tickets.
 * @throws throws an error on network or parsing failure.
 * @sideeffects triggers a network request.
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<Issue[]>(`/api/integrations/issues${q}`);
}

/**
 * createIssue registers a new work item tracker against the remote third-party system.
 *
 * @summary Executes the createIssue API operation against the backend server.
 * @param {Object} body - work item details payload.
 * @returns {Promise<Issue>} an asynchronous Promise resolving to the generated ticket.
 * @throws throws an error on network or parsing failure.
 * @sideeffects creates remote tracker state.
 */
export function createIssue(body: {
  integrationId: string;
  project: string;
  title: string;
  description?: string;
  createdBy?: string;
  priority?: string;
  labels?: string[];
}): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/create", body);
}

/**
 * updateIssueStatus progresses an existing ticketing item across its lifecycle phases.
 *
 * @summary Executes the updateIssueStatus API operation against the backend server.
 * @param {string} issueId - identifier for the ticket.
 * @param {string} status - the target phase string to migrate into.
 * @returns {Promise<Issue>} an asynchronous Promise resolving to the updated state.
 * @throws throws an error on network or parsing failure.
 * @sideeffects mutates state in external task management platforms.
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}

/**
 * assignIssue transfers ownership and accountability of an external ticket to a different entity.
 *
 * @summary Executes the assignIssue API operation against the backend server.
 * @param {string} issueId - identifier for the ticket.
 * @param {string} assignee - identifier of the target owner.
 * @returns {Promise<Issue>} an asynchronous Promise resolving to the reassigned ticket.
 * @throws throws an error on network or parsing failure.
 * @sideeffects alters responsibility routing on a remote system.
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}
