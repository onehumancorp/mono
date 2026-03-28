# LangGraph Checkpointer Component

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Active Execution
**Last Updated:** 2026-03-28

This component, `checkpointer`, manages state persistence and recovery for AI agents operating within LangGraph workflows inside the Agentic OS.

## Overview

The `LangGraphCheckpointer` serves as a fundamental building block for the **Stateful Episodic Memory** mandate. By persisting agent thread states continuously to a scalable database backend, it prevents "Agent Amnesia" and allows complex, multi-step orchestrations to survive container evictions, network partitions, and node scaling events.

## Features

1.  **State Persistence (`SaveCheckpoint`)**: Serializes an agent's current state and thread context into JSONB and stores it securely using a high-throughput, conflict-resolving UPSERT strategy.
2.  **State Recovery (`LoadCheckpoint`)**: Retrieves a specific thread ID's state and perfectly reconstitutes the context for the LangGraph executor.
3.  **Transient Error Handling**: Includes robust exponential backoff and retry mechanisms (`withRetry`) to gracefully handle database locks, connection blips, and high-concurrency contention without dropping checkpoints.

## Implementation Details

*   **Language**: Go
*   **Backend Storage**: PostgreSQL (with compatibility for SQLite during testing scenarios).
*   **Interfaces**: Implements the `LangGraphCheckpointer` interface which defines core methods required by the central orchestration Hub.

## Developer Workflow

This module requires a running database instance for local testing.

```bash
# Build the component
bazelisk build //srcs/checkpointer/...

# Run all unit tests within this package
bazelisk test //srcs/checkpointer/...
```

## Architectural Context

*   **Design Document**: `docs/features/advanced-agentic-capabilities/stateful-episodic-memory/design-doc.md`
*   **Epic Context**: `docs/execution-plan.md` (Task 2.1: Implement Checkpointer Interface)
*   **Shared Context**: `docs/shared_context.md`