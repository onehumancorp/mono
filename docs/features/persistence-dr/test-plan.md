# Test Plan: Persistence & Disaster Recovery

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Persistence & Disaster Recovery feature, ensuring it meets the requirements defined in the Design Document (`snapshot-fabric.md`) and CUJs (`cuj-org-snapshot.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on parsing event logs, snapshot metadata, and validating point-in-time boundaries.
- **Integration Testing:** Verify communication between the Hub, PostgreSQL/LangGraph checkpointers, and the K8s CSI Snapshotter.
- **End-to-End (E2E) Testing:** Validate the complete org backup and recovery flow from a snapshot file.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Event Log | Parse `events.jsonl` | JSON deserializes accurately | Pending |
| UT-02 | Snapshot Meta| Validate snapshot ID structure | Timestamp and Org match format | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> DB | Commit checkpoint state | Row written successfully | Pending |
| IT-02 | Hub -> K8s CSI | Trigger volume snapshot | K8s returns VolumeSnapshot ID | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Snapshot Trigger | Admin clicks Backup | Snapshot created in AWS/GCP < 10s | Pending |
| E2E-02 | Org Restore | Restore from Snapshot ID | 100% state fidelity after < 5s | Pending |
| E2E-03 | Partial Corrupt | Try restoring corrupt metadata | Restorer aborts and warns UI | Pending |

## 4. Edge Cases & Error Handling
- **Database Partition:** Verify event log buffers to disk if the LangGraph checkpointer DB is unreachable.
- **Snapshot Conflict:** Ensure concurrent snapshot requests are queued and not overwritten.

## 5. Security & Safety
- **Encryption:** Verify all CSI snapshots are encrypted at rest with Customer-Managed Keys (CMKs).
- **Access Control:** Ensure only users with `Admin` privileges can trigger a global restore.

## 6. Environment & Prerequisites
- OHC Hub configured with local storage driver supporting CSI snapshots.

## Implementation Details
- **Architecture**: The Persistence & DR tests utilize Go 1.26 table-driven tests. Integration logic runs against a local PostgreSQL seeder and a mock Kubernetes CSI (Container Storage Interface) driver.
- **Execution**: Run hermetically under Bazel 9.0.0 (`bazelisk test //...`). The suite explicitly avoids mocking the database layer, running `pg_dump`/`pg_restore` commands directly against the seeded PostgreSQL instance to ensure true end-to-end coverage of the Snapshot Fabric.
- **Validation**: Strict >95% test coverage ensures that LangGraph checkpointer state logic and event serialization to `events.jsonl` are rock-solid.

## Edge Cases
- **In-Flight Tool Operations**: E2E tests trigger a snapshot restore while a mock agent is waiting for an external API call. The test ensures that the external state is gracefully orphaned and the agent restarts safely from the checkpointed state.
- **Corrupted Snapshots**: A test intentionally modifies a byte in the CSI snapshot file to verify the Hub calculates checksums before restoration. If corrupted, the system fails closed and aborts the restore to prevent partial org states.
- **Storage Limit Pruning**: Tests simulate a full storage quota to verify that automated snapshot pruning correctly deletes older snapshots, prioritizing those without a labeled "keep" tag.
