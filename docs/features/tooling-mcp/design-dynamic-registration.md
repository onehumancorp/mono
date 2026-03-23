# Design Doc: Dynamic Tool Registration via MCP

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** $(date +"%Y-%m-%d")

## 1. Overview
The "Dynamic Tool Registration via MCP" feature enables One Human Corp's Multi-Agent Orchestrator to dynamically discover, validate, and bind to external Model Context Protocol (MCP) tools at runtime. This avoids the limitations of hardcoding OpenAPI schemas and allows the "Agentic OS" to extend its capabilities dynamically on a per-need basis across multiple agents.

## 2. Goals & Non-Goals
### 2.1 Goals
- Provide an endpoint (`/api/mcp/tools/register`) on the Dashboard Server to accept incoming MCP Tool registrations.
- Validate incoming Tool Registration requests securely using `SPIFFE ID`s (specifically calling `interop.ValidateSPIFFEID`).
- Persist dynamically registered tools in-memory alongside the default pre-configured list of MCP Tools.
- Ensure that agents querying available tools (`/api/mcp/tools`) automatically discover dynamically registered ones.

### 2.2 Non-Goals
- Persistent storage (database) of dynamic tools. Currently, dynamic tools will be tracked in-memory by the central orchestrator `Server`.
- Automatic polling/discovery of remote MCP endpoints (discovery must be push-based via the `/register` endpoint).

## 3. Detailed Design
### 3.1 Server Updates
The central orchestrator `Server` struct will be extended to include a `dynamicMCPTools []MCPTool` slice. This slice will be initialized with a baseline list of default tools (`defaultMcpTools`).

### 3.2 Registration Endpoint
A new POST endpoint `/api/mcp/tools/register` will be exposed.
The payload must include:
- `tool`: an object containing `id`, `name`, `description`, `category`, and `status`.
- `spiffeId`: A valid SPIFFE identifier indicating the trust domain and source of the tool request.

### 3.3 Security & Validation
Before any tool is accepted into the `dynamicMCPTools` slice, the `spiffeId` must be checked using `interop.ValidateSPIFFEID`. If the ID originates from an untrusted domain or does not meet OHC structure guidelines, the request will be rejected with HTTP 403 Forbidden.

## 4. Alternatives Considered
- **File-based Configuration Watcher**: Monitoring `MCP_BUNDLE_DIR` for YAML changes. Rejected because it requires shared filesystem access across K8s pods, which breaks the API-first loosely-coupled design.
- **Database Persistence**: Storing tools in a PostgreSQL registry. Rejected for this phase to reduce initial complexity, though this will likely be needed for multi-cluster federation later.
