import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { App } from "./App";

type MockResponse = {
  ok: boolean;
  status: number;
  json: () => Promise<unknown>;
};

const orgPayload = {
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
};

const meetingsPayload = [
  {
    id: "kickoff",
    agenda: "Demo",
    participants: ["pm-1", "swe-1"],
    transcript: [
      {
        id: "m-1",
        fromAgent: "pm-1",
        toAgent: "swe-1",
        type: "task",
        content: "Review roadmap",
        meetingId: "kickoff",
        occurredAt: "2026-03-13T00:00:00Z",
      },
    ],
  },
];

const costsPayload = {
  organizationID: "org-1",
  totalTokens: 123,
  totalCostUSD: 0.012345,
  agents: [],
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

  it("renders dashboard data after loading", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/org") return mockJson(orgPayload);
      if (input === "/api/meetings") return mockJson(meetingsPayload);
      if (input === "/api/costs") return mockJson(costsPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await screen.findByText("Acme Software");
    expect(screen.getByText("One Human Corp Dashboard")).toBeInTheDocument();
    expect(screen.getByText("PRODUCT_MANAGER")).toBeInTheDocument();
    expect(screen.getByText("Review roadmap", { exact: false })).toBeInTheDocument();
  });

  it("shows API error state", async () => {
    const fetchMock = vi.fn(async () => mockJson({}, 500));
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await screen.findByText(/Failed to load data/i);
  });

  it("submits message form and refreshes", async () => {
    const fetchMock = vi.fn(async (input: string, init?: RequestInit) => {
      if (input === "/api/messages") {
        expect(init?.method).toBe("POST");
        return mockJson({}, 200);
      }
      if (input === "/api/org") return mockJson(orgPayload);
      if (input === "/api/meetings") return mockJson(meetingsPayload);
      if (input === "/api/costs") return mockJson(costsPayload);
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
  });

  it("shows fallback error when send fails with non-Error", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/messages") {
        throw "send failed";
      }
      if (input === "/api/org") return mockJson(orgPayload);
      if (input === "/api/meetings") return mockJson(meetingsPayload);
      if (input === "/api/costs") return mockJson(costsPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: "Send Message" }));

    await screen.findByText("Failed to send message");
  });

  it("shows send error message when API returns non-OK", async () => {
    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/messages") {
        return mockJson({}, 400);
      }
      if (input === "/api/org") return mockJson(orgPayload);
      if (input === "/api/meetings") return mockJson(meetingsPayload);
      if (input === "/api/costs") return mockJson(costsPayload);
      return mockJson({}, 404);
    });

    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await screen.findByText("Acme Software");
    fireEvent.click(screen.getByRole("button", { name: "Send Message" }));

    await screen.findByText("Failed to send message: 400");
  });

  it("shows unknown error when load fails with non-Error", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw "load failed";
    }));

    render(<App />);

    await screen.findByText("Failed to load data: Unknown error");
  });

  it("renders empty transcript state and handles input updates and refresh", async () => {
    const emptyMeetings = [
      {
        id: "kickoff",
        agenda: "Demo",
        participants: ["pm-1", "swe-1"],
        transcript: [],
      },
    ];

    const fetchMock = vi.fn(async (input: string) => {
      if (input === "/api/org") return mockJson(orgPayload);
      if (input === "/api/meetings") return mockJson(emptyMeetings);
      if (input === "/api/costs") return mockJson(costsPayload);
      return mockJson({}, 404);
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await screen.findByText("No messages yet.");

    fireEvent.change(screen.getByLabelText("From Agent"), { target: { value: "new-from" } });
    fireEvent.change(screen.getByLabelText("To Agent"), { target: { value: "new-to" } });
    fireEvent.change(screen.getByLabelText("Meeting ID"), { target: { value: "new-meeting" } });
    fireEvent.change(screen.getByLabelText("Message Type"), { target: { value: "status" } });
    fireEvent.change(screen.getByLabelText("Content"), { target: { value: "updated content" } });

    expect(screen.getByDisplayValue("new-from")).toBeInTheDocument();
    expect(screen.getByDisplayValue("new-to")).toBeInTheDocument();
    expect(screen.getByDisplayValue("new-meeting")).toBeInTheDocument();
    expect(screen.getByDisplayValue("status")).toBeInTheDocument();
    expect(screen.getByDisplayValue("updated content")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/org");
    });
  });
});