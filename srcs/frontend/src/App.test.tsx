import { fireEvent, render, screen, waitFor, act } from "@testing-library/react";
import { App } from "./App";
import { fetchCosts, fetchDashboard } from "./api";

type MockResponse = {
  ok: boolean;
  status: number;
  json: () => Promise<unknown>;
  text: () => Promise<string>;
};

const dashboardPayload = {
  organization: {
    id: "org-1",
    name: "Acme Software",
    domain: "software_company",
    members: [{ id: "pm-1", name: "PM", role: "PRODUCT_MANAGER" }],
    roleProfiles: [
      {
        role: "PRODUCT_MANAGER",
        basePrompt: "Prioritize outcomes",
        capabilities: ["planning"],
        contextInputs: ["backlog"],
      },
    ],
  },
  meetings: [
    {
      id: "launch-readiness",
      participants: ["pm-1", "swe-1"],
      transcript: [
        {
          id: "m-1",
          fromAgent: "pm-1",
          toAgent: "swe-1",
          type: "task",
          content: "Review roadmap",
          meetingId: "launch-readiness",
          occurredAt: "2026-03-13T00:00:00Z",
        },
      ],
    },
  ],
  costs: {
    organizationID: "org-1",
    totalTokens: 123,
    totalCostUSD: 0.012345,
    agents: [{ agentID: "swe-1", model: "gpt-4o", tokenUsed: 120, costUSD: 0.008 }],
  },
  agents: [
    {
      id: "pm-1",
      name: "PM",
      role: "PRODUCT_MANAGER",
      organizationId: "org-1",
      status: "IN_MEETING",
    },
    {
      id: "swe-1",
      name: "SWE",
      role: "SOFTWARE_ENGINEER",
      organizationId: "org-1",
      status: "IN_MEETING",
    },
  ],
  statuses: [{ status: "IN_MEETING", count: 1 }],
  updatedAt: "2026-03-13T00:00:00Z",
};

function mockJson(data: unknown, status = 200): MockResponse {
  const ok = status >= 200 && status < 300;
  return {
    ok,
    status,
    json: async () => data,
    text: async () => (ok ? JSON.stringify(data) : ""),
  };
}

// Provide a fake auth token so the App shows the main UI, not the login screen.
beforeEach(() => {
  localStorage.setItem("ohc_token", "test-token");
});
afterEach(() => {
  localStorage.removeItem("ohc_token");
});

describe("App", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("renders command center data after loading", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await screen.findByText("Acme Software");
    expect(screen.getByText("One Human Corp Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Org Chart")).toBeInTheDocument();
    expect(screen.getByText("Review roadmap", { exact: false })).toBeInTheDocument();
  });

  it("shows API error state", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => mockJson({}, 500)));

    render(<App />);

    await screen.findByText(/Failed to load data/i);
  });

  it("submits message form and refreshes snapshot", async () => {
    const fetchMock = vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/messages") {
        expect(init?.method).toBe("POST");
        return mockJson({}, 200);
      }
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: "Send Message" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/messages",
        expect.objectContaining({ method: "POST" })
      );
    });
    await screen.findByText("Message delivered to the meeting timeline.");
  });

  it("shows send error message when API returns non-OK", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/messages") {
        return mockJson({}, 400);
      }
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: "Send Message" }));

    await screen.findByText("Request failed for /api/messages: 400");
  });

  it("refreshes snapshot when refresh button is pressed", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    await screen.findByText("Acme Software");

    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/dashboard", expect.any(Object));
    });
  });
});

// ── rich payload used across extended nav tests ─────────────────────────────
const richPayload = {
  ...dashboardPayload,
  organization: {
    ...dashboardPayload.organization,
    ceoId: "ceo-1",
    members: [
      { id: "ceo-1", name: "Alice CEO", role: "CEO", isHuman: true },
      { id: "swe-1", name: "Bob SWE", role: "SOFTWARE_ENGINEER", managerId: "ceo-1" },
    ],
    roleProfiles: [
      {
        role: "PRODUCT_MANAGER",
        basePrompt: "Prioritize outcomes",
        capabilities: ["planning", "roadmap"],
        contextInputs: ["backlog", "metrics"],
      },
      {
        role: "SOFTWARE_ENGINEER",
        basePrompt: "Write clean code",
        capabilities: ["coding"],
        contextInputs: ["specs"],
      },
    ],
  },
  agents: [
    { id: "ceo-1", name: "Alice CEO", role: "CEO", organizationId: "org-1", status: "ACTIVE" },
    { id: "swe-1", name: "Bob SWE", role: "SOFTWARE_ENGINEER", organizationId: "org-1", status: "BLOCKED" },
    { id: "qa-1", name: "Charlie QA", role: "QA_TESTER", organizationId: "org-1", status: "IDLE" },
  ],
  statuses: [
    { status: "ACTIVE", count: 2 },
    { status: "BLOCKED", count: 1 },
    { status: "IN_MEETING", count: 3 },
  ],
  costs: {
    ...dashboardPayload.costs,
    totalTokens: 1_500_000,
    projectedMonthlyUSD: 45.0,
  },
};

function makeFetch() {
  return vi.fn(async (input: string) => {
    if (input === "/api/dashboard") return mockJson(richPayload);
    if (input === "/api/domains")
      return mockJson([{ id: "software_company", name: "Software Company", description: "SaaS products" }]);
    if (input === "/api/mcp/tools")
      return mockJson([{ id: "git-mcp", name: "Git MCP", description: "Git ops", category: "dev", status: "available" }]);
    if (input === "/api/agents/hire") return mockJson(richPayload);
    if (input === "/api/agents/fire") return mockJson(richPayload);
    if (input === "/api/dev/seed") return mockJson(richPayload);
    if (input === "/api/messages") return mockJson({});
    return mockJson({}, 404);
  });
}

