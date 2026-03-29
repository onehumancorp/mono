# CUJ: Marketplace Journey

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
A high-level view of how the human CEO navigates the "One Human Corp" Marketplace to find, import, and deploy a specialized community-created AI agent template.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to "Marketplace" | Dashboard calls `/api/marketplace/list` | UI displays grid of available templates | Grid populates |
| 2 | Select "Marketing Agency" pack | CEO clicks "Install" | UI calls `/api/marketplace/import` with `pack_id` | `Downloading...` |
| 3 | Resolve missing tools | UI prompts for missing Slack integration | Hub pauses import until tool is linked | Setup wizard appears |
| 4 | Confirm Import | UI confirms successful validation | New agents registered in Hub DB | Agents visible in Org Chart |

## 3. Implementation Details
- **Architecture**: A centralized index of `SkillBlueprints` (JSON) that the `ohc-operator` fetches and provisions as `RoleProfile` CRDs in the local Kubernetes cluster.
- **Stack**: Go backend serving Flutter frontend pages. External Marketplace Index via HTTP APIs.
- **Security Check**: The downloaded payload is parsed for malicious scripts and strictly validated against the internal schema before being saved to Postgres.

## 4. Edge Cases
- **Network Failures**: If the external Marketplace is unreachable during import, the Dashboard displays an error and allows a retry, caching any partial downloads.
- **Dependency Missing**: A purchased template requires an MCP tool (e.g., Salesforce) not installed. The Hub places the new agent in `WAITING_FOR_TOOLS` state until the CEO configures the necessary connections.
- **Duplicate Names**: Importing a pack with roles already existing in the cluster causes the backend to automatically prefix the new roles (e.g., `marketplace_v1/growth_hacker`), avoiding clashes with custom organizational configurations.