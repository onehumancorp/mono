import type { ApiPipeline, ApiPipelinePromoteRequest } from "./proto_types";
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


/**
 * @summary Fetches the current organization's hierarchical and structural state.
 * @param None
 * @returns Promise<Organization>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchOrganization(): Promise<Organization> {
  return authedGetJSON<Organization>("/api/org");
}
/**
 * @summary Fetches all active virtual meeting rooms and their transcripts.
 * @param None
 * @returns Promise<MeetingRoom[]>
 * @throws May throw an error
 * @remarks Side Effects: None
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
 * @summary Fetches the accumulated API token costs and usage metrics for the organization.
 * @param None
 * @returns Promise<CostSummary>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export async function fetchCosts(): Promise<CostSummary> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}
/**
 * @summary Fetches a complete, normalized snapshot of the organization's current orchestration state.
 * @param None
 * @returns Promise<DashboardSnapshot>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export async function fetchDashboard(): Promise<DashboardSnapshot> {
  const response = await authedGetJSON<Record<string, unknown>>("/api/dashboard");
  return normalizeDashboard(response);
}
/**
 * @summary Dispatches a message or task from one agent to another within a specific meeting.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
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
 * @summary Executes the hireAgent operation.
 * @param name
 * @param role
 * @param providerType - Optional AI provider type (e.g. "minimax", "claude", "builtin")
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function hireAgent(name: string, role: string, providerType?: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role, ...(providerType ? { providerType } : {}) });
}
/**
 * @summary Executes the delegateTask operation.
 * @param fromAgentId
 * @param toAgentId
 * @param content
 * @param meetingId
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function delegateTask(
  fromAgentId: string,
  toAgentId: string,
  content: string,
  meetingId?: string,
): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/delegate", {
    fromAgentId,
    toAgentId,
    content,
    meetingId,
  });
}
/**
 * @summary Terminates an agent's process and removes it from the orchestration hub.
 * @param agentId
 * @returns Promise<DashboardSnapshot>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}
/**
 * @summary Retrieves available organizational domain templates (e.g., Software Company).
 * @param None
 * @returns Promise<DomainInfo[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchDomains(): Promise<DomainInfo[]> {
  return authedGetJSON<DomainInfo[]>("/api/domains");
}
/**
 * @summary Retrieves the catalog of active tools registered in the MCP gateway.
 * @param None
 * @returns Promise<MCPTool[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchMCPTools(): Promise<MCPTool[]> {
  return authedGetJSON<MCPTool[]>("/api/mcp/tools");
}
/**
 * @summary Overrides current state with a predefined scenario for demonstration purposes.
 * @param scenario
 * @returns Promise<DashboardSnapshot>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────
/**
 * @summary Retrieves all pending and resolved confidence gating approval requests.
 * @param None
 * @returns Promise<ApprovalRequest[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return authedGetJSON<ApprovalRequest[]>("/api/approvals");
}
/**
 * @summary Submits a new request for human manager sign-off on a high-risk action.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function requestApproval(body: {
  agentId: string;
  action: string;
  reason?: string;
  estimatedCostUsd?: number;
  riskLevel?: string;
}): Promise<ApprovalRequest> {
  return authedPostJSON<ApprovalRequest>("/api/approvals/request", body);
}
/**
 * @summary Submits the human manager's decision (approve/reject) for an approval request.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function decideApproval(
  approvalId: string,
  decision: "approve" | "reject",
  decidedBy?: string,
): Promise<ApprovalRequest[]> {
  return authedPostJSON<ApprovalRequest[]>("/api/approvals/decide", { approvalId, decision, decidedBy });
}

// ── Warm Handoff ──────────────────────────────────────────────────────────────
/**
 * @summary Retrieves all warm handoff escalations across the organization.
 * @param None
 * @returns Promise<HandoffPackage[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return authedGetJSON<HandoffPackage[]>("/api/handoffs");
}
/**
 * @summary Escalates a complex task from an autonomous agent to a human manager.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function createHandoff(body: {
  fromAgentId: string;
  toHumanRole?: string;
  intent: string;
  failedAttempts?: number;
  currentState?: string;
  visualGroundTruth?: string;
}): Promise<HandoffPackage> {
  return authedPostJSON<HandoffPackage>("/api/handoffs", body);
}

/**
 * @summary Executes the resolveHandoff operation.
 * @param handoffId
 * @param status
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function resolveHandoff(handoffId: string, status: "acknowledged" | "resolved"): Promise<HandoffPackage[]> {
  return authedPostJSON<HandoffPackage[]>("/api/handoffs/resolve", { handoffId, status });
}

// ── Pipelines (Automated SDLC) ────────────────────────────────────────────────
/**
 * @summary Executes the fetchPipelines operation.
 * @param None
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function fetchPipelines(): Promise<ApiPipeline[]> {
  return authedGetJSON<ApiPipeline[]>("/api/pipelines");
}

/**
 * @summary Executes the createPipeline operation.
 * @param body
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function createPipeline(body: { name: string; branch: string; initiatedBy: string }): Promise<ApiPipeline> {
  return authedPostJSON<ApiPipeline>("/api/pipelines", body);
}

/**
 * @summary Executes the promotePipeline operation.
 * @param body
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function promotePipeline(body: ApiPipelinePromoteRequest): Promise<ApiPipeline> {
  return authedPostJSON<ApiPipeline>("/api/pipelines/promote", body);
}

// ── Identity Management ───────────────────────────────────────────────────────
/**
 * @summary Retrieves the SPIFFE SVID certificates issued to the current workforce.
 * @param None
 * @returns Promise<AgentIdentity[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchIdentities(): Promise<AgentIdentity[]> {
  return authedGetJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────
/**
 * @summary Retrieves all imported skill packs available for agent instantiation.
 * @param None
 * @returns Promise<SkillPack[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchSkillPacks(): Promise<SkillPack[]> {
  return authedGetJSON<SkillPack[]>("/api/skills");
}
/**
 * @summary Imports a new specialized skill pack into the organization's domain.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function importSkillPack(body: {
  name: string;
  domain: string;
  description?: string;
  source?: string;
  author?: string;
}): Promise<SkillPack> {
  return authedPostJSON<SkillPack>("/api/skills/import", body);
}

// ── Snapshots ─────────────────────────────────────────────────────────────────
/**
 * @summary Retrieves all point-in-time recovery snapshots for the organization.
 * @param None
 * @returns Promise<OrgSnapshot[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return authedGetJSON<OrgSnapshot[]>("/api/snapshots");
}
/**
 * @summary Captures a point-in-time snapshot of the entire organization's memory and state.
 * @param label?
 * @returns Promise<OrgSnapshot>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return authedPostJSON<OrgSnapshot>("/api/snapshots/create", { label });
}
/**
 * @summary Restores the organization to a specific point-in-time snapshot.
 * @param snapshotId
 * @returns Promise<DashboardSnapshot>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────
/**
 * @summary Retrieves the catalog of community-published agents and tools.
 * @param None
 * @returns Promise<MarketplaceItem[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return authedGetJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────
/**
 * @summary Fetches real-time operational and health metrics for the organization.
 * @param None
 * @returns Promise<AnalyticsSummary>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchAnalytics(): Promise<AnalyticsSummary> {
  return authedGetJSON<AnalyticsSummary>("/api/analytics");
}

// ── External Integrations ─────────────────────────────────────────────────────

import type {
  ChatMessage,
  Integration,
  Issue,
  PullRequest,
} from "./types";
/**
 * @summary Retrieves external service connections, optionally filtered by category.
 * @param category?
 * @returns Promise<Integration[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchIntegrations(category?: string): Promise<Integration[]> {
  const q = category ? `?category=${category}` : "";
  return authedGetJSON<Integration[]>(`/api/integrations${q}`);
}
/**
 * @summary Connects and authenticates a specific external integration.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
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
  return authedPostJSON<Integration>("/api/integrations/connect", {
    integrationId,
    baseUrl: config?.baseUrl,
    botToken: config?.botToken,
    chatId: config?.chatId,
    webhookUrl: config?.webhookUrl,
    apiToken: config?.apiToken,
  });
}
/**
 * @summary Disconnects an active external integration.
 * @param integrationId
 * @returns Promise<Integration>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function disconnectIntegration(integrationId: string): Promise<Integration> {
  return authedPostJSON<Integration>("/api/integrations/disconnect", { integrationId });
}
/**
 * @summary Sends a test message to validate credentials before saving them.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function testChatIntegration(
  integrationId: string,
  config: { botToken?: string; chatId?: string; webhookUrl?: string },
): Promise<{ success: boolean }> {
  return authedPostJSON<{ success: boolean }>("/api/integrations/chat/test", {
    integrationId,
    ...config,
  });
}
/**
 * @summary Fetches recorded chat messages from the integration registry.
 * @param integrationId?
 * @returns Promise<ChatMessage[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchChatMessages(integrationId?: string): Promise<ChatMessage[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return authedGetJSON<ChatMessage[]>(`/api/integrations/chat/messages${q}`);
}
/**
 * @summary Dispatches a message to an external chat platform.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function sendChatMessage(body: {
  integrationId: string;
  channel: string;
  fromAgent: string;
  content: string;
  threadId?: string;
}): Promise<ChatMessage> {
  return authedPostJSON<ChatMessage>("/api/integrations/chat/send", body);
}
/**
 * @summary Fetches pull requests opened via the git integrations.
 * @param integrationId?
 * @returns Promise<PullRequest[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchPullRequests(integrationId?: string): Promise<PullRequest[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return authedGetJSON<PullRequest[]>(`/api/integrations/git/prs${q}`);
}
/**
 * @summary Opens a pull request/merge request on a connected git platform.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
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
  return authedPostJSON<PullRequest>("/api/integrations/git/pr/create", body);
}
/**
 * @summary Merges an open pull request on a connected git platform.
 * @param prId
 * @returns Promise<PullRequest>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function mergePullRequest(prId: string): Promise<PullRequest> {
  return authedPostJSON<PullRequest>("/api/integrations/git/pr/merge", { prId });
}
/**
 * @summary Closes an open pull request on a connected git platform without merging.
 * @param prId
 * @returns Promise<PullRequest>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function closePullRequest(prId: string): Promise<PullRequest> {
  return authedPostJSON<PullRequest>("/api/integrations/git/pr/close", { prId });
}
/**
 * @summary Fetches tickets from connected issue trackers.
 * @param integrationId?
 * @returns Promise<Issue[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchIssues(integrationId?: string): Promise<Issue[]> {
  const q = integrationId ? `?integrationId=${integrationId}` : "";
  return authedGetJSON<Issue[]>(`/api/integrations/issues${q}`);
}
/**
 * @summary Creates a ticket in a connected issue tracker.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
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
  return authedPostJSON<Issue>("/api/integrations/issues/create", body);
}
/**
 * @summary Executes the updateIssueStatus operation.
 * @param issueId
 * @param status
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function updateIssueStatus(issueId: string, status: string): Promise<Issue> {
  return authedPostJSON<Issue>("/api/integrations/issues/status", { issueId, status });
}
/**
 * @summary Executes the assignIssue operation.
 * @param issueId
 * @param assignee
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function assignIssue(issueId: string, assignee: string): Promise<Issue> {
  return authedPostJSON<Issue>("/api/integrations/issues/assign", { issueId, assignee });
}
/**
 * @summary Invokes an MCP tool with the given action and parameters. Communication tools route to the underlying connected integration. Git/issue tools create PRs or tickets in the connected platform.
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function invokeMCPTool(
  toolId: string,
  action: string,
  params: Record<string, string>,
): Promise<Record<string, unknown>> {
  return authedPostJSON<Record<string, unknown>>("/api/mcp/tools/invoke", { toolId, action, params });
}
/**
 * @summary Fetches the user's or organization's global settings and preferences.
 * @param None
 * @returns Promise<Settings>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchSettings(): Promise<Settings> {
  return authedGetJSON<Settings>("/api/settings");
}
/**
 * @summary Saves and updates the global settings and preferences.
 * @param settings
 * @returns Promise<Settings>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function saveSettings(settings: Settings): Promise<Settings> {
  return authedPostJSON<Settings>("/api/settings", settings);
}

// ── Auth ──────────────────────────────────────────────────────────────────────

const TOKEN_KEY = "ohc_token";
/**
 * @summary Retrieves the currently stored authentication JWT token from local storage.
 * @param None
 * @returns string | null
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}
/**
 * @summary Persists an authentication JWT token in local storage.
 * @param token
 * @returns void
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}
/**
 * @summary Removes the stored authentication JWT token from local storage.
 * @param None
 * @returns void
 * @throws May throw an error
 * @remarks Side Effects: None
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
 * @summary Executes the login operation.
 * @param username
 * @param password
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
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
 * @summary Invalidates the current session and clears the stored authentication token.
 * @param None
 * @returns Promise<void>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export async function logout(): Promise<void> {
  try {
    await authedPostJSON<void>("/api/auth/logout", {});
  } finally {
    clearStoredToken();
  }
}
/**
 * @summary Retrieves the public profile information of the currently authenticated user.
 * @param None
 * @returns Promise<UserPublic>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchMe(): Promise<UserPublic> {
  return authedGetJSON<UserPublic>("/api/auth/me");
}
/**
 * @summary Retrieves a list of all registered users in the system (requires Admin role).
 * @param None
 * @returns Promise<UserPublic[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchUsers(): Promise<UserPublic[]> {
  return authedGetJSON<UserPublic[]>("/api/users");
}
/**
 * @summary Creates a new user account within the system (requires Admin role).
 * @param None
 * @returns None
 * @throws May throw an error
 * @remarks Side Effects: None
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
 * @summary Deletes an existing user account from the system (requires Admin role).
 * @param id
 * @returns Promise<void>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export async function deleteUser(id: string): Promise<void> {
  const token = getStoredToken();
  await fetch(`/api/users/${id}`, {
    method: "DELETE",
    headers: { Authorization: token ? `Bearer ${token}` : "" },
  });
}
/**
 * @summary Retrieves the list of available operational roles and their associated permissions.
 * @param None
 * @returns Promise<Role[]>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function fetchRoles(): Promise<Role[]> {
  return authedGetJSON<Role[]>("/api/roles");
}
/**
 * @summary Creates a new custom role with an optional set of permissions.
 * @param body
 * @returns Promise<Role>
 * @throws May throw an error
 * @remarks Side Effects: None
 */
export function createRole(body: { name: string; permissions?: string[] }): Promise<Role> {
  return authedPostJSON<Role>("/api/roles", body);
}
/**
 * @summary Executes the scaleAgents operation.
 * @param role
 * @param count
 * @returns Promise with operation result
 * @throws May throw an error if the API request fails
 * @remarks Side Effects: Mutates server state
 */
export function scaleAgents(
  role: string,
  count: number,
): Promise<{ status: string; role: string; count: number }> {
  return authedPostJSON<{ status: string; role: string; count: number }>("/api/v1/scale", { role, count });
}
