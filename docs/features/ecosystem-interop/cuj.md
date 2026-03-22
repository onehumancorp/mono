# CUJ: Ecosystem Interoperability (Framework Adapters)

**Persona:** TPM Agent | **Context:** Integrating and orchestrating agents from diverse frameworks (OpenClaw, AutoGen, CrewAI, Semantic Kernel) within the unified One Human Corp (OHC) Agentic OS.
**Success Metrics:** Cross-framework task execution with zero state loss, successful MCP tool usage by external framework agents, and verified SPIFFE identity propagation.

## 1. User Journey Overview
The CEO spins up a hybrid agent swarm containing a native OHC Manager Agent, an AutoGen Planner, a CrewAI Writer, and an OpenClaw Researcher. The OHC control plane coordinates their communication, state, and tool access seamlessly via the Universal Agent interface.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | CEO uploads a multi-framework Skill Blueprint | Hub parses `blueprint.json` | Framework agents provisioned | Pods span up for AutoGen/CrewAI |
| 2 | CEO delegates a complex task | Hub creates LangGraph thread | Universal Agent Interface engaged | `thread_id` created in Postgres |
| 3 | AutoGen Planner maps subtasks | AutoGen adapter intercepts | OHC pub/sub event generated | AutoGen conversational state persisted to LangGraph |
| 4 | OpenClaw Researcher requests tool access | OpenClaw adapter calls MCP Switchboard | SPIFFE ID validated by MCP | Real-time state check-pointed to K8s CRD |
| 5 | CrewAI Writer executes role task | CrewAI adapter maps task to LangGraph | Document generation completed | Task execution logged to immutable audit trail |
| 6 | Swarm converges on goal | Manager agent distills the final output | Final state committed | Successful integration verified in CEO Dashboard |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Adapter Connection Loss
- **Detection**: Heartbeat ping to a framework adapter fails.
- **Auto-Recovery**: The OHC Hub immediately suspends the LangGraph thread, queues incoming events in Redis, and provisions a replacement adapter pod.
- **Manual Intervention**: If the adapter cannot be restored, the CEO is notified via the Human-in-the-Loop Handoff UI.

### 3.2 Scenario: Tool Access Violation
- **Detection**: A Semantic Kernel agent attempts an unauthorized MCP tool request.
- **Auto-Recovery**: The MCP Switchboard rejects the JSON-RPC call based on the agent's SPIFFE SVID. The SK adapter logs the failure to the LangGraph state.
- **Manual Intervention**: The CEO can update the agent's RBAC policy and replay the LangGraph checkpoint to allow the workflow to proceed.

### 3.3 Scenario: Context Payload Bloat
- **Detection**: AutoGen adapter detects the conversational array size exceeding the token threshold.
- **Auto-Recovery**: The adapter enforces immediate semantic summarization before committing to the LangGraph checkpointer.

## 4. Security & Privacy
- All intra-swarm communications across adapters require cryptographically verified SPIRE identities.
- Framework execution environments are containerized and strictly isolated to prevent lateral movement.
