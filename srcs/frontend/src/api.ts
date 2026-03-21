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
 * Summary: Fetches the current organization's hierarchical and structural state.
 * Intent: Fetches the current organization's hierarchical and structural state.
 * Params: None
 * Returns: Promise<Organization>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchOrganization(): Promise<Organization> {
  return authedGetJSON<Organization>("/api/org");
}

/**
 * Summary: Fetches all active virtual meeting rooms and their transcripts.
 * Intent: Fetches all active virtual meeting rooms and their transcripts.
 * Params: None
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
 * Summary: Fetches the accumulated API token costs and usage metrics for the organization.
 * Intent: Fetches the accumulated API token costs and usage metrics for the organization.
 * Params: None
 * Returns: Promise<CostSummary>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function fetchCosts(): Promise<CostSummary> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

/**
 * Summary: Fetches a complete, normalized snapshot of the organization's current orchestration state.
 * Intent: Fetches a complete, normalized snapshot of the organization's current orchestration state.
 * Params: None
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/dashboard");
  return normalizeDashboard(response);
}

/**
 * Summary: Dispatches a message or task from one agent to another within a specific meeting.
 * Intent: Dispatches a message or task from one agent to another within a specific meeting.
 * Params: None
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
 * Summary: Instantiates a new agent and assigns it to the organizational workforce.
 * Intent: Instantiates a new agent and assigns it to the organizational workforce.
 * Params: name, role
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

/**
 * Summary: Terminates an agent's process and removes it from the orchestration hub.
 * Intent: Terminates an agent's process and removes it from the orchestration hub.
 * Params: agentId
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

/**
 * Summary: Retrieves available organizational domain templates (e.g., Software Company).
 * Intent: Retrieves available organizational domain templates (e.g., Software Company).
 * Params: None
 * Returns: Promise<DomainInfo[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

/**
 * Summary: Retrieves the catalog of active tools registered in the MCP gateway.
 * Intent: Retrieves the catalog of active tools registered in the MCP gateway.
 * Params: None
 * Returns: Promise<MCPTool[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

/**
 * Summary: Overrides current state with a predefined scenario for demonstration purposes.
 * Intent: Overrides current state with a predefined scenario for demonstration purposes.
 * Params: scenario
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * Summary: Retrieves all pending and resolved confidence gating approval requests.
 * Intent: Retrieves all pending and resolved confidence gating approval requests.
 * Params: None
 * Returns: Promise<ApprovalRequest[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

/**
 * Summary: Submits a new request for human manager sign-off on a high-risk action.
 * Intent: Submits a new request for human manager sign-off on a high-risk action.
 * Params: None
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
 * Summary: Submits the human manager's decision (approve/reject) for an approval request.
 * Intent: Submits the human manager's decision (approve/reject) for an approval request.
 * Params: None
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
 * Summary: Retrieves all warm handoff escalations across the organization.
 * Intent: Retrieves all warm handoff escalations across the organization.
 * Params: None
 * Returns: Promise<HandoffPackage[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

/**
 * Summary: Escalates a complex task from an autonomous agent to a human manager.
 * Intent: Escalates a complex task from an autonomous agent to a human manager.
 * Params: None
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
 * Summary: Retrieves the SPIFFE SVID certificates issued to the current workforce.
 * Intent: Retrieves the SPIFFE SVID certificates issued to the current workforce.
 * Params: None
 * Returns: Promise<AgentIdentity[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

/**
 * Summary: Retrieves all imported skill packs available for agent instantiation.
 * Intent: Retrieves all imported skill packs available for agent instantiation.
 * Params: None
 * Returns: Promise<SkillPack[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

/**
 * Summary: Imports a new specialized skill pack into the organization's domain.
 * Intent: Imports a new specialized skill pack into the organization's domain.
 * Params: None
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
 * Summary: Retrieves all point-in-time recovery snapshots for the organization.
 * Intent: Retrieves all point-in-time recovery snapshots for the organization.
 * Params: None
 * Returns: Promise<OrgSnapshot[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

/**
 * Summary: Captures a point-in-time snapshot of the entire organization's memory and state.
 * Intent: Captures a point-in-time snapshot of the entire organization's memory and state.
 * Params: label?
 * Returns: Promise<OrgSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

/**
 * Summary: Restores the organization to a specific point-in-time snapshot.
 * Intent: Restores the organization to a specific point-in-time snapshot.
 * Params: snapshotId
 * Returns: Promise<DashboardSnapshot>
 * Errors: May throw an error
 * Side Effects: None
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * Summary: Retrieves the catalog of community-published agents and tools.
 * Intent: Retrieves the catalog of community-published agents and tools.
 * Params: None
 * Returns: Promise<MarketplaceItem[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * Summary: Fetches real-time operational and health metrics for the organization.
 * Intent: Fetches real-time operational and health metrics for the organization.
 * Params: None
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
 * Summary: Retrieves external service connections, optionally filtered by category.
 * Intent: Retrieves external service connections, optionally filtered by category.
 * Params: category?
 * Returns: Promise<Integration[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return getJSON<Integration[]>(`/api/integrations${q}`);
}

/**
 * Summary: Connects and authenticates a specific external integration.
 * Intent: Connects and authenticates a specific external integration.
 * Params: None
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
 * Summary: Disconnects an active external integration.
 * Intent: Disconnects an active external integration.
 * Params: integrationId
 * Returns: Promise<Integration>
 * Errors: May throw an error
 * Side Effects: None
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/disconnect", { integrationId });
}

/**
 * Summary: Sends a test message to validate credentials before saving them.
 * Intent: Sends a test message to validate credentials before saving them.
 * Params: None
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
 * Summary: Fetches recorded chat messages from the integration registry.
 * Intent: Fetches recorded chat messages from the integration registry.
 * Params: integrationId?
 * Returns: Promise<ChatMessage[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}

/**
 * Summary: Dispatches a message to an external chat platform.
 * Intent: Dispatches a message to an external chat platform.
 * Params: None
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
 * Summary: Fetches pull requests opened via the git integrations.
 * Intent: Fetches pull requests opened via the git integrations.
 * Params: integrationId?
 * Returns: Promise<PullRequest[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}

/**
 * Summary: Opens a pull request/merge request on a connected git platform.
 * Intent: Opens a pull request/merge request on a connected git platform.
 * Params: None
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
 * Summary: Merges an open pull request on a connected git platform.
 * Intent: Merges an open pull request on a connected git platform.
 * Params: prId
 * Returns: Promise<PullRequest>
 * Errors: May throw an error
 * Side Effects: None
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}

/**
 * Summary: Closes an open pull request on a connected git platform without merging.
 * Intent: Closes an open pull request on a connected git platform without merging.
 * Params: prId
 * Returns: Promise<PullRequest>
 * Errors: May throw an error
 * Side Effects: None
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}

/**
 * Summary: Fetches tickets from connected issue trackers.
 * Intent: Fetches tickets from connected issue trackers.
 * Params: integrationId?
 * Returns: Promise<Issue[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<Issue[]>(`/api/integrations/issues${q}`);
}

/**
 * Summary: Creates a ticket in a connected issue tracker.
 * Intent: Creates a ticket in a connected issue tracker.
 * Params: None
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
 * Summary: Updates the status phase of an existing ticket.
 * Intent: Updates the status phase of an existing ticket.
 * Params: issueId, status
 * Returns: Promise<Issue>
 * Errors: May throw an error
 * Side Effects: None
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}

/**
 * Summary: Assigns ownership of a ticket to a specific agent or human manager.
 * Intent: Assigns ownership of a ticket to a specific agent or human manager.
 * Params: issueId, assignee
 * Returns: Promise<Issue>
 * Errors: May throw an error
 * Side Effects: None
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}

/**
 * Summary: Invokes an MCP tool with the given action and parameters. Communication tools route to the underlying connected integration. Git/issue tools create PRs or tickets in the connected platform.
 * Intent: Invokes an MCP tool with the given action and parameters. Communication tools route to the underlying connected integration. Git/issue tools create PRs or tickets in the connected platform.
 * Params: None
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
 * Summary: Fetches the user's or organization's global settings and preferences.
 * Intent: Fetches the user's or organization's global settings and preferences.
 * Params: None
 * Returns: Promise<Settings>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchSettings(): Promise<Settings> {
  return getJSON<Settings>("/api/settings");
}

/**
 * Summary: Saves and updates the global settings and preferences.
 * Intent: Saves and updates the global settings and preferences.
 * Params: settings
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
 * Summary: Retrieves the currently stored authentication JWT token from local storage.
 * Intent: Retrieves the currently stored authentication JWT token from local storage.
 * Params: None
 * Returns: string | null
 * Errors: May throw an error
 * Side Effects: None
 */