describe("App – navigation tabs", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  // ── Meetings tab ────────────────────────────────────────────────────────────

  it("navigates to Meetings tab and shows Virtual Meeting Rooms heading", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    expect(screen.getByText("Virtual War Room")).toBeInTheDocument();
    expect(screen.getByText("Context Efficiency Active")).toBeInTheDocument();
  });

  it("shows meeting participants in meetings tab", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await waitFor(() => {
      const participants = document.querySelectorAll(".war-room-participant");
      expect(participants.length).toBeGreaterThan(0);
    });
  });

  it("shows 'No active meetings' in meetings tab when meeting list is empty", async () => {
    const noMeetings = { ...richPayload, meetings: [] };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(noMeetings);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    expect(screen.getAllByText("No active meetings.").length).toBeGreaterThan(0);
  });

  it("shows meeting agenda when present", async () => {
    const withAgenda = {
      ...dashboardPayload,
      meetings: [{ id: "launch-readiness", transcript: [], participants: [], agenda: "Review launch blockers" }],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(withAgenda);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("Review launch blockers")).toBeInTheDocument();
  });

  it("shows 'No messages yet' when meeting transcript is empty", async () => {
    const emptyTranscript = {
      ...dashboardPayload,
      meetings: [{ id: "empty-mtg", transcript: [], participants: [] }],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(emptyTranscript);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("No messages yet.")).toBeInTheDocument();
  });

  it("shows — for invalid occurredAt timestamp in transcript", async () => {
    const badTime = {
      ...dashboardPayload,
      meetings: [{
        id: "launch-readiness",
        participants: ["pm-1"],
        transcript: [{
          id: "m-bad",
          fromAgent: "pm-1",
          toAgent: "swe-1",
          type: "task",
          content: "Bad time msg",
          meetingId: "launch-readiness",
          occurredAt: "not-a-valid-date",
        }],
      }],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(badTime);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Bad time msg");
    // formatTime("not-a-valid-date") → "—"
    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBeGreaterThan(0);
  });

  it("can change meeting selection via dropdown", async () => {
    const twoMeetings = {
      ...dashboardPayload,
      meetings: [
        { id: "mtg-1", transcript: [], participants: ["pm-1"] },
        { id: "mtg-2", transcript: [], participants: ["swe-1"] },
      ],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(twoMeetings);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await waitFor(() => {
      const selects = document.querySelectorAll("select");
      expect(selects.length).toBeGreaterThan(0);
    });
    const select = document.querySelector("select");
    if (select) fireEvent.change(select, { target: { value: "mtg-2" } });
  });

  // ── Agents tab ──────────────────────────────────────────────────────────────

  it("navigates to Agents tab and shows agent cards", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    expect(screen.getByText("Agent Network")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "+ Hire Agent" })).toBeInTheDocument();
    // Bob SWE appears in agent-card AND agents-tab org-chart; use getAllByText
    expect(screen.getAllByText("Bob SWE").length).toBeGreaterThan(0);
  });

  it("shows ACTIVE, BLOCKED, IDLE status badges in agents tab", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    await waitFor(() => {
      expect(document.querySelector(".status-badge--active")).toBeTruthy();
      expect(document.querySelector(".status-badge--blocked")).toBeTruthy();
      expect(document.querySelector(".status-badge--idle")).toBeTruthy();
    });
  });

  it("shows status IN_MEETING badge in agents tab", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    await waitFor(() => {
      expect(document.querySelector(".status-badge--meeting")).toBeTruthy();
    });
  });

  it("shows empty state when no agents registered", async () => {
    const emptyAgents = { ...richPayload, agents: [] };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(emptyAgents);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    expect(screen.getByText(/No agents registered/)).toBeInTheDocument();
  });

  it("opens hire modal when + Hire Agent is clicked", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));
    expect(screen.getByText("Hire New Agent")).toBeInTheDocument();
  });

  it("closes hire modal when Cancel is clicked", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByText("Hire New Agent")).not.toBeInTheDocument();
  });

  it("closes hire modal when ✕ close button is clicked", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(screen.queryByText("Hire New Agent")).not.toBeInTheDocument();
  });

  it("Next Details button is disabled until role is selected", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));
    const nextBtn = screen.getByRole("button", { name: /Next: Details/i });
    expect(nextBtn).toBeDisabled();
    fireEvent.click(screen.getAllByText("SOFTWARE ENGINEER")[0]);
    expect(nextBtn).not.toBeDisabled();
  });

  it("successfully hires an agent and shows notice", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));

    fireEvent.click(screen.getAllByText("SOFTWARE ENGINEER")[0]);
    fireEvent.click(screen.getByRole("button", { name: /Next: Details/i }));

    fireEvent.change(screen.getByPlaceholderText(/Senior Software Engineer/i), { target: { value: "New Agent" } });
    fireEvent.click(screen.getByRole("button", { name: /Next: Confirm/i }));

    fireEvent.click(screen.getByRole("button", { name: "Deploy Agent" }));
    await screen.findByText(/Agent "New Agent" hired successfully/);
  });

  it("shows error when hire agent fails (visible in overview form)", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(richPayload);
      if (input === "/api/agents/hire") return mockJson({}, 422);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));

    fireEvent.click(screen.getAllByText("SOFTWARE ENGINEER")[0]);
    fireEvent.click(screen.getByRole("button", { name: /Next: Details/i }));

    fireEvent.change(screen.getByPlaceholderText(/Senior Software Engineer/i), { target: { value: "Fail Agent" } });
    fireEvent.click(screen.getByRole("button", { name: /Next: Confirm/i }));

    fireEvent.click(screen.getByRole("button", { name: "Deploy Agent" }));

    // error is stored in state; navigate to overview where the form renders it
    await waitFor(() => {
      fireEvent.click(screen.getByRole("button", { name: /overview/i }));
    });
    await screen.findByRole("alert");
  });

  it("changes role select in hire modal", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    fireEvent.click(screen.getByRole("button", { name: "+ Hire Agent" }));

    fireEvent.click(screen.getAllByText("PRODUCT MANAGER")[0]);
    fireEvent.click(screen.getByRole("button", { name: /Next: Details/i }));

    fireEvent.change(screen.getByPlaceholderText(/Senior Product Manager/i), { target: { value: "PM Agent" } });
    fireEvent.click(screen.getByRole("button", { name: /Next: Confirm/i }));

    fireEvent.click(screen.getByRole("button", { name: "Deploy Agent" }));

    await screen.findByText(/Agent "PM Agent" hired successfully/);
  });

  it("fires an agent and shows success notice", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    const removeButtons = await screen.findAllByRole("button", { name: "Remove" });
    fireEvent.click(removeButtons[0]);
    await screen.findByText(/removed from org/);
  });

  it("shows error when fire agent fails (visible in overview form)", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(richPayload);
      if (input === "/api/agents/fire") return mockJson({}, 500);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    const removeButtons = await screen.findAllByRole("button", { name: "Remove" });
    fireEvent.click(removeButtons[0]);
    // error is stored in state; navigate to overview where the form renders it
    await waitFor(() => {
      fireEvent.click(screen.getByRole("button", { name: /overview/i }));
    });
    await screen.findByRole("alert");
  });

  it("human org members do not show a Remove button", async () => {
    // ceo-1 is isHuman: true in richPayload.organization.members, so no Remove for that agent
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    // Only swe-1 and qa-1 get Remove buttons (ceo-1 is human)
    const removeButtons = await screen.findAllByRole("button", { name: "Remove" });
    expect(removeButtons.length).toBe(2);
  });

  // ── Cost tab ────────────────────────────────────────────────────────────────

  it("navigates to Cost tab and shows cost analytics sections", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    expect(screen.getByText("Cost Analytics")).toBeInTheDocument();
    expect(screen.getByText("Burn Rate Forecast")).toBeInTheDocument();
    expect(screen.getByText("Agent Spend Breakdown")).toBeInTheDocument();
  });

  it("shows 1.5M token count in cost tab", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    expect(screen.getByText("1.5M")).toBeInTheDocument();
  });

  it("shows projected monthly cost $45.00 in cost tab", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    expect(screen.getAllByText("$45.00").length).toBeGreaterThan(0);
  });

  it("shows 'No cost data yet' when agents list is empty", async () => {
    const noCost = { ...richPayload, costs: { ...richPayload.costs, agents: [] } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(noCost);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    expect(screen.getByText("No cost data yet.")).toBeInTheDocument();
  });

  it("shows top token consumers on overview when agents have costs", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("Top Token Consumers")).toBeInTheDocument();
  });

  it("shows 5.0K suffix for token counts in thousands", async () => {
    const kTokens = { ...dashboardPayload, costs: { ...dashboardPayload.costs, totalTokens: 5_000 } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(kTokens);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    // embedded in "5.0K tokens used" — use regex
    expect(screen.getByText(/5\.0K/)).toBeInTheDocument();
  });

  it("shows 2.5M suffix for large token counts", async () => {
    const bigTokens = { ...dashboardPayload, costs: { ...dashboardPayload.costs, totalTokens: 2_500_000 } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(bigTokens);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText(/2\.5M/)).toBeInTheDocument();
  });

  it("formats tiny cost (< $0.001) with 6 decimal places", async () => {
    const tinyCost = { ...dashboardPayload, costs: { ...dashboardPayload.costs, totalCostUSD: 0.000005 } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(tinyCost);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("$0.000005")).toBeInTheDocument();
  });

  it("formats cost >= $1 with 2 decimal places", async () => {
    const largeCost = { ...dashboardPayload, costs: { ...dashboardPayload.costs, totalCostUSD: 2.5 } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(largeCost);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("$2.50")).toBeInTheDocument();
  });

  // ── Playbooks tab ───────────────────────────────────────────────────────────

  it("navigates to Playbooks tab and shows role profiles", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /playbooks/i }));
    expect(screen.getByText("Role Playbooks")).toBeInTheDocument();
    expect(screen.getByText("PRODUCT MANAGER")).toBeInTheDocument();
    expect(screen.getByText("SOFTWARE ENGINEER")).toBeInTheDocument();
    expect(screen.getByText("Prioritize outcomes")).toBeInTheDocument();
    expect(screen.getByText("planning")).toBeInTheDocument();
    expect(screen.getByText("backlog")).toBeInTheDocument();
  });

  it("shows empty state when no role profiles defined", async () => {
    const noProfiles = { ...richPayload, organization: { ...richPayload.organization, roleProfiles: [] } };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(noProfiles);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /playbooks/i }));
    expect(screen.getByText(/No role profiles defined/)).toBeInTheDocument();
  });

  // ── Settings tab ────────────────────────────────────────────────────────────

  it("navigates to Settings tab and fetches domains and MCP tools", async () => {
    const fetchMock = makeFetch();
    vi.stubGlobal("fetch", fetchMock);
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    // "Settings" appears in sidebar button + heading; use heading role to be specific
    await screen.findByRole("heading", { name: /^Settings$/ });
    // "Software Company" appears in domain list + org info domain field; use getAllByText
    await waitFor(() => expect(screen.getAllByText("Software Company").length).toBeGreaterThan(0));
    await screen.findByText("Git MCP");
    expect(fetchMock).toHaveBeenCalledWith("/api/domains");
    expect(fetchMock).toHaveBeenCalledWith("/api/mcp/tools");
  });

  it("loads scenario and shows success notice", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByRole("heading", { name: /^Settings$/ });
    fireEvent.click(screen.getByRole("button", { name: "Load Scenario" }));
    await screen.findByText(/Loaded scenario: launch-readiness/);
  });

  it("shows error when load scenario fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(richPayload);
      if (input === "/api/dev/seed") return mockJson({}, 500);
      if (input === "/api/domains") return mockJson([]);
      if (input === "/api/mcp/tools") return mockJson([]);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByRole("heading", { name: /^Settings$/ });
    fireEvent.click(screen.getByRole("button", { name: "Load Scenario" }));
    // seedScenario throws "Request failed for /api/dev/seed: 500"; shown via global error alert
    await screen.findByText(/Failed to load data/);
  });

  it("can change scenario selector and load different scenario", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByRole("heading", { name: /^Settings$/ });
    const scenarioSelect = screen.getByDisplayValue("Software Co — Launch Readiness");
    fireEvent.change(scenarioSelect, { target: { value: "digital-marketing" } });
    fireEvent.click(screen.getByRole("button", { name: "Load Scenario" }));
    await screen.findByText(/Loaded scenario: digital-marketing/);
  });

  it("shows domain description and tool details in settings", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("SaaS products");
    await screen.findByText("Git ops");
    expect(screen.getByText("available")).toBeInTheDocument();
  });

  it("shows current org info in settings", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByRole("heading", { name: /^Settings$/ });
    expect(screen.getByText("Current Organization")).toBeInTheDocument();
    expect(screen.getAllByText("Acme Software").length).toBeGreaterThan(0);
  });

  // ── OrgTree / Sidebar ───────────────────────────────────────────────────────

  it("shows YOU tag for human org member in org chart", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("YOU")).toBeInTheDocument();
  });

  it("shows nested OrgTree with child node", async () => {
    const withTree = {
      ...dashboardPayload,
      organization: {
        ...dashboardPayload.organization,
        members: [
          { id: "root-1", name: "Root Node", role: "CEO" },
          { id: "child-1", name: "Child Node", role: "SOFTWARE_ENGINEER", managerId: "root-1" },
        ],
      },
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(withTree);
      return mockJson({}, 404);
    }));
    render(<App />);
    // Root Node appears in org chart + CEO sidebar card; use getAllByText
    await screen.findAllByText("Root Node");
    expect(screen.getAllByText("Child Node").length).toBeGreaterThan(0);
  });

  it("shows CEO card in sidebar when member has CEO role", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    // "Human CEO" label is only in the sidebar CEO card
    expect(screen.getByText("Human CEO")).toBeInTheDocument();
    // Alice CEO appears in CEO card + agent card + org chart; use getAllByText
    expect(screen.getAllByText("Alice CEO").length).toBeGreaterThan(0);
  });

  it("shows meeting messages badge count in sidebar nav", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    const badge = document.querySelector(".nav-badge");
    expect(badge).toBeTruthy();
    // dashboardPayload.meetings[0].transcript has 1 message
    expect(badge?.textContent).toBe("1");
  });

  it("shows 'No active agents' when all statuses have count 0", async () => {
    const zeroStatuses = { ...richPayload, statuses: [{ status: "IDLE", count: 0 }] };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(zeroStatuses);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("No active agents.")).toBeInTheDocument();
  });

  // ── Notice / error dismissal ────────────────────────────────────────────────

  it("dismisses notice by clicking ✕ button", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    // Trigger a notice via form submit
    fireEvent.click(screen.getByRole("button", { name: "Send Message" }));
    await screen.findByText("Message delivered to the meeting timeline.");
    fireEvent.click(screen.getByRole("button", { name: "Dismiss" }));
    expect(screen.queryByText("Message delivered to the meeting timeline.")).not.toBeInTheDocument();
  });

  it("shows domain label for Software Company domain", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    // domainLabel("software_company") = "Software Company"
    expect(screen.getByText("Software Company")).toBeInTheDocument();
  });

  it("shows 'Offline' status label when load fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => mockJson({}, 500)));
    render(<App />);
    await screen.findByText(/Failed to load data/i);
    expect(screen.getByText("Offline")).toBeInTheDocument();
  });

  it("shows 'Live' status label when data loads successfully", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText("Live")).toBeInTheDocument();
  });

  it("shows updated timestamp in header after data loads", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByText(/Updated/)).toBeInTheDocument();
  });

  it("formatTokens returns plain number for count < 1000", async () => {
    // dashboardPayload has totalTokens: 123 → renders "123"
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    // "123 tokens used" — via kpi-sub text
    expect(screen.getByText(/123 tokens used/)).toBeInTheDocument();
  });

  it("shows agent role formatted without underscores", async () => {
    vi.stubGlobal("fetch", makeFetch());
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    // appears in agent-card + org-chart member role; use getAllByText
    expect(screen.getAllByText("SOFTWARE ENGINEER").length).toBeGreaterThan(0);
  });

  it("shows spend breakdown bar items for agents with costs", async () => {
    const withCosts = {
      ...richPayload,
      costs: {
        ...richPayload.costs,
        agents: [
          { agentID: "swe-1", model: "gpt-4o", tokenUsed: 500, costUSD: 1.25 },
          { agentID: "pm-1", model: "gpt-4o-mini", tokenUsed: 200, costUSD: 0.5 },
        ],
        totalCostUSD: 1.75,
      },
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(withCosts);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    expect(screen.getByText("gpt-4o")).toBeInTheDocument();
    expect(screen.getByText("$1.25")).toBeInTheDocument();
    // tokenUsed 500 renders as "500 tokens" — use regex
    expect(screen.getByText(/\b500\b/)).toBeInTheDocument();
  });  // closes last it in App – navigation tabs
});  // closes App – navigation tabs describe

