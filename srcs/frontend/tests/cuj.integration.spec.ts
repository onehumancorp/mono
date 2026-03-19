import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
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