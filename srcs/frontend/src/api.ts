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
 * Summary: fetchOrganization functionality.
 * Parameters: None
 * Returns: Promise<Organization>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchOrganization(): Promise<Organization> {
  return authedGetJSON<Organization>("/api/org");
}

/**
 * Summary: fetchMeetings functionality.
 * Parameters: None
 * Returns: Promise<MeetingRoom[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMeetings(): Promise<MeetingRoom[]> {
  return authedGetJSON<MeetingRoom[]>("/api/meetings").then(normalizeMeetings);
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

/**
 * Summary: fetchCosts functionality.
 * Parameters: None
 * Returns: Promise<CostSummary>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function fetchCosts(): Promise<CostSummary> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

/**
 * Summary: fetchDashboard functionality.
 * Parameters: None
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/dashboard");
  return normalizeDashboard(response);
}

/**
 * Summary: sendMessage functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
 */
export async function sendMessage(form: {
  fromAgent: string;
  toAgent: string;
  meetingId: string;
  messageType: string;
  content: string;
}): Promise<DashboardSnapshot> {
  const params = new URLSearchParams(form);
  const token = getStoredToken();
  const response = await fetch("/api/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      Accept: "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: params.toString(),
    redirect: "follow",
  });

  if (!response.ok) {
    if (response.status === 401) {
      clearStoredToken();
      throw new Error("Unauthorized");
    }
    const text = await response.text().catch(() => "");
    let errorMessage = text;
    try {
      const parsed = JSON.parse(text);
      if (parsed.error) {
        errorMessage = parsed.error;
      }
    } catch {
      // Not JSON, use raw text
    }
    throw new Error(errorMessage || `Request failed for /api/messages: ${response.status}`);
  }
  const raw = await response.json() as Record<string, unknown>;
  return normalizeDashboard(raw);
}

/**
 * Summary: hireAgent functionality.
 * Parameters: name, role
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

/**
 * Summary: fireAgent functionality.
 * Parameters: agentId
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

/**
 * Summary: fetchDomains functionality.
 * Parameters: None
 * Returns: Promise<DomainInfo[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

/**
 * Summary: fetchMCPTools functionality.
 * Parameters: None
 * Returns: Promise<MCPTool[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

/**
 * Summary: seedScenario functionality.
 * Parameters: scenario
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * Summary: fetchApprovals functionality.
 * Parameters: None
 * Returns: Promise<ApprovalRequest[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

/**
 * Summary: requestApproval functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: decideApproval functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: fetchHandoffs functionality.
 * Parameters: None
 * Returns: Promise<HandoffPackage[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

/**
 * Summary: createHandoff functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: fetchIdentities functionality.
 * Parameters: None
 * Returns: Promise<AgentIdentity[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

/**
 * Summary: fetchSkillPacks functionality.
 * Parameters: None
 * Returns: Promise<SkillPack[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

/**
 * Summary: importSkillPack functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: fetchSnapshots functionality.
 * Parameters: None
 * Returns: Promise<OrgSnapshot[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

/**
 * Summary: createSnapshot functionality.
 * Parameters: label?
 * Returns: Promise<OrgSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

/**
 * Summary: restoreSnapshot functionality.
 * Parameters: snapshotId
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * Summary: fetchMarketplace functionality.
 * Parameters: None
 * Returns: Promise<MarketplaceItem[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * Summary: fetchAnalytics functionality.
 * Parameters: None
 * Returns: Promise<AnalyticsSummary>
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: fetchIntegrations functionality.
 * Parameters: category?
 * Returns: Promise<Integration[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return getJSON<Integration[]>(`/api/integrations${q}`);
}

/**
 * Summary: connectIntegration functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: disconnectIntegration functionality.
 * Parameters: integrationId
 * Returns: Promise<Integration>
 * Errors: May throw an error
 * Side Effects: None
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/disconnect", { integrationId });
}

/**
 * Summary: testChatIntegration functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
 */
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
 * Summary: fetchChatMessages functionality.
 * Parameters: integrationId?
 * Returns: Promise<ChatMessage[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}

