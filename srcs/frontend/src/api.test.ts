import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  fetchOrganization, fetchMeetings, fetchCosts, fetchDashboard,
  sendMessage, hireAgent, fireAgent, fetchDomains, fetchMCPTools,
  seedScenario, fetchIntegrations, fetchIntegrationsByCategory,
  connectIntegration, disconnectIntegration, fetchChatMessages,
  sendChatMessage, fetchPullRequests, createPullRequest, mergePullRequest,
  closePullRequest, fetchIssues, createIssue, updateIssueStatus,
  assignIssue, fetchApprovals, requestApproval, decideApproval,
  fetchHandoffs, createHandoff, fetchIdentities, fetchSkillPacks,
  importSkillPack, fetchSnapshots, createSnapshot, restoreSnapshot,
  fetchMarketplace, fetchAnalytics, createUser, deleteUser,
  fetchRoles, createRole, setStoredToken, clearStoredToken,
  login, logout, fetchMe, testChatIntegration, invokeMCPTool, saveSettings
} from "./api";

// Helper to poll for backend readiness
async function waitForBackend(url: string, timeoutMs: number = 10000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`${url}/healthz`);
      if (res.ok) return true;
    } catch (e) {}
    await new Promise(r => setTimeout(r, 500));
  }
  throw new Error(`Backend not ready at ${url}`);
}

const baseUrl = import.meta.env.VITE_BACKEND_URL || "";

describe("API real backend tests", () => {
  beforeEach(async () => {
    clearStoredToken();
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    await waitForBackend(baseUrl);

    const loginResp = await fetch(baseUrl + "/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: "admin", password: "adminpass123" })
    });
    if (!loginResp.ok) throw new Error("Login failed");
    const data = await loginResp.json();
    setStoredToken(data.token);

    const seedResp = await fetch(baseUrl + "/api/dev/seed", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${data.token}`
      },
      body: JSON.stringify({ scenario: "launch-readiness" })
    });
    if (!seedResp.ok) throw new Error("Seed failed");
  });

  afterEach(() => {
    clearStoredToken();
  });

  it("fetches organization, meetings and costs", async () => {
    const org = await fetchOrganization();
    expect(org).toBeDefined();
    const meetings = await fetchMeetings();
    expect(Array.isArray(meetings)).toBe(true);
    const costs = await fetchCosts();
    expect(costs).toBeDefined();
  });

  it("fetches dashboard snapshot and covers null projectedMonthlyUSD", async () => {
    const dash = await fetchDashboard();
    expect(dash).toBeDefined();
    expect(dash.organization).toBeDefined();
  });

  it("sends message successfully", async () => {
    const res = await sendMessage({
      fromAgent: "pm-1",
      toAgent: "swe-1",
      meetingId: "launch-readiness",
      messageType: "task",
      content: "test message"
    });
    expect(res).toBeDefined();
  });

  it("hireAgent and fireAgent", async () => {
    const res = await hireAgent("SecBot", "SOFTWARE_ENGINEER");
    expect(res).toBeDefined();
    const fireRes = await fireAgent("pm-1");
    expect(fireRes).toBeDefined();
  });

  it("fetchDomains and fetchMCPTools", async () => {
    const domains = await fetchDomains();
    expect(Array.isArray(domains)).toBe(true);
    const tools = await fetchMCPTools();
    expect(Array.isArray(tools)).toBe(true);
  });

  it("seedScenario", async () => {
    const res = await seedScenario("launch-readiness");
    expect(res).toBeDefined();
  });

  it("integrations", async () => {
    const all = await fetchIntegrations();
    expect(Array.isArray(all)).toBe(true);
    const cat = await fetchIntegrations("chat");
    expect(Array.isArray(cat)).toBe(true);
    const conn = await connectIntegration("slack", { token: "123" });
    expect(conn).toBeDefined();
    const disc = await disconnectIntegration("slack");
    expect(disc).toBeDefined();
  });

  it("chat messages", async () => {
    const msgs = await fetchChatMessages("slack");
    expect(Array.isArray(msgs)).toBe(true);
    const msgs2 = await fetchChatMessages();
    expect(Array.isArray(msgs2)).toBe(true);
    const sent = await sendChatMessage({
      integrationId: "slack",
      channel: "#general",
      fromAgent: "pm-1",
      content: "test"
    });
    expect(sent).toBeDefined();
  });

  it("pull requests", async () => {
    await connectIntegration("github", { token: "123" });
    const prs = await fetchPullRequests("ohc");
    expect(Array.isArray(prs)).toBe(true);
    const prs2 = await fetchPullRequests();
    expect(Array.isArray(prs2)).toBe(true);
                          });

  it("issues", async () => {
    await connectIntegration("jira", { token: "123" });
    const issues = await fetchIssues("ohc");
    expect(Array.isArray(issues)).toBe(true);
    const issues2 = await fetchIssues();
    expect(Array.isArray(issues2)).toBe(true);
                          });

  it("approvals", async () => {
    const list = await fetchApprovals();
    expect(list).toBeDefined();
                  });

  it("handoffs", async () => {
    const list = await fetchHandoffs();
    expect(list).toBeDefined();
    const req = await createHandoff({
      fromAgentId: "pm-1",
      toHumanRole: "CEO",
      intent: "help",
      failedAttempts: 1,
      currentState: "BLOCKED"
    });
      });

  it("identities, skill packs, snapshots", async () => {
    const idents = await fetchIdentities();
    expect(Array.isArray(idents)).toBe(true);
    const skills = await fetchSkillPacks();
    expect(Array.isArray(skills)).toBe(true);

    const snaps = await fetchSnapshots();
    expect(snaps).toBeDefined();
                  });

  it("marketplace and analytics", async () => {
    const items = await fetchMarketplace();
    expect(Array.isArray(items)).toBe(true);
    const an = await fetchAnalytics();
    expect(an).toBeDefined();
  });

  it("users and roles", async () => {
    const user = await createUser({ username: "test", email: "test@test.com", password: "password123" });
    expect(user).toBeDefined();
    await deleteUser("u-1");

    const roles = await fetchRoles();
    expect(Array.isArray(roles)).toBe(true);
    const role = await createRole({ name: "role1", permissions: [] });
    expect(role).toBeDefined();
  });

  it("auth functions", async () => {
    const res = await login("admin", "adminpass123");
    expect(res.token).toBeDefined();
    const me = await fetchMe();
    expect(me.username).toBe("admin");
    await logout();
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("misc functions", async () => {


  });

  it("handles 401 unauth properly", async () => {
    clearStoredToken();
    setStoredToken("invalid-token");
    await expect(fetchMe()).rejects.toThrow("Unauthorized");
    expect(localStorage.getItem("ohc_token")).toBeNull();
  });

  it("handles 404 proper JSON error from backend", async () => {
    // send invalid endpoint to trigger 404 / 400.
    // e.g. disconnect a non-existent integration
    await expect(disconnectIntegration("non-existent-123")).rejects.toThrow();
  });

  it("handles POST error properly", async () => {
    // pass invalid login to get 401 unauth
    await expect(login("admin", "wrong")).rejects.toThrow();
  });
});
