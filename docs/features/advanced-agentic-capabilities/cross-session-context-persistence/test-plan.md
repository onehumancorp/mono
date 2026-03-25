# Test Plan: Cross Session Context Persistence

## 1. Testing Strategy
Validate the end-to-end functionality, security boundaries, and performance constraints of the Cross Session Context Persistence feature using hermetic, table-driven tests. Ensure we use the Database Seeder pattern to establish deterministic starting states.

## 2. Test Cases
### 2.1 E2E Integration Test: Standard Execution Flow
- **Setup:** A mock environment with a deterministic database state via `/api/dev/seed`.
- **Action:** Complete a conversation in Session A. Start Session B and ask a question referencing a fact from Session A. Verify the agent successfully retrieves and utilizes the persisted context.
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
