# Checkpointer Module

## Identity
The `checkpointer` module manages the persistent state of LangGraph agent threads, enabling episodic memory and fault tolerance within the One Human Corp orchestration hub.

## Architecture
This module provides the `LangGraphCheckpointer` interface and its PostgreSQL-backed implementation. It serializes and deserializes the dynamic state (dictionaries/maps) of AI agents, storing them as JSONB within the `ohc.db` (or a remote PostgreSQL instance). This allows workflows to be paused, snapshotted, and resumed, solving "Agent Amnesia".

## Developer Workflow
The checkpointer is tested using standard Go testing tools and is tightly integrated with the `orchestration` module.

- **Test Checkpointer**: `bazel test //srcs/checkpointer/...`

## Configuration
Requires an active PostgreSQL connection (or local SQLite mock for tests) passed down from the central Hub configuration.
