# CUJ: Semantic Tool Search & Routing

**Persona:** CEO / Manager / Specialized Agent
**Context:** Initiating or managing workflows that require Semantic Tool Search & Routing within the Tool Discovery domain.
**Success Metrics:** Sub-50ms execution, zero "Agent Amnesia" occurrences, and fully validated SPIFFE identities.

## 1. User Journey Overview
When a complex Epic is initiated, the system automatically provisions necessary sub-agents. During execution, the agents leverage Semantic Tool Search & Routing to optimize their workflows, resolve ambiguities, and maintain state without exceeding token quotas.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action / System Trigger | Resulting State | Verification |
|------|------------------------------|-----------------|--------------|
| 1 | Agent encounters requirement for Semantic Tool Search & Routing | Feature module initialized | Log entry created in Postgres |
| 2 | Component requests context or execution | Validation via SPIFFE | Valid SVID confirmed |
| 3 | Core logic of Semantic Tool Search & Routing processes task | Graph state updated | New Checkpoint created |
| 4 | Task completes successfully | Semantic vector updated | CSI Snapshot synced |

## 3. Edge Cases & Recovery
### 3.1 Operation Timeout
- **Detection**: The Semantic Tool Search & Routing operation exceeds threshold latency.
- **Recovery**: Halts operation, falls back to general reasoning, and flags for review.

### 3.2 Resource Exhaustion
- **Detection**: VRAM quota or context token limit approached.
- **Recovery**: Triggers background distillation and checkpointing before proceeding.

## 4. UI/UX Details
- **Dashboard Visibility**: The CEO Dashboard clearly displays the status of Semantic Tool Search & Routing operations within the active thread's visualization graph.
- **Alerting**: Non-critical warnings are batched; critical failures instantly trigger a HITL Handoff.
