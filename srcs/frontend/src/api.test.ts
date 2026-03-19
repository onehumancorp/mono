import { fetchCosts, fetchDashboard, fetchDomains, fetchMCPTools, fetchMeetings, fetchOrganization, fireAgent, hireAgent, seedScenario, sendMessage } from "./api";

describe("api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetches organization, meetings and costs", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/org") {
        return { ok: true, status: 200, json: async () => ({ id: "org" }) };
      }
      if (input === "/api/meetings") {
        return { ok: true, status: 200, json: async () => [{ id: "kickoff", transcript: null }] };
      }
      if (input === "/api/costs") {
        return {
          ok: true,
          status: 200,
          json: async () => ({
            organizationId: "org",
            totalTokens: 1,
            totalCostUsd: 0.5,
            agents: [{ agentId: "swe-1", tokenUsed: 1, costUsd: 0.5 }],
          }),
        };
      }
      return { ok: false, status: 404, json: async () => ({}) };
    });

    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchOrganization()).resolves.toEqual({ id: "org" });
    await expect(fetchMeetings()).resolves.toEqual([{ id: "kickoff", transcript: [] }]);
    await expect(fetchCosts()).resolves.toEqual({
      organizationID: "org",
      totalTokens: 1,
      totalCostUSD: 0.5,
      agents: [{ agentID: "swe-1", model: "", tokenUsed: 1, costUSD: 0.5 }],
    });
  });

  it("normalizes missing costs fields", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => ({}) })));

    await expect(fetchCosts()).resolves.toEqual({
      organizationID: "",
      totalTokens: 0,
      totalCostUSD: 0,
      agents: [],
    });
  });

  it("supports uppercase cost field variants", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        status: 200,
        json: async () => ({
          organizationID: "org-up",
          totalTokens: 3,
          totalCostUSD: 1.2,
          agents: [{ agentID: "pm-1", model: "gpt-4o", tokenUsed: 3, costUSD: 1.2 }],
        }),
      }))
    );

    await expect(fetchCosts()).resolves.toEqual({
      organizationID: "org-up",
      totalTokens: 3,
      totalCostUSD: 1.2,
      agents: [{ agentID: "pm-1", model: "gpt-4o", tokenUsed: 3, costUSD: 1.2 }],
    });
  });

  it("fills default values for incomplete agent payload", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        status: 200,
        json: async () => ({
          organizationId: "org",
          totalTokens: 0,
          totalCostUsd: 0,
          agents: [{}],
        }),
      }))
    );

    await expect(fetchCosts()).resolves.toEqual({
      organizationID: "org",
      totalTokens: 0,
      totalCostUSD: 0,
      agents: [{ agentID: "", model: "", tokenUsed: 0, costUSD: 0 }],
    });
  });

  it("throws for failed GET request", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 500, json: async () => ({}) })));

    await expect(fetchOrganization()).rejects.toThrow("Request failed for /api/org: 500");
  });

  it("fetches dashboard snapshot and normalizes null transcripts", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        status: 200,
        json: async () => ({
          organization: {
            id: "org-1",
            name: "Demo Software Company",
            domain: "software_company",
            members: [],
            roleProfiles: [],
          },
          meetings: [{ id: "launch", transcript: null }],
          costs: {
            organizationID: "org-1",
            totalTokens: 3,
            totalCostUSD: 1.2,
            agents: [],
          },
          agents: [],
          statuses: [{ status: "IDLE", count: 1 }],
          updatedAt: "2026-03-13T00:00:00Z",
        }),
      }))
    );

    await expect(fetchDashboard()).resolves.toEqual({
      organization: {
        id: "org-1",
        name: "Demo Software Company",
        domain: "software_company",
        members: [],
        roleProfiles: [],
      },
      meetings: [{ id: "launch", transcript: [] }],
      costs: {
        organizationID: "org-1",
        totalTokens: 3,
        totalCostUSD: 1.2,
        agents: [],
      },
      agents: [],
      statuses: [{ status: "IDLE", count: 1 }],
      updatedAt: "2026-03-13T00:00:00Z",
    });
  });

  it("sends message with URL encoded body", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => ({}) }));
    vi.stubGlobal("fetch", fetchMock);

    await sendMessage({
      fromAgent: "pm-1",
      toAgent: "swe-1",
      meetingId: "kickoff",
      messageType: "task",
      content: "Review roadmap",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/messages",
      expect.objectContaining({
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          Accept: "application/json",
        },
        redirect: "follow",
      })
    );

    const firstCall = (fetchMock.mock.calls as unknown[][])[0];
    if (!firstCall) {
      throw new Error("expected at least one fetch call");
    }
    const init = firstCall[1] as RequestInit | undefined;
    expect(String(init?.body)).toContain("fromAgent=pm-1");
    expect(String(init?.body)).toContain("content=Review+roadmap");
  });

  it("throws for failed message send", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 400, json: async () => ({}), text: async () => "" })));

    await expect(
      sendMessage({
        fromAgent: "pm-1",
        toAgent: "swe-1",
        meetingId: "kickoff",
        messageType: "task",
        content: "x",
      })
    ).rejects.toThrow("Failed to send message: 400");
  });
});

