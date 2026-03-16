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

export function fetchOrganization(): Promise<Organization> {
  return getJSON<Organization>("/api/org");
}

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

export async function fetchCosts(): Promise<CostSummary> {
  const response = await getJSON<Record<string, unknown>>("/api/costs");
  return normalizeCosts(response);
}

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

export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}

export function fireAgent(agentId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/agents/fire", { agentId });
}

export function fetchDomains(): Promise<DomainInfo[]> {
  return getJSON<DomainInfo[]>("/api/domains");
}

export function fetchMCPTools(): Promise<MCPTool[]> {
  return getJSON<MCPTool[]>("/api/mcp/tools");
}

export function seedScenario(scenario: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/dev/seed", { scenario });
}
// ── Approval / Confidence Gating ─────────────────────────────────────────────

export function fetchApprovals(): Promise<ApprovalRequest[]> {
  return getJSON<ApprovalRequest[]>("/api/approvals");
}

export function requestApproval(body: {
  agentId: string;
  action: string;
  reason?: string;
  estimatedCostUsd?: number;
  riskLevel?: string;
}): Promise<ApprovalRequest> {
  return postJSON<ApprovalRequest>("/api/approvals/request", body);
}

export function decideApproval(
  approvalId: string,
  decision: "approve" | "reject",
  decidedBy?: string,
): Promise<ApprovalRequest[]> {
  return postJSON<ApprovalRequest[]>("/api/approvals/decide", { approvalId, decision, decidedBy });
}

// ── Warm Handoff ──────────────────────────────────────────────────────────────

export function fetchHandoffs(): Promise<HandoffPackage[]> {
  return getJSON<HandoffPackage[]>("/api/handoffs");
}

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

export function fetchIdentities(): Promise<AgentIdentity[]> {
  return getJSON<AgentIdentity[]>("/api/identities");
}

// ── Skill Packs ───────────────────────────────────────────────────────────────

export function fetchSkillPacks(): Promise<SkillPack[]> {
  return getJSON<SkillPack[]>("/api/skills");
}

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

export function fetchSnapshots(): Promise<OrgSnapshot[]> {
  return getJSON<OrgSnapshot[]>("/api/snapshots");
}

export function createSnapshot(label?: string): Promise<OrgSnapshot> {
  return postJSON<OrgSnapshot>("/api/snapshots/create", { label });
}

export function restoreSnapshot(snapshotId: string): Promise<DashboardSnapshot> {
  return postJSON<DashboardSnapshot>("/api/snapshots/restore", { snapshotId });
}

// ── Marketplace ───────────────────────────────────────────────────────────────

export function fetchMarketplace(): Promise<MarketplaceItem[]> {
  return getJSON<MarketplaceItem[]>("/api/marketplace");
}

// ── Real-time Analytics ───────────────────────────────────────────────────────

export function fetchAnalytics(): Promise<AnalyticsSummary> {
  return getJSON<AnalyticsSummary>("/api/analytics");
}
