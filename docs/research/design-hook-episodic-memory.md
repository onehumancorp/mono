# Design Hook: Stateful Episodic Memory & Checkpointing via LangGraph

## Executive Summary
Current mainstream AI orchestration frameworks suffer from "Agent Amnesia"—a failure to maintain long-term, token-efficient state across disjointed sessions. As operations scale, injecting massive historical contexts into LLM prompts leads to unacceptable token burn rates, latency spikes, and eventual context collapse.

This document proposes a radical "OHC Advantage" implementation: **Stateful Episodic Memory & Checkpointing**, driven entirely by LangGraph event streams backed by Kubernetes CSI Snapshotting and robust Vector databases.

## The Architecture
OHC's Agentic OS functions as a "Universal Bus" for multi-framework AI swarms. We will replace massive in-memory chat arrays with an append-only, distributed event log architecture and semantic similarity retrieval.

### Component 1: LangGraph Checkpointers
1. **The Checkpointer Store**: We deploy a dedicated LangGraph Checkpointer connected to a persistent PostgreSQL backend.
2. **State Threads**: Every virtual meeting room or long-running objective is assigned a distinct `thread_id`.
3. **Graph State Sync**: As agents progress through the LangGraph execution path, state is snapshotted iteratively. Agents only receive the most recent checkpoint state and active node transitions, rather than the raw, unfiltered conversation history.

### Component 2: Semantic Memory Distillation
Instead of carrying raw tokens forward:
1. **Background Distillation Worker**: A specialized, low-tier LLM asynchronously distills older checkpoints into concise, semantic summaries.
2. **Vector Retrieval Layer**: These summaries are embedded and stored in a high-performance Vector Database (e.g. Pinecone, Qdrant) managed as a native K8s StatefulSet. Active agents query this semantic layer via `semantic_vector_search` dynamically when historical context is required (e.g., "What was our decision on the caching layer in the previous meeting?").

### Component 3: The K8s CSI Snapshot Fabric
1. **Volume Level State**: Agent checkpointer data, scratchpads, and distilation logs are mapped to a `PersistentVolumeClaim`.
2. **CEO Point-in-Time Recovery**: We integrate Kubernetes Container Storage Interface (CSI) snapshots, allowing the human CEO to arbitrarily "roll back" the state of a specific `Subsidiary` CRD (including its LangGraph checkpoints) within 5 seconds in case of an orchestrated hallucination or incorrect branch execution.

## Token Efficiency & Performance Constraints
- **Zero Raw Injection**: No more injecting >50k context windows for simple interactions.
- **Lazy Loading Context**: Agents only load working memory and retrieve distilled episodic memory strictly on a need-to-know basis through the semantic retrieval layer.
- **Fail-Safe Constraints**: The system strictly limits prompt size to the core `thread_id` snapshot payload, significantly reducing inference latency (sub-50ms routing goals) and API burn rate.

## Next Steps
- Execute a proof-of-concept integrating a LangGraph Postgres Checkpointer directly into the K8s Operator cluster.
- Define the explicit gRPC streaming payload structure for checkpoint transmission to eliminate polling loops.
- Provision the Redis/Pinecone backings for the short/long term semantic memory stores.