describe("api – new endpoints", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const dashSnap = {
    organization: { id: "o", name: "N", domain: "d", members: [], roleProfiles: [] },
    meetings: [],
    costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] },
    agents: [],
    statuses: [],
    updatedAt: "2026-01-01T00:00:00Z",
  };

  it("hireAgent posts to /api/agents/hire and returns dashboard snapshot", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => dashSnap }));
    vi.stubGlobal("fetch", fetchMock);
    const result = await hireAgent("Alice", "SOFTWARE_ENGINEER");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/agents/hire",
      expect.objectContaining({ method: "POST" })
    );
    const body = JSON.parse(String((fetchMock.mock.calls[0] as any)[1]?.body ?? "{}"));
    expect(body).toEqual({ name: "Alice", role: "SOFTWARE_ENGINEER" });
    expect(result.organization.id).toBe("o");
  });

  it("hireAgent throws on non-OK response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 422, json: async () => ({}), text: async () => "" })));
    await expect(hireAgent("X", "Y")).rejects.toThrow("422");
  });

  it("fireAgent posts to /api/agents/fire and returns dashboard snapshot", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => dashSnap }));
    vi.stubGlobal("fetch", fetchMock);
    const result = await fireAgent("swe-1");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/agents/fire",
      expect.objectContaining({ method: "POST" })
    );
    const body = JSON.parse(String((fetchMock.mock.calls[0] as any)[1]?.body ?? "{}"));
    expect(body).toEqual({ agentId: "swe-1" });
    expect(result.organization.id).toBe("o");
  });

  it("fireAgent throws on non-OK response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 404, json: async () => ({}), text: async () => "" })));
    await expect(fireAgent("nobody")).rejects.toThrow("404");
  });

  it("fetchDomains returns domain list", async () => {
    const domains = [{ id: "software_company", name: "Software Company", description: "SaaS" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => domains })));
    await expect(fetchDomains()).resolves.toEqual(domains);
  });

  it("fetchDomains throws on non-OK response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 500, json: async () => ({}) })));
    await expect(fetchDomains()).rejects.toThrow("500");
  });

  it("fetchMCPTools returns tools list", async () => {
    const tools = [{ id: "git-mcp", name: "Git MCP", description: "Git", category: "dev", status: "available" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => tools })));
    await expect(fetchMCPTools()).resolves.toEqual(tools);
  });

  it("fetchMCPTools throws on non-OK response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 503, json: async () => ({}) })));
    await expect(fetchMCPTools()).rejects.toThrow("503");
  });

  it("seedScenario posts scenario name and returns dashboard snapshot", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => dashSnap }));
    vi.stubGlobal("fetch", fetchMock);
    const result = await seedScenario("digital-marketing");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/dev/seed",
      expect.objectContaining({ method: "POST" })
    );
    const body = JSON.parse(String((fetchMock.mock.calls[0] as any)[1]?.body ?? "{}"));
    expect(body).toEqual({ scenario: "digital-marketing" });
    expect(result.organization.id).toBe("o");
  });

  it("seedScenario throws on non-OK response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 400, json: async () => ({}), text: async () => "" })));
    await expect(seedScenario("bad")).rejects.toThrow("400");
  });

  it("fetchDashboard normalizes projectedMonthlyUSD field", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        status: 200,
        json: async () => ({
          organization: { id: "o", name: "N", domain: "d", ceoId: "ceo-1", members: [], roleProfiles: [] },
          meetings: [],
          costs: {
            organizationID: "o",
            totalTokens: 0,
            totalCostUSD: 1.5,
            projectedMonthlyUSD: 45.0,
            agents: [],
          },
          agents: [{ id: "a1", name: "Alice", role: "PM", organizationID: "o", status: "ACTIVE" }],
          statuses: [],
          updatedAt: "2026-03-01T00:00:00Z",
        }),
      }))
    );
    const snap = await fetchDashboard();
    expect(snap.costs.projectedMonthlyUSD).toBe(45.0);
    expect(snap.organization.ceoId).toBe("ceo-1");
    expect(snap.agents[0].id).toBe("a1");
  });

  it("fetchDashboard handles missing optional fields gracefully", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        status: 200,
        json: async () => ({
          organization: { id: "o2", name: "X", domain: "d2" },
          costs: {},
          agents: [{}],
          statuses: [{}],
        }),
      }))
    );
    const snap = await fetchDashboard();
    expect(snap.organization.members).toEqual([]);
    expect(snap.organization.roleProfiles).toEqual([]);
    expect(snap.costs.projectedMonthlyUSD).toBeUndefined();
    expect(snap.statuses[0].status).toBe("UNKNOWN");
    expect(snap.statuses[0].count).toBe(0);
  });
});

