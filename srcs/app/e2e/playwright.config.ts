import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for Flutter web e2e tests.
 *
 * The Flutter web app is served by a local HTTP server started by the Bazel
 * sh_test wrapper (flutter_web_e2e_test.sh).  The base URL is passed via the
 * PLAYWRIGHT_BASE_URL environment variable; it defaults to localhost:8765 when
 * running outside Bazel.
 */
export default defineConfig({
  testDir: __dirname,
  testMatch: ['web.spec.ts'],
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list'], ['json', { outputFile: 'playwright-results.json' }]],
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:8765',
    screenshot: 'only-on-failure',
    video: 'off',
    actionTimeout: 15_000,
    navigationTimeout: 30_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // Do NOT start any web server here – it is started by the Bazel test wrapper.
  webServer: undefined,
});