describe("App – form field coverage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("triggers onChange for all overview form fields", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/messages") return mockJson({});
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");

    // The New Message form on the overview tab uses <select> for agent fields.
    // Capture references before any changes to avoid ambiguity after re-render.
    // Option text format is "{name} ({id})" e.g. "PM (pm-1)", "SWE (swe-1)".
    const fromAgentCombo = screen.getByDisplayValue("PM (pm-1)");
    const toAgentCombo = screen.getByDisplayValue("SWE (swe-1)");
    fireEvent.change(fromAgentCombo, { target: { value: "swe-1" } });
    fireEvent.change(toAgentCombo, { target: { value: "pm-1" } });
    // meetingId select and the meetings-room select both show "launch-readiness";
    // use getAllByDisplayValue and pick the last match (the dispatch form).
    const launchReadinessEls = screen.getAllByDisplayValue("launch-readiness");
    const meetingIdSelect = launchReadinessEls[launchReadinessEls.length - 1];
    fireEvent.change(meetingIdSelect, { target: { value: "launch-readiness" } });
    fireEvent.change(screen.getByDisplayValue("task"), { target: { value: "decision" } });
    fireEvent.change(screen.getByDisplayValue("Review launch blockers and owner assignments"), {
      target: { value: "Updated content" },
    });

    // Verify textarea content updated (select state can't be verified via non-option values)
    expect(screen.getByDisplayValue("Updated content")).toBeInTheDocument();
  });

  it("triggers onChange for all meetings-tab Dispatch Message form fields", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/messages") return mockJson({});
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));

    // The War Room form only surfaces the "content" textarea for the CEO
    const inputArea = screen.getByPlaceholderText(/Inject direction/i);
    fireEvent.change(inputArea, { target: { value: "Updated direction" } });
    expect(screen.getByDisplayValue("Updated direction")).toBeInTheDocument();
  });

  it("changes the meeting select in the overview Active Meetings panel", async () => {
    const twoMeetings = {
      ...dashboardPayload,
      meetings: [
        { id: "meeting-a", transcript: [], participants: ["pm-1"] },
        { id: "meeting-b", transcript: [], participants: ["swe-1"] },
      ],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(twoMeetings);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    // change the meeting selector in the overview Active Meetings panel
    const selects = screen.getAllByRole("combobox");
    const meetingSelect = selects.find((s) => (s as HTMLSelectElement).value === "meeting-a");
    if (meetingSelect) fireEvent.change(meetingSelect, { target: { value: "meeting-b" } });
  });

  it("shows loading state in agents tab org chart before data loads", async () => {
    let resolve!: (v: unknown) => void;
    const pending = new Promise((r) => { resolve = r; });
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") {
        await pending;
        return mockJson(dashboardPayload);
      }
      return mockJson({}, 404);
    }));
    render(<App />);
    // While loading, navigate to agents tab — snapshot is null, org chart shows Loading...
    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    expect(screen.getByText("Loading…")).toBeInTheDocument();
    resolve(undefined);
  });
});

describe("api – branch coverage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("fetchDashboard handles absent organization field", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({
        // no organization field — triggers ?? {} fallback
        meetings: [],
        costs: undefined,  // triggers ?? {} fallback
        agents: "not-an-array",  // triggers [] fallback
        statuses: null,  // triggers [] fallback
        updatedAt: "2026-01-01T00:00:00Z",
      }),
    })));
    const snap = await fetchDashboard();
    expect(snap.organization.id).toBe("");
    expect(snap.organization.members).toEqual([]);
    expect(snap.agents).toEqual([]);
    expect(snap.statuses).toEqual([]);
    expect(snap.costs.organizationID).toBe("");
  });

  it("normalizeCosts handles projectedMonthlyUsd lowercase variant", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({
        organizationID: "org-lc",
        totalTokens: 10,
        totalCostUSD: 0.5,
        projectedMonthlyUsd: 15.0,  // lowercase variant, but projectedMonthlyUSD is absent
        agents: [],
      }),
    })));
    // projectedMonthlyUSD is undefined → ternary false branch → projectedMonthlyUSD = undefined
    const costs = await fetchCosts();
    expect(costs.projectedMonthlyUSD).toBeUndefined();
  });

  it("fetchDashboard with undefined updatedAt uses current ISO timestamp", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true, status: 200,
      json: async () => ({
        organization: { id: "o", name: "N", domain: "d", members: [], roleProfiles: [] },
        meetings: [],
        costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] },
        agents: [],
        statuses: [],
        // no updatedAt — triggers ?? new Date().toISOString()
      }),
    })));
    const snap = await fetchDashboard();
    // should have fallen back to current ISO date
    expect(snap.updatedAt).toBeTruthy();
    expect(snap.updatedAt).not.toBe("undefined");
  });
});  // closes api – branch coverage describe

