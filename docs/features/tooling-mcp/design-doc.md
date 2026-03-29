# Design Doc: Tooling & MCP Gateway


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The MCP (Model Context Protocol) Gateway enables AI agents to interface with external tools (GitHub, AWS, Jira) consistently via standard JSON-RPC without exposing API secrets directly to the LLM context.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Standardized Access**: Unified protocol for any external API integration.
- **Secret Isolation**: The Gateway intercepts and appends authorization headers.
### 2.2 Non-Goals
- Building proprietary integrations (we rely on the MCP ecosystem).

## 3. Implementation Details
- **Architecture**: The OHC MCP Gateway serves as a unified proxy written in Go 1.26. It intercepts all JSON-RPC calls from agents to underlying tool servers.
- **Protocol**: Exposes the standard Model Context Protocol (MCP) spec over gRPC/SSE, allowing external developer tools to integrate natively.
- **Security**: The Gateway strips and manages tokens. Agents are never given raw API keys. SSVID-based routing strictly enforces RBAC policies (e.g., ensuring only Finance agents access QuickBooks).

## 4. Edge Cases
- **Malicious URLs/SSRF**: The Gateway explicitly blocks connections to loopback (`127.0.0.1`, `::1`), private (`10.0.0.0/8`, `192.168.0.0/16`), and link-local (`169.254.x.x`) IPs to prevent Server-Side Request Forgery and TOCTOU vulnerabilities.
- **Rate Limits**: If an external API (like GitHub) rate-limits the tool, the MCP Gateway interprets the 429 response and issues a backoff command to the calling agent to prevent thrashing.
- **Schema Drift**: If an external tool changes its payload schema, the MCP proxy's strict type-checking will reject invalid LLM-generated JSON, failing closed to prevent corrupt data entry.