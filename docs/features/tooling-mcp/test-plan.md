# Test Plan: Tooling & MCP Gateway

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
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
