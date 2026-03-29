# CUJ: Persistence & DR Journey


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO utilizes the "Snapshot Fabric" to capture the current state of the organization before trying out a risky new strategic pivot.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Create Snapshot" | API triggers CSI Snapshot | Snapshot saved to S3/Disk | UI displays success toast |
| 2 | CEO "fires" entire Dept | Hub deletes 5 agent pods | Postgres logs firings | Dashboard updates |
| 3 | CEO decides to "Undo" | CEO clicks "Restore" | Hub suspends K8s reconciler | Reconciler paused |
| 4 | Restoration completes | Postgres re-loads dump | Org reverts to Step 1 state | Dashboard shows 5 agents back |

## 3. Implementation Details
- **Architecture**: A Go 1.26 sidecar service manages the storage interface.
- **Stack**: Postgres, K8s CSI plugins.
- **Rollback Speed**: Targets sub-5 second organizational state recovery.

## 4. Edge Cases
- **Concurrent DB Writes**: The snapshot process temporarily halts new writes to the `events.jsonl` log, queueing them in Redis to ensure snapshot consistency.
- **Expired Certificates**: SVIDs issued before the snapshot may be expired when the snapshot is restored. Agents will immediately fail mTLS checks and be forced to re-attest on boot.
- **Loss of Hub Node**: If the main Hub orchestrator node crashes, it automatically rebuilds its state from the latest Postgres `events.jsonl` checkpoint upon scheduling on a new node.