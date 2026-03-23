const fs = require('fs');

const content = `import { render, screen, waitFor, fireEvent, act, within } from "@testing-library/react";
import { expect, vi, describe, it, beforeAll, afterAll } from "vitest";
import { App } from "./App";
import { login } from "./api";

// Use real backend with seeded database
describe("App Full Integration", () => {
  beforeAll(async () => {
    // Wait for backend
    for (let i = 0; i < 20; i++) {
      try {
        const res = await fetch("http://127.0.0.1:8080/app");
        if (res.ok) break;
      } catch (e) {}
      await new Promise(r => setTimeout(r, 500));
    }

    // Login
    try {
      const res = await fetch("http://127.0.0.1:8080/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username: "admin", password: "adminpass123" })
      });
      const data = await res.json();
      localStorage.setItem("ohc_token", data.token);
    } catch(e) {}
  });

  afterAll(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const seed = async (scenario) => {
    await fetch("http://127.0.0.1:8080/api/dev/seed", {
      method: "POST",
      headers: { "Content-Type": "application/json", "Authorization": "Bearer " + localStorage.getItem("ohc_token") },
      body: JSON.stringify({ scenario })
    });
  };

  it("covers full happy path UI interactions", async () => {
    await seed("launch-readiness");

    await act(async () => {
      render(<App />);
    });

    // 1. Dashboard Tab
    await waitFor(() => {
      expect(screen.getByText(/Dashboard/i)).toBeInTheDocument();
    }, { timeout: 10000 });

    expect(screen.getByText(/Demo Software Company/i)).toBeInTheDocument();

    // Check initial content
    expect(screen.getByText(/Demo Software Company/i)).toBeInTheDocument();

    // Test tabs navigation
    fireEvent.click(screen.getByRole("button", { name: "Meetings" }));
    await screen.findByText(/Virtual War Room/i);

    // Test sending message
    const input = screen.getByPlaceholderText(/Inject direction or approve actions/i);
    fireEvent.change(input, { target: { value: "Hello from vitest" } });
    fireEvent.click(screen.getByRole("button", { name: "Send" }));

    // Test Integrations tab
    fireEvent.click(screen.getByRole("button", { name: "Integrations" }));
    await screen.findByText(/Connected Services/i);

    // Test Organization map
    fireEvent.click(screen.getByRole("button", { name: "Organization" }));
    await screen.findByText(/Hire Agent/i);

    // Test Users
    fireEvent.click(screen.getByRole("button", { name: "Users" }));
    await screen.findByText(/Add User/i);

    // Test Analytics
    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));
    await screen.findByText(/Cost/i);

    // Test Dynamic Scaling
    fireEvent.click(screen.getByRole("button", { name: "Dynamic Scaling" }));
    await screen.findByText(/Apply Scaling/i);

    // Test Settings
    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    await screen.findByText(/API Keys/i);
  });
});
`;
fs.writeFileSync("srcs/frontend/src/App.test.tsx", content);
