# CUJ: Org Snapshot and Restore

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
