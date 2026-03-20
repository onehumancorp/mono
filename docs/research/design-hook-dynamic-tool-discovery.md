# Design Hook: Dynamic Tool Discovery via MCP & SPIFFE

## Executive Summary
Agents in existing AI frameworks (like AutoGen and CrewAI) are tightly coupled to a static list of tools injected at initialization. When a novel problem arises that requires a new tool, the agent fails out, creating a high-friction loop.

This design proposes a native "Just-In-Time" tool synthesis workflow via One Human Corp's K8s/LangGraph architecture, leveraging the Model Context Protocol (MCP) Gateway and Zero-Trust SPIFFE/SPIRE authentication.

## The Architecture
We will migrate our static Switchboard gateway to a dynamically queryable registry pattern.

### Component 1: The Tool Registry API
1. **Dynamic Search Endpoint**: The Switchboard will expose an internal `/v1/tools/search` endpoint.
2. **Semantic Matching**: Agents can send a natural language description (e.g., "I need a tool to query AWS S3 bucket policies"). The Switchboard queries a vector database containing OpenAPI specifications of all internal and external registered MCP tools.
3. **Synthesis Response**: The API responds with the `tool_name`, `schema_url`, and required `RBAC` role to execute the tool.

### Component 2: Zero-Trust Tool Access (SPIFFE)
Tool discovery alone is a massive security risk without strict access controls.
1. **Role Negotiation**: When the agent discovers a new tool, it must request temporary access.
2. **SPIRE Integration**: The agent sends a request to the K8s Operator, which verifies the agent's current task (`thread_id`) against the tool's required permissions.
3. **Short-Lived SVID Generation**: If approved, SPIRE issues a short-lived x509 SVID (certificate) specifically authorizing the agent to call that tool for the next 15 minutes.

### Component 3: The LangGraph Node
We introduce a `DynamicToolDiscovery` node directly into our standard LangGraph execution template.
1. **Execution Failure Recovery**: If a standard tool call fails with a `ToolNotFound` or `NotImplemented` error, the execution graph automatically routes to the `DynamicToolDiscovery` node.
2. **Autonomous Fetch**: The agent pauses its main reasoning loop, fetches the new tool schema from the Switchboard, incorporates the strict JSON schema into its prompt, and re-enters the execution loop to retry the action.

## Token Efficiency & Performance Constraints
- **Lean Prompts**: Agents boot with an extremely minimal set of core tools (e.g., `search_tools`, `write_file`). All specific implementation tools (like the Jira API or Stripe API) are discovered lazily on a per-task basis.
- **Latency Optimization**: The `/v1/tools/search` endpoint must respond in sub-50ms using optimized vector retrieval (Sync.Pool buffer reuse), ensuring that the dynamic discovery loop does not noticeably slow down the execution graph.

## Next Steps
- Implement the `Semantic Tool Search & Routing` endpoint within the Switchboard.
- Test the LangGraph recovery node using a simulated "Agent finds a bug in AWS, needs CloudWatch tool" scenario.
- Finalize the integration between SPIRE and the MCP Gateway for short-lived tool certificates.
