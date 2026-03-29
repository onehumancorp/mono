# Test Plan: Tooling & MCP Gateway


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Tooling & MCP Gateway feature, ensuring it meets the requirements defined in the Design Document (`mcp-gateway.md`) and CUJs (`cuj-mcp-integration.md`, `cuj-skill-import.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components for parsing tool manifests, routing logic, and validating JSON/YAML Skill Packs.
- **Integration Testing:** Verify communication between the Hub, MCP Switchboard, and standard external tools (e.g., Gitea, Jira).
- **End-to-End (E2E) Testing:** Validate the complete flow of importing a skill and successfully invoking the new tool as an AI agent.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Skill Parser | Parse valid YAML "Skill Pack" | Roles and tools correctly instantiated | Pending |
| UT-02 | MCP Router | Route `git.commit()` tool call | Call redirected to correct Git server | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Gateway -> Tool | Gateway connects to GitHub MCP | Capabilities list returned | Pending |
| IT-02 | Agent -> Gateway| Agent executes a read command | Tool result returned as string | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | MCP Integration| Register new Tool | Tool available to agents < 1 min | Pending |
| E2E-02 | Skill Import | CEO uploads Legal Firm Pack | New roles appear in "Hire" modal | Pending |
| E2E-03 | Skill Execution| Agent uses newly imported skill | Success notification < 3s | Pending |

## 4. Edge Cases & Error Handling
- **Rate Limits:** Verify the gateway automatically throttles and queues calls if the external SaaS API rate-limits the agent.
- **Malformed Pack:** Ensure a YAML file with missing tool endpoints is rejected with descriptive validation errors.

## 5. Security & Safety
- **Tool Sandbox:** Ensure agents cannot use the `mcp-gateway` to escape the network namespace.
- **Secret Masking:** Ensure the gateway sanitizes any sensitive keys or tokens from the final LLM transcript.

## 6. Environment & Prerequisites
- A mock MCP Server container simulating a generic REST API.

## Implementation Details
- **Architecture**: The MCP Gateway tests utilize Go 1.26 table-driven unit tests. Integration tests spin up a mock HTTP server conforming to the Model Context Protocol (MCP) spec over gRPC/SSE, ensuring the Hub router correctly proxies JSON-RPC payloads.
- **Execution**: All tests run hermetically under Bazel 9.0.0 remote execution (`bazelisk test //...`). Tests enforce >95% coverage on the proxy layer and validation middleware.
- **Validation**: Strict validation of SSVID-based routing and RBAC ensures tests verify that an agent acting as a "SWE" cannot access tools assigned strictly to a "Finance" role.

## Edge Cases
- **Malicious URLs/SSRF**: The test suite actively attempts Server-Side Request Forgery by feeding the Gateway loopback (`127.0.0.1`, `::1`), private (`10.0.0.0/8`, `192.168.0.0/16`), and link-local (`169.254.x.x`) IPs. The Gateway is verified to explicitly block these and fail closed on DNS resolution errors.
- **Rate Limits**: Tests mock a `429 Too Many Requests` response from an external tool. They verify the MCP Gateway intercepts this and issues a backoff command to the LangGraph node to prevent LLM thrashing.
- **Schema Drift**: An integration test alters the mocked tool's response payload schema mid-execution. It verifies the MCP proxy's strict type-checking rejects the invalid LLM-generated JSON and fails closed to prevent database corruption.

### 3.4 Dynamic Tool Registration Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| DTR-01 | MCP Gateway | Dynamic search for capability | Correct capability mapped | Pending |
| DTR-02 | Agent Binding | Request tool bind with valid SVID | Tool successfully bound | Pending |
| DTR-03 | Agent Reject | Request bind with invalid SVID | Bind rejected securely | Pending |
