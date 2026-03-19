# CUJ: Warm Handoff to Human Manager

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** AI Agent → Human Manager
**Goal:** An agent that cannot complete a task escalates gracefully to a human.
**Success Metrics:** Human manager receives full context and takes over seamlessly.

## Context
An agent encountered an ambiguous goal or a technical blocker it cannot resolve alone.

## Journey Breakdown
### Step 1: Agent Triggers Handoff
- **User Input:** N/A (Agent Action).
- **System Action:** `POST /api/handoffs` is called with intent and failure context.
- **Outcome:** Handoff package is created.

### Step 2: Human Acknowledges Handoff
- **User Input:** Manager clicks "Acknowledge" in the Handoffs queue.
- **System Action:** UI displays the handoff details (intent, failed attempts, current state).
- **Outcome:** Manager is now "in charge" of the task.

## Error Modes & Recovery
### Failure 1: Lost Handoff
- **System Behavior:** Handoff remains in `PENDING` status.
- **Recovery Step:** Automated alert triggers for the human manager.

## Security & Privacy Considerations
- Handoff data contains snapshots of agent memory/state, which may contain sensitive info.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.

## Edge Cases
- **Timeout:** Task aborts and escalates to human CEO.
- **Rate Limit:** Agent backoffs using exponential retry.
- **Loss of Context:** Supervisor agent reconstructs state from snapshot.