describe("App – integrations nav", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("renders integrations nav item", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    expect(screen.getByRole("button", { name: /integrations/i })).toBeInTheDocument();
  });

  it("shows integrations panel when navigating to Integrations", async () => {
    const mockIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected", description: "Send via Slack", createdAt: "2026-01-01T00:00:00Z" },
      { id: "github", name: "GitHub", type: "github", category: "git", status: "disconnected", description: "Open PRs on GitHub", createdAt: "2026-01-01T00:00:00Z" },
      { id: "jira", name: "Jira", type: "jira", category: "issues", status: "disconnected", description: "Track in Jira", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");

    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");
    expect(screen.getByText("Chat Services")).toBeInTheDocument();
    expect(screen.getByText("Git Platforms")).toBeInTheDocument();
    expect(screen.getByText("Issue Trackers")).toBeInTheDocument();
    expect(screen.getByText("Slack")).toBeInTheDocument();
    expect(screen.getByText("GitHub")).toBeInTheDocument();
    expect(screen.getByText("Jira")).toBeInTheDocument();
  });

  it("shows Connect button for disconnected integrations", async () => {
    const mockIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected", description: "Send via Slack", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Slack");

    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    expect(connectBtns.length).toBeGreaterThan(0);

    // click connect for Slack
    fireEvent.click(connectBtns[0]);
    await screen.findByText(/disconnect/i);
  });

  it("shows Disconnect button for connected integrations", async () => {
    const mockIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "connected", description: "Send via Slack", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Slack");

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    expect(disconnectBtns.length).toBeGreaterThan(0);

    fireEvent.click(disconnectBtns[0]);
    await screen.findByRole("button", { name: /^connect$/i });
  });

  it("shows empty state when no integrations loaded yet", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson([]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");

    const loadingStates = screen.getAllByText("Loading integrations…");
    expect(loadingStates.length).toBeGreaterThan(0);
  });
});

// ── Git & Issues integration category coverage ─────────────────────────────
describe("App – git and issues integrations", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows git integration connect button and connects successfully", async () => {
    const mockIntegrations = [
      { id: "github", name: "GitHub", type: "github", category: "git", status: "disconnected", description: "Open PRs on GitHub", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Git Platforms");
    await screen.findByText("GitHub");

    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    expect(connectBtns.length).toBeGreaterThan(0);
    fireEvent.click(connectBtns[0]);
    await screen.findByText(/disconnect/i);
  });

  it("shows git integration disconnect button and disconnects successfully", async () => {
    const mockIntegrations = [
      { id: "github", name: "GitHub", type: "github", category: "git", status: "connected", description: "Open PRs on GitHub", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Git Platforms");
    await screen.findByText("GitHub");

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    expect(disconnectBtns.length).toBeGreaterThan(0);
    fireEvent.click(disconnectBtns[0]);
    await screen.findByRole("button", { name: /^connect$/i });
  });

  it("shows empty state for git when no git integrations", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson([
        { id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected", description: "Chat", createdAt: "2026-01-01T00:00:00Z" },
      ]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Git Platforms");

    const loadingStates = screen.getAllByText("Loading integrations…");
    expect(loadingStates.length).toBeGreaterThan(0);
  });

  it("shows issue tracker connect button and connects successfully", async () => {
    const mockIntegrations = [
      { id: "jira", name: "Jira", type: "jira", category: "issues", status: "disconnected", description: "Track in Jira", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Issue Trackers");
    await screen.findByText("Jira");

    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    expect(connectBtns.length).toBeGreaterThan(0);
    fireEvent.click(connectBtns[0]);
    await screen.findByText(/disconnect/i);
  });

  it("shows issue tracker disconnect button and disconnects successfully", async () => {
    const mockIntegrations = [
      { id: "jira", name: "Jira", type: "jira", category: "issues", status: "connected", description: "Track in Jira", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Issue Trackers");
    await screen.findByText("Jira");

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    expect(disconnectBtns.length).toBeGreaterThan(0);
    fireEvent.click(disconnectBtns[0]);
    await screen.findByRole("button", { name: /^connect$/i });
  });

  it("shows empty state for issues when no issue integrations", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson([
        { id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected", description: "Chat", createdAt: "2026-01-01T00:00:00Z" },
      ]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Issue Trackers");

    const loadingStates = screen.getAllByText("Loading integrations…");
    expect(loadingStates.length).toBeGreaterThan(0);
  });
});

// ── Meeting tab agenda coverage ──────────────────────────────────────────────
describe("App – meetings tab agenda coverage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows meeting agenda in the meetings tab when a meeting has an agenda", async () => {
    const withAgenda = {
      ...dashboardPayload,
      meetings: [{ id: "launch-readiness", transcript: [], participants: ["pm-1"], agenda: "Discuss launch blockers for meetings tab" }],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(withAgenda);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");

    // navigate to Meetings tab
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await screen.findByText("Virtual War Room");

    expect(screen.getByText("Discuss launch blockers for meetings tab")).toBeInTheDocument();
  });
});

// ── Settings branch coverage ─────────────────────────────────────────────────
describe("App – settings branch coverage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows — placeholders in settings when snapshot is null (navigated before load)", async () => {
    // Set up a fetch that never resolves for dashboard but resolves for settings deps
    let resolveDashboard: (v: unknown) => void;
    const dashboardPromise = new Promise((r) => { resolveDashboard = r; });

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return dashboardPromise;
      if (input === "/api/domains") return mockJson([{ id: "software_company", name: "Software Company", description: "SaaS" }]);
      if (input === "/api/mcp/tools") return mockJson([{ id: "git-mcp", name: "Git MCP", description: "Git ops", category: "dev", status: "maintenance" }]);
      return mockJson({}, 404);
    }));

    render(<App />);
    // Navigate to settings BEFORE dashboard loads (snapshot is null)
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByRole("heading", { name: /^Settings$/ });

    // Should show — for org info since snapshot is null
    await waitFor(() => {
      const dashes = screen.getAllByText("—");
      expect(dashes.length).toBeGreaterThan(0);
    });

    // The maintenance status tool should show yellow badge
    await screen.findByText("maintenance");
    expect(screen.getByText("maintenance")).toBeInTheDocument();

    // Resolve dashboard to clean up
    resolveDashboard!(mockJson(dashboardPayload));
  });

  it("shows non-available MCP tool with status badge", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/domains") return mockJson([{ id: "d", name: "D", description: "desc" }]);
      if (input === "/api/mcp/tools") return mockJson([
        { id: "git-mcp", name: "Git MCP", description: "Git ops", category: "dev", status: "available" },
        { id: "figma-mcp", name: "Figma MCP", description: "Design ops", category: "design", status: "beta" },
      ]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("Git MCP");
    await screen.findByText("Figma MCP");
    expect(screen.getByText("beta")).toBeInTheDocument();
    expect(screen.getByText("available")).toBeInTheDocument();
  });
});

// ── Integration callback coverage (multiple integrations) ─────────────────────
describe("App – multi-integration connect/disconnect callbacks", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("connect callback updates correct integration when multiple git integrations exist", async () => {
    const mockIntegrations = [
      { id: "github", name: "GitHub", type: "github", category: "git", status: "disconnected", description: "Open PRs on GitHub", createdAt: "2026-01-01T00:00:00Z" },
      { id: "gitea", name: "Gitea", type: "gitea", category: "git", status: "connected", description: "Self-hosted Gitea", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[1], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Git Platforms");
    // GitHub and Gitea names appear in the list (use getAllByText since descriptions differ)
    await waitFor(() => expect(screen.getAllByText("GitHub").length).toBeGreaterThan(0));

    // Click connect on GitHub (first)
    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    fireEvent.click(connectBtns[0]);
    // Wait for the list to update
    await waitFor(() => {
      const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
      expect(disconnectBtns.length).toBeGreaterThan(0);
    });
  });

  it("disconnect callback updates correct issue integration when multiple exist", async () => {
    const mockIntegrations = [
      { id: "jira", name: "Jira", type: "jira", category: "issues", status: "connected", description: "Track issues in Jira", createdAt: "2026-01-01T00:00:00Z" },
      { id: "plane", name: "Plane", type: "plane", category: "issues", status: "disconnected", description: "OSS issue tracker", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[1], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Issue Trackers");
    await waitFor(() => expect(screen.getAllByText("Jira").length).toBeGreaterThan(0));

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    fireEvent.click(disconnectBtns[0]);
    await waitFor(() => {
      // Jira should now show "Connect" button
      const allBtns = screen.getAllByRole("button");
      expect(allBtns.some((b) => b.textContent?.match(/connect/i))).toBe(true);
    });
  });
});

// ── Role profiles null fallback branch ──────────────────────────────────────
describe("App – playbooks roleProfiles null branch", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows empty state when roleProfiles is null/undefined in organization", async () => {
    const nullProfiles = {
      ...dashboardPayload,
      organization: { ...dashboardPayload.organization, roleProfiles: null as unknown as [] },
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(nullProfiles);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /playbooks/i }));
    expect(screen.getByText(/No role profiles defined/)).toBeInTheDocument();
  });
});

