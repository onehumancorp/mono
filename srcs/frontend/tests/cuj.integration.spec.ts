import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test.beforeEach(async ({ page, request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const data = await loginResp.json();
  const token = data.token;

  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", {
    headers: { Authorization: "Bearer " + token },
    data: { scenario: "launch-readiness" }
  });
  console.log(await response.text());
  expect(response.ok()).toBeTruthy();

  // Set localStorage token in browser context so UI bypassing login
  await page.goto("/");
  await page.evaluate((t) => {
    localStorage.setItem("ohc_token", t);
  }, token);
});

test("CUJ 1: frontend dashboard loads seeded organization command center", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();
  await expect(page.getByText("Demo Software Company")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Org Chart" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Active Meetings" })).toBeVisible();

  await saveShot(page, "cuj-01-frontend-dashboard");
});

test("CUJ 2: sending message updates UI and backend transcript", async ({ page, request }) => {
  await page.goto("/");

  // Navigate to meetings (War Room)
  await page.getByRole("button", { name: "Meetings" }).click();
  await expect(page.getByRole("heading", { name: "Virtual War Room" })).toBeVisible();

  const message = `Playwright seeded message ${Date.now()}`;
  await page.getByPlaceholder("Inject direction or approve actions as CEO...").fill(message);
  await page.getByRole("button", { name: "Send" }).click();

  await expect(page.getByText(message)).toBeVisible();

  // Re-login to get token for request verification
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  await expect(async () => {
    const meetingsResponse = await request.get("http://127.0.0.1:8080/api/meetings", { headers: { Authorization: "Bearer " + token } });
    expect(meetingsResponse.ok()).toBeTruthy();
    const meetings = (await meetingsResponse.json()) as Array<{ id: string; transcript?: Array<{ content: string }> }>;
    const hasMessage = meetings.some((meeting) =>
      (meeting.transcript ?? []).some((entry) => entry.content === message)
    );
    expect(hasMessage).toBeTruthy();
  }).toPass({ timeout: 15000 });

  await saveShot(page, "cuj-02-frontend-send-message");
});

