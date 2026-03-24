# Test Plan: LangGraph Checkpointing

## 1. Testing Strategy
Validate the correct serialization, persistence, and deserialization of complex agent states across multiple simulated sessions.

## 2. Test Cases
### 2.1 E2E Integration Test: Pause and Resume
- **Setup:** An agent begins a workflow and is explicitly paused after step 1.
- **Action:** A new agent instance is instantiated with the same `thread_id`.
- **Assertion:** Verify the new agent resumes precisely from step 2 with full context.

### 2.2 Edge Case: Large State Payload
- **Setup:** Inject a massive 10MB JSON state payload.
- **Action:** Trigger a checkpoint save.
- **Assertion:** Verify PostgreSQL handles the JSONB insert without performance degradation or truncation.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for the `pg-checkpointer` package.
