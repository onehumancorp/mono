# CUJ: VRAM Quota Management

**Persona:** DevOps / SRE Agent
**Context:** The organization's multi-agent workflows are scaling, risking out-of-memory errors on GPU nodes or runaway billing costs.
**Success Metrics:** Strict enforcement of GPU memory budgets without crashing active critical workflows.

## 1. User Journey Overview
Enforce strict department-level GPU memory budgets to prevent runaway compute costs during complex agent operations. The orchestration hub monitors real-time VRAM usage per agent group and gracefully pauses or terminates non-critical sub-agents when limits are reached.

## 2. Detailed Step-by-Step Breakdown
| Step | Action | System Trigger | Resulting State |
|------|--------|----------------|-----------------|
| 1 | Agent requests new sub-agent | Hub evaluates current VRAM usage | Usage vs Quota checked | Approval or denial state |
| 2 | Usage within quota | Sub-agent pod scheduled | VRAM allocated | Agent begins execution |
| 3 | Usage exceeds quota | Hub identifies idle/low-priority agents | Preemption logic triggered | Lower priority agent paused |
| 4 | VRAM freed | Original request retried | Sub-agent scheduled | Workflow continues |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: No Preemptible Agents
- **Detection**: Quota exceeded and all active agents are marked critical.
- **Auto-Recovery**: Request is queued until VRAM naturally frees up.
### 3.2 Scenario: Zombie Agents
- **Detection**: Agents that have stalled but hold VRAM.
- **Resolution**: A garbage collection routine forcefully evicts agents inactive for > 1 hour.

## 4. UI/UX Details
- **Compute Dashboard**: Real-time graphs showing VRAM utilization per department (e.g., Engineering, Marketing).

## 5. Security & Privacy
- **Resource Isolation**: Prevent noisy-neighbor issues by enforcing hard limits via K8s resource quotas and limits.
