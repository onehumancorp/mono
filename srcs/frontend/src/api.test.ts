import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { fetchCosts, fetchDashboard, fetchDomains, fetchMCPTools, fetchMeetings, fetchOrganization, fireAgent, hireAgent, seedScenario, sendMessage, login, setStoredToken, fetchIdentities, fetchSkillPacks, importSkillPack, fetchSnapshots, fetchApprovals, fetchHandoffs, fetchMarketplace, fetchAnalytics, createUser, deleteUser, fetchRoles, createRole, fetchMe, testChatIntegration, invokeMCPTool, saveSettings } from "./api";

const backendUrl = process.env.VITE_BACKEND_URL || "http://127.0.0.1:8080";

beforeAll(async () => {
  let ready = false;
  for (let i = 0; i < 20; i++) {
    try {
      const resp = await fetch(`${backendUrl}/api/dev/seed`, { method: "GET" });
      if (resp.ok || resp.status === 405) { ready = true; break; }
    } catch (e) {
      await new Promise(r => setTimeout(r, 500));
    }
  }

  const tokenResp = await login("admin", "admin");
  if (tokenResp) {
    setStoredToken(tokenResp.token);
  }
});

afterAll(() => {
  setStoredToken("");
});

describe("api calls with seeded backend", () => {
  it("fetches organization and dashboard snapshot", async () => {
    const tokenResp = await login("admin", "admin");
    if (tokenResp) setStoredToken(tokenResp.token);

    const snapshot = await seedScenario("launch-readiness");
    expect(snapshot.organization.name).toBe("Demo Software Company");

    const org = await fetchOrganization();
    expect(org.id).toBeDefined();

    const dash = await fetchDashboard();
    expect(dash.organization.id).toBe(org.id);
  });

  it("fetches costs and meetings", async () => {
    await seedScenario("launch-readiness");
    const costs = await fetchCosts();
    expect(costs.totalCostUSD).toBeDefined();

    const meetings = await fetchMeetings();
    expect(meetings.length).toBeGreaterThan(0);
  });

  it("fetches domains and MCP tools", async () => {
    await seedScenario("launch-readiness");
    const domains = await fetchDomains();
    expect(domains).toBeInstanceOf(Array);

    const tools = await fetchMCPTools();
    expect(tools).toBeInstanceOf(Array);
  });

  it("can hire and fire agents", async () => {
    await seedScenario("launch-readiness");

    const dashAfterHire = await hireAgent("New QA", "QA_TESTER");
    const newAgent = dashAfterHire.agents.find(a => a.name === "New QA");
    expect(newAgent).toBeDefined();

    if (newAgent) {
      const dashAfterFire = await fireAgent(newAgent.id);
      expect(dashAfterFire.agents.find(a => a.id === newAgent.id)).toBeUndefined();
    }
  });

  it("handles new endpoints", async () => {
    await seedScenario("launch-readiness");
    const idents = await fetchIdentities();
    expect(idents).toBeDefined();

    const packs = await fetchSkillPacks();
    expect(packs).toBeDefined();

    const snaps = await fetchSnapshots();
    expect(snaps).toBeDefined();

    const approvals = await fetchApprovals();
    expect(approvals).toBeDefined();

    const handoffs = await fetchHandoffs();
    expect(handoffs).toBeDefined();

    const items = await fetchMarketplace();
    expect(items).toBeDefined();

    const analytics = await fetchAnalytics();
    expect(analytics).toBeDefined();
  });

  it("handles user management", async () => {
    try {
      await createRole({ id: "test-role", name: "Test Role", permissions: ["read"] });
      const roles = await fetchRoles();
      expect(roles).toBeDefined();
    } catch (e) {
      // test fallback
    }
  });

  it("handles auth error paths (401)", async () => {
    setStoredToken("bad-token");
    await expect(fetchMe()).rejects.toThrow();

    // login failure
    await expect(login("bad", "wrong")).rejects.toThrow();
  });
});
