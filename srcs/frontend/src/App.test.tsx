import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { App } from "./App";

type MockResponse = {
  ok: boolean;
  status: number;
  json: () => Promise<unknown>;
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
  ],
  statuses: [{ status: "IN_MEETING", count: 1 }],
  updatedAt: "2026-03-13T00:00:00Z",
};

function mockJson(data: unknown, status = 200): MockResponse {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => data,
  };
}

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

    await screen.findByText("Failed to send message: 400");
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
      expect(fetchMock).toHaveBeenCalledWith("/api/dashboard");
    });
  });
});
