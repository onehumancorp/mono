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
 * Params: None
 * Returns: A Promise resolving to the Organization object.
 * Errors: None
 * Side Effects: None
 */
export function fetchOrganization(): Promise<Organization> {
  return authedGetJSON<Organization>("/api/org");
}

/**
 * Summary: Fetches all active virtual meeting rooms and their transcripts.
 * Params: None
 * Returns: A Promise resolving to an array of MeetingRoom objects.
 * Errors: None
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
 * Params: None
 * Returns: A Promise resolving to the CostSummary object containing breakdown by agent and model.
 * Errors: An Error if the request fails or returns a non-2xx status code.
 * Side Effects: Executes an HTTP GET request to /api/costs.
 */
export async function fetchCosts(): Promise<CostSummary> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

/**
 * Summary: Fetches a complete, normalized snapshot of the organization's current orchestration state.
 * Params: None
 * Returns: A Promise resolving to a DashboardSnapshot containing organization details, active meetings, and real-time costs.
 * Errors: An Error if the request fails or returns a non-2xx status code.
 * Side Effects: Executes an HTTP GET request to /api/dashboard.
 */
export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/dashboard");
  return normalizeDashboard(response);
}

/**
 * Summary: Dispatches a message or task from one agent to another within a specific meeting.
 * Params: form - An object containing the sender, recipient, meeting ID, type, and content of the message.
 * Returns: A Promise resolving when the message is successfully published.
 * Errors: None
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
 * Params: name - The display name for the new agent., role - The specific role profile the agent will assume.
 * Returns: A Promise resolving to the updated DashboardSnapshot.
 * Errors: None
 * Side Effects: None
 */
export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

