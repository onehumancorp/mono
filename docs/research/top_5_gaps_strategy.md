# OHC Strategy: Top 5 Urgent Capability Gaps


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


Based on triangulation of leading AI frameworks (OpenClaw, CrewAI, AutoGen, Claude Code), we have identified the top 5 urgent capability gaps. This document provides actionable designs to merge these features into One Human Corp's architecture, heavily utilizing the OHC Advantage (K8s / LangGraph).

## 1. Long-Term Episodic Memory
- **Gap:** Agents suffer from "Amnesia" across disjointed sessions. Passing full context arrays balloons token usage and degrades latency. Cross-session memory persistence using vector databases is missing to recall past user interactions, successful tool uses, and learned preferences.
- **OHC Advantage:** Native K8s CSI Snapshotting paired with LangGraph checkpointers, backed by scalable Vector Databases (Redis/Pinecone).
- **Actionable Design:** See `docs/research/design-hook-episodic-memory.md` for our Postgres-backed, snapshot-driven event stream implementation.

## 2. Dynamic Tool Discovery (MCP)
- **Gap:** Frameworks tightly couple agents to hardcoded OpenAPI schemas or Python functions. Agents cannot autonomously search a registry for new tools/APIs during runtime if existing tools fail to solve the task.
- **OHC Advantage:** OHC's existing Switchboard (MCP Gateway) enables zero-trust, RBAC-enforced runtime tool synthesis secured by SPIFFE/SPIRE dynamic RPC endpoints.
- **Actionable Design:** See `docs/research/design-hook-dynamic-tool-discovery.md` for blueprint on self-registering tool discovery via SPIFFE.

## 3. Native Vision & Multimodal Reasoning
- **Gap:** Over-reliance on OCR middleware introduces latency and loss of spatial context in UI reasoning. Agents cannot directly ingest and reason over screenshots, UI elements, and diagrams without OCR middleware.
- **OHC Advantage:** K8s sidecars capable of direct frame buffering, feeding multimodal payloads dynamically to capable models. Token-efficient multi-stream handling.
- **Actionable Design:** Deploy a lightweight image-to-text grounding worker pool. Agents can dispatch visual state diffing queries natively without leaving the LangGraph execution path.

## 4. Hierarchical Task Delegation
- **Gap:** Monolithic context bloat when a single agent attempts to orchestrate a massive software project. Manager agents cannot autonomously break down complex goals and spawn sub-agents with specific contexts.
- **OHC Advantage:** OHC's CRD structure (`TeamMember`, `Subsidiary`) inherently models hierarchies. Manager agents can spin up specialized sub-agents dynamically via K8s Operator.
- **Actionable Design:** Implement a `/scale` endpoint trigger where Manager agents can define specific context bounds and dynamically allocate VRAM quotas to spawn sub-agents on demand.

## 5. Stateful Execution Graph (LangGraph)
- **Gap:** Traditional frameworks fail at handling cyclic workflows (looping, retrying, reflecting) gracefully, often crashing or repeating identical hallucinated loops. There are no cyclic, stateful workflows that allow agents to loop, reflect, and retry failed actions based on a persistent graph state.
- **OHC Advantage:** Direct integration of LangGraph state machines natively managed by our Control Plane.
- **Actionable Design:** Migrate all core workflows from static prompting chains to LangGraph state transitions. Use deterministic state syncing to ensure we can instantly halt and recover via CSI Snapshots.