// ── Integration API tests ─────────────────────────────────────────────────────

import {
  assignIssue,
  closePullRequest,
  connectIntegration,
  createIssue,
  createPullRequest,
  disconnectIntegration,
  fetchChatMessages,
  fetchIntegrations,
  fetchIssues,
  fetchPullRequests,
  mergePullRequest,
  sendChatMessage,
  updateIssueStatus,
} from "./api";

describe("integration api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetches all integrations", async () => {
    const list = [{ id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => list })));
    await expect(fetchIntegrations()).resolves.toEqual(list);
  });

  it("fetches integrations by category", async () => {
    const list = [{ id: "github", category: "git" }];
    vi.stubGlobal("fetch", vi.fn(async (url: string) => {
      expect(url).toBe("/api/integrations?category=git");
      return { ok: true, status: 200, json: async () => list };
    }));
    await expect(fetchIntegrations("git")).resolves.toEqual(list);
  });

  it("connects an integration", async () => {
    const updated = { id: "slack", status: "connected" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => updated })));
    await expect(connectIntegration("slack", { baseUrl: "https://hooks.slack.com/test" })).resolves.toEqual(updated);
  });

  it("disconnects an integration", async () => {
    const updated = { id: "slack", status: "disconnected" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => updated })));
    await expect(disconnectIntegration("slack")).resolves.toEqual(updated);
  });

  it("fetches chat messages", async () => {
    const msgs = [{ id: "msg-1", integrationId: "slack", channel: "#eng", fromAgent: "swe-1", content: "hi" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => msgs })));
    await expect(fetchChatMessages("slack")).resolves.toEqual(msgs);
  });

  it("fetches all chat messages without filter", async () => {
    const msgs: unknown[] = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string) => {
      expect(url).toBe("/api/integrations/chat/messages");
      return { ok: true, status: 200, json: async () => msgs };
    }));
    await expect(fetchChatMessages()).resolves.toEqual(msgs);
  });

  it("sends a chat message", async () => {
    const msg = { id: "msg-1", integrationId: "slack", channel: "#eng", fromAgent: "swe-1", content: "PR ready" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => msg })));
    const result = await sendChatMessage({ integrationId: "slack", channel: "#eng", fromAgent: "swe-1", content: "PR ready" });
    expect(result).toEqual(msg);
  });

  it("sends a chat message with threadId", async () => {
    const msg = { id: "msg-2", threadId: "t-42" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => msg })));
    const result = await sendChatMessage({ integrationId: "discord", channel: "general", fromAgent: "pm-1", content: "hi", threadId: "t-42" });
    expect(result.threadId).toBe("t-42");
  });

  it("fetches pull requests", async () => {
    const prs = [{ id: "pr-1", status: "open" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => prs })));
    await expect(fetchPullRequests("github")).resolves.toEqual(prs);
  });

  it("fetches all pull requests without filter", async () => {
    vi.stubGlobal("fetch", vi.fn(async (url: string) => {
      expect(url).toBe("/api/integrations/git/prs");
      return { ok: true, status: 200, json: async () => [] };
    }));
    await expect(fetchPullRequests()).resolves.toEqual([]);
  });

  it("creates a pull request", async () => {
    const pr = { id: "pr-1", status: "open", title: "feat: billing" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => pr })));
    const result = await createPullRequest({
      integrationId: "github",
      repository: "org/repo",
      title: "feat: billing",
      sourceBranch: "feature/billing",
      targetBranch: "main",
    });
    expect(result).toEqual(pr);
  });

  it("merges a pull request", async () => {
    const pr = { id: "pr-1", status: "merged" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => pr })));
    await expect(mergePullRequest("pr-1")).resolves.toEqual(pr);
  });

  it("closes a pull request", async () => {
    const pr = { id: "pr-1", status: "closed" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => pr })));
    await expect(closePullRequest("pr-1")).resolves.toEqual(pr);
  });

  it("fetches issues", async () => {
    const issues = [{ id: "issue-1", status: "open" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => issues })));
    await expect(fetchIssues("jira")).resolves.toEqual(issues);
  });

  it("fetches all issues without filter", async () => {
    vi.stubGlobal("fetch", vi.fn(async (url: string) => {
      expect(url).toBe("/api/integrations/issues");
      return { ok: true, status: 200, json: async () => [] };
    }));
    await expect(fetchIssues()).resolves.toEqual([]);
  });

  it("creates an issue", async () => {
    const issue = { id: "issue-1", status: "open", priority: "high" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => issue })));
    const result = await createIssue({
      integrationId: "jira",
      project: "PROJ",
      title: "Billing dashboard",
      priority: "high",
    });
    expect(result).toEqual(issue);
  });

  it("updates issue status", async () => {
    const issue = { id: "issue-1", status: "in_progress" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => issue })));
    await expect(updateIssueStatus("issue-1", "in_progress")).resolves.toEqual(issue);
  });

  it("assigns an issue", async () => {
    const issue = { id: "issue-1", assignedTo: "swe-1" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => issue })));
    await expect(assignIssue("issue-1", "swe-1")).resolves.toEqual(issue);
  });
});

