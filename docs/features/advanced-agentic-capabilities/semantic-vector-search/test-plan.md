# Test Plan: Semantic Vector Search

## 1. Testing Strategy
Validate the distillation pipeline, embedding generation, and accuracy of vector similarity searches.

## 2. Test Cases
### 2.1 E2E Integration Test: Semantic Retrieval
- **Setup:** Seed the database with 10 varied historical summaries and their embeddings.
- **Action:** Perform a search query related to one specific summary.
- **Assertion:** Verify the correct summary is returned as the top result with a high similarity score.

### 2.2 Edge Case: Worker Concurrency
- **Setup:** Trigger distillation for 1,000 stale checkpoints simultaneously.
- **Action:** Monitor worker queues.
- **Assertion:** Verify the worker processes tasks without exceeding memory limits or causing database deadlocks.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for the `DistillationWorker` and search logic.
