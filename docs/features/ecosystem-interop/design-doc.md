# Design Doc: Ecosystem Interoperability (Framework Adapters)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-22

## 1. Overview
The "Ecosystem Interoperability" feature establishes the One Human Corp (OHC) "Agentic OS" control plane as the Universal Bus for AI agent swarms. It introduces native framework adapters allowing agents from heterogeneous platforms (OpenClaw, AutoGen, CrewAI, Semantic Kernel) to collaborate seamlessly within a unified OHC environment.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Framework Agnosticism**: Provide native adapters for OpenClaw, AutoGen, CrewAI, and Semantic Kernel.
- **Unified State Management**: Synchronize agent states across frameworks using LangGraph checkpointers, enabling true multi-framework swarms.
- **Identity & Security**: Secure all intra-swarm and inter-framework communications via cryptographically verified SPIFFE/SPIRE identities.
- **Universal Tooling**: Enable third-party framework agents to consume tools seamlessly via the OHC MCP Switchboard.

### 2.2 Non-Goals
- Native execution environments for Python-based frameworks (adapters act as a bridge, relying on external API hooks or sidecar containers for non-Go execution).

## 3. Detailed Architecture
### 3.1 Universal Interface (`srcs/interop/types.go`)
The core OHC control plane exposes a `UniversalAgent` interface. Every supported framework has a corresponding adapter that translates framework-specific constructs into OHC events.

### 3.2 Framework Adapters
- **OpenClaw Adapter (`openclaw_adapter.go`)**: Syncs real-time state check-pointing via append-only K8s custom resources and LangGraph event streams.
- **AutoGen Adapter (`autogen_adapter.go`)**: Maps AutoGen's multi-agent conversational model to OHC's event-driven pub/sub architecture.
- **CrewAI Adapter (`crewai_adapter.go`)**: Translates CrewAI roles, tasks, and team assignments into LangGraph states.
- **Semantic Kernel Adapter (`semantickernel_adapter.go`)**: Integrates SK's function calling and prompt orchestration directly into the shared state manager and agent command executor.

### 3.3 Core Infrastructure Components
- **MCP Switchboard**: Acts as the central proxy for all JSON-RPC tool calls originating from any framework, providing unified rate-limiting and authorization.
- **LangGraph Checkpointers**: The ultimate source of truth for persistent agent state. Adapter events are serialized into LangGraph checkpoints to guarantee fault tolerance and resumability across the cluster.
- **SPIRE Identity Mesh**: Every framework adapter operates with a dedicated SPIFFE SVID, ensuring zero-trust identity propagation across the multi-agent swarm.

## 4. Edge Cases
- **State Desync**: If an external framework agent drops a connection, the adapter forces a localized rollback to the last known-good LangGraph checkpoint.
- **Tool Access Denial**: Framework agents attempting to access restricted tools via MCP will be blocked by the Switchboard based on their specific SPIFFE ID's RBAC policy.
- **Payload Bloat**: To prevent AutoGen conversational histories from exceeding memory limits, the adapter enforces active summarization before writing to the LangGraph state.
