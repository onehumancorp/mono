# Interop Blueprint: Agentic OS Universal Standard

## Overview
This document outlines the state of AI agent framework interoperability within the OHC "Agentic OS" control plane. Our goal is to serve as the "Universal Bus" for swarms of autonomous agents running on Kubernetes, leveraging standard protocols like MCP (Model Context Protocol), LangGraph for state management, and SPIFFE/SPIRE for zero-trust identity propagation.

## Supported Frameworks

1. **OpenClaw**
   - **Status:** Supported
   - **Integration:** Implements the OHC universal agent interface. State synchronization is managed via LangGraph event streams, allowing OpenClaw agents to seamlessly participate in multi-framework swarms.
   - **Feature:** Real-time state check-pointing via append-only K8s custom resources.

2. **AutoGen**
   - **Status:** Supported
   - **Integration:** Adapters map AutoGen's multi-agent conversation model to OHC's event-driven architecture.
   - **Feature:** Seamless conversation delegation across heterogeneous agent frameworks, secured by SPIRE identities.

3. **CrewAI**
   - **Status:** Supported
   - **Integration:** Implements the OHC universal agent interface to map CrewAI roles and tasks to LangGraph state.
   - **Feature:** Seamless task execution and role-based agent assignment across the swarm.

4. **Semantic Kernel**
   - **Status:** Supported
   - **Integration:** Implements the OHC universal agent interface for Semantic Kernel. Integrates with the shared state manager and agent command executor.
   - **Feature:** Enhanced function calling and prompt orchestration directly on the control plane.

## Architecture

- **MCP Switchboard:** Acts as a proxy for JSON-RPC calls, managing rate-limiting and authorization.
  - **New Feature:** Supports MCP Pagination for handling large context windows and iterative data streaming across agents.
- **State Management:** LangGraph checkpointers persist agent state to the cluster, ensuring fault tolerance and resumability.
  - **New Feature:** Supports LangGraph Time-Travel Debugging to allow rewinding and re-evaluating shared state dynamically.
- **Identity:** All intra-swarm communications require cryptographically verified SPIRE identities.

## Future Work
- Optimize K8s operator for large-scale swarm scheduling.
