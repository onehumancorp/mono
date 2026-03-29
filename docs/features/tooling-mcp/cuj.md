# CUJ: Tooling & MCP Integrations


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO expands the capability of their AI workforce by linking an external tool (e.g., Slack or Linear) through the MCP Gateway.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to Tools | UI renders marketplace | API lists default servers | Verify 200 OK |
| 2 | Connect new tool | CEO enters server URL | Gateway tests connection | Success toast visible |
| 3 | Provide OAuth token | CEO authorizes app | Secrets saved to Vault | Vault stores token |
| 4 | Assign tool | CEO grants to PR Agent | RBAC mapping created | Agent sees tool in list |

## 3. Implementation Details
- **Architecture**: Integrated via Go 1.26 MCP Gateway acting as a reverse-proxy for standard MCP JSON-RPC messages.
- **Stack**: OHC Kubernetes Operator for managing deployed MCP sidecars.
- **Security Check**: Employs strictly fail-closed policies on network timeouts.

## 4. Edge Cases
- **Stale Contexts**: Tool metadata is updated dynamically. If a tool becomes unavailable mid-task, the MCP Gateway returns an explicit "ToolOffline" error to the agent, prompting it to wait or re-plan.
- **Authentication Handshake Failures**: If the OAuth token expires and refresh fails, the Gateway intercepts all tool calls and escalates a notification to the CEO for re-authentication, preventing agents from endlessly looping on 401s.
- **Payload Limits**: Extremely large API responses (e.g., a massive JSON dump from a database) are truncated by the Gateway before returning to the agent to avoid context window explosion.