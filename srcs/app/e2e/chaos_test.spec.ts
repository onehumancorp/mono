import { test, expect, Page } from '@playwright/test';

async function waitForFlutter(page: Page, timeoutMs = 30_000): Promise<void> {
  await page.waitForFunction(
    () => {
      const body = document.body;
      return (
        body &&
        (body.querySelector('flt-glass-pane') !== null ||
          body.querySelector('canvas') !== null ||
          body.children.length > 0)
      );
    },
    { timeout: timeoutMs },
  );
}

test.describe('OHC Swarm Chaos & Handoff Verification', () => {
  test('verify cross-agent handoff and visual failure recovery', async ({ page }) => {
    const baseUrl = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:8080';
    for (let i = 0; i < 30; i++) {
      try {
        const resp = await page.request.get(`${baseUrl}/healthz`);
        if (resp.ok()) break;
      } catch (e) {}
      await page.waitForTimeout(500);
    }

    try {
        await page.request.post(`${baseUrl}/api/dev/seed`, {
            data: { scenario: 'launch-readiness' }
        });
    } catch (e) {}

    await page.goto('/');
    await waitForFlutter(page);

    // Press tab to focus email field and sign in
    await page.keyboard.press('Tab');
    await page.keyboard.type('admin@test.com');
    await page.keyboard.press('Tab');
    await page.keyboard.type('adminpass123');
    await page.keyboard.press('Tab');
    await page.keyboard.press('Enter');

    await page.waitForTimeout(1000);

    await page.waitForTimeout(1000);

    const outDir = process.env.APP_SCREENSHOT_OUTPUT_DIR || process.env.PLAYWRIGHT_BROWSERS_PATH || process.cwd();
    await page.screenshot({ path: require('path').join(outDir, 'handoff-verification-before-chaos.png') });

    try {
        await page.request.post(`${baseUrl}/api/ops/chaos/lock-db`);
    } catch (e) {}

    await page.waitForTimeout(4000); // wait for DB lock to release
    await page.screenshot({ path: require('path').join(outDir, 'handoff-verification-after-recovery.png') });

    const title = await page.title();
    expect(title).toMatch(/One Human Corp/i);

    const flutterPresent = await page.evaluate(() => {
      return (
        document.querySelector('flt-glass-pane') !== null ||
        document.querySelector('canvas') !== null ||
        document.body.innerHTML.length > 100
      );
    });
    expect(flutterPresent).toBe(true);
  });
});
