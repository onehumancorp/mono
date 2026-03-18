# CUJ: Securely Extend Capabilities via MCP

**Persona:** Platform Admin / Org Owner
**Goal:** Add a new tool (e.g., Slack) to the OHC ecosystem via MCP.
**Success Metrics:** Tool is registered and available to agents in <1 minute.

## Context
The organisation needs to communicate externally via Slack, and the CEO wants the Marketing Agent to post updates automatically.

## Journey Breakdown
### Step 1: Register MCP Server
- **User Input:** Admin enters the MCP server URL for Slack (`http://mcp-slack:3000`).
- **System Action:** Hub calls the MCP "Handshake" to discover available tools (`post_message`, `create_channel`).
- **Outcome:** Slack tools are added to the integration registry.

### Step 2: Assign Tool to Agent
- **User Input:** Admin assigns "Slack" capability to the "Marketing Agent".
- **System Action:** Hub updates agent's tool permissions.
- **Outcome:** Marketing Agent can now post to Slack.

## Error Modes & Recovery
### Failure 1: Tool Timeout
- **System Behavior:** Agent reports "Tool Slack is unavailable".
- **Recovery Step:** Admin checks the status of the MCP server pod.

## Security & Privacy Considerations
- Tool calls are scoped by agent role.
- All secrets (API tokens) are managed by the MCP server, never exposed to the agent.