// ── Approval / Handoff / Identity / Skill / Snapshot / Marketplace / Analytics ─

import {
  fetchApprovals,
  requestApproval,
  decideApproval,
  fetchHandoffs,
  createHandoff,
  fetchIdentities,
  fetchSkillPacks,
  importSkillPack,
  fetchSnapshots,
  createSnapshot,
  restoreSnapshot,
  fetchMarketplace,
  fetchAnalytics,
} from "./api";

describe("approval api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchApprovals returns list", async () => {
    const approvals = [{ id: "a-1", status: "PENDING" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => approvals })));
    await expect(fetchApprovals()).resolves.toEqual(approvals);
  });

  it("requestApproval posts and returns approval", async () => {
    const approval = { id: "a-2", status: "PENDING" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => approval })));
    const result = await requestApproval({ agentId: "swe-1", action: "deploy", reason: "prod release", estimatedCostUsd: 200, riskLevel: "high" });
    expect(result).toEqual(approval);
  });

  it("decideApproval posts decision and returns updated approvals", async () => {
    const approvals = [{ id: "a-1", status: "APPROVED" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => approvals })));
    const result = await decideApproval("a-1", "approve", "CEO");
    expect(result).toEqual(approvals);
  });
});

describe("handoff api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchHandoffs returns list", async () => {
    const handoffs = [{ id: "h-1", status: "pending" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => handoffs })));
    await expect(fetchHandoffs()).resolves.toEqual(handoffs);
  });

  it("createHandoff posts and returns handoff package", async () => {
    const handoff = { id: "h-2", status: "pending" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => handoff })));
    const result = await createHandoff({ fromAgentId: "swe-1", intent: "Need design review", failedAttempts: 2, currentState: "awaiting" });
    expect(result).toEqual(handoff);
  });
});

