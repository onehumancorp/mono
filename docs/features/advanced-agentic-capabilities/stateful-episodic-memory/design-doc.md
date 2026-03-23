# Design Doc: Stateful Episodic Memory & Checkpointing

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-21

## 1. Overview
The "Stateful Episodic Memory & Checkpointing" feature provides long-term, token-efficient state tracking across disjointed sessions. This directly addresses "Agent Amnesia" by replacing in-memory chat arrays with an append-only, distributed event log architecture backed by K8s CSI Snapshots.

## 2. Goals & Non-Goals
### 2.1 Goals
- Implement a LangGraph Checkpointer connected to a persistent PostgreSQL backend.
- Ensure every virtual meeting room or long-running objective operates within a distinct `thread_id`.
- Enable semantic distillation to summarize older checkpoints for vector retrieval.
- Support arbitrary "rollbacks" of specific `Subsidiary` CRDs via K8s CSI Snapshots.

### 2.2 Non-Goals
- Building a custom vector database. The system will integrate with standard solutions (e.g., pgvector).

## 3. Detailed Design
### 3.1 Checkpointer Store
The LangGraph Checkpointer persists the execution graph's state after every node transition. This state includes the `thread_id` and the current active context payload.

### 3.2 State Threads
Every task or meeting room is assigned a unique `thread_id`. This allows multiple independent workflows to execute concurrently without state collision.

### 3.3 Graph State Sync
As agents progress, their execution path is iteratively snapshotted. To maintain token efficiency, agents only receive the most recent checkpoint state and active transitions, not the entire historical transcript.

### 3.4 Semantic Distillation
A background worker asynchronously processes older checkpoints. It uses an LLM to distill these checkpoints into semantic summaries, embeds them, and stores them in a vector database. Active agents query this vector layer when historical context is needed.

### 3.5 CSI Snapshots
Kubernetes CSI Snapshots allow the human CEO to capture the entire state of an organization (file system + agent memory). This enables rolling back a specific department to a previous "known-good" state within 5 seconds.

## 4. Security & Privacy
- Semantic distillation ensures that raw, potentially sensitive historical data is summarized before embedding, limiting exposure.
- Checkpoints are stored in isolated Postgres schemas per `Subsidiary` CRD.

## 5. Alternatives Considered
- **In-Memory Context Arrays**: Relying entirely on the LLM's growing context window. Rejected due to unacceptable token burn rates, latency spikes, and eventual context collapse.
