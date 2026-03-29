# Design Doc: Agent Delegate Mode


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
Agent Delegate Mode allows an AI agent to act as a routing proxy. It inspects an incoming task, selects the best-fit specialist agent from the registry, forwards the task, and surfaces the result back to the originating caller.

## 2. Goals & Non-Goals
### 2.1 Goals
- Enable dynamic task delegation between agents without human intervention.
- Provide a clear interface `DelegateTask` on the orchestration `Hub`.
### 2.2 Non-Goals
- Complex multi-agent negotiations during delegation.

## 3. Implementation Details
- **Architecture**: A new method `DelegateTask(fromAgentID, toAgentID string, task Message) error` will be added to the `Hub` struct.
- **Routing**: The method will update the task's `FromAgent` and `ToAgent` fields and route it via the existing `Publish` method.

## 4. Edge Cases
- **Agent Not Found**: If either the source or destination agent is not registered, the delegation will fail with an error.