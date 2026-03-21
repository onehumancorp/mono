# OHC Strategy: Top 5 Urgent Capability Gaps

Based on triangulation of leading AI frameworks (OpenClaw, CrewAI, AutoGen, Claude Code), we have identified the top 5 urgent capability gaps. This document provides actionable designs to merge these features into One Human Corp's architecture, heavily utilizing the OHC Advantage (K8s / LangGraph).

## 1. Stateful Episodic Memory
- **Gap:** Agents suffer from "Amnesia" across disjointed sessions. Passing full context arrays balloons token usage and degrades latency.
- **OHC Advantage:** Native K8s CSI Snapshotting paired with LangGraph checkpointers.
- **Actionable Design:** See `docs/research/design-hook-stateful-episodic-memory.md` for our Postgres-backed, snapshot-driven event stream implementation.

## 2. Dynamic Tool Discovery via MCP
- **Gap:** Frameworks tightly couple agents to hardcoded OpenAPI schemas or Python functions.
- **OHC Advantage:** OHC's existing Switchboard (MCP Gateway) enables zero-trust, RBAC-enforced runtime tool synthesis.
- **Actionable Design:** Deploy a `/v1/tools/search` endpoint. Agents can natively discover tools via semantic search. Connect SPIRE to grant short-lived execution SVIDs dynamically when an agent requires a new API.

## 3. Native Vision & Multimodal Reasoning
- **Gap:** Over-reliance on OCR middleware introduces latency and loss of spatial context in UI reasoning.
- **OHC Advantage:** K8s sidecars capable of direct frame buffering, feeding multimodal payloads dynamically to capable models.
- **Actionable Design:** Deploy a lightweight image-to-text grounding worker pool. Agents can dispatch visual state diffing queries natively without leaving the LangGraph execution path. Use single `http.Client` pools for low-latency multi-stream handling.

## 4. Multi-Agent Collaboration/Swarm
- **Gap:** Monolithic context bloat when a single agent attempts to orchestrate a massive software project.
- **OHC Advantage:** LangGraph orchestrated via gRPC Hub.
- **Actionable Design:** Implement hierarchical task delegation via the gRPC Hub. Manager agents will define specific context bounds and dynamically allocate VRAM quotas to spawn specialized sub-agents via the K8s Operator.

## 5. Human-in-the-Loop (HITL) Workflows
- **Gap:** Agents taking autonomous actions with high risk (e.g. executing code, transferring funds) without human checkpoints.
- **OHC Advantage:** Zero Trust authenticated approval gates.
- **Actionable Design:** Introduce LangGraph checkpoint pause mechanisms. Require cryptographic SPIRE identity validation for the human approver before advancing to high-risk states in the execution graph.