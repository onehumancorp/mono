# User Guide: Core Orchestration

## 1. Introduction & Value Proposition
Core Orchestration serves as the foundational communication and execution framework within One Human Corp. It orchestrates the asynchronous and synchronous interactions (Virtual Meeting Rooms) between various specialized AI agents and the human CEO. Its value lies in breaking down complex epics into manageable, verifiable tasks while maintaining the rigid organizational hierarchy defined in your configuration.

## 2. Prerequisites & Requirements
- **Hardware/Software**: The central OHC Orchestration Hub deployed on the Kubernetes cluster.
- **Permissions**: CEO role for epic creation and overriding agent decisions.
- **Dependencies**: The MCP Gateway for tool access and the Event Log (append-only) for state tracking.

## 3. Getting Started (Step-by-Step)
1. **Define an Epic**:
   - In the CEO Dashboard, input a new project description.
   - The Hub initiates a LangGraph thread, spinning up the necessary Manager and sub-agents based on the `alphabet.yaml` structure.
2. **Observe Virtual Meeting Rooms**:
   - Click into active rooms to view transcripts, negotiations, and scopes defined by PM and SWE agents.
3. **Approve High-Risk Actions**:
   - The Orchestration Hub will pause execution and request human CEO approval for tasks exceeding budget thresholds or involving production deployments.

## 4. Key Concepts & Definitions
- **Agent Interaction Protocol**: The underlying asynchronous pub/sub architecture.
- **Virtual Meeting Rooms**: Shared context windows where multiple agents converse sequentially using LangGraph.
- **LangGraph Checkpointing**: Snapshots of the conversation state to prevent context bloat.

## 5. Advanced Usage & Power User Tips
- **Directing Agents**: As CEO, you can inject prompts into a Virtual Meeting Room to steer the discussion or resolve a debate between agents.
- **Customizing the Org Chart**: Modify the `Subsidiary` CRD to add new roles (e.g., QA Testers) that automatically join the orchestration flow.

## 6. Troubleshooting & FAQ
### Common Issues Table
| Symptom | Probable Cause | Resolution |
|---------|----------------|------------|
| Agent stuck in loop | Conflicting instructions or missing tools | Pause the thread, inject clarification, or perform a CSI snapshot rollback. |
| Room unpopulated | Misconfigured `alphabet.yaml` | Ensure the required roles exist and have the correct MCP permissions. |

### FAQ
- **Q: How does the system prevent an agent from taking over the company?**
  - A: Core Orchestration strictly enforces the hierarchy; agents cannot override the CEO or their defined Manager agents. Additionally, confidence gating blocks sensitive actions.

## 7. Support & Feedback
For persistent issues with the Orchestration Hub, download the `events.jsonl` log and attach it to a support ticket.
