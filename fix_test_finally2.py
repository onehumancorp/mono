import re

# I will write a script that tests these functions directly without rendering App!
# Is it possible? No, they are inside the App component.
# I can just write a VERY robust UI test.

with open("srcs/frontend/src/App.test.tsx", "r") as f:
    content = f.read()

test_code = """
  it("probes and enables tools", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url === "/api/dashboard" || url === "/api/costs") return mockJson({ organization: { id: "o", name: "n", domain: "d", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: new Date().toISOString() });
      if (url === "/api/mcp/tools" || url === "/api/domains" || url === "/api/integrations") return mockJson([]);
      if (url === "/api/mcp/probe") return mockJson([{ name: "test_tool", description: "test desc" }]);
      if (url.includes("/api/roles/")) return mockJson({ status: "ok" });
      return mockJson({});
    });

    const { container } = render(<App />);

    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Settings" })); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Add New Integration" })); });

    await act(async () => { fireEvent.change(await screen.findByPlaceholderText("http://slack-mcp:3000"), { target: { value: "https://test-mcp:3000" } }); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Probe Server" })); });

    const checkbox = await screen.findByRole("checkbox");
    await act(async () => { fireEvent.click(checkbox); });

    // Select role
    const select = container.querySelector("select")!;
    await act(async () => { fireEvent.change(select, { target: { value: "software_engineer" } }); });

    // Enable
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: /Enable for Role/ })); });
  });

  it("handles errors", async () => {
    vi.stubGlobal("alert", vi.fn());
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url === "/api/dashboard" || url === "/api/costs") return mockJson({ organization: { id: "o", name: "n", domain: "d", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: new Date().toISOString() });
      if (url === "/api/mcp/tools" || url === "/api/domains" || url === "/api/integrations") return mockJson([]);
      if (url === "/api/mcp/probe") return new Response("Unencrypted tool connection", { status: 400 });
      if (url.includes("/api/roles/")) return new Response("Error", { status: 500 });
      return mockJson({});
    });

    render(<App />);

    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Settings" })); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Add New Integration" })); });

    await act(async () => { fireEvent.change(await screen.findByPlaceholderText("http://slack-mcp:3000"), { target: { value: "http://slack-mcp:3000" } }); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Probe Server" })); });

    expect(await screen.findByText("Unencrypted tool connection")).toBeInTheDocument();
  });
"""

# append
content = content.replace("  });\n});\n", "  });\n\n" + test_code + "});\n")

if "import { fireEvent, render, screen, waitFor, act }" not in content:
    content = content.replace('import { fireEvent, render, screen, waitFor }', 'import { fireEvent, render, screen, waitFor, act }')

with open("srcs/frontend/src/App.test.tsx", "w") as f:
    f.write(content)