describe("identity api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchIdentities returns identity list", async () => {
    const identities = [{ agentId: "swe-1", svid: "spiffe://corp/swe-1" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => identities })));
    await expect(fetchIdentities()).resolves.toEqual(identities);
  });
});

describe("skill pack api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchSkillPacks returns list", async () => {
    const packs = [{ id: "sp-1", name: "Marketing Pack" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => packs })));
    await expect(fetchSkillPacks()).resolves.toEqual(packs);
  });

  it("importSkillPack posts and returns imported pack", async () => {
    const pack = { id: "sp-2", name: "Legal Pack", domain: "legal", description: "Legal consulting", source: "custom" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => pack })));
    const result = await importSkillPack({ name: "Legal Pack", domain: "legal", description: "Legal consulting", source: "custom", author: "admin" });
    expect(result).toEqual(pack);
  });
});

describe("snapshot api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchSnapshots returns list", async () => {
    const snaps = [{ id: "snap-1", label: "v1.0" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => snaps })));
    await expect(fetchSnapshots()).resolves.toEqual(snaps);
  });

  it("createSnapshot posts label and returns snapshot", async () => {
    const snap = { id: "snap-2", label: "pre-launch" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => snap })));
    const result = await createSnapshot("pre-launch");
    expect(result).toEqual(snap);
  });

  it("restoreSnapshot posts snapshotId and returns dashboard", async () => {
    const dashboard = { organization: { id: "org-1", name: "Acme", domain: "software_company", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "org-1", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: "2026-03-01T00:00:00Z" };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => dashboard })));
    const result = await restoreSnapshot("snap-1");
    expect(result).toEqual(dashboard);
  });
});

describe("marketplace api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchMarketplace returns items", async () => {
    const items = [{ id: "item-1", name: "TikTok Virality Expert", type: "agent" }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => items })));
    await expect(fetchMarketplace()).resolves.toEqual(items);
  });
});

describe("analytics api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchAnalytics returns analytics summary", async () => {
    const analytics = { humanAgentRatio: 0.05, totalAgents: 20, totalHumans: 1, auditFidelityPct: 98.5, resumptionLatencyMs: 4200, pendingApprovals: 2, activeHandoffs: 1, tokenVelocity: 1500 };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => analytics })));
    await expect(fetchAnalytics()).resolves.toEqual(analytics);
  });
});

// ── Branch coverage: projectedMonthlyUsd lowercase fallback ─────────────────

import { fetchCosts as fetchCosts2 } from "./api";

