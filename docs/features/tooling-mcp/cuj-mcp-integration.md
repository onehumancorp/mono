# CUJ: Securely Extend Capabilities via MCP (Tooling Integration)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** Platform Admin | **Context:** Integrating a new corporate tool (e.g., Slack).
**Success Metrics:** Handshake success < 2s, 100% tool discovery, Zero exposed secrets.

## 1. User Journey Overview
The Admin needs to give the "Support Agent" access to Slack. They register a custom MCP Server. The OHC Gateway performs a dynamic handshake to discover tools (`send_message`, `list_channels`). Once approved, these tools are "mapped" to the Support Agent's capabilities.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Add New Integration". | FE: `openMCPForm()` | UI: URL & Token input modal. | Modal ID `#mcp-integration-form`. |
| 2 | Enter `http://slack-mcp:3000`. | BE: `POST /api/mcp/probe` | Gateway: Initiates JSON-RPC ListTools. | HTTP 200 with Tool Schema. |
| 3 | Review discovered tools. | N/A | UI: List with checkboxes for permissions. | DOM check for `.tool-checkbox`. |
| 4 | Click "Enable for Role: Support".| BE: `PUT /api/roles/support/tools`| Hub: Updates RoleProfile ACL. | Registry reflects new tool mapping. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: MCP Gateway Timeout
- **Detection**: Probe fails with `context.DeadlineExceeded`.
- **Recovery Step**: Check cluster network policies; UI suggests checking `mcp-slack` pod logs.
### 3.2 Scenario: Insecure Tool Discovery
- **Detection**: MCP server doesn't support mTLS.
- **User Feedback**: "Caution: Unencrypted tool connection. Proceed only in local environment."

## 4. UI/UX Details
- **Component IDs**: `IntegrationList`, `ToolPermissionMatrix`.
- **Visual Cues**: Success adds a "Slack" icon to the registered integrations bar with a green "Live" badge.

## 5. Security & Privacy
- **Token Masking**: API keys for Slack are stored exclusively in the MCP Server environment, never in the OHC Hub.
- **Audit Log**: `Admin[kevin] ENABLED Tool[slack.post] for Role[SUPPORT]` logged.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.