/**
 * Summary: sendChatMessage functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: fetchPullRequests functionality.
 * Parameters: integrationId?
 * Returns: Promise<PullRequest[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}

/**
 * Summary: createPullRequest functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: mergePullRequest functionality.
 * Parameters: prId
 * Returns: Promise<PullRequest>
 * Errors: May throw an error
 * Side Effects: None
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}

/**
 * Summary: closePullRequest functionality.
 * Parameters: prId
 * Returns: Promise<PullRequest>
 * Errors: May throw an error
 * Side Effects: None
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}

/**
 * Summary: fetchIssues functionality.
 * Parameters: integrationId?
 * Returns: Promise<Issue[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<Issue[]>(`/api/integrations/issues${q}`);
}

/**
 * Summary: createIssue functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
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
 * Summary: updateIssueStatus functionality.
 * Parameters: issueId, status
 * Returns: Promise<Issue>
 * Errors: May throw an error
 * Side Effects: None
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}

/**
 * Summary: assignIssue functionality.
 * Parameters: issueId, assignee
 * Returns: Promise<Issue>
 * Errors: May throw an error
 * Side Effects: None
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}

/**
 * Summary: invokeMCPTool functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
 */
export function invokeMCPTool(
  toolId: string,
  action: string,
  params: Record<string, string>,
): Promise<Record<string, unknown>> {
  return postJSON<Record<string, unknown>>("/api/mcp/tools/invoke", { toolId, action, params });
}

/**
 * Summary: fetchSettings functionality.
 * Parameters: None
 * Returns: Promise<Settings>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSettings(): Promise<Settings> {
  return getJSON<Settings>("/api/settings");
}

/**
 * Summary: saveSettings functionality.
 * Parameters: settings
 * Returns: Promise<Settings>
 * Errors: May throw an error
 * Side Effects: None
 */
export function saveSettings(settings: Settings): Promise<Settings> {
  return postJSON<Settings>("/api/settings", settings);
}

// ── Auth ──────────────────────────────────────────────────────────────────────

const TOKEN_KEY = "ohc_token";

/**
 * Summary: getStoredToken functionality.
 * Parameters: None
 * Returns: string | null
 * Errors: May throw an error
 * Side Effects: None
 */
export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Summary: setStoredToken functionality.
 * Parameters: token
 * Returns: void
 * Errors: May throw an error
 * Side Effects: None
 */
export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Summary: clearStoredToken functionality.
 * Parameters: None
 * Returns: void
 * Errors: May throw an error
 * Side Effects: None
 */
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
    const text = await response.text().catch(() => "");
    let errorMessage = text;
    try {
      const parsed = JSON.parse(text);
      if (parsed.error) {
        errorMessage = parsed.error;
      }
    } catch {
      // Not JSON, use raw text
    }
    throw new Error(errorMessage || `Request failed for ${path}: ${response.status}`);
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

/**
 * Summary: login functionality.
 * Parameters: username, password
 * Returns: Promise<LoginResponse>
 * Errors: May throw an error
 * Side Effects: None
 */
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

/**
 * Summary: logout functionality.
 * Parameters: None
 * Returns: Promise<void>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function logout(): Promise<void> {
  try {
    await authedPostJSON<void>("/api/auth/logout", {});
  } finally {
    clearStoredToken();
  }
}

/**
 * Summary: fetchMe functionality.
 * Parameters: None
 * Returns: Promise<UserPublic>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMe(): Promise<UserPublic> {
  return authedGetJSON<UserPublic>("/api/auth/me");
}

/**
 * Summary: fetchUsers functionality.
 * Parameters: None
 * Returns: Promise<UserPublic[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchUsers(): Promise<UserPublic[]> {
  return authedGetJSON<UserPublic[]>("/api/users");
}

/**
 * Summary: createUser functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
 */
export function createUser(body: {
  username: string;
  email: string;
  password: string;
  roles?: string[];
}): Promise<UserPublic> {
  return authedPostJSON<UserPublic>("/api/users", body);
}

/**
 * Summary: deleteUser functionality.
 * Parameters: id
 * Returns: Promise<void>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function deleteUser(id: string): Promise<void> {
  const token = getStoredToken();
  await fetch(`/api/users/${id}`, {
    method: "DELETE",
    headers: { Authorization: token ? `Bearer ${token}` : "" },
  });
}

/**
 * Summary: fetchRoles functionality.
 * Parameters: None
 * Returns: Promise<Role[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchRoles(): Promise<Role[]> {
  return authedGetJSON<Role[]>("/api/roles");
}

/**
 * Summary: createRole functionality.
 * Parameters: body
 * Returns: Promise<Role>
 * Errors: May throw an error
 * Side Effects: None
 */
export function createRole(body: { name: string; permissions?: string[] }): Promise<Role> {
  return authedPostJSON<Role>("/api/roles", body);
}


/**
 * Summary: scaleAgents functionality.
 * Parameters: None
 * Returns: None
 * Errors: May throw an error
 * Side Effects: None
 */
export function scaleAgents(
  role: string,
  count: number,
): Promise<{ status: string; role: string; count: number }> {
  return authedPostJSON<{ status: string; role: string; count: number }>("/api/v1/scale", { role, count });
}
