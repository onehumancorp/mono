# CUJ: Human-in-the-Loop (HITL) Handoff UI

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

**Persona:** Human Manager / CEO
**Goal:** Seamlessly take over a task from an AI agent, review context and visual evidence, and provide SPIFFE-gated approval to continue.
**Success Metrics:** Sub-second UI loading, full preservation of agent state, zero unauthorized approvals.

## 1. User Journey Overview
When an AI agent reaches an ambiguous decision point or requires a high-risk action approval, it triggers a "Warm Handoff" to a Human Manager, passing full context and pausing execution.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Agent encounters blocker | Agent calls `/api/handoffs` | Handoff created, Agent pauses | Handoff visible in CEO Dashboard |
| 2 | Manager reviews handoff | UI fetches handoff payload | Displays intent, state, and screenshots | UI diffs render correctly |
| 3 | Manager approves/rejects | UI submits approval via OIDC | Hub validates SPIFFE/OIDC claims | Approval recorded |
| 4 | Agent resumes execution | Hub triggers agent resume | LangGraph state transitions | Agent executes approved action |

## 3. Implementation Details
- **Architecture**: The handoff payloads are saved in Postgres and pushed to the CEO Dashboard via Server-Sent Events (SSE).
- **Stack**: Flutter/Dart frontend, Go backend, Postgres DB.
- **Visual Integration**: Screenshots and UI diffs are stored as blobs, with their URLs sent in the handoff payload.

## 4. Edge Cases
- **Simultaneous Approvals**: If two managers attempt to approve the same handoff concurrently, the database enforces a lock, rejecting the second attempt with a conflict error.
- **Handoff Timeout**: If a handoff is not actioned within a specific period (e.g., 2 hours), the UI flags it as "Expired" and the agent falls back to a safe default action or escalates.
- **Context Size Limit**: If the agent's LangGraph state is excessively large, it is summarized before being sent to the UI to maintain performance and prevent browser memory issues.