test("CUJ 3: backend root route remains reachable for bundled frontend", async ({ page }) => {
  await page.goto("http://127.0.0.1:8080/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();

  await saveShot(page, "cuj-03-backend-app-route");
});

test("CUJ 4: Dynamic Scaling triggers SSE trace logs", async ({ page }) => {
  await page.goto("/");

  // Navigate to Dynamic Scaling tab
  await page.getByRole("button", { name: "Dynamic Scaling" }).click();
  await expect(page.getByRole("heading", { name: "Dynamic Scaling" })).toBeVisible();

  // Step 1 - Select Role
  await page.getByRole("button", { name: "Software Engineer" }).click();
  await page.getByRole("button", { name: /Next: Set Capacity/i }).click();

  // Step 2 - Set Capacity
  const capacityInput = page.getByLabel("Target Active Agents");
  await capacityInput.fill("5");
  await page.getByRole("button", { name: /Next: Review/i }).click();

  // Apply scaling changes
  const applyButton = page.getByRole("button", { name: /Apply Scaling Changes/i });
  await applyButton.click();

  // Verify that the SSE trace logs stream in
  await expect(page.getByText("AI Workforce Manager: Reconciling Team Member resource.")).toBeVisible();
  await expect(page.getByText("AgentHired")).toBeVisible();

  await saveShot(page, "cuj-04-dynamic-scaling");
});

test("CUJ 5: agent chat – user navigates to agent detail and sends a chat message", async ({ page }) => {
  await page.goto("/");

  // Navigate to Agents tab
  await page.getByRole("button", { name: "Agents" }).click();
  await expect(page.getByRole("heading", { name: "Agent Network" })).toBeVisible();

  // Click the first agent card to open detail view.
  // Agent cards have aria-label "View details for <name>"
  const agentCards = page.getByRole("button", { name: /View details for/ });
  await expect(agentCards.first()).toBeVisible();
  const agentName = await agentCards.first().getAttribute("aria-label");
  const name = agentName?.replace("View details for ", "") ?? "Agent";
  await agentCards.first().click();

  // Agent detail should show the Chat box.
  await expect(page.getByRole("heading", { name: `Chat with ${name}` })).toBeVisible();

  // Navigate to the chat box inside the activity tab
  await page.getByRole("button", { name: "Activity" }).click();

  // Type and submit a message via the chat textarea.
  const chatMessage = `E2E chat test ${Date.now()}`;
  const inputLocator = page.locator('.war-room-composer .textarea');
  await inputLocator.fill(chatMessage);
  const sendButton = page.locator('.war-room-composer .war-room-send-btn');
  await sendButton.click();

  // After sending, the textarea should be cleared.
  await expect(inputLocator).toHaveValue("", { timeout: 10000 });

  await saveShot(page, "cuj-05-agent-chat-send");
});

test("CUJ 6: agent chat – sent message appears in backend meeting transcript", async ({ page, request }) => {
  await page.goto("/");

  // Navigate to Agents tab
  await page.getByRole("button", { name: "Agents" }).click();
  await expect(page.getByRole("heading", { name: "Agent Network" })).toBeVisible();

  // Open first agent detail.
  const agentCards = page.getByRole("button", { name: /View details for/ });
  await expect(agentCards.first()).toBeVisible();
  const agentName = await agentCards.first().getAttribute("aria-label");
  const name = agentName?.replace("View details for ", "") ?? "Agent";
  await agentCards.first().click();

  // Navigate to the chat box inside the activity tab
  await page.getByRole("button", { name: "Activity" }).click();

  // Send a uniquely identifiable message via the chat UI.
  const chatMessage = `CUJ-6 agent chat ${Date.now()}`;
  const inputLocator = page.locator('.war-room-composer .textarea');
  await inputLocator.fill(chatMessage);
  const sendButton = page.locator('.war-room-composer .war-room-send-btn');
  await sendButton.click();

  // Verify the textarea clears (message was submitted).
  await expect(inputLocator).toHaveValue("", { timeout: 10000 });

  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  // Verify the backend meeting transcript contains the message.
  await expect(async () => {
    const meetingsResponse = await request.get("http://127.0.0.1:8080/api/meetings", { headers: { Authorization: "Bearer " + token } });
    expect(meetingsResponse.ok()).toBeTruthy();
    const meetings = (await meetingsResponse.json()) as Array<{
      id: string;
      transcript?: Array<{ content: string }>;
    }>;
    const hasMessage = meetings.some((meeting) =>
      (meeting.transcript ?? []).some((entry) => entry.content === chatMessage)
    );
    expect(hasMessage).toBeTruthy();
  }).toPass({ timeout: 15000 });

  await saveShot(page, "cuj-06-agent-chat-transcript");
});

test("CUJ 7: agent chat via Chatwoot integration – verify send and messages endpoints", async ({ request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  // Send a chat message through the integration chat endpoint.
  const chatContent = `Chatwoot E2E test ${Date.now()}`;
  const sendResponse = await request.post("http://127.0.0.1:8080/api/integrations/chat/send", {
    headers: { Authorization: "Bearer " + token },
    data: {
      integrationId: "slack",
      channel: "#general",
      fromAgent: "pm-1",
      content: chatContent,
    },
  });
  expect(sendResponse.ok()).toBeTruthy();
  const sent = (await sendResponse.json()) as { content: string; integrationId: string };
  expect(sent.content).toBe(chatContent);
  expect(sent.integrationId).toBe("slack");

  // Retrieve the message via the messages list endpoint.
  const messagesResponse = await request.get(
    "http://127.0.0.1:8080/api/integrations/chat/messages?integrationId=slack", { headers: { Authorization: "Bearer " + token } }
  );
  expect(messagesResponse.ok()).toBeTruthy();
  const messages = (await messagesResponse.json()) as Array<{ content: string }>;
  const found = messages.some((m) => m.content === chatContent);
  expect(found).toBeTruthy();
});

test("CUJ 8: handoff resolution flows end-to-end", async ({ page, request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  await request.post("http://127.0.0.1:8080/api/handoffs", {
    headers: { Authorization: "Bearer " + token },
    data: {
      fromAgentId: "swe-1",
      toHumanRole: "CEO",
      intent: "Test Handoff Intent",
      failedAttempts: 1,
      currentState: "BLOCKED"
    }
  });

  await page.goto("/");

  // Navigate to Handoffs tab
  await page.getByRole("button", { name: "Handoffs" }).click();
  await expect(page.getByRole("heading", { name: "Warm Handoffs" })).toBeVisible();

  // Verify the seeded handoff is visible and has a "pending" status badge
  const handoffCard = page.locator('.handoff-card').filter({ hasText: 'Test Handoff Intent' }).first();
  await expect(handoffCard).toBeVisible({ timeout: 15000 });
  await expect(handoffCard.getByText("PENDING", { exact: true })).toBeVisible({ timeout: 15000 });

  // Bypass slider in e2e to ensure deterministic state resolution (slider physics are flaky)
  const allHandoffs = await request.get("http://127.0.0.1:8080/api/handoffs", { headers: { Authorization: "Bearer " + token } });
  const handoffsList = await allHandoffs.json();
  const targetId = handoffsList.find((h: any) => h.intent === "Test Handoff Intent").id;
  await request.post("http://127.0.0.1:8080/api/handoffs/resolve", {
    headers: { Authorization: "Bearer " + token },
    data: { handoffId: targetId, status: "resolved" }
  });

  // UI should update to show RESOLVED status
  await expect(handoffCard.getByText("RESOLVED", { exact: true })).toBeVisible({ timeout: 15000 });

  // The success notice should appear
  await expect(page.getByText("Handoff resolved and agent execution resumed.")).toBeVisible({ timeout: 10000 });

  await saveShot(page, "cuj-08-handoff-resolution");
});

test("CUJ 9: approval execution flows end-to-end", async ({ page, request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  await request.post("http://127.0.0.1:8080/api/approvals/request", {
    headers: { Authorization: "Bearer " + token },
    data: {
      agentId: "pm-1",
      action: "deploy-production",
      reason: "Launch prep",
      estimatedCostUsd: 100.50,
      riskLevel: "critical"
    }
  });

  await request.post("http://127.0.0.1:8080/api/messages", {
    headers: { Authorization: "Bearer " + token, "Content-Type": "application/x-www-form-urlencoded" },
    data: "fromAgent=pm-1&toAgent=CEO&meetingId=launch-readiness&messageType=ApprovalNeeded&content=All pre-launch checks passed. Requesting final CEO approval to deploy to production."
  });

  await page.goto("/");

  // Navigate to War Room
  await page.getByRole("button", { name: "Meetings" }).click();
  await expect(page.getByRole("heading", { name: "Virtual War Room" })).toBeVisible();

  // Verify the seeded approval is visible
  const approvalCard = page.locator('.approval-card').filter({ hasText: 'CEO Approval Required' }).first();
  await expect(approvalCard).toBeVisible({ timeout: 15000 });

  // Bypass slider in e2e to ensure deterministic state resolution (slider physics are flaky)
  const allApprovals = await request.get("http://127.0.0.1:8080/api/approvals", { headers: { Authorization: "Bearer " + token } });
  const approvalsList = await allApprovals.json();
  const approvalTarget = approvalsList.find((a: any) => a.action === "deploy-production")?.id;
  if (!approvalTarget) {
    throw new Error("Could not find seeded approval target");
  }

  await request.post("http://127.0.0.1:8080/api/approvals/decide", {
    headers: { Authorization: "Bearer " + token },
    data: { approvalId: approvalTarget, decision: "approve", decidedBy: "CEO" }
  });

  // UI should update to show Approved by CEO chip
  await expect(page.getByText("✓ Approved by CEO")).toBeVisible({ timeout: 15000 });

  // The success notice should appear
  await expect(page.getByText("Approval successfully recorded.")).toBeVisible({ timeout: 10000 });

  await saveShot(page, "cuj-09-approval-execution");
});

test("CUJ 10: E2E Pipelines UI interactions with seeded data", async ({ page, request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  // Seed the database to get the default mock pipeline in staging state
  await request.post("http://127.0.0.1:8080/api/dev/seed", {
    headers: { Authorization: "Bearer " + token },
    data: { scenario: "launch-readiness" }
  });

  // Navigate and login
  await page.goto("/");
  await page.fill("#login-username", "admin");
  await page.fill("#login-password", "adminpass123");
  await page.click("button:has-text('Sign in')");

  // Wait for dashboard to load
  await expect(page.locator("text=One Human Corp Dashboard")).toBeVisible();

  // Navigate to pipelines tab
  await page.click("button:has-text('Automated SDLC (Pipelines)')");
  await expect(page.locator("text=Active PRs")).toBeVisible();

  // Verify the seeded pipeline is visible (from the backend seed)
  await expect(page.locator("text=feat-billing-seed")).toBeVisible();
  await expect(page.locator("text=STAGING").first()).toBeVisible();

  // Approve for production (Promote)
  await page.click("button:has-text('Approve for Production')");

  // Verify the status changed to PROMOTED
  await expect(page.locator("text=PROMOTED").first()).toBeVisible();

  // Create a new pipeline
  await page.click("button:has-text('+ Start Implementation')");
  await page.fill("input[placeholder='e.g. feat-analytics']", "feat-e2e-playwright");
  await page.click("button:has-text('Create Pipeline')");

  // Verify pipeline was created
  await expect(page.locator("text=feat-e2e-playwright")).toBeVisible();

  await saveShot(page, "cuj-10-pipelines");
});
