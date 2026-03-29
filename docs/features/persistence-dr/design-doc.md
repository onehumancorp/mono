# Design Doc: Persistence & Disaster Recovery (DR)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Persistence and DR framework ("Snapshot Fabric") enables One Human Corp to save its entire organizational state at any point in time and restore it deterministically within 5 seconds.

## 2. Goals & Non-Goals
### 2.1 Goals
- Unified state persistence across the entire multi-agent framework.
- "Undo Button" for complex corporate decisions (hiring, strategy changes).
### 2.2 Non-Goals
- Granular, message-level "undo" (snapshots work on an epoch level).

## 3. Implementation Details
- **Architecture**: The Snapshot Fabric leverages Kubernetes CSI (Container Storage Interface) volume snapshots combined with Postgres `pg_dump`/`pg_restore` for database state.
- **State Management**: LangGraph checkpointers serialize the exact multi-agent state (memory, pending tasks, active tools) into the append-only `events.jsonl` log. Restoring a snapshot replays or truncates this log deterministically.
- **Execution**: Orchestrated via Go 1.26 in the OHC Hub. The Operator pauses the `HoldingCompany` CRD reconciliation loop during the restoration process.

## 4. Edge Cases
- **In-Flight Tool Operations**: If an agent is executing a long-running external API call (e.g., provisioning AWS infrastructure) during a snapshot restore, the external state might become orphaned from the restored internal state.
- **Corrupted Snapshots**: Checksums are validated before restoration. If a CSI snapshot is corrupted, the system fails closed and aborts the restore to prevent partial org states.
- **Storage Limits**: Automated snapshot pruning deletes older snapshots when cluster storage limits are reached, prioritizing explicitly labeled "keep" snapshots.