# Test Plan: Ecosystem Interoperability (Framework Adapters)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-22

## 1. Overview
A high-level summary of the testing strategy for the Ecosystem Interoperability feature set, validating the Universal Agent interface and native framework adapters (OpenClaw, AutoGen, CrewAI, Semantic Kernel) in `srcs/interop/`. It ensures seamless multi-agent swarm state synchronization via LangGraph checkpointers and secure MCP access.

## 2. Test Strategy
- **Unit Testing:** Focus on verifying individual adapter state mapping, message translation logic, and context summarization boundaries.
- **Integration Testing:** Validate end-to-end communication between the OHC Hub pub/sub system and simulated framework adapters, including K8s CRD updates.
- **End-to-End (E2E) Testing:** Simulate a heterogeneous swarm executing a complex multi-step workflow via the MCP Switchboard and Postgres checkpointer.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | `autogen_adapter.go` | Translates AutoGen conversational payload to OHC event | Proper `AgentMessage` generated | Pending |
| UT-02 | `crewai_adapter.go` | Maps CrewAI task result to LangGraph state checkpoint | JSON state schema successfully validated | Pending |
| UT-03 | `openclaw_adapter.go` | Triggers K8s CRD update upon state change | `patch` payload generated correctly | Pending |
| UT-04 | `semantickernel_adapter.go` | Maps SK function call request to MCP JSON-RPC format | Correct JSON-RPC schema returned | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Adapter -> Hub | Verify OHC Pub/Sub routing across frameworks | Events routed strictly via SPIFFE SVID | Pending |
| IT-02 | Adapter -> DB | LangGraph state persistence for third-party agents | Checkpoints stored correctly in Postgres | Pending |
| IT-03 | Adapter -> MCP Gateway | Framework agent triggers an MCP tool call | MCP request processed and response returned | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Ecosystem Interoperability | CEO deploys multi-framework swarm. Agents successfully complete delegated task via Universal Agent interface | Task resolved, state snapshotted via CSI | Pending |
| E2E-02 | Adapter Failure | Simulate framework adapter crash mid-task | Hub correctly suspends thread and recovers | Pending |

## 4. Edge Cases & Error Handling
- **Context Limit Breaches**: Ensure `autogen_adapter.go` automatically forces semantic distillation if payloads exceed token boundaries.
- **Malformed Payloads**: Ensure invalid JSON responses from the third-party frameworks do not crash the OHC Hub (fail gracefully).
- **Unauthorized MCP Access**: Test that an agent adapter lacking specific MCP privileges correctly fails closed at the Switchboard.

## 5. Security & Safety
- **SPIFFE Validation**: Test that all adapter instances undergo strict SPIFFE authentication before accessing the Hub or MCP Switchboard.
- **Isolation**: Verify that adapter payloads do not leak sensitive state data outside of their designated `thread_id`.

## 6. Implementation Details
- **Execution**: Run via `bazelisk test //srcs/interop/...` under the Bazel sandbox.
- **Mocks**: External framework agent executions are thoroughly mocked using the existing test harness.
- **Validation**: Strict enforcement of >95% test coverage.
