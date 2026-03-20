# Design Doc: Core Orchestration

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Core Orchestration Engine (the "Hub") transforms high-level CEO mandates into actionable workflows, delegating sub-tasks across a dynamic hierarchy of AI agents.

## 2. Goals & Non-Goals
### 2.1 Goals
- Automate task delegation and dependency graph management.
- Provide a persistent memory context across Virtual Meeting Rooms.
### 2.2 Non-Goals
- Human chat application features (e.g., emojis, threads).

## 3. Implementation Details
- **Architecture**: A Go 1.26 monolith operating within the `Hub` module handles the orchestration loop. It uses goroutines for lightweight concurrent task management and LangGraph checkpointers for saving/loading context.
- **Data Model**: The K8s Custom Resource Definition (`HoldingCompany`) natively syncs with internal Postgres records tracking current active directives and agent assignments.
- **Messaging Pipeline**: Multi-agent delegation pushes state messages to Redis. The Hub reads from Redis to evaluate if dependencies are met for downstream tasks.

## 4. Edge Cases
- **Cyclic Dependencies**: If a PM mistakenly assigns two tasks that depend on each other, the Orchestration Engine's DAG evaluator detects the cycle and raises a `DependencyCycleError` for human review.
- **Node Failures**: If an underlying K8s node dies, the agent pod restarts on a new node and recovers its exact position in the workflow by rehydrating state from the Postgres append-only log.
- **Unreachable LLM Provider**: If the external AI API (e.g., OpenAI/Gemini) is down, the engine pauses active workflows and places them in a `ProviderRetry` queue rather than silently dropping tasks.