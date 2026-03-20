# User Guide: MCP Tool Integrations

## Introduction
MCP (Model Context Protocol) is how your agents "interact with the world". It connects them to tools like GitHub, Slack, and your internal databases.

## Setup
### 1. Browse Integrations
Open the "Marketplace" and filter for "Tools".

### 2. Connect a Tool
Click "Connect" and enter the MCP server URL. For popular tools, we provide pre-configured templates.

### 3. Permissions
Assign the tool to specific agents. Do not give "Delete Database" access to a junior intern agent!

## Advanced Usage
### Building Custom Tools
You can build your own MCP server in any language. Just provide a manifest file that OHC can read.

## Troubleshooting
**Tool is disconnected**
- Check the "Gateway" status icon.
- Ensure the MCP server is running on a reachable network path.

## Implementation Details
- **Architecture**: The OHC MCP Gateway serves as a unified proxy written in Go 1.26. It intercepts all JSON-RPC calls from agents to underlying tool servers.
- **Protocol**: Exposes the standard Model Context Protocol (MCP) spec over gRPC/SSE, allowing external developer tools to integrate natively.
- **Security**: The Gateway strips and manages tokens. Agents are never given raw API keys. SSVID-based routing strictly enforces RBAC policies (e.g., ensuring only Finance agents access QuickBooks).

## Edge Cases
- **Malicious URLs/SSRF**: The Gateway explicitly blocks connections to loopback (`127.0.0.1`, `::1`), private (`10.0.0.0/8`, `192.168.0.0/16`), and link-local (`169.254.x.x`) IPs to prevent Server-Side Request Forgery and TOCTOU vulnerabilities.
- **Rate Limits**: If an external API (like GitHub) rate-limits the tool, the MCP Gateway interprets the 429 response and issues a backoff command to the calling agent to prevent thrashing.
- **Schema Drift**: If an external tool changes its payload schema, the MCP proxy's strict type-checking will reject invalid LLM-generated JSON, failing closed to prevent corrupt data entry.
