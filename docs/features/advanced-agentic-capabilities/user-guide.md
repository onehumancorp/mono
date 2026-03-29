# User Guide: Advanced Agentic Capabilities


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Introduction & Value Proposition
The Advanced Agentic Capabilities feature set represents a paradigm shift in how One Human Corp handles complex, long-running workflows. It directly addresses "Agent Amnesia" and static tool limitations by introducing Stateful Episodic Memory, Dynamic Tool Discovery, and Native Multimodal Reasoning. For the CEO, this means massive epics can be delegated with confidence, knowing agents will retain context across disjointed sessions, autonomously bind to required tools, and effectively parse visual data without context bloat or runaway token costs.

## 2. Prerequisites & Requirements
- **Hardware/Software**: A standard web browser for the CEO Dashboard; backend Kubernetes cluster with CSI snapshot capabilities and persistent storage (PostgreSQL for checkpointers).
- **Permissions**: CEO role required for manual rollback or VRAM quota adjustment.
- **Dependencies**: The core OHC Orchestration Hub and MCP Gateway must be deployed and operational.

## 3. Getting Started (Step-by-Step)
1. **Initiate an Epic**:
   - In the CEO Dashboard, click "New Epic" and provide a high-level goal (e.g., "Build and deploy a new frontend dashboard").
   - The system automatically spawns a Manager agent, which then provisions necessary sub-agents and allocates VRAM.
2. **Monitor Execution**:
   - Navigate to the "Active Workflows" tab. Here, you can visualize the hierarchy of sub-agents and their current LangGraph checkpoints.
   - If a sub-agent needs a specific tool not currently available, observe the "Dynamic Tools" panel to see the MCP Gateway register and bind the new tool via SPIFFE in real-time.
3. **Multimodal Grounding**:
   - For tasks requiring UI validation, sub-agents will automatically request multimodal scans. You can view the parsed results in the agent's activity log.
4. **Historical Context Retrieval**:
   - If an agent needs past context, it automatically performs a semantic vector search against distilled checkpoints. This is transparent but visible in the agent's trace logs.

## 4. Key Concepts & Definitions
- **Stateful Episodic Memory**: Instead of growing the context window indefinitely, the execution path is periodically snapshotted (LangGraph checkpoints) and older states are distilled into vector embeddings for semantic retrieval.
- **Dynamic Tool Discovery**: Agents are not bound to hardcoded OpenAPI schemas. They query the MCP Gateway at runtime to find and securely bind to the tools they need.
- **Hierarchical Delegation**: A Manager agent dynamically scales out by spawning sub-agents with narrow, focused contexts, preventing context bloat.
- **CSI Snapshots**: Kubernetes-level snapshots that allow the CEO to roll back the entire state of a subsidiary or workflow instantly.

## 5. Advanced Usage & Power User Tips
- **VRAM Quota Adjustments**: If a task is stalling due to context limits, the CEO can manually adjust the VRAM quota for a specific department via the settings panel.
- **Forcing Rollbacks**: In the event of a "hallucination loop," the CEO can halt the execution graph and select a previous LangGraph checkpoint to trigger a full CSI snapshot rollback, safely resetting the workflow.
- **Custom Tool Registration**: Users can pre-register custom MCP bundles if they know a specific esoteric tool will be needed, though the system will attempt to discover it dynamically.

## 6. Troubleshooting & FAQ
### Common Issues Table
| Symptom | Probable Cause | Resolution |
|---------|----------------|------------|
| Agent fails to bind tool | MCP Gateway cannot locate the requested tool | Manually verify the tool is registered in the OHC registry or provide explicit instructions. |
| Workflow halts with "Quota Exceeded" | VRAM limits reached for the spawned sub-agents | Increase VRAM quota or trigger a semantic distillation to free up memory. |
| Hallucination Loop detected | Cyclic redundancy in the LangGraph execution path | Halt execution, perform a CSI rollback to a previous checkpoint, and refine the initial prompt. |

### FAQ
- **Q: Are my custom tools secure?**
  - A: Yes, all dynamic tool bindings are authenticated and authorized via SPIFFE/SPIRE, ensuring zero-trust security even for dynamically discovered endpoints.
- **Q: How does Episodic Memory save costs?**
  - A: By distilling older checkpoints into semantic embeddings, the active context window remains small. Agents only pull in historical data when strictly necessary, drastically reducing LLM token burn.

## 7. Support & Feedback
If you encounter persistent issues, please file a bug report via the CEO Dashboard's "Support" tab or directly in the One Human Corp issue tracker, including the relevant `thread_id` and checkpoint logs.
