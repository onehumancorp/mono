import { expect, test } from "@playwright/test";

test.describe("Agent Hiring Flow", () => {
  test("Should allow human CEO to hire a new agent via wizard and see it in the roster", async ({ page }) => {
    await page.goto("/");
    await page.waitForTimeout(2000);

    // 1. Log in
    const signInHeading = page.locator("h1", { hasText: "Sign in to One Human Corp" });
    if (await signInHeading.isVisible()) {
      await page.fill('input[id="login-username"]', "admin");
      await page.fill('input[id="login-password"]', "admin");
      await page.click('button[type="submit"]');
      await page.waitForTimeout(2000);
    }

    await expect(page.locator("h1", { hasText: "One Human Corp Dashboard" })).toBeVisible();

    // 2. Load demo scenario
    await page.getByRole("button", { name: "Settings" }).click();
    await page.getByRole("combobox", { name: "Scenario" }).selectOption("hr-dashboard-demo");
    await page.getByRole("button", { name: "Load Scenario" }).click();

    // Wait for scenario to load (dismiss notification)
    await page.waitForTimeout(2000);
    const closeBtn = page.locator(".alert-close");
    if (await closeBtn.isVisible()) {
        await closeBtn.click();
    }

    // 3. Navigate to Agents tab
    await page.getByRole("button", { name: "Agents" }).click();
    await expect(page.locator("h2", { hasText: "Agent Network" })).toBeVisible();

    // 4. Click Hire
    await page.getByRole("button", { name: "+ Hire Agent" }).click();
    await expect(page.locator("h2", { hasText: "Hire New Agent" })).toBeVisible();

    // 5. Wizard Step 1 - Select Role
    const pmRole = page.locator("button.role-select-card").filter({ hasText: "PRODUCT MANAGER" });
    await pmRole.click();

    // 6. Wizard Step 2 - Configure
    await page.fill('input[placeholder="e.g. Senior Engineer 3"]', "E2E Test PM");
    await page.fill('input[placeholder="e.g. gpt-4o, claude-3.5-sonnet"]', "gpt-4-turbo");
    await page.getByRole("button", { name: "Next: Review →" }).click();

    // 7. Wizard Step 3 - Deploy
    await page.getByRole("button", { name: "Deploy Agent" }).click();
    await expect(page.locator("h2", { hasText: "Hire New Agent" })).not.toBeVisible();

    // 8. Verify the new agent is in the list
    await expect(page.locator("p.agent-card__name", { hasText: "E2E Test PM" })).toBeVisible();
    await expect(page.locator("p.agent-card__role", { hasText: "PRODUCT MANAGER" })).toBeVisible();
  });
});
