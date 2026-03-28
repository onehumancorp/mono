import { test, expect } from "@playwright/test";

test.describe("Proactive Insights Widget", () => {
  test("dynamically surfaces insights to the CEO on the Overview tab", async ({ page, request }) => {
    // Wait for the backend to be fully initialized to prevent flakiness
    await expect.poll(async () => {
      const res = await request.get("/api/health");
      return res.status();
    }, { timeout: 15000 }).toBe(404); // /api/health doesn't exist, use /healthz

    await expect.poll(async () => {
      const res = await request.get("/healthz");
      return res.status();
    }, { timeout: 15000 }).toBe(200);

    // Navigate to the dashboard
    await page.goto("/");

    // Assuming we need to log in first.
    // The test might already be logged in or we might need to perform login.
    // Let's check if the login form is present
    const loginTitle = page.locator(".login-title");
    if (await loginTitle.isVisible()) {
      await page.fill("#login-username", "admin");
      await page.fill("#login-password", "admin");
      await page.click(".login-btn");
    }

    // Wait for main dashboard to load
    await expect(page.locator(".page-title")).toBeVisible();

    // The overview tab should be selected
    await expect(page.locator(".nav-item.active").filter({ hasText: "Overview" })).toBeVisible();

    // Proactive Insights might not be visible if there are no insights yet.
    // Let's seed the scenario to ensure we have insights.
    const token = await page.evaluate(() => localStorage.getItem("ohc_token"));

    // Switch to settings to trigger seed if necessary, or just call API
    const response = await request.post("/api/dev/seed", {
      data: { scenario: "launch-readiness" },
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ""
      }
    });

    // We expect the seed to succeed
    expect(response.ok()).toBeTruthy();

    // Wait a bit, then refresh the page to fetch the new insights.
    await page.reload();

    // Verify the Proactive Insights panel is visible
    const panel = page.locator(".proactive-insights-panel");
    await expect(panel).toBeVisible({ timeout: 10000 });

    // Verify it contains insight cards
    const cards = panel.locator(".insight-card");
    expect(await cards.count()).toBeGreaterThan(0);

    // Verify specific content in the insight card
    const firstCardText = await cards.first().textContent();
    expect(firstCardText).toContain("pending handoff(s) requiring human intervention");
  });
});