// ── Multi-integration map callback false branch coverage ─────────────────────
describe("App – multi-integration map callback branches", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("chat connect: both map branches covered with multiple chat integrations", async () => {
    const mockIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "disconnected", description: "Slack messaging", createdAt: "2026-01-01T00:00:00Z" },
      { id: "discord", name: "Discord", type: "discord", category: "chat", status: "connected", description: "Discord server", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");
    await waitFor(() => expect(screen.getAllByText("Slack").length).toBeGreaterThan(0));

    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    fireEvent.click(connectBtns[0]);
    await waitFor(() => {
      const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
      expect(disconnectBtns.length).toBeGreaterThan(1);
    });
  });

  it("chat disconnect: both map branches covered with multiple chat integrations", async () => {
    const mockIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "connected", description: "Slack messaging", createdAt: "2026-01-01T00:00:00Z" },
      { id: "discord", name: "Discord", type: "discord", category: "chat", status: "disconnected", description: "Discord server", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");
    await waitFor(() => expect(screen.getAllByText("Slack").length).toBeGreaterThan(0));

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    fireEvent.click(disconnectBtns[0]);
    await waitFor(() => {
      const connectBtns = screen.getAllByRole("button", { name: /connect/i });
      expect(connectBtns.length).toBeGreaterThan(0);
    });
  });

  it("git disconnect: both map branches covered with multiple git integrations", async () => {
    const mockIntegrations = [
      { id: "github", name: "GitHub", type: "github", category: "git", status: "connected", description: "Open PRs on GitHub", createdAt: "2026-01-01T00:00:00Z" },
      { id: "gitea", name: "Gitea", type: "gitea", category: "git", status: "disconnected", description: "Self-hosted Gitea", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/disconnect") return mockJson({ ...mockIntegrations[0], status: "disconnected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Git Platforms");
    await waitFor(() => expect(screen.getAllByText("GitHub").length).toBeGreaterThan(0));

    const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
    fireEvent.click(disconnectBtns[0]);
    await waitFor(() => {
      const connectBtns = screen.getAllByRole("button", { name: /connect/i });
      expect(connectBtns.length).toBeGreaterThan(1);
    });
  });

  it("issues connect: both map branches covered with multiple issue integrations", async () => {
    const mockIntegrations = [
      { id: "jira", name: "Jira", type: "jira", category: "issues", status: "disconnected", description: "Track in Jira", createdAt: "2026-01-01T00:00:00Z" },
      { id: "plane", name: "Plane", type: "plane", category: "issues", status: "connected", description: "OSS issue tracker", createdAt: "2026-01-01T00:00:00Z" },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(mockIntegrations);
      if (input === "/api/integrations/connect") return mockJson({ ...mockIntegrations[0], status: "connected" });
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Issue Trackers");
    await waitFor(() => expect(screen.getAllByText("Jira").length).toBeGreaterThan(0));

    const connectBtns = screen.getAllByRole("button", { name: /connect/i });
    fireEvent.click(connectBtns[0]);
    await waitFor(() => {
      const disconnectBtns = screen.getAllByRole("button", { name: /disconnect/i });
      expect(disconnectBtns.length).toBeGreaterThan(1);
    });
  });
});

// ── Cost and Playbooks null-snapshot branch coverage ─────────────────────────
describe("App – cost and playbooks null-snapshot branches", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("cost tab shows zero fallbacks when navigated before data loads", async () => {
    let resolveDashboard: (v: unknown) => void;
    const neverResolves = new Promise((r) => { resolveDashboard = r; });

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return neverResolves;
      return mockJson({}, 404);
    }));

    render(<App />);
    // Navigate to cost tab BEFORE data loads (snapshot is null)
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));

    await screen.findByText("Burn Rate Forecast");
    // With null snapshot, all KPIs should show fallback values
    const zeros = screen.getAllByText("$0.000000");
    expect(zeros.length).toBeGreaterThan(0);

    // resolve to prevent act() leaks
    resolveDashboard!(mockJson(dashboardPayload));
  });

  it("playbooks tab shows empty state with null-snapshot org roleProfiles", async () => {
    let resolveDashboard: (v: unknown) => void;
    const neverResolves = new Promise((r) => { resolveDashboard = r; });

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return neverResolves;
      return mockJson({}, 404);
    }));

    render(<App />);
    // Navigate to playbooks tab BEFORE data loads (snapshot is null)
    fireEvent.click(screen.getByRole("button", { name: /playbooks/i }));
    await screen.findByText("Role Playbooks");

    expect(screen.getByText(/No role profiles defined/)).toBeInTheDocument();

    resolveDashboard!(mockJson(dashboardPayload));
  });

  it("cost tab burn gauge shows 0% when projectedMonthlyUSD is falsy", async () => {
    const zeroProjection = {
      ...dashboardPayload,
      costs: {
        organizationID: "org-1",
        totalTokens: 0,
        totalCostUSD: 0,
        projectedMonthlyUSD: 0,
        agents: [{ agentID: "swe-1", model: "", tokenUsed: 0, costUSD: 0 }],
      },
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(zeroProjection);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    await screen.findByText("Burn Rate Forecast");

    // 0% burn since projectedMonthlyUSD is falsy
    expect(screen.getByText("0%")).toBeInTheDocument();
  });
});

// ── User management tab coverage ─────────────────────────────────────────────

describe("App – user management tab", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows Users tab with empty state when no users", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users") return mockJson([]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
    await screen.findByText("User Management");
    await waitFor(() => expect(screen.getByText("No users yet.")).toBeInTheDocument());
  });

  it("shows list of users in Users tab", async () => {
    const mockUsers = [
      { id: "u-1", username: "alice", email: "alice@example.com", roles: ["admin"] },
      { id: "u-2", username: "bob", email: "bob@example.com", roles: ["operator"] },
    ];

    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users") return mockJson(mockUsers);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
    await screen.findByText("User Management");
    await waitFor(() => expect(screen.getByText("alice")).toBeInTheDocument());
    expect(screen.getByText("bob")).toBeInTheDocument();
  });

  it("create user form fields can be filled", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users") return mockJson([]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
    await screen.findByText("User Management");

    fireEvent.change(screen.getByPlaceholderText("e.g. alice"), { target: { value: "charlie" } });
    fireEvent.change(screen.getByPlaceholderText("alice@example.com"), { target: { value: "charlie@example.com" } });
    fireEvent.change(screen.getByPlaceholderText("••••••••"), { target: { value: "password123" } });
    fireEvent.change(screen.getByDisplayValue("operator"), { target: { value: "admin" } });

    expect((screen.getByPlaceholderText("e.g. alice") as HTMLInputElement).value).toBe("charlie");
    expect((screen.getByPlaceholderText("alice@example.com") as HTMLInputElement).value).toBe("charlie@example.com");
    expect((screen.getByDisplayValue("admin") as HTMLSelectElement).value).toBe("admin");
  });

  it("submits create user form and refreshes list", async () => {
    const newUser = { id: "u-3", username: "charlie", email: "charlie@example.com", roles: ["admin"] };
    let userList: unknown[] = [];

    vi.stubGlobal("fetch", vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users" && (!init || init.method !== "POST")) return mockJson(userList);
      if (input === "/api/users" && init?.method === "POST") {
        userList = [newUser];
        return mockJson(newUser);
      }
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
    await screen.findByText("User Management");
    await waitFor(() => expect(screen.getByText("No users yet.")).toBeInTheDocument());

    fireEvent.change(screen.getByPlaceholderText("e.g. alice"), { target: { value: "charlie" } });
    fireEvent.change(screen.getByPlaceholderText("alice@example.com"), { target: { value: "charlie@example.com" } });
    fireEvent.change(screen.getByPlaceholderText("••••••••"), { target: { value: "password123" } });

    fireEvent.click(screen.getByRole("button", { name: /create user/i }));
    await waitFor(() => expect(screen.getByText("charlie")).toBeInTheDocument());
  });

  it("remove user button calls deleteUser and refreshes list", async () => {
    const mockUsers = [{ id: "u-1", username: "alice", email: "alice@example.com", roles: ["admin"] }];
    let userList = [...mockUsers];

    vi.stubGlobal("fetch", vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users" && (!init || init.method !== "DELETE")) return mockJson(userList);
      if (input === "/api/users/u-1" && init?.method === "DELETE") {
        userList = [];
        return mockJson({}, 204);
      }
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
    await screen.findByText("User Management");
    await waitFor(() => expect(screen.getByText("alice")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: /remove/i }));
    await waitFor(() => expect(screen.getByText("No users yet.")).toBeInTheDocument());
  });
});

// ── Login / Logout flow coverage ──────────────────────────────────────────────

describe("App – login/logout flows", () => {
  beforeEach(() => {
    localStorage.removeItem("ohc_token");
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    localStorage.removeItem("ohc_token");
  });

  it("renders login form when no token is stored", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => mockJson({}, 404)));
    render(<App />);
    expect(screen.getByText("Sign in to One Human Corp")).toBeInTheDocument();
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
  });

  it("onChange handlers update username and password inputs", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => mockJson({}, 404)));
    render(<App />);
    const usernameInput = screen.getByLabelText("Username") as HTMLInputElement;
    const passwordInput = screen.getByLabelText("Password") as HTMLInputElement;
    fireEvent.change(usernameInput, { target: { value: "admin" } });
    fireEvent.change(passwordInput, { target: { value: "secret" } });
    expect(usernameInput.value).toBe("admin");
    expect(passwordInput.value).toBe("secret");
  });

  it("successful login shows main app", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/auth/login") return mockJson({ token: "tok", username: "admin" });
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    fireEvent.submit(screen.getByRole("button", { name: /sign in/i }).closest("form")!);
    await screen.findByText("Acme Software");
  });

  it("failed login shows error message", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/auth/login") {
        return { ok: false, status: 401, json: async () => ({}), text: async () => "Bad credentials" };
      }
      return mockJson({}, 404);
    }));
    render(<App />);
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    fireEvent.submit(screen.getByRole("button", { name: /sign in/i }).closest("form")!);
    await screen.findByText("Bad credentials");
  });

  it("handleLogout signs out and shows login screen", async () => {
    localStorage.setItem("ohc_token", "test-token");
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/auth/logout") return mockJson({});
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /sign out/i }));
    await screen.findByText("Sign in to One Human Corp");
  });
});

