# CUJ: Org Snapshot and Restore


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** Org Admin / DR Engineer
**Goal:** Create a snapshot of the current org state and later restore from it.
**Success Metrics:** 100% state fidelity after restore. Restore completes in <5s.

## Context
A major reorganisation is planned, and the admin needs a "safe point" to return to if things go wrong.

## Journey Breakdown
### Step 1: Create Snapshot
- **User Input:** Admin clicks "Create Snapshot" and labels it "Before Reorg".
- **System Action:** `POST /api/snapshots` creates a full state dump in PG.
- **Outcome:** Snapshot ID is returned.

### Step 2: Restore Snapshot
- **User Input:** After an unsuccessful reorg, Admin clicks "Restore" on the "Before Reorg" snapshot.
- **System Action:** `POST /api/snapshots/restore` overwrites current memory with snapshot data.
- **Outcome:** Org name and agent list revert to the previous state.

## Error Modes & Recovery
### Failure 1: Corrupt Snapshot
- **System Behavior:** Restore fails with "Invalid checksum".
- **Recovery Step:** Admin picks an earlier snapshot.

## Security & Privacy Considerations
- Snapshots contain the entire org state, including sensitive transcripts.
- Only users with "Snapshot" permissions can restore.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.

## Edge Cases
- **Timeout:** Task aborts and escalates to human CEO.
- **Rate Limit:** Agent backoffs using exponential retry.
- **Loss of Context:** Supervisor agent reconstructs state from snapshot.
