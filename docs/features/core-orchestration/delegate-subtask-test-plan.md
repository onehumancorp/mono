# Test Plan: Hierarchical Task Delegation (Delegate SubTask)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-27

## 1. Overview
A high-level summary of the testing strategy for the Hierarchical Task Delegation (`DelegateSubTask`) feature, ensuring it meets the requirements defined in the Design Document (`delegate-subtask-design-doc.md`) and CUJ (`delegate-subtask-cuj.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on verifying the `DelegateSubTask` method logic directly on the `HubServiceServer` and `Hub` instance, validating quota enforcement and agent provisioning.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | `HubServiceServer` | `DelegateSubTask` with valid arguments and under quota limit | Sub-agent provisioned, task message published | Passed |
| UT-02 | `HubServiceServer` | `DelegateSubTask` with missing `task_id` | Returns `InvalidArgument` | Passed |
| UT-03 | `HubServiceServer` | `DelegateSubTask` with missing `target_role` | Returns `InvalidArgument` | Passed |
| UT-04 | `HubServiceServer` | `DelegateSubTask` when VRAM quota limit is met (>=10 agents) | Returns `ResourceExhausted`, no sub-agent created | Passed |

## 4. Edge Cases & Error Handling
- Validate that the lack of `task_id` or `target_role` is handled gracefully.
- Validate that the quota logic properly identifies the current agent count and rejects spawning if the VRAM limit is reached or exceeded.

## 5. Security & Safety
- Ensure `SYSTEM` agent is properly verified or created to avoid publishing errors.

## 6. Environment & Prerequisites
- Standard unit testing environment via `bazelisk test //srcs/orchestration/...`.