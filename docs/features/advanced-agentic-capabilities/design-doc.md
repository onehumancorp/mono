# Design Doc: Advanced Agentic Capabilities

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
The "Advanced Agentic Capabilities" initiative represents Phase 8 of the One Human Corp (OHC) Strategic Roadmap. Its objective is to integrate the most critical features identified across leading AI frameworks (CrewAI, AutoGen, LangGraph) to establish OHC as the definitive orchestration platform. This design addresses the core challenge of scaling autonomous multi-agent systems efficiently, directly mitigating "Agent Amnesia," static tool binding, and orchestration context bloat.

## 2. Goals & Non-Goals
### 2.1 Goals
- Implement **Stateful Episodic Memory & Checkpointing** using LangGraph and Kubernetes CSI Snapshots for robust, token-efficient, cross-session state persistence.
- Enable **Dynamic Tool Registration via MCP Gateway**, allowing runtime synthesis of tools without hardcoded schemas.
- Introduce **Native Vision & Multimodal Reasoning** capabilities, enabling agents to parse visual data natively (e.g., screenshots, UI elements) without OCR latency.
- Provide **Hierarchical Task Delegation** mechanisms, allowing Manager agents to dynamically spawn sub-agents with narrow, highly focused contexts.
- Migrate complex cyclic workflows to a **Stateful Execution Graph** (LangGraph model), enabling deterministic retries and structured execution.

### 2.2 Non-Goals
- Real-time video stream processing for multimodal reasoning (initial focus is on static image/UI frames).
- Creating new foundational multimodal models (OHC will consume existing capabilities via API providers).
- Developing a custom graph execution engine from scratch (OHC will leverage LangGraph-style checkpointing patterns integrated natively with K8s).

## 3. Detailed Design
### 3.1 Stateful Episodic Memory & Checkpointing
To resolve "Agent Amnesia," OHC shifts from massive in-memory chat arrays to an append-only, distributed event log architecture.
- **Checkpointer Store**: A dedicated LangGraph Checkpointer connected to a persistent PostgreSQL backend.
- **State Threads**: Every virtual meeting room or long-running objective operates within a distinct `thread_id`.
- **Graph State Sync**: As agents progress, the execution path is iteratively snapshotted. Agents only receive the most recent checkpoint state and active transitions.
- **Semantic Distillation**: A background worker asynchronously distills older checkpoints into semantic summaries, which are embedded and stored. Active agents query this vector layer when historical context is needed, keeping active memory clean and token-efficient.
- **CSI Snapshots**: K8s CSI Snapshots allow the CEO to arbitrarily "roll back" the state of a specific `Subsidiary` CRD (including LangGraph checkpoints) within 5 seconds.

### 3.2 Dynamic Tool Discovery via MCP
Current frameworks couple agents to hardcoded OpenAPI schemas. OHC utilizes the unified **MCP Gateway (Switchboard)** to allow zero-trust, RBAC-enforced runtime tool synthesis.
- **Dynamic Registration**: Tools are discovered and registered via the SPIFFE-gated MCP Gateway.
- **Runtime Binding**: Agents can query the registry and dynamically bind to necessary tools based on task requirements, minimizing error loops caused by missing hardcoded configurations.

### 3.3 Native Vision & Multimodal Reasoning
Over-reliance on OCR middleware introduces latency and loss of spatial context.
- **Multimodal Workers**: Deploy a lightweight image-to-text grounding worker pool or pass multimodal payloads directly to capable LLM endpoints.
- **Native Parsing**: Agents can dispatch visual state diffing queries natively without leaving the execution path, drastically improving UI verification workflows and frontend verification.

### 3.4 Hierarchical Task Delegation
Monolithic context bloat occurs when a single agent attempts to orchestrate a massive software project.
- **Dynamic Provisioning**: OHC's CRD structure (`TeamMember`, `Subsidiary`) inherently models hierarchies. Manager agents can trigger a `/scale` endpoint to dynamically allocate VRAM quotas and spawn specialized sub-agents.
- **Narrow Contexts**: Sub-agents operate with strictly defined context bounds, ensuring optimal token allocation per task.

### 3.5 Stateful Execution Graph (LangGraph)
Traditional static prompting chains fail at handling cyclic workflows (looping, retrying, reflecting).
- **Graph Migration**: Migrate all core workflows to stateful execution models.
- **Deterministic State**: Use deterministic state syncing to ensure workflows can be instantly halted and recovered via CSI Snapshots.

## 4. Cross-cutting Concerns
### 4.1 Security & Identity
- All dynamic tool discovery and inter-agent communication must be authenticated via SPIFFE SVIDs.
- VRAM Quota Management prevents runaway compute costs during hierarchical task delegation.
### 4.2 Scalability & Performance
- The system strictly limits prompt size to the core `thread_id` snapshot payload.
- Lazy Loading Context ensures agents retrieve distilled episodic memory strictly on a need-to-know basis.

## 5. Alternatives Considered
- **In-Memory Context Arrays**: Relying entirely on the LLM's growing context window. Rejected due to unacceptable token burn rates, latency spikes, and eventual context collapse.
- **Hardcoded Tool Chains**: Defining specific tools per agent role in static configuration. Rejected as it severely limits flexibility and extensibility when importing new Skill Blueprints.
