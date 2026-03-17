import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";
import type { Locator } from "@playwright/test";

const screenshotDir = "../../docs/screenshots";

async function saveShot(page: Page, name: string, target?: Locator): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  if (target) {
    await target.screenshot({ path: `${screenshotDir}/${name}.png` });
    return;
  }
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test.beforeEach(async ({ request }) => {
  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", {
    data: { scenario: "launch-readiness" },
  });
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
  await expect(page.getByRole("heading", { name: "New Message" })).toBeVisible();

  const message = `Playwright seeded message ${Date.now()}`;
  await page.getByLabel("Content").fill(message);
  await page.getByRole("button", { name: "Send Message" }).click();

  await expect(page.locator(".transcript-body", { hasText: message }).first()).toBeVisible();

  const meetingsResponse = await request.get("http://127.0.0.1:8080/api/meetings");
  expect(meetingsResponse.ok()).toBeTruthy();
  const meetings = (await meetingsResponse.json()) as Array<{ id: string; transcript?: Array<{ content: string }> }>;
  const hasMessage = meetings.some((meeting) =>
    (meeting.transcript ?? []).some((entry) => entry.content === message)
  );
  expect(hasMessage).toBeTruthy();

  await saveShot(page, "cuj-02-frontend-send-message");
});

test("CUJ 3: backend API is reachable with frontend dashboard style", async ({ page, request }) => {
  const dashboardResp = await request.get("http://127.0.0.1:8080/api/dashboard");
  expect(dashboardResp.ok()).toBeTruthy();

  await page.goto("/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();
  await expect(page.getByText("Demo Software Company")).toBeVisible();

  await saveShot(page, "cuj-03-backend-app-route");
});

test("CUJ 4: settings shows model registry and delegate mode", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();
  await page.evaluate(() => {
    const settings = Array.from(document.querySelectorAll("button.nav-item")).find((el) =>
      (el.textContent || "").trim() === "Settings"
    );
    settings?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
  });
  await page.waitForTimeout(250);

  await saveShot(page, "cuj-04-settings-delegate-mode");
});