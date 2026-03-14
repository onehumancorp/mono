import { fetchCosts, fetchDashboard, fetchMeetings, fetchOrganization, sendMessage } from "./api";

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
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 400, json: async () => ({}) })));

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