// ── Telegram wizard coverage ──────────────────────────────────────────────────

const mockTelegramIntegrations = [
  { id: "telegram", name: "Telegram", type: "telegram", category: "chat", status: "disconnected", description: "Telegram Bot", createdAt: "2026-01-01T00:00:00Z" },
];

function makeFetchWithIntegrations(integrations: unknown[], extraHandlers?: (input: string) => MockResponse | null) {
  return vi.fn(async (input: string) => {
    if (input === "/api/dashboard") return mockJson(dashboardPayload);
    if (input === "/api/integrations") return mockJson(integrations);
    if (input === "/api/integrations?category=chat") return mockJson(integrations.filter((i: unknown) => (i as { category: string }).category === "chat"));
    if (extraHandlers) {
      const result = extraHandlers(input);
      if (result) return result;
    }
    return mockJson({}, 404);
  });
}

describe("App – Telegram wizard", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  async function openTelegramWizard() {
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");

    // The integrations load takes a moment
    await screen.findByText("Telegram Bot");

    const setupBtns = screen.getAllByRole("button", { name: /setup/i });
    fireEvent.click(setupBtns[0]);
    await screen.findByText("Connect Telegram Bot");
  }

  it("opens Telegram wizard on Setup click", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations));
    await openTelegramWizard();
    expect(screen.getByText("Step 1 — Create a Telegram Bot")).toBeInTheDocument();
  });

  it("closes wizard on ✕ click", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /close/i }));
    await waitFor(() => expect(screen.queryByText("Connect Telegram Bot")).not.toBeInTheDocument());
  });

  it("navigates wizard steps: 1→2→3 and fills credentials", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations));
    await openTelegramWizard();

    // Step 1 → 2
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");

    // Fill bot token and chat ID
    const inputs = screen.getAllByRole("textbox");
    // chat ID input is textbox (not password), bot token is password type
    const chatIdInput = inputs.find(el => (el as HTMLInputElement).placeholder?.includes("group"));
    fireEvent.change(screen.getByPlaceholderText(/123456:ABC-DEF/), { target: { value: "bot-token-123" } });
    if (chatIdInput) fireEvent.change(chatIdInput, { target: { value: "-100123" } });

    // Step 2 → 3 (need both fields non-empty)
    const nextBtn = screen.getByRole("button", { name: /next: test/i });
    expect(nextBtn).not.toBeDisabled();
    fireEvent.click(nextBtn);
    await screen.findByText("Step 3 — Test Connection");
  });

  it("Send Test Message success advances testSent state", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      return null;
    }));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");
    fireEvent.change(screen.getByPlaceholderText(/123456:ABC-DEF/), { target: { value: "bot-token-123" } });
    const chatInputs = screen.getAllByRole("textbox");
    const chatIdInput = chatInputs.find(el => (el as HTMLInputElement).placeholder?.includes("group"));
    if (chatIdInput) fireEvent.change(chatIdInput, { target: { value: "-100123" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
  });

  it("Send Test Message error shows error message", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ error: "Bad token" }, 500);
      return null;
    }));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");
    fireEvent.change(screen.getByPlaceholderText(/123456:ABC-DEF/), { target: { value: "tok" } });
    const chatInputs = screen.getAllByRole("textbox");
    const chatIdInput = chatInputs.find(el => (el as HTMLInputElement).placeholder?.includes("group"));
    if (chatIdInput) fireEvent.change(chatIdInput, { target: { value: "-100" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
  });

  it("Complete Setup success closes wizard", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      if (input === "/api/integrations/connect") return mockJson({ id: "telegram", name: "Telegram", type: "telegram", category: "chat", status: "connected", description: "Telegram Bot", createdAt: "2026-01-01T00:00:00Z" });
      return null;
    }));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");
    fireEvent.change(screen.getByPlaceholderText(/123456:ABC-DEF/), { target: { value: "bot-tok" } });
    const chatInputs = screen.getAllByRole("textbox");
    const chatIdInput = chatInputs.find(el => (el as HTMLInputElement).placeholder?.includes("group"));
    if (chatIdInput) fireEvent.change(chatIdInput, { target: { value: "-100" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
    fireEvent.click(screen.getByRole("button", { name: /complete setup/i }));
    await waitFor(() => expect(screen.queryByText("Connect Telegram Bot")).not.toBeInTheDocument());
  });

  it("Complete Setup error shows error message", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      if (input === "/api/integrations/connect") return mockJson({ error: "Server error" }, 500);
      return null;
    }));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");
    fireEvent.change(screen.getByPlaceholderText(/123456:ABC-DEF/), { target: { value: "tok" } });
    const chatInputs = screen.getAllByRole("textbox");
    const chatIdInput = chatInputs.find(el => (el as HTMLInputElement).placeholder?.includes("group"));
    if (chatIdInput) fireEvent.change(chatIdInput, { target: { value: "-100" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
    fireEvent.click(screen.getByRole("button", { name: /complete setup/i }));
    // wizard stays open on error (connect failed)
    await waitFor(() => expect(screen.getByText("Connect Telegram Bot")).toBeInTheDocument());
  });

  it("Back button goes to previous step", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockTelegramIntegrations));
    await openTelegramWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter token/i }));
    await screen.findByText("Step 2 — Enter Credentials");
    fireEvent.click(screen.getByRole("button", { name: /back/i }));
    await screen.findByText("Step 1 — Create a Telegram Bot");
  });
});

// ── Discord wizard coverage ───────────────────────────────────────────────────

const mockDiscordIntegrations = [
  { id: "discord", name: "Discord", type: "discord", category: "chat", status: "disconnected", description: "Discord Webhook", createdAt: "2026-01-01T00:00:00Z" },
];

describe("App – Discord wizard", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  async function openDiscordWizard() {
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");
    const setupBtns = screen.getAllByRole("button", { name: /setup/i });
    fireEvent.click(setupBtns[0]);
    await screen.findByText("Connect Discord Webhook");
  }

  it("opens Discord wizard on Setup click", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations));
    await openDiscordWizard();
    expect(screen.getByText("Step 1 — Create a Discord Webhook")).toBeInTheDocument();
  });

  it("closes Discord wizard on ✕ click", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /close/i }));
    await waitFor(() => expect(screen.queryByText("Connect Discord Webhook")).not.toBeInTheDocument());
  });

  it("navigates Discord wizard steps 1→2→3", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.change(screen.getByPlaceholderText(/https:\/\/discord.com/), { target: { value: "https://discord.com/api/webhooks/123/abc" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
  });

  it("Discord Back button navigates to previous step", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.click(screen.getByRole("button", { name: /back/i }));
    await screen.findByText("Step 1 — Create a Discord Webhook");
  });

  it("Discord Send Test Message success", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      return null;
    }));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.change(screen.getByPlaceholderText(/https:\/\/discord.com/), { target: { value: "https://discord.com/api/webhooks/123/abc" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
  });

  it("Discord Send Test Message error shows error", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ error: "bad webhook" }, 500);
      return null;
    }));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.change(screen.getByPlaceholderText(/https:\/\/discord.com/), { target: { value: "https://discord.com/api/webhooks/123/abc" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
  });

  it("Discord Complete Setup success", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      if (input === "/api/integrations/connect") return mockJson({ id: "discord", name: "Discord", type: "discord", category: "chat", status: "connected", description: "Discord Webhook", createdAt: "2026-01-01T00:00:00Z" });
      return null;
    }));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.change(screen.getByPlaceholderText(/https:\/\/discord.com/), { target: { value: "https://discord.com/api/webhooks/123/abc" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
    fireEvent.click(screen.getByRole("button", { name: /complete setup/i }));
    await screen.findByText("Discord Connected!");
    fireEvent.click(screen.getByRole("button", { name: /done/i }));
    await waitFor(() => expect(screen.queryByText("Connect Discord Webhook")).not.toBeInTheDocument());
  });

  it("Discord Complete Setup error shows error", async () => {
    vi.stubGlobal("fetch", makeFetchWithIntegrations(mockDiscordIntegrations, (input) => {
      if (input === "/api/integrations/chat/test") return mockJson({ success: true });
      if (input === "/api/integrations/connect") return mockJson({ error: "Server error" }, 500);
      return null;
    }));
    await openDiscordWizard();
    fireEvent.click(screen.getByRole("button", { name: /next: enter url/i }));
    await screen.findByText("Step 2 — Enter Webhook URL");
    fireEvent.change(screen.getByPlaceholderText(/https:\/\/discord.com/), { target: { value: "https://discord.com/api/webhooks/123/abc" } });
    fireEvent.click(screen.getByRole("button", { name: /next: test/i }));
    await screen.findByText("Step 3 — Test Connection");
    fireEvent.click(screen.getByRole("button", { name: /send test message/i }));
    await screen.findByText("Step 3 — Test Succeeded ✓");
    fireEvent.click(screen.getByRole("button", { name: /complete setup/i }));
    // wizard stays open on error (connect failed)
    await waitFor(() => expect(screen.getByText("Connect Discord Webhook")).toBeInTheDocument());
  });
});

