# Design Doc: The "One Human Corp" Marketplace


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The "One Human Corp" Marketplace is a community-driven ecosystem where users can buy, sell, and share highly specialized agents (e.g., a "TikTok Virality Expert Agent"), custom organizational templates (Skill Blueprints), and unique tool integrations (MCP Servers). It is the critical enabler for Phase 4 of the roadmap, providing a plug-and-play AI talent exchange.

## 2. Goals & Non-Goals
### 2.1 Goals
- Enable the discovery and import of specialized AI agents and organizational templates.
- Provide a standardized format (Skill Pack/Blueprint) for packaging agents and MCP tool mappings.
- Integrate securely with the OHC Hub Registry to instantiate imported blueprints into active workflows.

### 2.2 Non-Goals
- Monetization and fiat payment processing in v1 (initially a free, community-driven exchange).
- Providing raw LLM model weights (the marketplace distributes prompts, roles, and tool integrations, not foundational models).

## 3. Implementation Details
- **Architecture**: A public-facing Marketplace Registry (web UI and API) where creators upload `blueprint.json` files containing Role Profiles, Domain Prompts, and MCP definitions.
- **Integration**: The Dashboard UI will have a "Marketplace" tab. When a CEO clicks "Install", the frontend calls the Hub's `ImportSkillBlueprint` API, which validates the JSON schema and registers the new Role Profiles into the local Postgres database.
- **Verification**: All uploaded templates undergo automated static analysis to ensure they do not contain malicious MCP commands or SSRF vulnerabilities (leveraging the existing SSRF prevention mechanisms).

## 4. Edge Cases
- **Version Conflicts**: If a CEO installs a template that conflicts with an existing Role Profile name (e.g., `swe_agent`), the system automatically namespaces the new role (e.g., `vendorX/swe_agent`) to prevent overwriting active enterprise configurations.
- **Missing Tooling**: If an imported agent requires an MCP tool not present in the local cluster (e.g., `jira-mcp`), the Hub will flag the agent in an `INCOMPLETE` status until the human manager provides the necessary credentials/tooling via the Switchboard.