# User Guide: Core Orchestration Engine


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Introduction & Value Proposition
The Core Orchestration Engine is the "brain" of One Human Corp. It transforms high-level CEO mandates into actionable, multi-agent workflows. By automating task delegation, context propagation, and meeting management, it allows a single human to operate with the capacity of a 50-person department.

**Key Benefits:**
- **Reduced Context Switch**: Agents manage sub-tasks autonomously.
- **Persistent Memory**: Every decision is captured in meeting transcripts.
- **Scalable Execution**: Burst capacity by hiring specialists on-demand.

## 2. Prerequisites & Requirements
- **OHC Tier**: Professional or Enterprise for multi-department support.
- **Provider API Keys**: Configured in Settings (Gemini, OpenAI, or Anthropic).
- **Identity Service**: `spire-server` must be healthy for agent-to-agent communication.

## 3. Getting Started (Core Workflow)
1. **[Mission Control]**: Open the dashboard and locate the "Active Mission" bar.
2. **[Defining a Goal]**: Enter a complex objective like "Evaluate our Q3 security posture and draft a remediation plan."
3. **[Automated Scoping]**: The Product Manager agent will intercept the goal, create a `MeetingRoom`, and invite necessary specialists (e.g., Security Engineer, QA).
4. **[Verification]**: Monitor the transcript via `GET /api/meetings`. You can "Intervene" at any time by sending a message as `CEO`.

## 4. Key Concepts
- **The Hub**: A thread-safe Go registry that maintains the current state of all `Agents` and `MeetingRooms`.
- **Message Types**:
    - `task`: Actionable work items.
    - `status`: Progress updates (Blocked, In-Progress, Done).
    - `handoff`: Technical context package for human-in-the-loop escalation.
- **Role Profiles**: Immutable archetypes (CEO, SWE, PM) that define the "personality" and prompt-base of each agent.

## 5. Advanced Usage & Power User Tips
- **Hierarchy Overrides**: Use the `managerId` field in the Agent settings to create custom reporting lines (e.g., have 3 SWEs report to a specific "Lead Agent").
- **Custom Skill Packs**: Import `.json` skill packs via `/api/skills/import` to give agents domain-specific knowledge (e.g., "HIPAA Compliance" or "Advanced Kubernetes").
- **A2A Latency Tuning**: Adjust `REDIS_PUB_SUB_BUFFER` to optimize message delivery speed in high-concurrency environments.

## 6. Troubleshooting & FAQ
### Common Issues Table
| Symptom | Probable Cause | Resolution |
|---------|----------------|------------|
| Agent stuck in `IDLE` | No pending tasks in Hub | Check if the PM agent has broken down the main goal. |
| Message delivery failure | Redis connection lost | Verify `REDIS_ADDR` env var and pod status. |
| Identity Error | SPIRE SVID expired | Trigger a manual rotation via `/api/identities/rotate`. |

### FAQ
- **Q: Can I limit agent spend?**
    - A: Yes, set a `MaxMonthlyUSD` in the Billing Settings to auto-suspend the Orchestration Engine.
- **Q: How do I backup my org?**
    - A: Use the "Snapshot" feature which bundles all Hub state into a PG-backed record.

## 7. Support
For technical issues, please file a ticket via the "OHC Support" portal or contact your dedicated Engineering Director agent.

## Implementation Details
- **Architecture**: A Go 1.26 monolith operating within the `Hub` module handles the orchestration loop. It uses goroutines for lightweight concurrent task management and LangGraph checkpointers for saving/loading context.
- **Data Model**: The K8s Custom Resource Definition (`HoldingCompany`) natively syncs with internal Postgres records tracking current active directives and agent assignments.
- **Messaging Pipeline**: Multi-agent delegation pushes state messages to Redis. The Hub reads from Redis to evaluate if dependencies are met for downstream tasks.

## Edge Cases
- **Cyclic Dependencies**: If a PM mistakenly assigns two tasks that depend on each other, the Orchestration Engine's DAG evaluator detects the cycle and raises a `DependencyCycleError` for human review.
- **Node Failures**: If an underlying K8s node dies, the agent pod restarts on a new node and recovers its exact position in the workflow by rehydrating state from the Postgres append-only log.
- **Unreachable LLM Provider**: If the external AI API (e.g., OpenAI/Gemini) is down, the engine pauses active workflows and places them in a `ProviderRetry` queue rather than silently dropping tasks.
