import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test.beforeEach(async ({ request, page, context }) => {
  // Perform login to ensure the app functions and doesn't get stuck on auth guard
  const loginResponse = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "admin" },
  });
  expect(loginResponse.ok()).toBeTruthy();
  const { token } = await loginResponse.json();
  await context.addInitScript((val) => {
    window.localStorage.setItem("ohc_token", val);
  }, token);

  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", {
    data: { scenario: "launch-readiness" },
    headers: { Authorization: `Bearer ${token}` }
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

  const token = await page.evaluate(() => window.localStorage.getItem("ohc_token"));

  await expect(async () => {
    // Wait until the dashboard API returns the meetings with the injected message
    const dashResponse = await request.get("http://127.0.0.1:8080/api/dashboard", {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
    expect(dashResponse.ok()).toBeTruthy();
    const dashData = await dashResponse.json();
    const meetings = dashData.meetings as Array<{ id: string; transcript?: Array<{ content: string }> }>;
    const hasMessage = meetings.some((meeting) =>
      (meeting.transcript ?? []).some((entry) => entry.content === message)
    );
    expect(hasMessage).toBeTruthy();
  }).toPass({ timeout: 10_000 });

  await saveShot(page, "cuj-02-frontend-send-message");
});

test("CUJ 3: backend / route remains reachable for bundled frontend", async ({ page }) => {
  await page.goto("http://127.0.0.1:8080/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();

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
