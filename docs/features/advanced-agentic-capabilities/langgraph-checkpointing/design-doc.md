# Design Document: LangGraph Checkpointing

## 1. Executive Summary
**Objective:** Enable long-term agent memory persistence using LangGraph's checkpointer interface integrated with a highly available PostgreSQL backend.
**Scope:** Develop the Go wrapper `pg-checkpointer` implementing the LangGraph Checkpointer interface.

## 2. Architecture & Components
- **LangGraph Node:** Interacts with the interface.
- **PG Checkpointer Service:** The Go service handling Postgres serialization.
- **Database Layer:** Optimized JSONB columns for state storage.

## 3. Data Flow
1. Agent executes a node and triggers `SaveCheckpoint`.
2. Service serializes the current context and tool responses.
3. Data is persisted to Postgres.
4. When resuming, `LoadCheckpoint` retrieves the data.

## 4. API & Data Models
```go
type Checkpoint struct {
  ThreadID string `json:"thread_id"`
  State map[string]interface{} `json:"state"`
}
```

## 5. Implementation Details
- Use strict JSON decoding with `dec.DisallowUnknownFields()` when reloading checkpoints to ensure schema integrity.
- Optimize PostgreSQL indexes on `thread_id`.
- Maintain Zero-Lock stack compatibility.
