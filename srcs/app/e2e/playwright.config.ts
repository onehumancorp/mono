import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for Flutter web e2e tests.
 */
export default defineConfig({
  testDir: __dirname,
  testMatch: ['*.spec.ts'],
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  reporter: [
    ['list'],
    ['json', { outputFile: 'playwright-results.json' }],
    ['html', {
        outputFolder: 'playwright-report',
        open: 'never'
    }]
  ],
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
  webServer: undefined,
});
