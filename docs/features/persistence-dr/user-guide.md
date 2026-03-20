# User Guide: Persistence & Snapshots

## Introduction
The Snapshot Fabric allows you to "Undo" complex organizational changes or recover from a disaster with a single click.

## Usage
### 1. Creating a Snapshot
Before performing a major reorganisation (e.g., merging two departments), click "Create Snapshot" in the Mission Control panel.

### 2. Labeling
Always give your snapshots a descriptive label (e.g. `pre-frontend-refactor`).

### 3. Restoring
If something goes wrong, navigate to the Snapshots log and click "Restore". Your organisation will revert to its previous state within 5 seconds.

## Best Practices
- Create automated "Weekly Snapshots" in the settings.
- Use snapshots to "Test Scenarios" without affecting your long-term production state.

## Troubleshooting
**Snapshot restoration failed**
- Check the database logs for the OHC cluster.
- Ensure you have enough storage space in your Kubernetes cluster.

## Implementation Details
- **Architecture**: The Snapshot Fabric leverages Kubernetes CSI (Container Storage Interface) volume snapshots combined with Postgres `pg_dump`/`pg_restore` for database state.
- **State Management**: LangGraph checkpointers serialize the exact multi-agent state (memory, pending tasks, active tools) into the append-only `events.jsonl` log. Restoring a snapshot replays or truncates this log deterministically.
- **Execution**: Orchestrated via Go 1.26 in the OHC Hub. The Operator pauses the `HoldingCompany` CRD reconciliation loop during the restoration process.

## Edge Cases
- **In-Flight Tool Operations**: If an agent is executing a long-running external API call (e.g., provisioning AWS infrastructure) during a snapshot restore, the external state might become orphaned from the restored internal state.
- **Corrupted Snapshots**: Checksums are validated before restoration. If a CSI snapshot is corrupted, the system fails closed and aborts the restore to prevent partial org states.
- **Storage Limits**: Automated snapshot pruning deletes older snapshots when cluster storage limits are reached, prioritizing explicitly labeled "keep" snapshots.
