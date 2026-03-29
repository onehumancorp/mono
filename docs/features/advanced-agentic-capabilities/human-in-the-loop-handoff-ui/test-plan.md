# Test Plan: Human-in-the-Loop Handoff UI

## 1. Testing Strategy
Validate the end-to-end handoff lifecycle, including agent pausing, UI presentation, approval workflows, and thread resumption.

## 2. Test Cases
### 2.1 E2E Integration Test: Successful Handoff
- **Setup:** A mock agent triggers a handoff for an approval over a set threshold.
- **Action:** Manager accesses the UI and approves.
- **Assertion:** Verify the agent thread resumes seamlessly from the exact paused state.

### 2.2 Edge Case: Stale State Rejection
- **Setup:** Simulate the underlying data changing before the manager approves.
- **Action:** Manager attempts to approve.
- **Assertion:** Verify the system detects the divergence and prompts the manager to re-review or reject the handoff.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for the `HandoffService` and the Flutter frontend components.