/**
 * Summary: Terminates an agent's process and removes it from the orchestration hub.
 * Params: agentId - The unique identifier of the agent to fire.
 * Returns: A Promise resolving to the updated DashboardSnapshot.
 * Errors: None
 * Side Effects: None
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

/**
 * Summary: Retrieves available organizational domain templates (e.g., Software Company).
 * Params: None
 * Returns: A Promise resolving to an array of DomainInfo objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

/**
 * Summary: Retrieves the catalog of active tools registered in the MCP gateway.
 * Params: None
 * Returns: A Promise resolving to an array of MCPTool objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

/**
 * Summary: Overrides current state with a predefined scenario for demonstration purposes.
 * Params: scenario - The identifier string of the scenario to seed.
 * Returns: A Promise resolving to the resulting DashboardSnapshot.
 * Errors: None
 * Side Effects: None
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

/**
 * Summary: Retrieves all pending and resolved confidence gating approval requests.
 * Params: None
 * Returns: A Promise resolving to an array of ApprovalRequest objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

/**
 * Summary: Submits a new request for human manager sign-off on a high-risk action.
 * Params: body - An object defining the action, reason, estimated cost, and risk level.
 * Returns: A Promise resolving to the newly created ApprovalRequest.
 * Errors: None
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
 * Params: approvalId - The unique ID of the pending approval request., decision - The decision status, either 'approve' or 'reject'., decidedBy - Optional identifier for the human manager who made the decision.
 * Returns: A Promise resolving to an updated array of ApprovalRequest objects.
 * Errors: None
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
 * Params: None
 * Returns: A Promise resolving to an array of HandoffPackage objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

/**
 * Summary: Escalates a complex task from an autonomous agent to a human manager.
 * Params: body - The handoff package containing context, intent, and failed attempts.
 * Returns: A Promise resolving to the newly created HandoffPackage.
 * Errors: None
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
 * Params: None
 * Returns: A Promise resolving to an array of AgentIdentity objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

/**
 * Summary: Retrieves all imported skill packs available for agent instantiation.
 * Params: None
 * Returns: A Promise resolving to an array of SkillPack objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

/**
 * Summary: Imports a new specialized skill pack into the organization's domain.
 * Params: body - An object defining the skill pack metadata and capabilities.
 * Returns: A Promise resolving to the successfully imported SkillPack.
 * Errors: None
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
 * Params: None
 * Returns: A Promise resolving to an array of OrgSnapshot objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

/**
 * Summary: Captures a point-in-time snapshot of the entire organization's memory and state.
 * Params: label - Optional user-defined label for the snapshot.
 * Returns: A Promise resolving to the created OrgSnapshot.
 * Errors: None
 * Side Effects: None
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

/**
 * Summary: Restores the organization to a specific point-in-time snapshot.
 * Params: snapshotId - The unique ID of the snapshot to restore.
 * Returns: A Promise resolving to the restored DashboardSnapshot.
 * Errors: None
 * Side Effects: None
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

/**
 * Summary: Retrieves the catalog of community-published agents and tools.
 * Params: None
 * Returns: A Promise resolving to an array of MarketplaceItem objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

/**
 * Summary: Fetches real-time operational and health metrics for the organization.
 * Params: None
 * Returns: A Promise resolving to the AnalyticsSummary.
 * Errors: None
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
 * Params: category - Optional integration category to filter by (e.g., 'chat', 'git').
 * Returns: A Promise resolving to an array of Integration objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return getJSON<Integration[]>(`/api/integrations${q}`);
}

/**
 * Summary: Connects and authenticates a specific external integration.
 * Params: integrationId - The identifier for the service to connect to., config - Optional configuration including base URL and credentials.
 * Returns: A Promise resolving to the connected Integration.
 * Errors: None
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
 * Params: integrationId - The identifier for the service to disconnect from.
 * Returns: A Promise resolving to the disconnected Integration.
 * Errors: None
 * Side Effects: None
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return postJSON<Integration>("/api/integrations/disconnect", { integrationId });
}

/**
 * Summary: Sends a test message to validate credentials before saving them.
 * Params: integrationId - The identifier of the chat service being tested., config - The credential parameters required to send the test message.
 * Returns: A Promise resolving to an object indicating success.
 * Errors: An Error if the test message fails to send.
 * Side Effects: Executes an HTTP POST request to /api/integrations/chat/test.
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
 * Params: integrationId - Optional integration ID to filter the messages by.
 * Returns: A Promise resolving to an array of ChatMessage objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}

/**
 * Summary: Dispatches a message to an external chat platform.
 * Params: body - An object defining the target integration, channel, sender, and message context.
 * Returns: A Promise resolving to the recorded ChatMessage.
 * Errors: None
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
 * Params: integrationId - Optional integration ID to filter the pull requests by.
 * Returns: A Promise resolving to an array of PullRequest objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}

/**
 * Summary: Opens a pull request/merge request on a connected git platform.
 * Params: body - The parameters required to open a pull request (repo, branches, etc).
 * Returns: A Promise resolving to the successfully created PullRequest.
 * Errors: None
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
 * Params: prId - The unique ID of the pull request to merge.
 * Returns: A Promise resolving to the merged PullRequest.
 * Errors: None
 * Side Effects: None
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}

/**
 * Summary: Closes an open pull request on a connected git platform without merging.
 * Params: prId - The unique ID of the pull request to close.
 * Returns: A Promise resolving to the closed PullRequest.
 * Errors: None
 * Side Effects: None
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return postJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}

/**
 * Summary: Fetches tickets from connected issue trackers.
 * Params: integrationId - Optional integration ID to filter the tickets by.
 * Returns: A Promise resolving to an array of Issue objects.
 * Errors: None
 * Side Effects: None
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return getJSON<Issue[]>(`/api/integrations/issues${q}`);
}

/**
 * Summary: Creates a ticket in a connected issue tracker.
 * Params: body - The details needed to create the issue (project, title, description, priority).
 * Returns: A Promise resolving to the generated Issue object.
 * Errors: None
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
 * Params: issueId - The unique ID of the ticket., status - The new status to transition to.
 * Returns: A Promise resolving to the updated Issue.
 * Errors: None
 * Side Effects: None
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}

/**
 * Summary: Assigns ownership of a ticket to a specific agent or human manager.
 * Params: issueId - The unique ID of the ticket., assignee - The identifier of the agent or manager taking ownership.
 * Returns: A Promise resolving to the assigned Issue.
 * Errors: None
 * Side Effects: None
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return postJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}

/**
 * Summary: Invokes an MCP tool with the given action and parameters. Communication tools route to the underlying connected integration. Git/issue tools create PRs or tickets in the connected platform.
 * Params: toolId - The unique identifier of the target MCP tool., action - The specific action or operation to execute., params - A key-value map of parameters required by the tool.
 * Returns: A Promise resolving to the opaque JSON result returned by the tool.
 * Errors: An Error if the tool invocation fails.
 * Side Effects: Executes an HTTP POST request to /api/mcp/tools/invoke.
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
 * Params: None
 * Returns: A Promise resolving to the Settings object.
 * Errors: An Error if the request fails.
 * Side Effects: Executes an HTTP GET request to /api/settings.
 */
