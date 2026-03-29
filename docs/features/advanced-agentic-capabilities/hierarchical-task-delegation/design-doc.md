# Design Document: Hierarchical Task Delegation


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Executive Summary
**Objective:** Reduce context bloat and improve token efficiency by enabling agents to dynamically spawn specialized sub-agents for distinct parallelizable workstreams.
**Scope:** Implement the `DelegationService` within the Orchestration Hub and integrate it with the Kubernetes Operator to dynamically provision agent pods.

## 2. Architecture & Components
- **Manager Agent Node:** The LangGraph node representing the supervising agent. It utilizes a structured LLM output (e.g., JSON Schema) to decompose an epic into a list of `SubTask` objects.
- **Delegation Service:** A gRPC service in the Orchestration Hub that receives `SubTask` requests and communicates with the `ohc-operator`.
- **Kubernetes Operator (`ohc-operator`):** Reconciles the dynamically created `TeamMember` CRDs, spinning up new agent pods with specific roles (e.g., `SWE`, `QA`).
- **Asynchronous Event Bus:** Utilizes the existing Pub/Sub infrastructure for agents to report `TaskCompleted` or `TaskFailed` events back to the Manager Agent.

## 3. Data Flow
1. **Epic Deconstruction:** Manager Agent outputs `{"subtasks": [{"role": "SWE", "prompt": "Implement login component"}]}`.
2. **Provisioning:** The Hub receives the output and creates a temporary `TeamMember` CRD in the cluster.
3. **Execution:** The Operator spins up the Sub-Agent Pod. The Sub-Agent receives its isolated LangGraph thread ID and prompt.
4. **Completion:** Upon finishing, the Sub-Agent writes its final state to the Checkpointer and emits a `TaskCompleted` event.
5. **Aggregation:** The Manager Agent wakes up, reads the distilled checkpoint of the Sub-Agent, and proceeds.

## 4. API & Data Models
### 4.1 `SubTask` Protobuf Definition
```protobuf
message SubTask {
  string task_id = 1;
  string target_role = 2; // e.g., "frontend-swe"
  string instruction = 3;
  string parent_thread_id = 4;
}
```

## 5. Implementation Details
- **VRAM Quota Enforcement:** Before spawning a sub-agent, the `DelegationService` must check the organization's VRAM limits. If limits are exceeded, the task is queued.
- **Garbage Collection:** Once a sub-agent completes its task and the Manager Agent acknowledges it, the temporary `TeamMember` CRD is deleted by the Hub to free compute resources.
- **Security:** Sub-agents must not receive the full context of the epic, only the specific `instruction` provided, to prevent token waste and limit exposure.
