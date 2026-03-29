# CUJ: Hierarchical Task Delegation


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Persona:** Manager Agent | **Context:** Orchestrating a complex epic requiring multiple specialized sub-tasks.
**Success Metrics:** Sub-agents spawned correctly, tasks parallelized efficiently without token bloat, and sub-task results integrated into the main workflow.

## 1. User Journey Overview
When a Manager Agent is assigned a complex epic (e.g., "Build and deploy a new frontend dashboard"), it recognizes that the scope exceeds a single context window or requires parallel, specialized execution. The Manager Agent dynamically spawns sub-agents (e.g., UI/UX Agent, Frontend SWE Agent, QA Agent), assigns them narrow, specific sub-tasks, and orchestrates their execution. Once sub-tasks are complete, the Manager Agent aggregates the results and finalizes the epic.

## 2. Detailed Step-by-Step Breakdown

| Step | User/Agent Action | System Trigger | Resulting State | Verification |
|------|-------------------|----------------|-----------------|--------------|
| 1 | Manager receives complex epic | Sub-task decomposition logic initiated | Epic broken down into discrete JSON objects | JSON array of sub-tasks |
| 2 | Manager delegates Sub-Task A | Hub calls `/scale` API to spawn Sub-Agent | Specialized Sub-Agent (e.g., Frontend SWE) spawned | K8s Pod running |
| 3 | Sub-Agent executes Sub-Task A | LangGraph execution thread created for Sub-Agent | Code written and pushed to branch | Git commit hash logged |
| 4 | Sub-Agent finishes Sub-Task A | Event `TaskCompleted` published to message bus | Manager Agent notified of completion | Event log entry |
| 5 | Manager aggregates results | Semantic Vector Search queries sub-agent checkpoints | Results distilled into Manager's context | Distilled summary available |
| 6 | Manager finalizes epic | Manager marks epic as `RESOLVED` | Epic completed successfully | CEO Dashboard update |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Sub-Agent Failure
- **Detection**: Sub-Agent encounters a terminal error or hits a retry limit on a specific tool call.
- **Auto-Recovery**: Sub-Agent terminates and publishes a `TaskFailed` event. Manager Agent analyzes the failure reason and either respawns the Sub-Agent with modified instructions or escalates to the human CEO via the Handoff UI.
### 3.2 Scenario: Context Window Overflow
- **Detection**: A Sub-Agent's context exceeds the defined VRAM quota.
- **Resolution**: The orchestration hub triggers the Semantic Distillation Worker to summarize older checkpoints, freeing up the context window before resuming the Sub-Agent's thread.

## 4. UI/UX Details
- **Dashboard Integration**: The CEO Dashboard displays a visual tree graph showing the Manager Agent at the root and all spawned Sub-Agents as branches, updating their status in real-time.

## 5. Security & Privacy
- **Sub-Agent Scoping**: Sub-agents inherit only a strict subset of the Manager Agent's permissions via narrowed SPIFFE SVIDs, ensuring least privilege access during execution.
