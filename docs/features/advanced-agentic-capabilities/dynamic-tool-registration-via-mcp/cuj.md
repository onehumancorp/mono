# CUJ: Dynamic Tool Registration via MCP

**Persona:** SWE Agent / QA Agent | **Context:** An agent needs a specific capability (e.g., executing a GitHub API call) that is not currently bound to its default prompt context.
**Success Metrics:** The agent successfully queries the MCP Gateway, registers the necessary tool schema dynamically, and executes the tool without hardcoded dependencies.

## 1. User Journey Overview
When an AI agent (e.g., SWE Agent) realizes it lacks a specific capability to complete a task, it dynamically queries the centralized MCP (Model Context Protocol) Gateway. The Gateway searches its registry, synthesizes the OpenAPI schema for the requested tool (e.g., `tools.git.commit()`), and securely binds it to the agent's context using a short-lived SPIFFE certificate. The agent then executes the tool successfully and continues its workflow.

## 2. Detailed Step-by-Step Breakdown

| Step | User/Agent Action | System Trigger | Resulting State | Verification |
|------|-------------------|----------------|-----------------|--------------|
| 1 | Agent encounters missing tool | Hub evaluates available context | `ToolNotFound` error logged | Error log entry |
| 2 | Agent queries MCP Gateway | Hub forwards query to MCP Registry | Available tools list returned | JSON list of capabilities |
| 3 | Agent selects required tool | Hub requests short-lived SPIFFE cert for the tool | Tool schema injected into agent's prompt | Updated system prompt |
| 4 | Agent executes the tool | Hub routes payload to specific MCP server | Tool executes (e.g., Git commit) | API response received |
| 5 | Tool results returned | Hub passes results to Checkpointer | Agent state updated | Result logged in `events.jsonl` |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Tool Not Found in Registry
- **Detection**: The agent queries the MCP Gateway for a specific capability, but it is not installed or configured in the organization's settings.
- **Auto-Recovery**: The Hub replies with an empty list. The agent either attempts a Fallback Strategy (e.g., writing a custom script) or places itself in a `WAITING_FOR_TOOLS` state.
- **Human Escalation**: The Manager Agent alerts the human CEO via the Handoff UI that a required MCP bundle (e.g., `jira-mcp`) must be installed via the Switchboard.
### 3.2 Scenario: Invalid Tool Parameters
- **Detection**: The agent attempts to execute the newly bound tool, but the payload fails schema validation (e.g., missing required fields).
- **Auto-Recovery**: The MCP Gateway returns a detailed 400 Bad Request error. The agent's `Tool Parameter Auto-Correction` logic parses the error, fixes the payload, and retries the execution.

## 4. UI/UX Details
- **Dashboard Representation**: The "Active Agents" view in the CEO Dashboard dynamically updates to show an icon representing the newly bound tool attached to the agent's profile.

## 5. Security & Privacy
- **Dynamic Access Control**: Tool execution is strictly gated by SPIFFE/SPIRE. An agent can only bind to a tool if its `RoleProfile` is explicitly authorized to use it, preventing a compromised SWE Agent from executing Finance/Accounting tool calls.
