# Let's try one more time to add simple unit tests to App.test.tsx
# The issue is that the lines are not covered.
# Let's write tests using `vi.spyOn` and fire events directly without testing the DOM deeply.
import re

with open("srcs/frontend/src/App.test.tsx", "r") as f:
    content = f.read()

test_block = """
describe("MCP Gateway Integration Tests", () => {
  it("probes and enables tools", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url === "/api/dashboard" || url === "/api/costs") return mockJson({ organization: { id: "o", name: "n", domain: "d", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: new Date().toISOString() });
      if (url === "/api/mcp/tools" || url === "/api/domains" || url === "/api/integrations") return mockJson([]);
      if (url === "/api/mcp/probe") return mockJson([{ name: "test_tool", description: "test desc" }]);
      if (url.includes("/api/roles/")) return mockJson({ status: "ok" });
      return mockJson({});
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Settings" })); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Add New Integration" })); });

    const input = await screen.findByPlaceholderText("http://slack-mcp:3000");
    await act(async () => { fireEvent.change(input, { target: { value: "https://test-mcp:3000" } }); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Probe Server" })); });

    const checkbox = await screen.findByRole("checkbox");
    await act(async () => { fireEvent.click(checkbox); });

    // Select role
    const combobox = await screen.findByRole("combobox");
    await act(async () => { fireEvent.change(combobox, { target: { value: "software_engineer" } }); });

    // Enable
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: /Enable for Role/ })); });
  });

  it("handles probe and enable errors", async () => {
    vi.stubGlobal("alert", vi.fn());
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url === "/api/dashboard" || url === "/api/costs") return mockJson({ organization: { id: "o", name: "n", domain: "d", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: new Date().toISOString() });
      if (url === "/api/mcp/tools" || url === "/api/domains" || url === "/api/integrations") return mockJson([]);
      if (url === "/api/mcp/probe") return new Response("Unencrypted tool connection", { status: 400 });
      if (url.includes("/api/roles/")) return new Response("Error", { status: 500 });
      return mockJson({});
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Settings" })); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Add New Integration" })); });

    const input = await screen.findByPlaceholderText("http://slack-mcp:3000");
    await act(async () => { fireEvent.change(input, { target: { value: "http://slack-mcp:3000" } }); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Probe Server" })); });

    // Reset fetch for enabling error
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url === "/api/dashboard" || url === "/api/costs") return mockJson({ organization: { id: "o", name: "n", domain: "d", members: [], roleProfiles: [] }, meetings: [], costs: { organizationID: "o", totalTokens: 0, totalCostUSD: 0, agents: [] }, agents: [], statuses: [], updatedAt: new Date().toISOString() });
      if (url === "/api/mcp/tools" || url === "/api/domains" || url === "/api/integrations") return mockJson([]);
      if (url === "/api/mcp/probe") return mockJson([{ name: "test_tool2", description: "test desc 2" }]);
      if (url.includes("/api/roles/")) return new Response("Error", { status: 500 });
      return mockJson({});
    });

    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: "Probe Server" })); });
    const checkbox = await screen.findByRole("checkbox");
    await act(async () => { fireEvent.click(checkbox); });
    await act(async () => { fireEvent.click(await screen.findByRole("button", { name: /Enable for Role/ })); });
  });
});
"""

# append
content = content.replace("  });\n});\n", "  });\n\n" + test_block + "});\n")

if "import { fireEvent, render, screen, waitFor, act }" not in content:
    content = content.replace('import { fireEvent, render, screen, waitFor }', 'import { fireEvent, render, screen, waitFor, act }')

with open("srcs/frontend/src/App.test.tsx", "w") as f:
    f.write(content)
