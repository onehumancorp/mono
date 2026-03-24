# CUJ: Semantic Vector Search

**Persona:** Data / ML Agent
**Context:** An agent needs to access historical decisions and context from previous sessions without overloading its current context window.
**Success Metrics:** Fast, highly relevant retrieval of historical checkpoint data via semantic similarity.

## 1. User Journey Overview
Implement an asynchronous semantic distillation worker to summarize older checkpoints and store them as vector embeddings. This provides agents with efficient, long-term memory retrieval without exceeding LLM context window limits.

## 2. Detailed Step-by-Step Breakdown
| Step | Action | System Trigger | Resulting State |
|------|--------|----------------|-----------------|
| 1 | Checkpointer flags stale session | Distillation worker awakened | Checkpoint loaded | Distillation prompt prepared |
| 2 | Distillation worker summarizes state | Calls embedding model | Vector embedding generated | Float array created |
| 3 | Worker stores embedding | pgvector index updated | Data persisted | Row committed |
| 4 | Active agent queries memory | Vector similarity search executed | Relevant summaries retrieved | Context injected |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Embedding Service Outage
- **Detection**: Calls to the embedding model endpoint fail.
- **Auto-Recovery**: Distillation worker queues the tasks and backs off exponentially.
### 3.2 Scenario: Irrelevant Search Results
- **Detection**: Similarity scores fall below a minimum threshold.
- **Resolution**: Search returns empty, preventing hallucination based on loosely related context.

## 4. UI/UX Details
- **Memory Explorer**: A Dashboard tool allowing the CEO to query the vector database using natural language to find past agent decisions.

## 5. Security & Privacy
- **Tenant Isolation**: Embeddings must be strictly segregated by tenant ID using Row Level Security (RLS) in PostgreSQL.
