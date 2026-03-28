import { chromium, devices } from "@playwright/test";
import { mkdir } from "node:fs/promises";
import path from "node:path";

const baseUrl = process.env.PLAYWRIGHT_BASE_URL;
const outputRoot = process.env.APP_SCREENSHOT_OUTPUT_DIR;

if (!baseUrl) {
  throw new Error("PLAYWRIGHT_BASE_URL is required.");
}

if (!outputRoot) {
  throw new Error("APP_SCREENSHOT_OUTPUT_DIR is required.");
}

const desktopContext = (userAgent) => ({
  viewport: { width: 1512, height: 982 },
  deviceScaleFactor: 1,
  isMobile: false,
  hasTouch: false,
  colorScheme: "light",
  userAgent,
});

const profiles = [
  {
    name: "web",
    context: desktopContext(
      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
    ),
  },
  {
    name: "linux",
    context: desktopContext(
      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
    ),
  },
  {
    name: "windows",
    context: desktopContext(
      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
    ),
  },
  {
    name: "macos",
    context: desktopContext(
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
    ),
  },
  {
    name: "android",
    context: {
      ...devices["Pixel 7"],
      colorScheme: "light",
    },
  },
  {
    name: "ios",
    context: {
      ...devices["iPhone 14"],
      colorScheme: "light",
    },
  },
];

const browser = await chromium.launch({ headless: true });

try {
  for (const profile of profiles) {
    const context = await browser.newContext(profile.context);
    const page = await context.newPage();

    await page.goto(baseUrl, { waitUntil: "networkidle" });
    await page.waitForFunction(
      () =>
        Boolean(
          document.querySelector("flutter-view") ||
            document.querySelector("flt-glass-pane") ||
            document.querySelector("canvas"),
        ),
      { timeout: 60000 },
    );
    await page.waitForTimeout(1500);

    const targetDir = path.join(outputRoot, profile.name);
    await mkdir(targetDir, { recursive: true });
    await page.screenshot({
      path: path.join(targetDir, "login.png"),
      fullPage: true,
    });

    await context.close();
  }
} finally {
  await browser.close();
}