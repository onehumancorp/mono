import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  webServer: [
    {
      command: process.env.PLAYWRIGHT_BACKEND_COMMAND ?? "go run ../cmd/ohc",
      port: 8080,
      reuseExistingServer: false,
      timeout: 120_000,
    },
    {
      command: "npm run dev -- --host 127.0.0.1 --port 8081",
      port: 8081,
      reuseExistingServer: false,
      timeout: 120_000,
    },
  ],
  use: {
    baseURL: "http://127.0.0.1:8081",
    headless: true,
    viewport: { width: 1440, height: 900 },
  },
  reporter: [["list"]],
});
