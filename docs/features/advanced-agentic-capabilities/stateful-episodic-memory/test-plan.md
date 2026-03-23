# Test Plan: Stateful Episodic Memory & Checkpointing

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-21

## 1. Overview
A high-level summary of the testing strategy for the Stateful Episodic Memory & Checkpointing feature, ensuring it meets the requirements defined in the Design Document and CUJs.

## 2. Test Strategy
- **Unit Testing:** Focus on individual components like state serialization, checkpoint payload constraints, and vector database querying.
- **Integration Testing:** Verify communication between the LangGraph Checkpointer, Postgres, the background semantic distillation worker, and the LLM API.
- **End-to-End (E2E) Testing:** Validate an agent retrieving cross-session memory in a new context and confirm successful K8s CSI Snapshots via mock integration.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Checkpoint Saving | Checkpointer serializes and saves state | State saved to DB | Pending |
| UT-02 | Distillation Trigger | State payload exceeds limit | Distillation worker queued | Pending |
| UT-03 | Semantic Summary | LLM distills past transcript into summary | Text length significantly reduced | Pending |
| UT-04 | Vector Retrieval | Agent queries memory bank | Relevant summary returned | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Checkpointer -> Postgres | Multi-turn interaction recorded | Sequential states logged with `thread_id` | Pending |
| IT-02 | Worker -> Vector DB | Worker embeddings saved to pgvector | Embedded vector searchable | Pending |
| IT-03 | Graph -> Checkpointer | Context payload injection | Retrieved summary correctly formatted | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Cross-Session Recall | Agent asks question about past interaction | Agent answers using retrieved distilled memory | Pending |
| E2E-02 | Org Rollback | Simulate CEO triggering rollback via UI | State reverts to selected checkpoint | Pending |

## 4. Edge Cases & Error Handling
- **Database Limits**: Simulate the Checkpointer hitting storage limits and test eviction/distillation strategies.
- **Distillation Failures**: Verify the system falls back gracefully if the summarization LLM fails during distillation.

## 5. Security & Safety
- **Cross-Thread Leakage**: Ensure that agents operating in `thread_id=1` cannot retrieve checkpointer states from `thread_id=2` unless explicitly authorized.
- **Isolated Namespaces**: Validate that vector embeddings are isolated by `Subsidiary` CRD boundaries to prevent inter-organizational memory leakage.

## 6. Implementation Details
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox.
- **Mocks**: Mock the LLM, external vector database, and the K8s CSI Snapshot APIs.
- **Validation**: Strict enforcement of >95% test coverage for the memory persistence and retrieval engine.
