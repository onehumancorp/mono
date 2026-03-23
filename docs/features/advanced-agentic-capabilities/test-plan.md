# Test Plan: Advanced Agentic Capabilities

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-20

## 1. Overview
A high-level summary of the testing strategy for the Advanced Agentic Capabilities feature set, including Stateful Episodic Memory, Dynamic Tool Discovery, Native Vision, Hierarchical Task Delegation, and Stateful Execution Graph integration.

## 2. Test Strategy
- **Unit Testing:** Focus on verifying the LangGraph checkpointer state logic and semantic distillation triggers.
- **Integration Testing:** Verify communication between the OHC Hub, Postgres DB, and MCP Gateway for dynamic tool binding.
- **End-to-End (E2E) Testing:** Validate the entire hierarchical task delegation, multimodal parsing, and CSI snapshot recovery flow.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | LangGraph Checkpointer | Verify state is snapshotted iteratively | State transition logged | Pending |
| UT-02 | Semantic Distillation | Validate summarization payload size | Payload size reduced | Pending |
| UT-03 | Hierarchical Delegation | Verify VRAM quota allocation | Quota enforced during `/scale` | Pending |
| UT-04 | Native Vision | Parse mock image to text | Correct text extracted | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> DB | Verify Postgres checkpointer | `thread_id` and state stored | Pending |
| IT-02 | Hub -> MCP Gateway | Dynamic tool registration | SPIFFE SVID validated | Pending |
| IT-03 | Sub-agent -> Hub | Retrieve semantic vector search | Correct context returned | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Advanced Capabilities | CEO spawns complex task, sub-agents dynamically discover tools, maintain state across checkpoints | Task completed within token limits | Pending |
| E2E-02 | Hallucination Recovery | CEO halts cyclic execution, reviews checkpoints, and forces CSI rollback | Rollback successful, agent redirected | Pending |

## 4. Edge Cases & Error Handling
- **Context Limit Breaches**: Ensure the system gracefully handles scenarios where context payloads exceed maximum token limits.
- **Tool Resolution Failures**: Verify sub-agents fall back to general reasoning or escalate when dynamic tool discovery fails.
- **VRAM Exhaustion**: Confirm the system prevents spawning sub-agents when the VRAM quota is depleted.

## 5. Security & Safety
- **SPIFFE Validation**: Test that all dynamically registered MCP tools undergo strict SPIFFE authentication.
- **Isolation**: Verify sub-agent contexts remain isolated and do not leak data across separate `thread_id` sessions.

## 6. Implementation Details
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox.
- **Mocks**: External MCP endpoints and LLM reasoning engines are mocked for deterministic testing.
- **Validation**: Strict enforcement of >95% test coverage.

### 3.4 Handoff UI Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| HITL-01 | Hub -> DB | Verify Handoff package creation | Handoff state saved as PENDING | Pending |
| HITL-02 | Webhook | Validate Slack/Mattermost hook | Valid payload sent | Pending |
| HITL-03 | Approval Gate | Verify cryptographically signed token | Task resumes only on valid token | Pending |
