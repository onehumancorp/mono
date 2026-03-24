# Test Plan: VRAM Quota Management

## 1. Testing Strategy
Validate that hard resource limits are enforced, queuing logic works under pressure, and preemption of lower priority agents functions correctly.

## 2. Test Cases
### 2.1 Integration Test: Quota Enforcement
- **Setup:** Set a mock organization's VRAM limit to 1 GPU.
- **Action:** Attempt to spawn two agents requiring 1 GPU each.
- **Assertion:** Verify the first agent is scheduled and the second agent remains in `Pending` or is rejected by the admission controller.

### 2.2 Edge Case: Priority Preemption
- **Setup:** A critical task requires VRAM held by an idle background task.
- **Action:** Critical task initiated.
- **Assertion:** Verify the idle task is gracefully paused, and the critical task acquires the necessary resources.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for the `QuotaManager`.
