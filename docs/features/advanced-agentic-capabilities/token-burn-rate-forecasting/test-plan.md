# Test Plan: Token Burn-Rate Forecasting

## 1. Testing Strategy
Validate the end-to-end functionality, security boundaries, and performance constraints of the Token Burn-Rate Forecasting feature using hermetic, table-driven tests.

## 2. Test Cases
### 2.1 E2E Integration Test: Standard Execution Flow
- **Setup:** A mock environment with a deterministic database state via `/api/dev/seed`.
- **Action:** Simulate an agent invoking the Token Burn-Rate Forecasting functionality.
- **Assertion:** Verify the operation completes successfully and the correct events are written to `events.jsonl`.

### 2.2 Edge Case: Strict Schema and Payload Validation
- **Setup:** Craft an invalid payload containing unknown JSON fields.
- **Action:** Submit the payload to the feature's API endpoint.
- **Assertion:** Verify the request is rejected immediately via `dec.DisallowUnknownFields()` and does not crash the server.

### 2.3 Edge Case: Memory and Resource Bounding
- **Setup:** Simulate a high-frequency barrage of requests.
- **Action:** Monitor the feature's map-based trackers and buffers.
- **Assertion:** Verify memory growth remains bounded and map entries are properly deleted after resolving tracked states.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST strictly exceed 95% for the corresponding Go packages.
- Tests will utilize lightweight dependency injection for fatal exit paths (`os.Exit`).