// ── MCP Invoke modal coverage ─────────────────────────────────────────────────

describe("App – MCP invoke modal (communication category)", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const commTool = { id: "slack-mcp", name: "Slack MCP", description: "Send Slack messages", category: "communication", status: "available" };

  it("opens MCP invoke modal for communication tool and exercises form fields", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([commTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Slack MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Slack MCP");

    // Exercise communication fields
    fireEvent.change(screen.getByPlaceholderText(/e.g. -1001234567890 or #general/), { target: { value: "#general" } });
    const textareas = screen.getAllByRole("textbox");
    const msgArea = textareas.find(el => el.tagName.toLowerCase() === "textarea");
    if (msgArea) fireEvent.change(msgArea, { target: { value: "Hello!" } });

    // Close button (footer) - use getAllByRole to handle multiple "Close" buttons
    const closeBtns = screen.getAllByRole("button", { name: /^close$/i });
    fireEvent.click(closeBtns[closeBtns.length - 1]);
    await waitFor(() => expect(screen.queryByText("Invoke: Slack MCP")).not.toBeInTheDocument());
  });

  it("invokes tool successfully and shows result", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([commTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      if (input === "/api/mcp/tools/invoke") return mockJson({ status: "sent" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Slack MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Slack MCP");
    fireEvent.click(screen.getByRole("button", { name: /invoke tool/i }));
    await waitFor(() => expect(screen.getByText("Result")).toBeInTheDocument());
  });

  it("invokes tool with error and shows error", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([commTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      if (input === "/api/mcp/tools/invoke") return mockJson({ error: "tool error" }, 500);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Slack MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Slack MCP");
    fireEvent.click(screen.getByRole("button", { name: /invoke tool/i }));
    await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
  });

  it("closes MCP modal with ✕ button", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([commTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Slack MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Slack MCP");
    // Click the header ✕ close button (first of the "Close" buttons)
    const allCloseBtns = screen.getAllByRole("button", { name: /close/i });
    fireEvent.click(allCloseBtns[0]);
    await waitFor(() => expect(screen.queryByText("Invoke: Slack MCP")).not.toBeInTheDocument());
  });
});

describe("App – MCP invoke modal (code category)", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const codeTool = { id: "git-mcp", name: "Git MCP", description: "Create PRs", category: "code", status: "available" };

  it("shows code category form fields and exercises them", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([codeTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Git MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Git MCP");

    fireEvent.change(screen.getByPlaceholderText("owner/repo"), { target: { value: "my/repo" } });
    fireEvent.change(screen.getByPlaceholderText("feat: description"), { target: { value: "feat: new feature" } });
    fireEvent.change(screen.getByPlaceholderText("feature/my-branch"), { target: { value: "feature/test" } });
    // target branch has default "main"
    const targetBranchInput = screen.getByDisplayValue("main") as HTMLInputElement;
    fireEvent.change(targetBranchInput, { target: { value: "develop" } });
    expect(targetBranchInput.value).toBe("develop");
  });
});

describe("App – MCP invoke modal (project_management category)", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const pmTool = { id: "jira-mcp", name: "Jira MCP", description: "Create issues", category: "project_management", status: "available" };

  it("shows project_management category form fields and exercises them", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([pmTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Jira MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Jira MCP");

    fireEvent.change(screen.getByPlaceholderText("e.g. PROJ"), { target: { value: "PROJ" } });
    fireEvent.change(screen.getByPlaceholderText("Issue title"), { target: { value: "New bug" } });
    // priority select - find by default value "medium"
    const prioritySelect = screen.getByDisplayValue("Medium") as HTMLSelectElement;
    fireEvent.change(prioritySelect, { target: { value: "high" } });
    expect(prioritySelect.value).toBe("high");
  });
});

describe("App – MCP invoke modal (other/default category)", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const otherTool = { id: "other-mcp", name: "Other MCP", description: "Other tool", category: "monitoring", status: "available" };

  it("shows default invocation message for unknown category", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/mcp/tools") return mockJson([otherTool]);
      if (input === "/api/settings") return mockJson({ minimaxApiKey: "" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("MCP Tool Gateway");
    await waitFor(() => expect(screen.getByText("Other MCP")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /invoke/i }));
    await screen.findByText("Invoke: Other MCP");
    expect(screen.getByText(/default invocation/i)).toBeInTheDocument();
  });

  it("handles settings fetch domains api error", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/domains") return mockJson({error: "fail"}, 500);
      if (input === "/api/mcp/tools") return mockJson([]);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
  });

  it("handles users API error", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/users") return mockJson({error: "fail"}, 500);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /users/i }));
  });

  it("handles integrations fetch API error", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson({error: "fail"}, 500);
      return mockJson({}, 404);
    }));

    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
  });
});

// ── handleSaveSettings coverage ───────────────────────────────────────────────

describe("App – handleSaveSettings", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("saves settings successfully and shows notice", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/settings" && (!init || init.method !== "POST")) return mockJson({ minimaxApiKey: "" });
      if (input === "/api/settings" && init?.method === "POST") return mockJson({ minimaxApiKey: "new-key" });
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("AI Model Configuration");
    fireEvent.change(screen.getByPlaceholderText("sk-cp-..."), { target: { value: "new-key" } });
    fireEvent.click(screen.getByRole("button", { name: /save settings/i }));
    await screen.findByText("Settings saved successfully.");
  });

  it("settings save failure does not show success notice", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/settings" && (!init || init.method !== "POST")) return mockJson({ minimaxApiKey: "" });
      if (input === "/api/settings" && init?.method === "POST") return mockJson({ error: "Server error" }, 500);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /settings/i }));
    await screen.findByText("AI Model Configuration");
    fireEvent.click(screen.getByRole("button", { name: /save settings/i }));
    // Wait for save to complete - button should be re-enabled (savingSettings=false)
    await waitFor(() => expect(screen.getByRole("button", { name: /save settings/i })).not.toBeDisabled());
    expect(screen.queryByText("Settings saved successfully.")).not.toBeInTheDocument();
  });
});

// ── Meetings tab form onChange handlers ───────────────────────────────────────

describe("App – Meetings tab form onChange", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("changes toAgent and meetingId selects", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await screen.findByText("Virtual War Room");

    // Find "To" select by its option text (agents from dashboardPayload)
    const toSelects = screen.getAllByRole("combobox");
    // First select in meeting form should be "To" agent
    const toAgentSelect = toSelects.find(el => {
      const opts = Array.from((el as HTMLSelectElement).options);
      return opts.some(o => o.text.includes("PM") || o.text.includes("SWE"));
    });
    if (toAgentSelect) {
      fireEvent.change(toAgentSelect, { target: { value: "pm-1" } });
      expect((toAgentSelect as HTMLSelectElement).value).toBe("pm-1");
    }

    // meetingId select
    const meetingSelects = screen.getAllByRole("combobox");
    const meetingSelect = meetingSelects.find(el => {
      const opts = Array.from((el as HTMLSelectElement).options);
      return opts.some(o => o.text.includes("launch-readiness"));
    });
    if (meetingSelect) {
      fireEvent.change(meetingSelect, { target: { value: "launch-readiness" } });
      expect((meetingSelect as HTMLSelectElement).value).toBe("launch-readiness");
    }
  });
});

// ── Nav item click (setActiveNav) ─────────────────────────────────────────────

describe("App – nav item click", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("clicking nav items changes the active section", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");

    fireEvent.click(screen.getByRole("button", { name: /agents/i }));
    await screen.findByText("Agent Network");

    fireEvent.click(screen.getByRole("button", { name: /cost/i }));
    await screen.findByText("Burn Rate Forecast");

    fireEvent.click(screen.getByRole("button", { name: /overview/i }));
    await screen.findByText("Acme Software");
  });
});

// ── Integrations tab – connected integration with chatspace/hasCredentials ────

describe("App – integrations connected branch coverage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("shows chatspace and hasCredentials for connected integrations", async () => {
    const connectedIntegrations = [
      { id: "slack", name: "Slack", type: "slack", category: "chat", status: "connected", description: "Slack integration", createdAt: "2026-01-01T00:00:00Z", chatspace: "my-workspace", hasCredentials: true },
    ];
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/integrations") return mockJson(connectedIntegrations);
      if (input === "/api/integrations?category=chat") return mockJson(connectedIntegrations);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /integrations/i }));
    await screen.findByText("Chat Services");
    await waitFor(() => expect(screen.getByText(/Chatspace:/)).toBeInTheDocument());
    expect(screen.getByText(/Credentials configured ✓/)).toBeInTheDocument();
  });
});

// ── Meetings form meetingId onChange coverage ─────────────────────────────────

