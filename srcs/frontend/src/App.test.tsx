import { render, screen, waitFor, act } from "@testing-library/react";
import { App } from "./App";
import { setStoredToken, clearStoredToken } from "./api";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

// Helper to poll for backend readiness
async function waitForBackend(url: string, timeoutMs: number = 10000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`${url}/healthz`);
      if (res.ok) return true;
    } catch (e) {}
    await new Promise(r => setTimeout(r, 500));
  }
  throw new Error(`Backend not ready at ${url}`);
}

describe("App Component against real backend", () => {
  beforeEach(async () => {
    clearStoredToken();
    vi.unstubAllGlobals();
    vi.clearAllMocks();

    const baseUrl = import.meta.env.VITE_BACKEND_URL || "";
    await waitForBackend(baseUrl);

    // Login as admin
    const loginResp = await fetch(baseUrl + "/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: "admin", password: "adminpass123" })
    });
    if (!loginResp.ok) throw new Error("Login failed");
    const data = await loginResp.json();
    setStoredToken(data.token);

    // Seed the DB
    const seedResp = await fetch(baseUrl + "/api/dev/seed", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${data.token}`
      },
      body: JSON.stringify({ scenario: "launch-readiness" })
    });
    if (!seedResp.ok) throw new Error("Seed failed");
  });

  afterEach(() => {
    clearStoredToken();
  });

  it("renders the dashboard after fetching real data", async () => {
    render(<App />);

    // Check if loading state resolves and main dashboard is visible.
    // The seeded scenario is "launch-readiness" (Demo Software Company)
    await waitFor(() => {
      expect(screen.getByText(/Demo Software Company/i)).toBeInTheDocument();
    }, { timeout: 5000 });
  });

  it("renders Agent Network tab with seeded agents", async () => {
    render(<App />);

    await waitFor(() => {
      expect(screen.getByText(/Demo Software Company/i)).toBeInTheDocument();
    }, { timeout: 5000 });

    // The "launch-readiness" seed creates a PM and SWE.
    const agentsBtn = screen.getByRole("button", { name: "Agents" });
    agentsBtn.click();

    await waitFor(() => {
      expect(screen.getByText(/pm-1/i)).toBeInTheDocument();
    }, { timeout: 5000 });
  });
});
