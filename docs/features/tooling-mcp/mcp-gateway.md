# Design Doc: External Tool Aggregation (MCP Gateway)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** Antigravity
**Status:** In Review
**Last Updated:** 2026-03-17

## 1. Overview
The OHC platform leverages the Model Context Protocol (MCP) to provide agents with a standardized, tool-agnostic interface to the external software ecosystem (GitHub, Jira, AWS, Slack, etc.). By abstracting vendor-specific APIs behind the MCP "Switchboard", we ensure that agents can focus on logic while the infrastructure handles authentication, routing, and rate-limiting.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Unified Tool Interface**: Enable agents to call `git.commit()` or `jira.create_ticket()` using the same RPC pattern.
- **Provider Agnosticism**: Switch between GitHub and Gitea, or AWS and GCP, without re-prompting or re-coding agents.
- **Dynamic Discovery**: Agents can query the `/api/mcp/tools` endpoint to discover available capabilities at runtime.
### 2.2 Non-Goals
- **Replacing Native Tooling**: We wrap existing tools; we do not build a new git client or Jira alternative.
- **End-User Tooling**: The OHC MCP Gateway is for *agent workload* access, not direct human-to-tool interaction.

## 3. Detailed Design

### 3.1 The MCP Switchboard Logic
The backend server (`srcs/dashboard/server.go`) maintains a `Registry` of available MCP servers. When an agent emits a tool call, the `Hub` routes it based on the `Category`:
```go
type MCPTool struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Category    string `json:"category"` // code, infra, docs, comms
    Status      string `json:"status"`   // available, busy, offline
}
```

### 3.2 Tool Mapping & Resolution
| Virtual Method | Mapping (Cloud) | Mapping (Self-Hosted) |
|----------------|-----------------|-----------------------|
| `git.create_pr`| GitHub GraphQL API | Gitea REST API |
| `infra.provision`| AWS CloudFormation | OpenStack / Proxmox |
| `comms.notify` | Slack Webhooks | Mattermost API |

### 3.3 Security & Governance
- **Role-Based Access (RBAC)**: Agents only "see" tools mapped to their role (e.g., SWEs can't access `billing.payout`).
- **Confidence Gating**: Destructive actions (e.g., `aws.terminate_instance`) are intercepted by the `Guardian Agent` and require a `POST /api/approvals` flow.
- **Audit Logging**: Every tool call, including its JSON-RPC payload and response, is logged to the CNPG `audit_logs` table for compliance.

## 4. Cross-cutting Concerns
### 4.1 Scalability
MCP servers are deployed as independent sidecars or centralized deployments in Kubernetes. The Gateway uses persistent gRPC/SSE connections to minimize tool-call latency (< 100ms overhead).
### 4.2 Error Handling
- **Circuit Breaking**: If an MCP server (e.g., `jira-mcp`) returns 5xx errors, the Gateway trips a circuit breaker and signals a `BLOCKER_RAISED` event to the Hub.
- **Fallback**: Limited support for "Secondary Providers" if the primary tool is unreachable.

## 5. Alternatives Considered
- **Direct Agent API Access**: Giving agents API keys for every service. **Rejected**: Unsafe (keys exposed in prompts), un-auditable, and requires constant code updates for every new tool.
- **Custom Plugin System**: Building our own binary plugin system. **Rejected**: MCP is an industry standard (pushed by Anthropic/Google/etc.), ensuring we can leverage community-built tool servers.

## 6. Implementation Roadmap
- **Phase 1**: Static registration and basic routing (COMPLETE).
- **Phase 2**: Dynamic discovery and mTLS-backed tool server connections (IN-PROGRESS).
- **Phase 3**: "Tool Marketplace" where users can install new community MCP servers (BACKLOG).

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.

### 3.4 Dynamic Tool Registration via MCP
Current frameworks tightly couple agents to hardcoded tool schemas. OHC utilizes our unified **MCP Gateway (Switchboard)**, allowing instant, secure, and dynamic tool synthesis across entire federated clusters.
- **Dynamic Registration**: Tools are discovered and registered via the SPIFFE-gated MCP Gateway.
- **Runtime Binding**: Agents can query the registry and dynamically bind to necessary tools based on task requirements, minimizing error loops caused by missing hardcoded configurations.
- **Seamless Integrations**: Supports a wide array of tools via standardized Model Context Protocols.
