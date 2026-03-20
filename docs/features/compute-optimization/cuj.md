# CUJ: Compute Optimization Journey

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO utilizes the Compute Optimization layer to ensure their highest-priority projects are allocated the most expensive LLM models and hardware.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Tag project "High Priority" | Dashboard sends `PATCH /api/projects/1` | Project metadata updated | Priority visible |
| 2 | Assign SWE Agent | Hub provisions new agent | `ohc-operator` reads priority | Node affinity set |
| 3 | Monitor GPU usage | CEO reviews Compute tab | Metrics scraped via OpenTelemetry | Heatmap visible |
| 4 | Throttle department | CEO edits quota | Scheduler evicts pods | Agents restart on CPU nodes |

## 3. Implementation Details
- **Architecture**: The OHC Kubernetes Operator watches for `TeamMember` resource changes and applies `nodeSelector` and `tolerations` dynamically based on the project priority.
- **Stack**: Go 1.26, OpenTelemetry for scraping VRAM usage.
- **State Serialization**: Checkpointers allow seamless movement of agents between GPU and CPU nodes during throttling.

## 4. Edge Cases
- **Quota Thrashing**: Rapidly changing the VRAM quota might cause the scheduler to repeatedly evict and spin up pods. The system uses a 5-minute cooldown window on quota changes.
- **Partial Checkpoints**: If a GPU node dies mid-inference before the event log writes to Postgres, the agent will resume from the last known complete checkpoint, losing at most a few seconds of 'thought'.
- **Network Bandwidth Issues**: Distributing heavily weighted model inferences across zones could cause latency; the scheduler prefers co-locating interacting agents on the same physical rack.