describe("api – projectedMonthlyUsd branch", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("uses projectedMonthlyUsd lowercase fallback when projectedMonthlyUSD is null", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        organizationID: "org-1",
        totalTokens: 0,
        totalCostUSD: 0,
        projectedMonthlyUSD: null,
        projectedMonthlyUsd: 30.0,
        agents: [],
      }),
    })));
    const result = await fetchCosts2();
    expect(result.projectedMonthlyUSD).toBe(30.0);
  });
});

// ── Branch coverage: projectedMonthlyUSD ?? 0 final fallback ─────────────────

describe("api – projectedMonthlyUSD zero fallback", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("returns 0 when both projectedMonthlyUSD and projectedMonthlyUsd are null", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        organizationID: "org-1",
        totalTokens: 0,
        totalCostUSD: 0,
        projectedMonthlyUSD: null,
        projectedMonthlyUsd: null,
        agents: [],
      }),
    })));
    const result = await fetchCosts2();
    expect(result.projectedMonthlyUSD).toBe(0);
  });
});

// ── User management API coverage ─────────────────────────────────────────────

import { createUser, deleteUser, fetchRoles, createRole } from "./api";

describe("user management api", () => {
  beforeEach(() => {
    localStorage.setItem("ohc_token", "test-token");
  });

  afterEach(() => {
    localStorage.removeItem("ohc_token");
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("createUser posts body and returns created user", async () => {
    const newUser = { id: "u-1", username: "alice", email: "alice@example.com", roles: ["operator"] };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => newUser })));
    const result = await createUser({ username: "alice", email: "alice@example.com", password: "secret", roles: ["operator"] });
    expect(result).toEqual(newUser);
  });

  it("deleteUser sends DELETE request and completes without error", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 204, json: async () => ({}) })));
    await expect(deleteUser("u-1")).resolves.toBeUndefined();
    const fetchMock = vi.mocked(fetch as ReturnType<typeof vi.fn>);
    expect(fetchMock).toHaveBeenCalledWith("/api/users/u-1", expect.objectContaining({ method: "DELETE" }));
  });

  it("fetchRoles returns list of roles", async () => {
    const roles = [{ id: "r-1", name: "admin", permissions: [] }];
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => roles })));
    await expect(fetchRoles()).resolves.toEqual(roles);
  });

  it("createRole posts body and returns created role", async () => {
    const role = { id: "r-2", name: "operator", permissions: ["read"] };
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => role })));
    const result = await createRole({ name: "operator", permissions: ["read"] });
    expect(result).toEqual(role);
  });
});

// ── Auth functions coverage ───────────────────────────────────────────────────

import { login, logout, fetchMe, setStoredToken, clearStoredToken, testChatIntegration, invokeMCPTool, saveSettings } from "./api";

describe("api – auth functions", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("setStoredToken stores token in localStorage", () => {
    setStoredToken("my-token");
    expect(localStorage.getItem("ohc_token")).toBe("my-token");
  });

  it("clearStoredToken removes token from localStorage", () => {
    localStorage.setItem("ohc_token", "tok");
    clearStoredToken();
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("login success sets token and returns response", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({ token: "abc123", username: "admin" }),
      text: async () => "",
    })));
    const result = await login("admin", "password");
    expect(result.token).toBe("abc123");
    expect(localStorage.getItem("ohc_token")).toBe("abc123");
  });

  it("login failure throws error with response text", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 401,
      json: async () => ({}),
      text: async () => "Bad credentials",
    })));
    await expect(login("admin", "wrong")).rejects.toThrow("Bad credentials");
  });

  it("login failure throws 'Login failed' when text is empty", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 401,
      json: async () => ({}),
      text: async () => "",
    })));
    await expect(login("admin", "wrong")).rejects.toThrow("Login failed");
  });

  it("logout calls /api/auth/logout and clears token", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({}),
      text: async () => "",
    })));
    await logout();
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("fetchMe calls /api/auth/me and returns user", async () => {
    localStorage.setItem("ohc_token", "tok");
    const user = { id: "u-1", username: "admin", email: "admin@example.com", roles: ["admin"] };
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => user,
      text: async () => "",
    })));
    await expect(fetchMe()).resolves.toEqual(user);
  });
});

