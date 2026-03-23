import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test.beforeEach(async ({ request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } });
  expect(response.ok()).toBeTruthy();
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

  const meetingsResponse = await request.get("http://127.0.0.1:8080/api/meetings");
  expect(meetingsResponse.ok()).toBeTruthy();
  const meetings = (await meetingsResponse.json()) as Array<{ id: string; transcript?: Array<{ content: string }> }>;
  const hasMessage = meetings.some((meeting) =>
    (meeting.transcript ?? []).some((entry) => entry.content === message)
  );
  expect(hasMessage).toBeTruthy();

  await saveShot(page, "cuj-02-frontend-send-message");
});

test("CUJ 3: backend /app route remains reachable for bundled frontend", async ({ page }) => {
  await page.goto("http://127.0.0.1:8080/app");
  await expect(page.getByRole("heading", { name: "React Frontend Route" })).toBeVisible();

  await saveShot(page, "cuj-03-backend-app-route");
});

test("CUJ 4: Dynamic Scaling triggers SSE trace logs", async ({ page }) => {
  await page.goto("/");

  // Navigate to Dynamic Scaling tab
  await page.getByRole("button", { name: "Dynamic Scaling" }).click();
  await expect(page.getByRole("heading", { name: "Dynamic Scaling" })).toBeVisible();

  // Apply scaling changes
  const applyButton = page.getByRole("button", { name: /Apply Scaling Changes/i });
  await applyButton.click();

  // Verify that the SSE trace logs stream in
  await expect(page.getByText("K8s Operator: Reconciling TeamMember resource.")).toBeVisible();
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

  // Type and submit a message via the chat textarea.
  const chatMessage = `E2E chat test ${Date.now()}`;
  await page.getByPlaceholder(`Send a message to ${name}…`).fill(chatMessage);
  await page.getByRole("button", { name: "Send" }).click();

  // After sending, the textarea should be cleared.
  await expect(page.getByPlaceholder(`Send a message to ${name}…`)).toHaveValue("");

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

  // Send a uniquely identifiable message via the chat UI.
  const chatMessage = `CUJ-6 agent chat ${Date.now()}`;
  await page.getByPlaceholder(`Send a message to ${name}…`).fill(chatMessage);
  await page.getByRole("button", { name: "Send" }).click();

  // Verify the textarea clears (message was submitted).
  await expect(page.getByPlaceholder(`Send a message to ${name}…`)).toHaveValue("");

  // Verify the backend meeting transcript contains the message.
  const meetingsResponse = await request.get("http://127.0.0.1:8080/api/meetings");
  expect(meetingsResponse.ok()).toBeTruthy();
  const meetings = (await meetingsResponse.json()) as Array<{
    id: string;
    transcript?: Array<{ content: string }>;
  }>;
  const hasMessage = meetings.some((meeting) =>
    (meeting.transcript ?? []).some((entry) => entry.content === chatMessage)
  );
  expect(hasMessage).toBeTruthy();

  await saveShot(page, "cuj-06-agent-chat-transcript");
});

test("CUJ 7: agent chat via Chatwoot integration – verify send and messages endpoints", async ({ request }) => {
  // This test verifies the Chatwoot-backed chat API endpoints work end-to-end.
  // No direct DB seeding – all data is provided via API interaction.

  // Send a chat message through the integration chat endpoint.
  const chatContent = `Chatwoot E2E test ${Date.now()}`;
  const sendResponse = await request.post("http://127.0.0.1:8080/api/integrations/chat/send", {
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
    "http://127.0.0.1:8080/api/integrations/chat/messages?integrationId=slack"
  );
  expect(messagesResponse.ok()).toBeTruthy();
  const messages = (await messagesResponse.json()) as Array<{ content: string }>;
  const found = messages.some((m) => m.content === chatContent);
  expect(found).toBeTruthy();
});
