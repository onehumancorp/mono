# Checkpointer

<div style="backdrop-filter: blur(20px) saturate(200%); background: rgba(255, 255, 255, 0.03); border: 1px solid rgba(255, 255, 255, 0.08); padding: 16px; border-radius: 8px; font-family: 'Outfit', 'Inter', sans-serif;">
  <strong>Overview:</strong> The Checkpointer package manages state snapshots for LangGraph threads within the Agentic OS. It provides hermetic and robust persistence using PostgreSQL/SQLite backends to ensure seamless state recovery.
</div>

## Architecture

The system utilizes an upsert strategy (`ON CONFLICT (thread_id) DO UPDATE SET state = excluded.state`) to store thread-specific conversational states efficiently. A resilient retry mechanism with exponential backoff and jitter is included to mitigate transient locking errors (e.g., `database is locked` in SQLite).

## Quick Start

```go
// Initialize the database connection
db, err := sql.Open("sqlite", ":memory:")
if err != nil {
    log.Fatal(err)
}

// Instantiate the checkpointer
p := checkpointer.NewPGCheckpointer(db)

// Ensure the checkpoints table is initialized
err = p.EnsureTableExists(context.Background())
if err != nil {
    log.Fatal(err)
}

// Persist a conversational state snapshot
err = p.SaveCheckpoint(context.Background(), "thread-123", map[string]interface{}{
    "messages": []string{"hello", "world"},
})
```

## Features
- **Absolute Autonomy:** Agents do not require manual intervention to persist state across interruptions.
- **Aesthetic Excellence:** Logs and data formats are strictly checked, and no arbitrary payloads are allowed due to `DisallowUnknownFields` strict decoding.
- **Hermetic Implementation:** Retries transient errors transparently for stability in K8s native environments.

---
_One Human Corp - Swarm Intelligence Protocol (OHC-SIP)_
