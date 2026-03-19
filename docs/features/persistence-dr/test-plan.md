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
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
