# OHC Strategy: Top 5 Urgent Capability Gaps

Based on triangulation of leading AI frameworks (OpenClaw, CrewAI, AutoGen, Claude Code), we have identified the top 5 urgent capability gaps. This document provides actionable designs to merge these features into One Human Corp's architecture, heavily utilizing the OHC Advantage (K8s / LangGraph).

## 1. Stateful Episodic Memory
- **Gap:** Agents suffer from "Amnesia" across disjointed sessions. Passing full context arrays balloons token usage and degrades latency.
- **OHC Advantage:** Native K8s CSI Snapshotting paired with LangGraph checkpointers.
- **Actionable Design:** See `docs/research/design-hook-episodic-memory.md` for our Postgres-backed, snapshot-driven event stream implementation.

## 2. Dynamic Tool Registration via MCP
- **Gap:** Frameworks tightly couple agents to hardcoded OpenAPI schemas or Python functions.
- **OHC Advantage:** OHC's existing Switchboard (MCP Gateway) enables zero-trust, RBAC-enforced runtime tool synthesis.
- **Actionable Design:** See `docs/research/design-hook-dynamic-tool-discovery.md` for blueprint on self-registering tool discovery via SPIFFE.

## 3. Native Vision & Multimodal Reasoning
- **Gap:** Over-reliance on OCR middleware introduces latency and loss of spatial context in UI reasoning.
- **OHC Advantage:** K8s sidecars capable of direct frame buffering, feeding multimodal payloads dynamically to capable models.
- **Actionable Design:** Deploy a lightweight image-to-text grounding worker pool. Agents can dispatch visual state diffing queries natively without leaving the LangGraph execution path.

## 4. Hierarchical Task Delegation
- **Gap:** Monolithic context bloat when a single agent attempts to orchestrate a massive software project.
- **OHC Advantage:** OHC's CRD structure (`TeamMember`, `Subsidiary`) inherently models hierarchies. Manager agents can spin up specialized sub-agents dynamically via K8s Operator.
- **Actionable Design:** Implement a `/scale` endpoint trigger where Manager agents can define specific context bounds and dynamically allocate VRAM quotas to spawn sub-agents on demand.

## 5. Stateful Execution Graph (LangGraph)
- **Gap:** Traditional frameworks fail at handling cyclic workflows (looping, retrying, reflecting) gracefully, often crashing or repeating identical hallucinated loops.
- **OHC Advantage:** Direct integration of LangGraph state machines natively managed by our Control Plane.
- **Actionable Design:** Migrate all core workflows from static prompting chains to LangGraph state transitions. Use deterministic state syncing to ensure we can instantly halt and recover via CSI Snapshots.