export function fetchSettings(): Promise<Settings> {
  return getJSON<Settings>("/api/settings");
}

/**
 * Summary: Saves and updates the global settings and preferences.
 * Params: settings - The updated settings object to persist.
 * Returns: A Promise resolving to the successfully saved Settings.
 * Errors: An Error if the save operation fails.
 * Side Effects: Executes an HTTP POST request to /api/settings.
 */
export function saveSettings(settings: Settings): Promise<Settings> {
  return postJSON<Settings>("/api/settings", settings);
}

// ── Auth ──────────────────────────────────────────────────────────────────────

const TOKEN_KEY = "ohc_token";

/**
 * Summary: Retrieves the currently stored authentication JWT token from local storage.
 * Params: None
 * Returns: The token string if it exists, otherwise null.
 * Errors: None
 * Side Effects: Reads from window.localStorage.
 */
export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Summary: Persists an authentication JWT token in local storage.
 * Params: token - The raw JWT string to store.
 * Returns: None
 * Errors: None
 * Side Effects: Writes to window.localStorage.
 */
export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Summary: Removes the stored authentication JWT token from local storage.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: Deletes the key from window.localStorage.
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
 * Params: username - The user's account identifier., password - The user's secret password.
 * Returns: A Promise resolving to the LoginResponse containing the issued token.
 * Errors: An Error if authentication fails or credentials are invalid.
 * Side Effects: Executes an HTTP POST request to /api/auth/login and stores the resulting token in local storage.
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
 * Params: None
 * Returns: A Promise resolving when the logout completes.
 * Errors: None
 * Side Effects: Executes an authenticated HTTP POST request to /api/auth/logout and clears local storage.
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
 * Params: None
 * Returns: A Promise resolving to the UserPublic profile.
 * Errors: An Error if the user is unauthenticated or the request fails.
 * Side Effects: Executes an authenticated HTTP GET request to /api/auth/me.
 */
export function fetchMe(): Promise<UserPublic> {
  return authedGetJSON<UserPublic>("/api/auth/me");
}

/**
 * Summary: Retrieves a list of all registered users in the system (requires Admin role).
 * Params: None
 * Returns: A Promise resolving to an array of UserPublic profiles.
 * Errors: An Error if the request fails or the caller lacks permissions.
 * Side Effects: Executes an authenticated HTTP GET request to /api/users.
 */
export function fetchUsers(): Promise<UserPublic[]> {
  return authedGetJSON<UserPublic[]>("/api/users");
}

/**
 * Summary: Creates a new user account within the system (requires Admin role).
 * Params: body - The parameters required to create the user, including username, email, password, and optional roles.
 * Returns: A Promise resolving to the newly created UserPublic profile.
 * Errors: An Error if the creation fails, validation fails, or permissions are insufficient.
 * Side Effects: Executes an authenticated HTTP POST request to /api/users.
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
 * Params: id - The unique identifier of the user to delete.
 * Returns: A Promise resolving when the deletion is successful.
 * Errors: An Error if the deletion fails or the user cannot be found.
 * Side Effects: Executes an authenticated HTTP DELETE request to /api/users/:id.
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
 * Params: None
 * Returns: A Promise resolving to an array of Role objects.
 * Errors: An Error if the request fails.
 * Side Effects: Executes an authenticated HTTP GET request to /api/roles.
 */
export function fetchRoles(): Promise<Role[]> {
  return authedGetJSON<Role[]>("/api/roles");
}

/**
 * Summary: Creates a new custom role with an optional set of permissions.
 * Params: body - The role configuration containing its name and permissions.
 * Returns: A Promise resolving to the newly created Role.
 * Errors: An Error if role creation fails.
 * Side Effects: Executes an authenticated HTTP POST request to /api/roles.
 */
export function createRole(body: { name: string; permissions?: string[] }): Promise<Role> {
  return authedPostJSON<Role>("/api/roles", body);
}