export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Summary: Persists an authentication JWT token in local storage.
 * Intent: Persists an authentication JWT token in local storage.
 * Params: token
 * Returns: void
 * Errors: May throw an error
 * Side Effects: None
 */
export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Summary: Removes the stored authentication JWT token from local storage.
 * Intent: Removes the stored authentication JWT token from local storage.
 * Params: None
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
 * Summary: Authenticates a user and retrieves a JWT token.
 * Intent: Authenticates a user and retrieves a JWT token.
 * Params: username, password
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
 * Summary: Invalidates the current session and clears the stored authentication token.
 * Intent: Invalidates the current session and clears the stored authentication token.
 * Params: None
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
 * Summary: Retrieves the public profile information of the currently authenticated user.
 * Intent: Retrieves the public profile information of the currently authenticated user.
 * Params: None
 * Returns: Promise<UserPublic>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchMe(): Promise<UserPublic> {
  return authedGetJSON<UserPublic>("/api/auth/me");
}

/**
 * Summary: Retrieves a list of all registered users in the system (requires Admin role).
 * Intent: Retrieves a list of all registered users in the system (requires Admin role).
 * Params: None
 * Returns: Promise<UserPublic[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchUsers(): Promise<UserPublic[]> {
  return authedGetJSON<UserPublic[]>("/api/users");
}

/**
 * Summary: Creates a new user account within the system (requires Admin role).
 * Intent: Creates a new user account within the system (requires Admin role).
 * Params: None
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
 * Summary: Deletes an existing user account from the system (requires Admin role).
 * Intent: Deletes an existing user account from the system (requires Admin role).
 * Params: id
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
 * Summary: Retrieves the list of available operational roles and their associated permissions.
 * Intent: Retrieves the list of available operational roles and their associated permissions.
 * Params: None
 * Returns: Promise<Role[]>
 * Errors: May throw an error
 * Side Effects: None
 */
export function fetchRoles(): Promise<Role[]> {
  return authedGetJSON<Role[]>("/api/roles");
}

/**
 * Summary: Creates a new custom role with an optional set of permissions.
 * Intent: Creates a new custom role with an optional set of permissions.
 * Params: body
 * Returns: Promise<Role>
 * Errors: May throw an error
 * Side Effects: None
 */
export function createRole(body: { name: string; permissions?: string[] }): Promise<Role> {
  return authedPostJSON<Role>("/api/roles", body);
}


/**
 * Summary: Scales the number of agents for a specific role dynamically.
 * Intent: Scales the number of agents for a specific role dynamically.
 * Params: role, count
 * Returns: Promise<{ status: string; role: string; count: number }>
 * Errors: May throw an error
 * Side Effects: None
 */
export function scaleAgents(
  role: string,
  count: number,
): Promise<{ status: string; role: string; count: number }> {
  return authedPostJSON<{ status: string; role: string; count: number }>("/api/v1/scale", { role, count });
}
