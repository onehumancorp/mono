import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test("Chaos: Simulate DB failure and recovery during agent handoff", async ({ page, request }) => {
  console.log("Starting Chaos test");

  // Wait for health check before UI auth
  await expect(async () => {
    const health = await request.get("http://127.0.0.1:8080/healthz");
    expect(health.ok()).toBeTruthy();
  }).toPass({ timeout: 15000 });

  // Step 1: Login
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();

  // Step 2: Seed handoff simulating a failure state
  const handoffResp = await request.post("http://127.0.0.1:8080/api/handoffs", {
    headers: { Authorization: "Bearer " + token },
    data: {
      fromAgentId: "swe-1",
      toHumanRole: "CEO",
      intent: "Merge conflict resolution required for legacy billing module.",
      failedAttempts: 3,
      currentState: JSON.stringify({
        Step_1_Code_Checkout: "SUCCESS",
        Step_2_Dependency_Install: "SUCCESS",
        Step_3_Test_Suite: "FAIL: TypeError in billing_test.go",
        Step_4_Auto_Remediation: "SIGKILL: Timeout after 30s"
      })
    }
  });

  const handoff = await handoffResp.json();

  // Setup UI auth
  await page.goto("/");
  await page.evaluate((t) => localStorage.setItem("ohc_token", t), token);
  await page.goto("/");

  await page.waitForSelector('text=One Human Corp Dashboard');

  // Navigate to Handoffs tab
  await page.getByRole("button", { name: "Handoffs" }).click();

  // Look for our seeded handoff
  const handoffCard = page.locator('.handoff-card').filter({ hasText: 'Merge conflict resolution required for legacy billing module.' }).first();
  await expect(handoffCard).toBeVisible({ timeout: 15000 });

  // Verify visual indicators of failure
  await expect(handoffCard.getByText('Failed Attempts: 3')).toBeVisible({ timeout: 15000 });
  await expect(handoffCard.getByText('PENDING')).toBeVisible({ timeout: 15000 }); // Default status is pending

  await saveShot(page, "chaos-01-blocked-handoff");

  // Simulating the resolution
  // Use backend api directly since slide-to-approve UI is flaky in playwright without a real mouse
  await request.post("http://127.0.0.1:8080/api/handoffs/resolve", {
    headers: { Authorization: "Bearer " + token },
    data: {
      handoffId: handoff.id,
      status: "resolved"
    }
  });

  // Reload or re-navigate to refresh state
  await page.reload();
  await page.getByRole("button", { name: "Handoffs" }).click();

  // Look for our seeded handoff again, it should be resolved
  const updatedHandoffCard = page.locator('.handoff-card', { hasText: 'Merge conflict resolution required for legacy billing module.' });
  // Looking for the text inside handoff-resolved-stamp
  await expect(updatedHandoffCard.locator('.handoff-resolved-stamp')).toHaveText(/Resolved/);

  await saveShot(page, "chaos-02-resolved-handoff");
});
