# Design Document: Semantic Vector Search

## 1. Executive Summary
**Objective:** Provide scalable, long-term memory retrieval using vector embeddings stored directly in PostgreSQL with pgvector.
**Scope:** Build the `DistillationWorker` and integrate `pgvector` for similarity search within the `OrchestrationHub`.

## 2. Architecture & Components
- **Distillation Worker:** A background cron/worker that summarizes old LangGraph state.
- **Embedding Client:** Interfaces with models like `text-embedding-3-small`.
- **pgvector Database:** The underlying storage and indexing mechanism.

## 3. Data Flow
1. Worker identifies old threads and generates summaries.
2. Summaries are converted to embeddings and saved via `pgvector`.
3. An active agent executes a `search_memory("previous deployment issues")` tool call.
4. The Hub executes a cosine similarity search and returns the top K results.

## 4. API & Data Models
```sql
CREATE TABLE embeddings (
  id uuid PRIMARY KEY,
  tenant_id uuid,
  content text,
  embedding vector(1536)
);
```

## 5. Implementation Details
- Optimize `pgvector` queries using HNSW (Hierarchical Navigable Small World) indexes for fast retrieval.
- Maintain Zero-Lock stack compatibility.
