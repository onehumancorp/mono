import { test, expect } from "@playwright/test";

test("CUJ 10: handoff rejection conflict error handling", async ({ page, request }) => {
  // Login to get token
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  // Create a new handoff
  const createResp = await request.post("http://127.0.0.1:8080/api/handoffs", {
    headers: { Authorization: "Bearer " + token },
    data: {
      fromAgentId: "swe-1",
      toHumanRole: "CEO",
      intent: "Test Handoff Rejection Conflict",
      failedAttempts: 1,
      currentState: "BLOCKED"
    }
  });
  const newHandoff = await createResp.json();
  const targetId = newHandoff.id;

  await page.goto("/");

  // Navigate to Handoffs tab
  await page.getByRole("button", { name: "Handoffs" }).click();
  await expect(page.getByRole("heading", { name: "Warm Handoffs" })).toBeVisible();

  // Verify the seeded handoff is visible
  const handoffCard = page.locator('.handoff-card').filter({ hasText: 'Test Handoff Rejection Conflict' }).first();
  await expect(handoffCard).toBeVisible({ timeout: 15000 });

  // Introduce a conflict by resolving the handoff in the background
  await request.post("http://127.0.0.1:8080/api/handoffs/resolve", {
    headers: { Authorization: "Bearer " + token },
    data: { handoffId: targetId, status: "resolved" }
  });

  // Intercept the next request to delay it so we can check for "Rejecting..." text, but since we already resolved it in the backend, the next resolve attempt will fail with 409 Conflict.
  await page.route("**/api/handoffs/resolve", async (route) => {
    // wait for 500ms so we can assert the "Rejecting..." text
    await new Promise(r => setTimeout(r, 500));
    route.continue();
  });

  // Attempt to Reject the handoff in the UI
  const rejectBtn = handoffCard.getByRole("button", { name: "Reject" });
  await rejectBtn.click();

  // Check if "Rejecting..." is displayed
  await expect(rejectBtn).toHaveText("Rejecting...");

  // Wait for the conflict error to appear
  const errorMsg = page.locator('.field-error');
  await expect(errorMsg).toBeVisible();
  await expect(errorMsg).toContainText("Conflict: This handoff has already been addressed by another director or the state has changed.");

  // Remove the route to clean up
  await page.unroute("**/api/handoffs/resolve");
});