describe("api – authedGetJSON error paths", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("authedGetJSON 401 clears token and throws Unauthorized", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 401,
      json: async () => ({}),
      text: async () => "",
    })));
    await expect(fetchMe()).rejects.toThrow("Unauthorized");
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("authedGetJSON non-401 throws Request failed error", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 500,
      json: async () => ({}),
      text: async () => "",
    })));
    await expect(fetchMe()).rejects.toThrow("Request failed for /api/auth/me: 500");
  });
});

describe("api – authedPostJSON error paths", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("authedPostJSON 401 clears token and throws Unauthorized", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 401,
      json: async () => ({}),
      text: async () => "Unauthorized response",
    })));
    const { createUser: createUserAuth } = await import("./api");
    await expect(createUserAuth({ username: "x", email: "x@x.com", password: "p" })).rejects.toThrow("Unauthorized");
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("authedPostJSON non-401 with error text throws that text", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 422,
      json: async () => ({}),
      text: async () => "Validation error",
    })));
    const { createUser: createUserAuth2 } = await import("./api");
    await expect(createUserAuth2({ username: "x", email: "x@x.com", password: "p" })).rejects.toThrow("Validation error");
  });

  it("authedPostJSON non-401 with empty text throws Request failed error", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 500,
      json: async () => ({}),
      text: async () => "",
    })));
    const { createUser: createUserAuth3 } = await import("./api");
    await expect(createUserAuth3({ username: "x", email: "x@x.com", password: "p" })).rejects.toThrow("Request failed for /api/users: 500");
  });
});

describe("api – testChatIntegration, invokeMCPTool, saveSettings", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("testChatIntegration posts and returns success", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({ success: true }),
      text: async () => "",
    })));
    await expect(testChatIntegration("telegram", { botToken: "tok", chatId: "123" })).resolves.toEqual({ success: true });
  });

  it("invokeMCPTool posts and returns result", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({ output: "done" }),
      text: async () => "",
    })));
    await expect(invokeMCPTool("git-mcp", "invoke", { repo: "my/repo" })).resolves.toEqual({ output: "done" });
  });

  it("saveSettings posts and returns updated settings", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({ minimaxApiKey: "new-key" }),
      text: async () => "",
    })));
    await expect(saveSettings({ minimaxApiKey: "new-key" })).resolves.toEqual({ minimaxApiKey: "new-key" });
  });
});

describe("api – authed calls without token (empty auth header branch)", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("authedGetJSON uses empty Authorization when no token stored", async () => {
    // no token in localStorage
    const user = { id: "u-1", username: "admin", email: "a@b.com", roles: [] };
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => user,
      text: async () => "",
    })));
    await expect(fetchMe()).resolves.toEqual(user);
  });

  it("authedPostJSON uses empty Authorization when no token stored", async () => {
    const { createUser: cu } = await import("./api");
    const newUser = { id: "u-2", username: "x", email: "x@x.com", roles: [] };
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => newUser,
      text: async () => "",
    })));
    await expect(cu({ username: "x", email: "x@x.com", password: "p" })).resolves.toEqual(newUser);
  });

  it("authedPostJSON text().catch branch when text() throws", async () => {
    localStorage.setItem("ohc_token", "tok");
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: false, status: 500,
      json: async () => ({}),
      text: async () => { throw new Error("no text"); },
    })));
    const { createUser: cu } = await import("./api");
    await expect(cu({ username: "x", email: "x@x.com", password: "p" })).rejects.toThrow("Request failed for /api/users: 500");
  });
});

describe("api – deleteUser without token branch", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("deleteUser uses empty auth header when no token", async () => {
    // no token stored
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 204, json: async () => ({}), text: async () => "" })));
    const { deleteUser: du } = await import("./api");
    await expect(du("u-99")).resolves.toBeUndefined();
  });
});
