import { expect, test, type Page } from "@playwright/test";
import { mkdir } from "node:fs/promises";

const screenshotDir = "tests/screenshots";

async function saveShot(page: Page, name: string): Promise<void> {
  await mkdir(screenshotDir, { recursive: true });
  await page.screenshot({ path: `${screenshotDir}/${name}.png`, fullPage: true });
}

test("CUJ 1: frontend dashboard loads and shows organization data", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();
  await expect(page.getByText("Demo Software Company")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Org Chart" })).toBeVisible();

  await saveShot(page, "cuj-01-frontend-dashboard");
});

test("CUJ 2: frontend send message flow updates active meetings", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Send Message" })).toBeVisible();

  const message = `Playwright frontend message ${Date.now()}`;
  await page.getByLabel("Content").fill(message);
  await page.getByRole("button", { name: "Send Message" }).click();

  await expect(page.getByText(message)).toBeVisible();
  await saveShot(page, "cuj-02-frontend-send-message");
});

test("CUJ 3: backend dashboard form posts and transcript updates", async ({ page }) => {
  await page.goto("http://127.0.0.1:8080/");
  await expect(page.getByRole("heading", { name: "One Human Corp Dashboard" })).toBeVisible();

  const message = `Playwright backend message ${Date.now()}`;
  await page.getByLabel("Content").fill(message);
  await page.getByRole("button", { name: "Send Message" }).click();

  await expect(page.getByText(message)).toBeVisible();
  await saveShot(page, "cuj-03-backend-send-message");
});

test("CUJ 4: backend app route is reachable", async ({ page }) => {
  await page.goto("http://127.0.0.1:8080/app");
  await expect(page.getByRole("heading", { name: "React Frontend Route" })).toBeVisible();

  await saveShot(page, "cuj-04-backend-app-route");
});
