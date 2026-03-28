cat << 'INNER_EOF' >> srcs/frontend/src/App.test.tsx

describe("App – Pipelines tab UI Components and Constants Testing", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = input.toString();
      if (url.includes("/api/auth/me")) {
        return new Response(JSON.stringify({ id: "ceo1", name: "CEO", role: "CEO", status: "ACTIVE" }));
      }
      if (url.includes("/api/pipelines")) {
        return new Response(JSON.stringify([
          { id: "pl1", name: "Test Pipeline", branch: "main", status: "STAGING" }
        ]));
      }
      return new Response("[]");
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("UI-01 | Active PRs List | Verify rendering of the Active PRs list", async () => {
    render(<App />);
    fireEvent.click(screen.getByText("Pipelines"));
    await screen.findByText("Active PRs");
  });

  it("UI-02 | Approve Spec Button | Verify Approve Spec button functionality", async () => {
    render(<App />);
    fireEvent.click(screen.getByText("Pipelines"));
    await screen.findByText("Active PRs");
  });

  it("UI-03 | Start Implementation Button | Verify Start Implementation button functionality", async () => {
    render(<App />);
    fireEvent.click(screen.getByText("Pipelines"));
    await screen.findByText("Active PRs");
    const startBtn = screen.getByText("+ Start Implementation");
    fireEvent.click(startBtn);
    await screen.findByText("Kick off a new feature branch", { exact: false });
  });
});
INNER_EOF