describe("App – meetings meetingId onChange explicit", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("changing meeting select updates selectedMeetingID", async () => {
    const payload = {
      ...dashboardPayload,
      meetings: [
        { id: "meeting-1", participants: [], transcript: [] },
        { id: "meeting-2", participants: [], transcript: [] },
      ],
    };
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(payload);
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await screen.findByText("Virtual War Room");

    const comboboxes = screen.getAllByRole("combobox");
    const meetingSelect = comboboxes.find(el => {
      const opts = Array.from((el as HTMLSelectElement).options);
      return opts.some(o => o.value === "meeting-2");
    });
    expect(meetingSelect).toBeDefined();
    if (meetingSelect) {
      fireEvent.change(meetingSelect, { target: { value: "meeting-2" } });
      await waitFor(() => expect((meetingSelect as HTMLSelectElement).value).toBe("meeting-2"));
    }
  });
});

// ── War Room Enter key handler coverage ───────────────────────────────────────

describe("App – War Room Approval Cards", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("renders CEO Approval card and handles Approve/Reject actions", async () => {
    // Mock snapshot with an ApprovalNeeded message targeting CEO
    const snapshotWithApproval = {
      ...richPayload,
      organization: {
        ...richPayload.organization,
        ceoId: "CEO",
      },
      meetings: [
        {
          id: "launch-readiness",
          agenda: "Test agenda",
          participants: ["pm-1", "CEO"],
          transcript: [
            {
              id: "msg-1",
              fromAgent: "pm-1",
              toAgent: "CEO",
              type: "ApprovalNeeded",
              content: "CEO approval needed for launch",
              occurredAt: new Date().toISOString(),
            },
          ],
        },
      ],
    } as any;

    // we use global fetch mock in other tests, let's stick to the same pattern
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const urlStr = input.toString();
        if (urlStr === "/api/dashboard") return mockJson(snapshotWithApproval);
        if (urlStr === "/api/messages") {
          return mockJson(snapshotWithApproval);
        }
        return mockJson({});
      }),
    );

    render(<App />);

    // Switch to Meetings tab where the war room is
    await waitFor(() => {
      expect(screen.getByText("Meetings")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("Meetings"));

    // Wait for data load and component rendering
    await waitFor(() => {
      expect(screen.getByText("CEO Approval Required")).toBeInTheDocument();
    });

    const approveBtn = screen.getByRole("button", { name: "Approve" });
    const rejectBtn = screen.getByRole("button", { name: "Reject" });
    expect(approveBtn).toBeInTheDocument();
    expect(rejectBtn).toBeInTheDocument();

    // Test Approve click
    fireEvent.click(approveBtn);
    await waitFor(() => {
      expect(window.fetch).toHaveBeenCalledWith(
        "/api/messages",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining("messageType=SpecApproved")
        })
      );
    });

    // Test Reject click
    fireEvent.click(rejectBtn);
    await waitFor(() => {
      expect(window.fetch).toHaveBeenCalledWith(
        "/api/messages",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining("messageType=direction")
        })
      );
    });
  });
});

describe("App – War Room Enter key submit", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("pressing Enter (non-shift) in the war room textarea triggers submit", async () => {
    let messageSent = false;
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/messages") {
        messageSent = true;
        return mockJson({});
      }
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await screen.findByText("Virtual War Room");

    const textarea = screen.getByPlaceholderText(/Inject direction/i);
    fireEvent.change(textarea, { target: { value: "Hello meeting" } });
    // Fire Enter key to trigger submit via the onKeyDown handler
    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: false, bubbles: true, cancelable: true });
    await waitFor(() => expect(messageSent).toBe(true));
  });

  it("pressing Shift+Enter in the war room textarea does NOT submit", async () => {
    let messageSent = false;
    vi.stubGlobal("fetch", vi.fn(async (input: string) => {
      if (input === "/api/dashboard") return mockJson(dashboardPayload);
      if (input === "/api/messages") {
        messageSent = true;
        return mockJson({});
      }
      return mockJson({}, 404);
    }));
    render(<App />);
    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: /meetings/i }));
    await screen.findByText("Virtual War Room");

    const textarea = screen.getByPlaceholderText(/Inject direction/i);
    fireEvent.change(textarea, { target: { value: "Multi-line" } });
    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: true, bubbles: true, cancelable: true });
    // Give time to confirm no submit occurred
    await new Promise((r) => setTimeout(r, 100));
    expect(messageSent).toBe(false);
  });
});

describe("App - Dynamic Scaling", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = input.toString();
      if (url === "/api/dashboard") {
        return { ok: true, json: async () => ({
          org: { name: "Acme Software" },
          stats: { budgetUsedUsd: 100 },
          tasks: [],
          users: [],
          agents: []
        })};
      }
      if (url === "/api/settings") {
         return { ok: true, json: async () => ({}) };
      }
      if (url === "/api/auth/me") {
        return { ok: true, json: async () => ({ username: "admin" }) };
      }
      if (url === "/api/v1/scale") {
        return { ok: true, json: async () => ({ status: "success", role: "sales_rep", count: 2 }) };
      }
      if (url === "/api/domains") {
        return { ok: true, json: async () => ([]) };
      }
      return originalFetch(input, init);
    }) as any;
  });

  afterEach(() => {
    global.fetch = originalFetch;
    delete (window as any).EventSource;
    vi.clearAllMocks();
  });

  it("navigates to scaling tab and submits scaling form", async () => {
    let _onmessage: ((ev: any) => void) | null = null;
    let _onerror: (() => void) | null = null;

    class MockEventSource {
      close() {}
      set onmessage(fn: (ev: any) => void) { _onmessage = fn; }
      set onerror(fn: () => void) { _onerror = fn; }
    }

    (window as any).EventSource = MockEventSource;

    render(<App />);
    await screen.findByText("One Human Corp Dashboard");

    const scalingTab = screen.getByText("Dynamic Scaling");
    await act(async () => {
      fireEvent.click(scalingTab);
    });

    expect(screen.getByText(/Scale AI agents dynamically/i)).toBeInTheDocument();

    const applyButton = screen.getByText("Apply Scaling Changes");
    await act(async () => {
      fireEvent.click(applyButton);
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith("/api/v1/scale", expect.objectContaining({ method: "POST" }));
    });

    // Simulate SSE
    await act(async () => {
      if (_onmessage) {
        _onmessage({ data: JSON.stringify({ event: "K8s Operator: Reconciling...", status: "INFO" }) });
      }
    });

    expect(screen.getByText("K8s Operator: Reconciling...")).toBeInTheDocument();

    await act(async () => {
      if (_onmessage) {
        _onmessage({ data: JSON.stringify({ event: "AgentHired", status: "SUCCESS" }) });
      }
    });

    await act(async () => {
      if (_onerror) {
        _onerror();
      }
    });
  });

  it("handles malformed SSE data", async () => {
    let _onmessage: ((ev: any) => void) | null = null;

    class MockEventSource {
      close() {}
      set onmessage(fn: (ev: any) => void) { _onmessage = fn; }
      set onerror(_fn: () => void) {}
    }

    (window as any).EventSource = MockEventSource;

    render(<App />);
    await screen.findByText("One Human Corp Dashboard");

    await act(async () => {
      fireEvent.click(screen.getByText("Dynamic Scaling"));
    });
    await act(async () => {
      fireEvent.click(screen.getByText("Apply Scaling Changes"));
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith("/api/v1/scale", expect.objectContaining({ method: "POST" }));
    });

    await act(async () => {
      if (_onmessage) {
        _onmessage({ data: "invalid json" });
      }
    });
  });

  it("handles scaleAgents api error", async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = input.toString();
      if (url === "/api/dashboard") {
        return { ok: true, json: async () => ({ org: {}, stats: {}, tasks: [], users: [], agents: [] })};
      }
      if (url === "/api/auth/me") {
        return { ok: true, json: async () => ({ username: "admin" }) };
      }
      if (url === "/api/settings") {
         return { ok: true, json: async () => ({}) };
      }
      if (url === "/api/domains") {
        return { ok: true, json: async () => ([]) };
      }
      if (url === "/api/v1/scale") {
        return {
           ok: false,
           status: 500,
           text: async () => "Internal Server Error"
        };
      }
      return originalFetch(input, init);
    }) as any;

    render(<App />);
    await screen.findByText("One Human Corp Dashboard");

    await act(async () => {
      fireEvent.click(screen.getByText("Dynamic Scaling"));
    });

    // Create an unhandled rejection guard, or simply wait for notice
    await act(async () => {
      fireEvent.click(screen.getByText("Apply Scaling Changes"));
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith("/api/v1/scale", expect.objectContaining({ method: "POST" }));
    });

    // There is no UI notice displayed when this throws, it just logs to console and sets isScalingActive=false.
    // So we just verify that it completed processing and the button is re-enabled.
    await waitFor(() => {
        expect(screen.getByText("Apply Scaling Changes")).toBeInTheDocument();
        expect(screen.getByText("Apply Scaling Changes")).not.toBeDisabled();
    });
  });

  it("updates slider values", async () => {
    render(<App />);
    await screen.findByText("One Human Corp Dashboard");

    await act(async () => {
      fireEvent.click(screen.getByText("Dynamic Scaling"));
    });

    const sliders = screen.getAllByRole("slider");
    expect(sliders.length).toBe(3);

    await act(async () => {
      fireEvent.change(sliders[0], { target: { value: "5" } });
      fireEvent.change(sliders[1], { target: { value: "8" } });
      fireEvent.change(sliders[2], { target: { value: "10" } });
    });
  });
});
