import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { App } from "./App";
import { login, seedScenario } from "./api";

describe("App Integration Tests against Real Backend", () => {
  beforeEach(async () => {
    localStorage.clear();
    // Use the real api, authenticate, and seed the scenario
    const loginResp = await login("admin", "adminpass123");
    localStorage.setItem("ohc_token", loginResp.token);
    await seedScenario("launch-readiness");
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("renders the dashboard successfully after loading", async () => {
    render(<App />);
    await waitFor(() => {
      expect(screen.getByText(/One Human Corp Dashboard/i)).toBeInTheDocument();
    }, { timeout: 15000 });
  });

  it("navigates through tabs", async () => {
    render(<App />);
    await waitFor(() => {
      expect(screen.getByText(/One Human Corp Dashboard/i)).toBeInTheDocument();
    }, { timeout: 15000 });

    const tabs = [
      { trigger: "Agents", wait: /Agent Network/i },
      { trigger: "Meetings", wait: /Virtual War Room/i },
      { trigger: "Handoffs", wait: /Warm Handoffs/i },
      { trigger: "Pipelines", wait: /Active PRs/i },
      { trigger: "Integrations", wait: /Connect your AI agents to external services/i },
      { trigger: "Playbooks", wait: /Role Playbooks/i },
      { trigger: "Cost", wait: /Cost Analytics/i },
      { trigger: "Users", wait: /User Management/i },
      { trigger: "Settings", wait: /Settings/i },
      { trigger: "Dynamic Scaling", wait: /Dynamic Scaling/i },
      { trigger: "Overview", wait: /Organization/i },
    ];

    for (const tab of tabs) {
      const el = screen.getAllByText(new RegExp("^" + tab.trigger + "$", "i")).find(e => e.closest("button"));
      if (el) {
        fireEvent.click(el.closest("button")!);
      } else {
        throw new Error(`Tab not found: ${tab.trigger}`);
      }
      try {
        await waitFor(() => {
          expect(screen.getAllByText(tab.wait).length).toBeGreaterThan(0);
        }, { timeout: 5000 });
      } catch (err) {
        throw new Error(`Failed to find wait text for tab ${tab.trigger}: ${tab.wait}`);
      }
    }
  }, 15000);

  it("interacts with agent hiring flow", async () => {
    render(<App />);
    await waitFor(() => {
      expect(screen.getByText(/One Human Corp Dashboard/i)).toBeInTheDocument();
    }, { timeout: 15000 });

    const el = screen.getByText(/^Agents$/i).closest("button");
    if (el) fireEvent.click(el);
    await waitFor(() => {
      expect(screen.getByText(/Agent Network/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("+ Hire Agent"));
    await waitFor(() => {
      expect(screen.getByText(/Hire New Agent/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Close" }).parentElement || screen.getByText("✕"));
  });

  it("renders the login screen when token is invalid", async () => {
    localStorage.removeItem("ohc_token");
    render(<App />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
    }, { timeout: 5000 });
  });
});
