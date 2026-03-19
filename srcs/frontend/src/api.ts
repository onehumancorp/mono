import type {
  AnalyticsSummary,
  AgentIdentity,
  ApprovalRequest,
  CostSummary,
  DashboardSnapshot,
  DomainInfo,
  HandoffPackage,
  LoginResponse,
  MarketplaceItem,
  MCPTool,
  MeetingRoom,
  OrgSnapshot,
  Organization,
  Role,
  Settings,
  SkillPack,
  UserPublic,
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
    const text = await response.text().catch(() => "");
    throw new Error(text || `Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

/**
 * Fetches the current organization's hierarchical and structural state.
 *
 * @returns A Promise resolving to the Organization object.
 */
export function fetchOrganization(): Promise<Organization> {
  return getJSON<Organization>("/api/org");
}

/**
 * Fetches all active virtual meeting rooms and their transcripts.
 *
 * @returns A Promise resolving to an array of MeetingRoom objects.
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

/** Normalise a raw dashboard JSON response into a typed DashboardSnapshot. */
function normalizeDashboard(response: Record<string, unknown>): DashboardSnapshot {
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

export async function fetchCosts(): Promise<CostSummary> {
  const response = await getJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await getJSON<Record<string, unknown>>("/api/dashboard");
  return normalizeDashboard(response);
}

/**
 * Dispatches a message or task from one agent to another within a specific meeting.
 *
 * @param form - An object containing the sender, recipient, meeting ID, type, and content of the message.
 * @returns A Promise resolving when the message is successfully published.
 */
export async function sendMessage(form: {
  fromAgent: string;
  toAgent: string;
  meetingId: string;
  messageType: string;
  content: string;
}): Promise<DashboardSnapshot> {
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
    const text = await response.text().catch(() => "");
    throw new Error(text || `Failed to send message: ${response.status}`);
  }
  const raw = await response.json() as Record<string, unknown>;
  return normalizeDashboard(raw);
}

/**
 * Instantiates a new agent and assigns it to the organizational workforce.
 *
 * @param name - The display name for the new agent.
 * @param role - The specific role profile the agent will assume.
 * @returns A Promise resolving to the updated DashboardSnapshot.
 */
export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

/**
 * Terminates an agent's process and removes it from the orchestration hub.
 *
 * @param agentId - The unique identifier of the agent to fire.
 * @returns A Promise resolving to the updated DashboardSnapshot.
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

/**
 * Retrieves available organizational domain templates (e.g., Software Company).
 *
 * @returns A Promise resolving to an array of DomainInfo objects.
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

/**
 * Retrieves the catalog of active tools registered in the MCP gateway.
 *
 * @returns A Promise resolving to an array of MCPTool objects.
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

/**
 * Overrides current state with a predefined scenario for demonstration purposes.
 *
 * @param scenario - The identifier string of the scenario to seed.
 * @returns A Promise resolving to the resulting DashboardSnapshot.
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * Retrieves all pending and resolved confidence gating approval requests.
 *
 * @returns A Promise resolving to an array of ApprovalRequest objects.
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

/**
 * Submits a new request for human manager sign-off on a high-risk action.
 *
 * @param body - An object defining the action, reason, estimated cost, and risk level.
 * @returns A Promise resolving to the newly created ApprovalRequest.
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
 * Submits the human manager's decision (approve/reject) for an approval request.
 *
 * @param approvalId - The unique ID of the pending approval request.
 * @param decision - The decision status, either 'approve' or 'reject'.
 * @param decidedBy - Optional identifier for the human manager who made the decision.
 * @returns A Promise resolving to an updated array of ApprovalRequest objects.
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
 * Retrieves all warm handoff escalations across the organization.
 *
 * @returns A Promise resolving to an array of HandoffPackage objects.
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

/**
 * Escalates a complex task from an autonomous agent to a human manager.
 *
 * @param body - The handoff package containing context, intent, and failed attempts.
 * @returns A Promise resolving to the newly created HandoffPackage.
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
 * Retrieves the SPIFFE SVID certificates issued to the current workforce.
 *
 * @returns A Promise resolving to an array of AgentIdentity objects.
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

/**
 * Retrieves all imported skill packs available for agent instantiation.
 *
 * @returns A Promise resolving to an array of SkillPack objects.
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

/**
 * Imports a new specialized skill pack into the organization's domain.
 *
 * @param body - An object defining the skill pack metadata and capabilities.
 * @returns A Promise resolving to the successfully imported SkillPack.
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
 * Retrieves all point-in-time recovery snapshots for the organization.
 *
 * @returns A Promise resolving to an array of OrgSnapshot objects.
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

/**
 * Captures a point-in-time snapshot of the entire organization's memory and state.
 *
 * @param label - Optional user-defined label for the snapshot.
 * @returns A Promise resolving to the created OrgSnapshot.
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

/**
 * Restores the organization to a specific point-in-time snapshot.
 *
 * @param snapshotId - The unique ID of the snapshot to restore.
 * @returns A Promise resolving to the restored DashboardSnapshot.
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * Retrieves the catalog of community-published agents and tools.
 *
 * @returns A Promise resolving to an array of MarketplaceItem objects.
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * Fetches real-time operational and health metrics for the organization.
 *
 * @returns A Promise resolving to the AnalyticsSummary.
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
 * Retrieves external service connections, optionally filtered by category.
 *
 * @param category - Optional integration category to filter by (e.g., 'chat', 'git').
 * @returns A Promise resolving to an array of Integration objects.
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return getJSON<Integration[]>(`/api/integrations${q}`);
}

/**
 * Connects and authenticates a specific external integration.
 *
 * @param integrationId - The identifier for the service to connect to.
 * @param config - Optional configuration including base URL and credentials.
 * @returns A Promise resolving to the connected Integration.
 */
export function connectIntegration(
  integrationId: string,
  config?: {
    baseUrl?: string;
    botToken?: string;
    chatId?: string;
    webhookUrl?: string;
    apiToken?: string;
  },
): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/connect", {
    integrationId,
    baseUrl: config?.baseUrl,
    botToken: config?.botToken,
    chatId: config?.chatId,
    webhookUrl: config?.webhookUrl,
    apiToken: config?.apiToken,
  });
}

/**
 * Disconnects an active external integration.
 *
 * @param integrationId - The identifier for the service to disconnect from.
 * @returns A Promise resolving to the disconnected Integration.
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/disconnect", { integrationId });
}

/** Send a test message to validate credentials before saving them. */
export function testChatIntegration(
  integrationId: string,
  config: { botToken?: string; chatId?: string; webhookUrl?: string },
): Promise<{ success: boolean }> {
  return postJSON<{ success: boolean }>("/api/integrations/chat/test", {
    integrationId,
    ...config,
  });
}

/**
 * Fetches recorded chat messages from the integration registry.
 *
 * @param integrationId - Optional integration ID to filter the messages by.
 * @returns A Promise resolving to an array of ChatMessage objects.
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}

/**
 * Dispatches a message to an external chat platform.
 *
 * @param body - An object defining the target integration, channel, sender, and message context.
 * @returns A Promise resolving to the recorded ChatMessage.
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
 * Fetches pull requests opened via the git integrations.
 *
 * @param integrationId - Optional integration ID to filter the pull requests by.
 * @returns A Promise resolving to an array of PullRequest objects.
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}

/**
 * Opens a pull request/merge request on a connected git platform.
 *
 * @param body - The parameters required to open a pull request (repo, branches, etc).
 * @returns A Promise resolving to the successfully created PullRequest.
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
 * Merges an open pull request on a connected git platform.
 *
 * @param prId - The unique ID of the pull request to merge.
 * @returns A Promise resolving to the merged PullRequest.
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}

/**
 * Closes an open pull request on a connected git platform without merging.
 *
 * @param prId - The unique ID of the pull request to close.
 * @returns A Promise resolving to the closed PullRequest.
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}

/**
 * Fetches tickets from connected issue trackers.
 *
 * @param integrationId - Optional integration ID to filter the tickets by.
 * @returns A Promise resolving to an array of Issue objects.
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<Issue[]>(`/api/integrations/issues${q}`);
}

/**
 * Creates a ticket in a connected issue tracker.
 *
 * @param body - The details needed to create the issue (project, title, description, priority).
 * @returns A Promise resolving to the generated Issue object.
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
 * Updates the status phase of an existing ticket.
 *
 * @param issueId - The unique ID of the ticket.
 * @param status - The new status to transition to.
 * @returns A Promise resolving to the updated Issue.
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}

/**
 * Assigns ownership of a ticket to a specific agent or human manager.
 *
 * @param issueId - The unique ID of the ticket.
 * @param assignee - The identifier of the agent or manager taking ownership.
 * @returns A Promise resolving to the assigned Issue.
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}

/** Invoke an MCP tool with the given action and parameters.
 *  Communication tools route to the underlying connected integration.
 *  Git/issue tools create PRs or tickets in the connected platform.
 */
export function invokeMCPTool(
  toolId: string,
  action: string,
  params: Record<string, string>,
): Promise<Record<string, unknown>> {
  return postJSON<Record<string, unknown>>("/api/mcp/tools/invoke", { toolId, action, params });
}

export function fetchSettings(): Promise<Settings> {
  return getJSON<Settings>("/api/settings");
}

export function saveSettings(settings: Settings): Promise<Settings> {
  return postJSON<Settings>("/api/settings", settings);
}

// ── Auth ──────────────────────────────────────────────────────────────────────

const TOKEN_KEY = "ohc_token";

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

async function authedGetJSON<T>(path: string): Promise<T> {
  const token = getStoredToken();
  const response = await fetch(path, {
    headers: {
      Authorization: token ? `Bearer ${token}` : "",
      Accept: "application/json",
    },
  });
  if (!response.ok) {
    if (response.status === 401) {
      clearStoredToken();
      throw new Error("Unauthorized");
    }
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

async function authedPostJSON<T>(path: string, body: unknown): Promise<T> {
  const token = getStoredToken();
  const response = await fetch(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      Authorization: token ? `Bearer ${token}` : "",
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    const text = await response.text().catch(() => "");
    if (response.status === 401) {
      clearStoredToken();
      throw new Error("Unauthorized");
    }
    throw new Error(text || `Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const resp = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!resp.ok) {
    const text = await resp.text().catch(() => "");
    throw new Error(text || "Login failed");
  }
  const result = (await resp.json()) as LoginResponse;
  setStoredToken(result.token);
  return result;
}

export async function logout(): Promise<void> {
  try {
    await authedPostJSON<void>("/api/auth/logout", {});
  } finally {
    clearStoredToken();
  }
}

export function fetchMe(): Promise<UserPublic> {
  return authedGetJSON<UserPublic>("/api/auth/me");
}

export function fetchUsers(): Promise<UserPublic[]> {
  return authedGetJSON<UserPublic[]>("/api/users");
}

export function createUser(body: {
  username: string;
  email: string;
  password: string;
  roles?: string[];
}): Promise<UserPublic> {
  return authedPostJSON<UserPublic>("/api/users", body);
}

export async function deleteUser(id: string): Promise<void> {
  const token = getStoredToken();
  await fetch(`/api/users/${id}`, {
    method: "DELETE",
    headers: { Authorization: token ? `Bearer ${token}` : "" },
  });
}

export function fetchRoles(): Promise<Role[]> {
  return authedGetJSON<Role[]>("/api/roles");
}

export function createRole(body: { name: string; permissions?: string[] }): Promise<Role> {
  return authedPostJSON<Role>("/api/roles", body);
}

