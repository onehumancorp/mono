# CUJ: Stateful Episodic Memory & Checkpointing

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-21

**Persona:** Software Engineer Agent | **Context:** Participating in a month-long product development cycle.
**Success Metrics:** Cross-session memory recall, minimal token burn, and robust rollback capability.

## 1. User Journey Overview
A SWE agent is tasked with a new feature that depends on an architectural decision made three weeks ago in a different virtual meeting room. The agent seamlessly queries the distilled semantic memory to recall the historical context without loading the entire past transcript into its active context window.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Agent encounters missing context | Agent queries semantic memory vector db | Vector query executed | Query logged in `events.jsonl` |
| 2 | System retrieves context | Vector db returns distilled summary | Summary payload prepared | Payload size verified < limit |
| 3 | System injects context | Summary injected into agent prompt | Agent possesses historical context | Agent references past decision |
| 4 | Agent executes task | Task completed | Checkpoint saved | State recorded in Postgres |
| 5 | CEO initiates rollback | CEO selects past checkpoint in UI | K8s CSI Snapshot restored | Org state reverts to prior point |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Semantic Distillation Failure
- **Detection**: The background worker fails to distill older checkpoints into a vector representation.
- **Auto-Recovery**: The worker retries with exponential backoff.
- **System Action**: Active agents continue using the raw checkpoint history until the distillation completes, temporarily consuming more tokens.

### 3.2 Scenario: Vector Search Hallucination
- **Detection**: The vector db returns irrelevant historical context.
- **System Action**: The agent evaluates the retrieved context, determines its irrelevance, and requests a broader or more specific query.

## 4. UI/UX Details
- **Memory Explorer**: The CEO Dashboard includes a "Memory Explorer" view to search and inspect the organization's semantic vector database.
- **Snapshot Timeline**: A visual timeline of K8s CSI Snapshots allows the CEO to click and instantly restore the entire state.

## 5. Security & Privacy
- **Isolated Embeddings**: Semantic embeddings are strictly isolated per `Subsidiary` CRD. Agents cannot query the memory banks of other organizations.
- **Immutable Log**: All state checkpoints and vector queries are appended to an immutable `events.jsonl` audit log.
