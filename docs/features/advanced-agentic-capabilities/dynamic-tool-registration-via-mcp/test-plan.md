# Test Plan: Dynamic Tool Registration via MCP

## 1. Testing Strategy
Validate the dynamic synthesis and injection of tool schemas, verify secure routing via the MCP Switchboard, and ensure strict parameter validation.

## 2. Test Cases
### 2.1 E2E Integration Test: Successful Binding and Execution
- **Setup:** A mock `math-mcp` server registered in the MCP Gateway providing an `add_numbers` tool. The agent's default prompt does *not* include this tool.
- **Action:** Provide the agent a prompt requiring addition (e.g., "What is 55 + 45?").
- **Assertion:** Verify the agent successfully calls `query_registry("addition")`, receives the schema, binds it, and executes `add_numbers(55, 45)`.
- **Assertion:** Verify the correct result (100) is returned and the execution was successful.

### 2.2 Edge Case: Tool Not Found
- **Setup:** A clean MCP registry with no registered tools.
- **Action:** Agent queries for a "database search" capability.
- **Assertion:** Verify the Hub returns an empty list or specific error, and the agent degrades gracefully (e.g., returns a message stating it lacks the tool or escalates).

### 2.3 Edge Case: Unauthorized Access (SPIFFE)
- **Setup:** A `finance-mcp` server restricted to the "Finance Director" role.
- **Action:** A "SWE Agent" dynamically binds the `process_payment` tool and attempts execution.
- **Assertion:** Verify the Hub intercepts the call, validates the SWE Agent's SPIFFE ID against the tool's `AuthRole`, and instantly drops the request with an HTTP 403 Forbidden.

### 2.4 Edge Case: Strict Schema Validation
- **Setup:** A `github-mcp` tool expecting a required string `repo_name`.
- **Action:** Agent executes the tool but includes an unknown field `force_push: true` in the JSON payload.
- **Assertion:** Verify the Hub's JSON decoder, utilizing `dec.DisallowUnknownFields()`, rejects the payload and returns an error without forwarding it to the MCP server.

## 3. Automation & CI/CD
- All unit and integration tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for `srcs/orchestration/mcp_gateway.go`.
- Avoid arbitrary `time.Sleep()` for async tool executions; strictly use deterministic polling loops.
